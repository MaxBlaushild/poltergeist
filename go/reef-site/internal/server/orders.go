package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /api/reef/orders/:token (R-8.1, R-8.2: no-login order status lookup).
func (s *server) getOrder(c *gin.Context) {
	order, err := s.deps.DbClient.ReefOrder().FindByToken(c.Request.Context(), c.Param("token"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}
