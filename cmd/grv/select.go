//go:build !darwin && !freebsd
// +build !darwin,!freebsd

package main

import (
	"syscall"
)

// SelectSyscallID returns the ID of the select system call
func SelectSyscallID() uintptr {
	return syscall.SYS_PSELECT6
}
