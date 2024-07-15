package nsenter

/*
#cgo CFLAGS: -Wall -D_AS_CGO_LIB
extern void nsexec();
void __attribute__((constructor)) init(void) {
    nsexec();
}
*/
import "C"
