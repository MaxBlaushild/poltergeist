package logger

import (
	"bytes"
	"context"
	stdlog "log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWithAuthenticatedUserIDAddsPrefix(t *testing.T) {
	ctx := WithAuthenticatedUserID(context.Background(), "user-123")

	if got := Prefix(ctx); got != "[user=user-123] " {
		t.Fatalf("expected prefixed user ID, got %q", got)
	}
}

func TestAttachAuthenticatedUserUpdatesGinAndRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("GET", "/test", nil)

	AttachAuthenticatedUser(ctx, "user-456")

	if got := Prefix(ctx); got != "[user=user-456] " {
		t.Fatalf("expected gin context prefix, got %q", got)
	}

	if got := Prefix(ctx.Request.Context()); got != "[user=user-456] " {
		t.Fatalf("expected request context prefix, got %q", got)
	}
}

func TestPrintfPrefixesAuthenticatedUserID(t *testing.T) {
	var buffer bytes.Buffer
	originalWriter := stdlog.Writer()
	originalFlags := stdlog.Flags()
	stdlog.SetOutput(&buffer)
	stdlog.SetFlags(0)
	defer stdlog.SetOutput(originalWriter)
	defer stdlog.SetFlags(originalFlags)

	ctx := WithAuthenticatedUserID(context.Background(), "user-789")
	Printf(ctx, "hello %s", "world")

	output := buffer.String()
	if !strings.Contains(output, "[user=user-789] hello world") {
		t.Fatalf("expected output to contain prefixed log line, got %q", output)
	}
}
