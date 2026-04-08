package logger

import (
	"context"
	stdlog "log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type contextKey string

const (
	authenticatedUserIDContextKey contextKey = "authenticated_user_id"
	authenticatedUserIDGinKey     string     = "authenticatedUserID"
)

func WithAuthenticatedUserID(ctx context.Context, userID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return ctx
	}

	return context.WithValue(ctx, authenticatedUserIDContextKey, trimmedUserID)
}

func AttachAuthenticatedUser(ctx *gin.Context, userID string) {
	if ctx == nil {
		return
	}

	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return
	}

	ctx.Set(authenticatedUserIDGinKey, trimmedUserID)
	if ctx.Request != nil {
		ctx.Request = ctx.Request.WithContext(
			WithAuthenticatedUserID(ctx.Request.Context(), trimmedUserID),
		)
	}
}

func AuthenticatedUserID(source any) (string, bool) {
	switch value := source.(type) {
	case *gin.Context:
		if value == nil {
			return "", false
		}

		if userID, ok := value.Get(authenticatedUserIDGinKey); ok {
			if userIDString, ok := userID.(string); ok {
				trimmed := strings.TrimSpace(userIDString)
				if trimmed != "" {
					return trimmed, true
				}
			}
		}

		if value.Request != nil {
			return AuthenticatedUserID(value.Request.Context())
		}
	case *http.Request:
		if value != nil {
			return AuthenticatedUserID(value.Context())
		}
	case context.Context:
		if value == nil {
			return "", false
		}

		if userID, ok := value.Value(authenticatedUserIDContextKey).(string); ok {
			trimmed := strings.TrimSpace(userID)
			if trimmed != "" {
				return trimmed, true
			}
		}
	}

	return "", false
}

func Prefix(source any) string {
	if userID, ok := AuthenticatedUserID(source); ok {
		return "[user=" + userID + "] "
	}
	return ""
}

func Printf(source any, format string, args ...any) {
	stdlog.Printf(Prefix(source)+format, args...)
}

func Println(source any, args ...any) {
	if prefix := strings.TrimSpace(Prefix(source)); prefix != "" {
		args = append([]any{prefix}, args...)
	}
	stdlog.Println(args...)
}
