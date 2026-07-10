package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /houses/:id/overview — the house page: members and the running House
// Favor log. Names and houses are already public (login dropdown, leaderboard),
// so there is no secret content here.
func (s *server) getHouseOverview(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid house id"})
		return
	}

	house, err := s.dbClient.Vampire().GetHouseByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if house == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "house not found"})
		return
	}

	members, err := s.dbClient.Vampire().ListCharactersByHouse(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log, err := s.dbClient.Vampire().ListHouseFavorLog(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Item House Favor is a live "+X" overlay (see houseItemFavor), not part of the
	// ledger — so it's excluded from the base sum and log here too.
	favor := 0.0
	logOut := make([]gin.H, 0, len(log))
	for _, e := range log {
		if e.Source == "item" {
			continue
		}
		favor += e.Delta
		logOut = append(logOut, gin.H{
			"id":        e.ID,
			"delta":     e.Delta,
			"reason":    e.Reason,
			"gmName":    e.GMName,
			"source":    e.Source,
			"createdAt": e.CreatedAt,
		})
	}
	itemFavor := s.houseItemFavor(ctx)[id.String()]

	memberOut := make([]gin.H, 0, len(members))
	for _, m := range members {
		memberOut = append(memberOut, gin.H{
			"id":    m.ID,
			"name":  m.Name,
			"title": m.Title,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"house":   gin.H{"id": house.ID, "name": house.Name, "favor": favor, "itemFavor": itemFavor},
		"members": memberOut,
		"log":     logOut,
	})
}
