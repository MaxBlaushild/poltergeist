package procexec

import (
	"context"
	"errors"
	"testing"
	"time"
)

func newWorkDir(t *testing.T) string {
	t.Helper()
	dir, err := NewWorkDir(t.TempDir())
	if err != nil {
		t.Fatalf("NewWorkDir: %v", err)
	}
	return dir
}

func TestRun_Success(t *testing.T) {
	result, err := Run(context.Background(), newWorkDir(t), "/bin/echo", []string{"hello"}, Limits{
		Timeout:  5 * time.Second,
		MemoryMB: 256,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := string(result.Stdout); got != "hello\n" {
		t.Fatalf("stdout = %q, want %q", got, "hello\n")
	}
	if result.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0", result.ExitCode)
	}
	defer Cleanup(result.WorkDir)
}

// R-10 acceptance criterion: "The subprocess runner enforces its timeout —
// verified by a test with a deliberately non-terminating input."
func TestRun_TimeoutEnforced(t *testing.T) {
	start := time.Now()
	result, err := Run(context.Background(), newWorkDir(t), "/bin/sh", []string{"-c", "while true; do :; done"}, Limits{
		Timeout:  300 * time.Millisecond,
		MemoryMB: 64,
	})
	elapsed := time.Since(start)
	defer Cleanup(result.WorkDir)

	var procErr *Error
	if !errors.As(err, &procErr) {
		t.Fatalf("expected *procexec.Error, got %v (%T)", err, err)
	}
	if procErr.Kind != KindTimeout {
		t.Fatalf("Kind = %q, want %q", procErr.Kind, KindTimeout)
	}
	// Generous upper bound: the busy loop must actually be killed, not left
	// running past the deadline plus a small cleanup grace period.
	if elapsed > 5*time.Second {
		t.Fatalf("timeout was not enforced promptly: took %s to return for a 300ms limit", elapsed)
	}
}

func TestRun_NonZeroExitIsFailed(t *testing.T) {
	result, err := Run(context.Background(), newWorkDir(t), "/bin/sh", []string{"-c", "exit 3"}, Limits{
		Timeout:  5 * time.Second,
		MemoryMB: 64,
	})
	defer Cleanup(result.WorkDir)

	var procErr *Error
	if !errors.As(err, &procErr) {
		t.Fatalf("expected *procexec.Error, got %v (%T)", err, err)
	}
	if procErr.Kind != KindFailed {
		t.Fatalf("Kind = %q, want %q", procErr.Kind, KindFailed)
	}
	if procErr.ExitCode != 3 {
		t.Fatalf("ExitCode = %d, want 3", procErr.ExitCode)
	}
}

func TestRun_SignalIsCrashed(t *testing.T) {
	result, err := Run(context.Background(), newWorkDir(t), "/bin/sh", []string{"-c", "kill -SEGV $$"}, Limits{
		Timeout:  5 * time.Second,
		MemoryMB: 64,
	})
	defer Cleanup(result.WorkDir)

	var procErr *Error
	if !errors.As(err, &procErr) {
		t.Fatalf("expected *procexec.Error, got %v (%T)", err, err)
	}
	if procErr.Kind != KindCrashed {
		t.Fatalf("Kind = %q, want %q", procErr.Kind, KindCrashed)
	}
}

func TestRun_WorkDirIsFreshAndIsolatedPerInvocation(t *testing.T) {
	base := t.TempDir()

	dir1, err := NewWorkDir(base)
	if err != nil {
		t.Fatalf("NewWorkDir: %v", err)
	}
	r1, err := Run(context.Background(), dir1, "/bin/sh", []string{"-c", "echo one > out.txt"}, Limits{Timeout: 5 * time.Second, MemoryMB: 64})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer Cleanup(r1.WorkDir)

	dir2, err := NewWorkDir(base)
	if err != nil {
		t.Fatalf("NewWorkDir: %v", err)
	}
	r2, err := Run(context.Background(), dir2, "/bin/sh", []string{"-c", "cat out.txt 2>/dev/null || echo missing"}, Limits{Timeout: 5 * time.Second, MemoryMB: 64})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer Cleanup(r2.WorkDir)

	if r1.WorkDir == r2.WorkDir {
		t.Fatalf("expected distinct work dirs per invocation, got the same: %s", r1.WorkDir)
	}
	if got := string(r2.Stdout); got != "missing\n" {
		t.Fatalf("second invocation should not see the first invocation's file, got %q", got)
	}
}

func TestRun_RequiresWorkDir(t *testing.T) {
	if _, err := Run(context.Background(), "", "/bin/echo", []string{"hi"}, Limits{Timeout: time.Second, MemoryMB: 64}); err == nil {
		t.Fatal("expected an error when workDir is empty")
	}
}
