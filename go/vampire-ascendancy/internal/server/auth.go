package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	pkghttp "github.com/MaxBlaushild/poltergeist/pkg/http"
	"github.com/gin-gonic/gin"
)

// Real-account auth for Hosts/Co-Hosts, delegating straight to the shared
// poltergeist authenticator (go/authenticator) — same pattern as reef-site
// (see go/reef-site/internal/server/auth.go): email+password or Google,
// no user table of our own. Player sigil auth (login.go) is unrelated and
// unaffected — guests never see this.

// POST /vampire-ascendancy/auth/register
func (s *server) registerUser(ctx *gin.Context) {
	var req auth.RegisterByEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := s.authClient.RegisterByEmail(ctx, &req)
	if err != nil {
		forwardAuthError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// POST /vampire-ascendancy/auth/login
func (s *server) loginUser(ctx *gin.Context) {
	var req auth.LoginByEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := s.authClient.LoginByEmail(ctx, &req)
	if err != nil {
		forwardAuthError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// POST /vampire-ascendancy/auth/google — body is the ID token from Google
// Identity Services' button on the frontend; the authenticator verifies it.
func (s *server) loginUserWithGoogle(ctx *gin.Context) {
	var req auth.LoginWithGoogleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := s.authClient.LoginWithGoogle(ctx, &req)
	if err != nil {
		forwardAuthError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// forwardAuthError surfaces the authenticator's own status code/message
// (e.g. 409 "already exists", 401 "invalid email or password") instead of
// flattening everything to a generic 500.
func forwardAuthError(ctx *gin.Context, err error) {
	if statusErr, ok := err.(*pkghttp.StatusError); ok {
		ctx.JSON(statusErr.StatusCode, gin.H{"error": statusErr.Message})
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
