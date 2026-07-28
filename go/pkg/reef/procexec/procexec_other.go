//go:build !linux

package procexec

import "syscall"

// networkSandboxSupported is false outside Linux — CLONE_NEWNET has no
// equivalent here, so Run never attempts it and always reports
// Result.NetworkSandboxed=false.
const networkSandboxSupported = false

func sandboxedSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}
