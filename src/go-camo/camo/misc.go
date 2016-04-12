package camo

import (
	"net"
	"os"
	"syscall"
)

func isBrokenPipe(err error) bool {
	if opErr, ok := err.(*net.OpError); ok {
		// go1.6 changed this again?
		if syscallErr, ok := opErr.Err.(*os.SyscallError); ok {
			switch syscallErr.Err {
			case syscall.EPIPE, syscall.ECONNRESET:
				return true
			default:
				return false
			}
		}

		// older go
		switch opErr.Err {
		case syscall.EPIPE, syscall.ECONNRESET:
			return true
		default:
			return false
		}
	}
	return false
}
