package server

import (
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	pkghttp "github.com/MaxBlaushild/poltergeist/pkg/http"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

// optionalCurrentUser is checkout's own lightweight auth check — guest
// checkout must keep working, so unlike /me/orders this never rejects a
// request; a missing or invalid token just means the order gets created
// without a UserID (same as before customer accounts existed).
func (s *server) optionalCurrentUser(c *gin.Context) *models.User {
	header := c.Request.Header.Get("Authorization")
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil
	}

	user, err := s.deps.AuthClient.VerifyToken(c.Request.Context(), &auth.VerifyTokenRequest{Token: parts[1]})
	if err != nil {
		return nil
	}
	return user
}

// POST /api/reef/auth/register. Delegates straight to the shared
// poltergeist authenticator (go/authenticator) — reef-site has no user
// table of its own, matching every other domain's pattern (see
// go/final-fete/internal/server/auth.go), just email+password instead of
// phone+SMS.
func (s *server) registerCustomer(c *gin.Context) {
	var req auth.RegisterByEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := s.deps.AuthClient.RegisterByEmail(c.Request.Context(), &req)
	if err != nil {
		forwardAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// POST /api/reef/auth/login.
func (s *server) loginCustomer(c *gin.Context) {
	var req auth.LoginByEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := s.deps.AuthClient.LoginByEmail(c.Request.Context(), &req)
	if err != nil {
		forwardAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// POST /api/reef/auth/google — body is the ID token from Google Identity
// Services' button on the frontend; the authenticator verifies it.
func (s *server) loginWithGoogle(c *gin.Context) {
	var req auth.LoginWithGoogleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := s.deps.AuthClient.LoginWithGoogle(c.Request.Context(), &req)
	if err != nil {
		forwardAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// GET /api/reef/me/orders — order history. Wrapped in
// middleware.WithAuthenticationWithoutLocation, so c.MustGet("user") is
// always a valid *models.User by the time this runs.
func (s *server) getMyOrders(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	orders, err := s.deps.DbClient.ReefOrder().FindByUserID(c.Request.Context(), user.ID)
	if err != nil {
		internalError(c, "list my orders", err)
		return
	}

	ctx := c.Request.Context()
	resp := make([]operatorOrderResponse, 0, len(orders))
	for _, order := range orders {
		resp = append(resp, s.toOperatorOrderResponse(ctx, order))
	}
	c.JSON(http.StatusOK, resp)
}

// forwardAuthError surfaces the authenticator's own status code/message
// (e.g. 409 "already exists", 401 "invalid email or password") instead of
// flattening everything to a generic 500 — these are customer-facing
// validation states, not internal errors.
func forwardAuthError(c *gin.Context, err error) {
	if statusErr, ok := err.(*pkghttp.StatusError); ok {
		c.JSON(statusErr.StatusCode, gin.H{"error": statusErr.Message})
		return
	}
	internalError(c, "auth request", err)
}
