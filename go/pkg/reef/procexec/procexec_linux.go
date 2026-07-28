//go:build linux

package procexec

import "syscall"

// networkSandboxSupported is true on Linux, where Run attempts a fresh
// network namespace (CLONE_NEWNET) before falling back to unsandboxed
// networking if the process lacks CAP_SYS_ADMIN.
const networkSandboxSupported = true

func sandboxedSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setpgid:    true,
		Cloneflags: syscall.CLONE_NEWNET,
	}
}
