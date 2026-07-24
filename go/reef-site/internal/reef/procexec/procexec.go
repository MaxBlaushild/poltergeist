// Package procexec is the one shared entry point for every external binary
// this module invokes — OpenSCAD and the slicer both (R-2.5). It centralizes
// the sandboxing contract so no caller can accidentally shell out without a
// timeout, a memory bound, and a per-invocation temp dir.
//
// What is and isn't enforced here, honestly:
//   - Hard wall-clock timeout: enforced directly (context deadline + group kill).
//   - Memory limit: enforced via `ulimit -v` (a real rlimit on the child's
//     virtual address space), which is portable and needs no special
//     privilege. It is coarser than a cgroup memory limit but does not
//     require the container runtime to grant cgroup delegation.
//   - No network access: attempted via a fresh network namespace
//     (CLONE_NEWNET), which requires CAP_SYS_ADMIN. When the calling process
//     doesn't have it (e.g. a non-root container, or this package's own
//     tests running unprivileged), Run degrades to unsandboxed networking
//     and returns Result.NetworkSandboxed=false so callers/ops can see the
//     degradation rather than silently trusting it. The durable enforcement
//     of "no network" is the container's egress policy (security groups /
//     NACLs), not this flag — this is defense in depth on top of that, not a
//     substitute for it.
//   - Read-only filesystem except a per-invocation temp dir: Run gives the
//     child a fresh, empty CWD (Result.WorkDir) and nothing else — but it
//     does not chroot or bind-mount, so this alone does not stop a
//     determined binary from writing elsewhere on disk. The durable
//     enforcement is the container's read-only root filesystem (Docker
//     `--read-only` / ECS `readonlyRootFilesystem: true`, with WorkDir's
//     parent mounted as the one writable volume) — again, defense in depth,
//     not a substitute.
package procexec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

// Kind classifies why a run did not succeed, per R-2.5's "clean error type
// distinguishing timeout / crash / invalid-input."
type Kind string

const (
	KindTimeout Kind = "timeout"
	// KindCrashed: the process was terminated by a signal (segfault, abort
	// from a failed allocation once the memory limit is hit, OOM-kill, or
	// our own group-kill on timeout cleanup).
	KindCrashed Kind = "crashed"
	// KindFailed: the process ran to completion and exited non-zero — the
	// "invalid input" bucket (e.g. OpenSCAD rejecting malformed parameters).
	KindFailed Kind = "failed"
)

type Error struct {
	Kind     Kind
	ExitCode int
	Stderr   string
	Bin      string
}

func (e *Error) Error() string {
	stderr := e.Stderr
	if len(stderr) > 500 {
		stderr = stderr[:500] + "…"
	}
	return fmt.Sprintf("procexec: %s (%s) exit=%d stderr=%q", e.Bin, e.Kind, e.ExitCode, stderr)
}

// Limits are always explicit at the call site — there is no implicit
// "trust the default" timeout, on purpose (R-2.5 is a hard requirement, not
// a suggestion).
type Limits struct {
	Timeout  time.Duration
	MemoryMB int
}

type Result struct {
	Stdout           []byte
	Stderr           []byte
	ExitCode         int
	Duration         time.Duration
	WorkDir          string
	NetworkSandboxed bool
}

// maxCapturedOutput bounds stdout/stderr capture so a runaway or malicious
// binary can't exhaust memory through output volume alone.
const maxCapturedOutput = 10 * 1024 * 1024

// NewWorkDir allocates a fresh per-invocation temp dir under baseTempDir.
// Callers that need to place input files before running (e.g. writing a
// .scad source file) create it with this first and pass it to Run; callers
// with nothing to stage upfront can skip straight to Run, which allocates
// its own via this same helper.
func NewWorkDir(baseTempDir string) (string, error) {
	dir, err := os.MkdirTemp(baseTempDir, "reef-exec-*")
	if err != nil {
		return "", fmt.Errorf("procexec: creating work dir: %w", err)
	}
	return dir, nil
}

// Run executes bin with args inside workDir, enforcing limits. workDir must
// already exist — callers that need to stage input files first (e.g. a
// .scad source) create it with NewWorkDir and write into it before calling
// Run; callers with nothing to stage can just pass NewWorkDir's result
// straight through. The dir is left in place on return — callers read their
// expected output files from Result.WorkDir and are responsible for
// removing it once done (kept around on failure deliberately, for
// debugging).
func Run(ctx context.Context, workDir, bin string, args []string, limits Limits) (*Result, error) {
	if limits.Timeout <= 0 {
		return nil, fmt.Errorf("procexec: Limits.Timeout must be set (got %s)", limits.Timeout)
	}
	if workDir == "" {
		return nil, fmt.Errorf("procexec: workDir must be set (see NewWorkDir)")
	}

	runCtx, cancel := context.WithTimeout(ctx, limits.Timeout)
	defer cancel()

	script := `ulimit -v ` + strconv.Itoa(limits.MemoryMB*1024) + `; exec "$0" "$@"`
	cmd := exec.CommandContext(runCtx, "sh", append([]string{"-c", script, bin}, args...)...)
	cmd.Dir = workDir
	cmd.Env = []string{"PATH=/usr/bin:/bin:/usr/local/bin"}

	var stdout, stderr limitedBuffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	networkSandboxed := true
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:    true,
		Cloneflags: syscall.CLONE_NEWNET,
	}
	// Kill the whole process group (not just the direct child) on timeout,
	// since the wrapping `sh -c ... exec` replaces itself with bin — a
	// group kill still catches anything bin itself forks.
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	cmd.WaitDelay = 5 * time.Second

	start := time.Now()
	runErr := cmd.Run()
	if runErr != nil && isPermissionDenied(runErr) {
		// No CAP_SYS_ADMIN for a fresh network namespace (e.g. unprivileged
		// dev/test environment) — degrade to unsandboxed networking rather
		// than failing every invocation outright, but tell the caller.
		networkSandboxed = false
		stdout.Reset()
		stderr.Reset()
		cmd = exec.CommandContext(runCtx, "sh", append([]string{"-c", script, bin}, args...)...)
		cmd.Dir = workDir
		cmd.Env = []string{"PATH=/usr/bin:/bin:/usr/local/bin"}
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Cancel = func() error {
			if cmd.Process == nil {
				return nil
			}
			return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		cmd.WaitDelay = 5 * time.Second
		start = time.Now()
		runErr = cmd.Run()
	}
	duration := time.Since(start)

	result := &Result{
		Stdout:           stdout.Bytes(),
		Stderr:           stderr.Bytes(),
		Duration:         duration,
		WorkDir:          workDir,
		NetworkSandboxed: networkSandboxed,
	}

	if runErr == nil {
		result.ExitCode = 0
		return result, nil
	}

	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		return result, &Error{Kind: KindTimeout, ExitCode: -1, Stderr: string(result.Stderr), Bin: bin}
	}

	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
		if exitErr.ProcessState != nil {
			if status, ok := exitErr.ProcessState.Sys().(syscall.WaitStatus); ok && status.Signaled() {
				return result, &Error{Kind: KindCrashed, ExitCode: result.ExitCode, Stderr: string(result.Stderr), Bin: bin}
			}
		}
		return result, &Error{Kind: KindFailed, ExitCode: result.ExitCode, Stderr: string(result.Stderr), Bin: bin}
	}

	return result, fmt.Errorf("procexec: %s: %w", bin, runErr)
}

func isPermissionDenied(err error) bool {
	return errors.Is(err, syscall.EPERM) || errors.Is(err, os.ErrPermission)
}

// Cleanup removes a Result's WorkDir. Callers call this once they've read
// whatever output files they need (STL, GLB, slicer JSON, ...).
func Cleanup(workDir string) error {
	if workDir == "" || workDir == "/" {
		return fmt.Errorf("procexec: refusing to remove suspicious work dir %q", workDir)
	}
	return os.RemoveAll(workDir)
}

// OutPath is a small convenience for building an output file path inside a
// Result's WorkDir.
func OutPath(workDir, name string) string {
	return filepath.Join(workDir, name)
}

type limitedBuffer struct {
	bytes.Buffer
	truncated bool
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.Buffer.Len() >= maxCapturedOutput {
		b.truncated = true
		return len(p), nil
	}
	remaining := maxCapturedOutput - b.Buffer.Len()
	if len(p) > remaining {
		b.truncated = true
		p = p[:remaining]
	}
	return b.Buffer.Write(p)
}
