package main

/*
#cgo CFLAGS: -Wall
#define _GNU_SOURCE
#include <unistd.h>
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <sys/wait.h>

// #define _DEBUG
#ifdef _DEBUG
#define DEBUG(...) printf(__VA_ARGS__)
#else
#define DEBUG(...)
#endif

#ifndef PATH_MAX
#define PATH_MAX 4096
#endif


static void wait_child(pid_t pid) {
	int status;
	pid_t ret;
	for(;;) {
		ret = waitpid(pid, &status, WUNTRACED);
		if (ret == pid && (WIFSTOPPED(status))) {
			kill(pid, SIGCONT);
		} else {
		 	break;
		}
	}
}

static inline ssize_t read_all(int fd, char *buf, size_t count)
{
	ssize_t ret;
	ssize_t c = 0;

	memset(buf, 0, count);
	while (count > 0) {
		ret = read(fd, buf, count);
		if (ret < 0) {
			return c ? c : -1;
		}
		if (ret == 0)
			return c;
		count -= ret;
		buf += ret;
		c += ret;
	}
	return c;
}

static inline ssize_t read_all_alloc(int fd, char **buf)
{
	size_t size = 1024, c = 0;
	ssize_t ret;

	*buf = malloc(size);
	if (!*buf)
		return -1;

	while (1) {
		ret = read_all(fd, *buf + c, size - c);
		if (ret < 0) {
			free(*buf);
			*buf = NULL;
			return -1;
		}

		if (ret == 0)
			return c;

		c += ret;
		if (c == size) {
			size *= 2;
			*buf = realloc(*buf, size);
			if (!*buf)
				return -1;
		}
	}
}

ssize_t ns_get_envbuf(char *pid, char **p) {
	char pathbuf[PATH_MAX];
	int fd;
	ssize_t rc = 0;

	snprintf(pathbuf, sizeof(pathbuf), "/proc/%s/environ", pid);
	fd = open(pathbuf, O_RDONLY);
	if (fd < 0) {
		printf("open file: %s failed, case: %s\n", pathbuf, strerror(errno));
		return -1;
	}

	rc = read_all_alloc(fd, p);
	(*p)[rc] = '\0';
	close(fd);

	return rc;
}
int ns_set_env(char *env_buf, ssize_t rc) {
	char *p, *val;

	p = env_buf;
	val = NULL;
	while(rc > 0) {
		DEBUG("%s\n", p);
		val = strchr(p, '=');
		if(!val) continue;
		*val = '\0';
		if (setenv(p, val+1, 0) != 0) {
			printf("setenv error: %s\n",strerror(errno));
		}
		*val = '=';
		rc -= strlen(p) + 1;
	}
	return 1;
}

void __attribute__((constructor)) nsenter() {
	DEBUG("into nsenter\n");
	char *ns_list = NULL, *pid, *env_buf, *nsdir;
	int ns_enum_list[] = {CLONE_NEWIPC , CLONE_NEWNET , CLONE_NEWPID , CLONE_NEWUTS , CLONE_NEWCGROUP, CLONE_NEWNS};

	char *ns_char_list[] = {"ipc", "net", "pid", "uts", "cgroup", "mnt"};
	int fds[6]={0};
	int ns_args = 0;
	int list_len = (int)(sizeof(ns_enum_list)/sizeof(int));
	int i;
	char nspath[1024 * 3];
	char nsprefix[2048];
	pid_t spid;
	ssize_t rc = 0;

	for(i = 0; i < list_len; i++) {
		fds[i] = -1;
	}

	ns_list = getenv("NSLIST");
	if (!ns_list) {
		DEBUG("ns list is null\n");
		return;
	}
	DEBUG("GET NSDIR\n");
	nsdir = getenv("NSDIR");
	if (nsdir && *nsdir != '\0') {
		DEBUG("NSDIR=%s\n", nsdir);
		sprintf(nsprefix, "%s", nsdir);
	}
	DEBUG("GET PID\n");
	pid = getenv("NSPID");
	if (pid && *pid != '\0') {
		sprintf(nsprefix,"/proc/%s/ns", pid);
	}
	if ((nsdir == NULL || *nsdir == '\0') && (pid == NULL || *pid == '\0')) {
		if (nsdir == NULL) {
			DEBUG("nsdir is null\n");
		}

		if (pid == NULL) {
			DEBUG("pid is null\n");
		}
		return;
	}
	DEBUG("NSPREFIX: %s\nNSLIST:%s\n", nsprefix,ns_list);
	if (sscanf(ns_list,"%d",&ns_args) < 1) {
		DEBUG("parse ns_list failed, ns_list value: %s\n", ns_list);
		exit(-1);
	}

	for (i = 0; i < list_len; i++) {
		if (ns_enum_list[i] & ns_args) {
			DEBUG("set ns: %s\n\n", ns_char_list[i]);
			sprintf(nspath, "%s/%s" , nsprefix, ns_char_list[i]);
			fds[i] = open(nspath, O_RDONLY);
			if (fds[i] < 0) {
				printf("open file: %s failed\n", nspath);
				continue;
			}
		}
	}
	DEBUG("after getns\n");
	if (pid && *pid != '\0') {
		env_buf = malloc(sizeof(char) * 256);
		rc = ns_get_envbuf(pid, &env_buf);
		DEBUG("rc=%ld\nenv=%s\n", rc, env_buf);
	}

	// spid = fork();
	// if (spid < 0) {
	// 	printf("fork error\n");
	// 	exit(-1);
	// }
	// if (spid > 0) {
	// 	free(env_buf);
	// 	wait_child(spid);
	// 	exit(0);
	// }

	char root_path[256]={0};
	sprintf(root_path, "/proc/%s/root" ,pid);
	int root_fd = open(root_path, O_RDONLY);


	for(i = 0; i < list_len; i++) {
		if(fds[i] < 0) {
			continue;
		}
		DEBUG("try to set ns: %s\n", ns_char_list[i]);
		if (setns(fds[i], ns_enum_list[i]) < 0) {
			printf("setns %s failed\n", ns_char_list[i]);
			exit(-1);
		}
		DEBUG("set ns: %s\n", ns_char_list[i]);

	}

	if(ns_args & CLONE_NEWNS) {
		int chflag = 0;

		if (root_fd > 0) {
			int wdfd = open(".", O_RDONLY);
			if (wdfd > 0) {
				if(fchdir(root_fd) >= 0 && chroot(".") == 0 && chdir("/") == 0){

					close(wdfd);
					chflag = 1;
				}
			}
		}
		if(!chflag) {
			printf("chroot error:%s\n", strerror(errno));
			exit(-1);
		}
	}

	if(ns_args & CLONE_NEWPID) {
		// ns_set_env(env_buf, rc);
		if(ns_args & CLONE_NEWNS) {
			clearenv();
			ns_set_env(env_buf, rc);
			setenv("NSPID", pid, 0);
			setenv("NSLIST", ns_list, 0);
			// system("cat /proc/self/environ");
		}

		free(env_buf);

		spid = fork();
		if(spid > 0) {
			wait_child(spid);
			exit(0);
		}
	}

	DEBUG("cur pid: %d\n",getpid());

	DEBUG("after fork\n");
	// system("cat /proc/self/environ");
	// printf("\n\n");

}
*/
import "C"
