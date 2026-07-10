package server

import (
	"context"
	"net/http"
	"sort"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// houseItemFavor computes, per house id, the live House Favor contributed by
// items players currently hold. It is NOT written to the ledger, so it updates
// the instant an HF item is assigned — shown as a "+X" overlay on the standings.
func (s *server) houseItemFavor(ctx context.Context) map[string]float64 {
	v := s.dbClient.Vampire()
	out := map[string]float64{}
	players, err := v.ListPlayers(ctx)
	if err != nil {
		return out
	}
	houseOf := map[string]*uuid.UUID{}
	for _, p := range players {
		if p.Character != nil {
			houseOf[p.ID.String()] = p.Character.HouseID
		}
	}
	pis, err := v.ListAllPlayerItems(ctx)
	if err != nil {
		return out
	}
	for _, pi := range pis {
		if pi.Item == nil || pi.Item.HFEffect == 0 {
			continue
		}
		if hid := houseOf[pi.PlayerID.String()]; hid != nil {
			out[hid.String()] += float64(pi.Item.HFEffect)
		}
	}
	return out
}

// leaderboardWithItems attaches the live item-HF overlay and ranks by the
// combined total (base ledger + item overlay).
func (s *server) leaderboardWithItems(ctx context.Context) ([]db.HouseFavorStanding, error) {
	standings, err := s.dbClient.Vampire().Leaderboard(ctx)
	if err != nil {
		return nil, err
	}
	itemFavor := s.houseItemFavor(ctx)
	for i := range standings {
		standings[i].ItemFavor = itemFavor[standings[i].HouseID.String()]
	}
	sort.SliceStable(standings, func(i, j int) bool {
		ti := standings[i].Favor + standings[i].ItemFavor
		tj := standings[j].Favor + standings[j].ItemFavor
		if ti != tj {
			return ti > tj
		}
		return standings[i].SortOrder < standings[j].SortOrder
	})
	return standings, nil
}

// getLeaderboard returns House Favor standings — always visible to players, even
// before content is unlocked. It is the authoritative live standing.
func (s *server) getLeaderboard(ctx *gin.Context) {
	standings, err := s.leaderboardWithItems(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"standings": standings})
}

// POST /gm/hf — award or deduct House Favor. Appends to the ledger; the
// leaderboard is the running sum.
func (s *server) gmAwardHouseFavor(ctx *gin.Context) {
	var body struct {
		HouseID string  `json:"houseId"`
		Delta   float64 `json:"delta"`
		Reason  string  `json:"reason"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	houseID, err := uuid.Parse(body.HouseID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid house id"})
		return
	}
	if body.Delta == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "delta must be non-zero"})
		return
	}

	gmName := gmNameFromContext(ctx)
	if err := s.dbClient.Vampire().AddHouseFavor(ctx, &models.VampireHouseFavorLedger{
		HouseID: houseID,
		Delta:   body.Delta,
		Reason:  body.Reason,
		GMName:  gmName,
		Source:  "manual",
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logGM(ctx, "award_house_favor", map[string]interface{}{
		"houseId": body.HouseID,
		"delta":   body.Delta,
		"reason":  body.Reason,
	})

	standings, err := s.leaderboardWithItems(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"standings": standings})
}

// getHouses lists the houses (for GM award dropdowns).
func (s *server) getHouses(ctx *gin.Context) {
	houses, err := s.dbClient.Vampire().ListHouses(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"houses": houses})
}
