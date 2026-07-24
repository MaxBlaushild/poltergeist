package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /api/reef/tanks (R-8.1). Only verified profiles are ever returned —
// R-3.4: unverified rows are research backlog and must not appear in
// configurator dropdowns. ReefTankProfileHandle.FindVerified already
// filters on verified=true.
func (s *server) listTanks(c *gin.Context) {
	profiles, err := s.deps.DbClient.ReefTankProfile().FindVerified(c.Request.Context())
	if err != nil {
		internalError(c, "list tanks", err)
		return
	}
	c.JSON(http.StatusOK, profiles)
}
