#define _GNU_SOURCE

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <setjmp.h>
#include <unistd.h>
#include <sys/types.h>
#include <sched.h>
#include <signal.h>
#include <errno.h>
#include <sys/socket.h>
#include <linux/types.h>

#ifdef _DEBUG
#define DEBUG(...) printf(__VA_ARGS__)
#else
#define DEBUG(...)
#endif
enum sync_t
{
    SYNC_RECVPID_PLS = 0x42,
    SYNC_RECVPID_ACK = 0x43,
    SYNC_GRANDCHILD = 0x44,
    SYNC_CHILD_FINISH = 0x45,
};

#define STAGE_1 0
#define STAGE_2 1
#define STAGE_3 2

struct clone_t
{
    char stack[4096] __attribute__((aligned(16)));
    char stack_ptr[0];
    jmp_buf *env;
    int jmpval;
};

static __attribute__((noinline)) int child_func(void *arg)
{
    struct clone_t *ca = (struct clone_t *)arg;
    longjmp(*ca->env, ca->jmpval);
}
static __attribute__((noinline)) int clone_parent(jmp_buf *env, int jmpval)
{
    struct clone_t ca = {
        .env = env,
        .jmpval = jmpval,
    };
    return clone(child_func, ca.stack_ptr, CLONE_PARENT | SIGCHLD, &ca);
}

void try_unshare(int flags, const char *msg)
{
    DEBUG("unshare %s\n", msg);
    int retries = 5;
    for (; retries > 0; retries--)
    {
        if (unshare(flags) == 0)
        {
            return;
        }
        DEBUG("unshare error: %s\n", strerror(errno));
        if (errno != EINVAL)
            break;
    }
    exit(-1);
}

int getenv_int(const char *name)
{
    char *val, *endptr;
    int ret;

    val = getenv(name);
    DEBUG("INIT PIPE: %s\n", val);
    if (val == NULL || *val == '\0')
        return -ENOENT;

    ret = strtol(val, &endptr, 10);
    if (val == endptr || *endptr != '\0')
        DEBUG("unable to parse %s=%s", name, val);

    if (ret < 0)
        DEBUG("bad value for %s=%s (%d)", name, val, ret);

    return ret;
}

void nsexec()
{
    jmp_buf env;
    int sync_child_pipe[2], sync_grandchild_pipe[2];
    int syncfd, pipenum;

    if ((pipenum = getenv_int("_INIT_PIPE")) < 0)
    {
        DEBUG("read init pipe error\n");
        return;
    }

    char msgbuf[4096];
    int nl = read(pipenum, msgbuf, 4096);
    if (nl < 0)
    {
        exit(-1);
    }
    DEBUG("find env var _INIT_PIPE\n");
    if (socketpair(AF_LOCAL, SOCK_STREAM, 0, sync_child_pipe) < 0)
    {
        exit(-1);
    }

    if (socketpair(AF_LOCAL, SOCK_STREAM, 0, sync_grandchild_pipe) < 0)
    {
        exit(-1);
    }

    switch (setjmp(env))
    {
    case STAGE_1:
        int stage1_complete, stage2_complete;
        pid_t stage2_pid;
        stage1_complete = 0;
        DEBUG("pid: %d get 0, syncfd: %d\n", getpid(), syncfd);
        pid_t stage1_pid = clone_parent(&env, STAGE_2);
        if (stage1_pid < 0)
        {
            exit(-1);
        }
        DEBUG("clone get new pid: %d\n", stage1_pid);
        syncfd = sync_child_pipe[1];
        if (close(sync_child_pipe[0]) < 0)
        {
            exit(-1);
        }
        enum sync_t s;
        int n;
        while (!stage1_complete)
        {
            if ((n = read(syncfd, &s, sizeof(s))) != sizeof(s))
            {
                DEBUG("read syncfd failed: %s, n = %d sizoef(s): %ld\n", strerror(errno), n, sizeof(s));
                exit(-1);
            }
            switch (s)
            {
            case SYNC_RECVPID_PLS:
                DEBUG("recving pid...\n");
                if (read(syncfd, &stage2_pid, sizeof(stage2_pid)) != sizeof(stage2_pid))
                {
                    kill(stage1_pid, SIGKILL);
                    exit(-1);
                }
                DEBUG("recved pid: %d\n", stage2_pid);
                s = SYNC_RECVPID_ACK;
                if (write(syncfd, &s, sizeof(s)) != sizeof(s))
                {
                    kill(stage1_pid, SIGKILL);
                    kill(stage2_pid, SIGKILL);
                    exit(-1);
                }
                int len = dprintf(pipenum, "{\"childpid\":%d,\"grandchildpid\":%d}", stage1_pid, stage2_pid);
                close(pipenum);
                if (len < 0)
                {
                    kill(stage1_pid, SIGKILL);
                    kill(stage2_pid, SIGKILL);
                    DEBUG("write back pid error\n");
                    exit(-1);
                }
                break;
            case SYNC_CHILD_FINISH:
                DEBUG("recv child process finished\n");
                stage1_complete = 1;
                break;
            default:
                exit(-1);
            }
        }
        stage2_complete = 0;
        syncfd = sync_grandchild_pipe[1];
        if (close(sync_grandchild_pipe[0]) < 0)
        {
            exit(-1);
        }
        s = SYNC_GRANDCHILD;
        if (write(syncfd, &s, sizeof(s)) != sizeof(s))
        {
            DEBUG("write sync_grandchild error\n");
            kill(stage2_pid, SIGKILL);
            exit(-1);
        }
        while (!stage2_complete)
        {
            enum sync_t s;
            if (read(syncfd, &s, sizeof(s)) != sizeof(s))
            {
                DEBUG("read grandchild's msg error\n");
                kill(stage2_pid, SIGKILL);
                exit(-1);
            }
            switch (s)
            {
            case SYNC_CHILD_FINISH:
                stage2_complete = 1;
                break;
            default:
                DEBUG("read grandchild's msg: unknown msg\n");
                kill(stage2_pid, SIGKILL);
                exit(-1);
            }
        }
        exit(0);
        break;
    case STAGE_2:
        DEBUG("pid: %d get 1\n", getpid());
        try_unshare(CLONE_NEWIPC | CLONE_NEWNET | CLONE_NEWPID | CLONE_NEWNS | CLONE_NEWUTS | CLONE_NEWCGROUP | CLONE_NEWTIME, "unshare");
        pid_t stage3_pid = clone_parent(&env, STAGE_3);
        if (stage3_pid < 0)
        {
            exit(-1);
        }
        syncfd = sync_child_pipe[0];
        if (close(sync_child_pipe[1]) < 0)
        {
            exit(-1);
        }
        s = SYNC_RECVPID_PLS;
        if (write(syncfd, &s, sizeof(s)) != sizeof(s))
        {
            DEBUG("write stage 3's pid error\n");
            exit(-1);
        }
        if (write(syncfd, &stage3_pid, sizeof(stage3_pid)) != sizeof(stage3_pid))
        {
            DEBUG("write stage 3's pid date error\n");
            exit(-1);
        }
        if (read(syncfd, &s, sizeof(s)) != sizeof(s))
        {
            DEBUG("read from syncfd %d error\n", syncfd);
            exit(-1);
        }
        if (s != SYNC_RECVPID_ACK)
        {
            DEBUG("read error status\n");
            exit(-1);
        }
        s = SYNC_CHILD_FINISH;
        DEBUG("write finish to parent, syncfd: %d\n", syncfd);
        if (write(syncfd, &s, sizeof(s)) != sizeof(s))
        {
            DEBUG("write finish to parent failed, %s\n", strerror(errno));
            exit(-1);
        }
        exit(0);
        break;
    case STAGE_3:
        DEBUG("stage3 pid: %d get 3\n", getpid());

        syncfd = sync_grandchild_pipe[0];
        if (close(sync_grandchild_pipe[1] < 0))
        {
            DEBUG("close error\n");
            exit(-1);
        }
        if (read(syncfd, &s, sizeof(s)) != sizeof(s))
        {
            DEBUG("read msg error\n");
            exit(-1);
        }
        if (s != SYNC_GRANDCHILD)
        {
            DEBUG("read status error\n");
            exit(-1);
        }
        if (setsid() < 0)
        {
            DEBUG("setsid failed\n");
            exit(-1);
        }

        if (setuid(0) < 0)
        {
            DEBUG("setuid failed\n");
            exit(-1);
        }

        if (setgid(0) < 0)
        {
            DEBUG("setgid failed\n");
            exit(-1);
        }

        s = SYNC_CHILD_FINISH;
        if (write(syncfd, &s, sizeof(s)) != sizeof(s))
        {
            DEBUG("write error\n");
            exit(-1);
        }
        break;
    }
}
#ifndef _AS_CGO_LIB
int main(int argc, char *argv[])
{
    setenv("_INIT_PIPE", "1", 1);
    nsexec();
    return 0;
}
#endif
