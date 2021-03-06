// +build dragonfly freebsd linux openbsd solaris

package sys

import "syscall"
import "golang.org/x/sys/unix"

var MlockSupported bool = true

func MlockAll() error {
	return unix.Mlockall(syscall.MCL_CURRENT | syscall.MCL_FUTURE)
}
