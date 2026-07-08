package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GET /gm/export — a full standings snapshot (House Favor per house + Blood Token
// totals per player) the GM can download and keep off-system. Combined with the
// archive-on-reset safety net, this ensures scores are never only in one place.
func (s *server) gmExportStandings(ctx *gin.Context) {
	v := s.dbClient.Vampire()

	standings, err := v.Leaderboard(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	btTotals, err := v.BloodTokenTotalsByPlayer(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	players, err := v.ListPlayers(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	btByPlayer := map[string]int{}
	for _, t := range btTotals {
		btByPlayer[t.PlayerID.String()] = t.Total
	}

	playerRows := make([]gin.H, 0, len(players))
	for _, p := range players {
		row := gin.H{
			"playerId":    p.ID,
			"playerName":  p.GuestLabel, // GM-only roster name
			"active":      p.Active,
			"bloodTokens": btByPlayer[p.ID.String()],
			"character":   nil,
			"house":       nil,
		}
		if p.Character != nil {
			row["character"] = p.Character.Name
			if p.Character.House != nil {
				row["house"] = p.Character.House.Name
			}
		}
		playerRows = append(playerRows, row)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"exportedAt": time.Now().UTC(),
		"houseFavor": standings,
		"players":    playerRows,
	})
}
