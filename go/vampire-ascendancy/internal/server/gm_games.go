package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Placement rewards, from the player rules. 1st/2nd/3rd earn Blood Tokens for the
// player and House Favor for their house; everyone else who played earns +1 BT.
var gameBTByPlace = map[int]int{1: 5, 2: 3, 3: 1}
var gameHFByPlace = map[int]float64{1: 5, 2: 3, 3: 2}

const gameParticipationBT = 1

// characterLookup returns a map of character id -> character (with house), for
// resolving winner names without an N+1 of per-game queries.
func (s *server) characterLookup(ctx *gin.Context) (map[string]models.VampireCharacter, error) {
	chars, err := s.dbClient.Vampire().ListCharacters(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string]models.VampireCharacter, len(chars))
	for _, c := range chars {
		m[c.ID.String()] = c
	}
	return m, nil
}

func winnerJSON(charID *uuid.UUID, byID map[string]models.VampireCharacter) gin.H {
	if charID == nil {
		return nil
	}
	c, ok := byID[charID.String()]
	if !ok {
		return nil
	}
	row := gin.H{"characterId": c.ID, "characterName": c.Name}
	if c.House != nil {
		row["house"] = c.House.Name
	}
	return row
}

func gamesResponse(games []models.VampireGame, byID map[string]models.VampireCharacter) []gin.H {
	out := make([]gin.H, 0, len(games))
	for _, g := range games {
		out = append(out, gin.H{
			"id":      g.ID,
			"ordinal": g.Ordinal,
			"name":    g.Name,
			"status":  g.Status,
			"first":   winnerJSON(g.FirstCharacterID, byID),
			"second":  winnerJSON(g.SecondCharacterID, byID),
			"third":   winnerJSON(g.ThirdCharacterID, byID),
		})
	}
	return out
}

// GET /gm/games — the game list with recorded finishers.
func (s *server) gmListGames(ctx *gin.Context) {
	games, err := s.dbClient.Vampire().ListGames(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	byID, err := s.characterLookup(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"games": gamesResponse(games, byID)})
}

// POST /gm/games — add a game to the list.
func (s *server) gmCreateGame(ctx *gin.Context) {
	var body struct {
		Name    string `json:"name"`
		Ordinal int    `json:"ordinal"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil || body.Name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "a game name is required"})
		return
	}
	game, err := s.dbClient.Vampire().UpsertGame(ctx, body.Ordinal, body.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "create_game", map[string]interface{}{"name": body.Name})
	ctx.JSON(http.StatusOK, gin.H{"id": game.ID})
}

// POST /gm/games/:id/result — record a game's finishers and apply the awards.
// Blood Tokens go to the finishing players; House Favor to their houses; other
// listed participants each earn +1 BT. Blocked once a game is already recorded so
// awards can't be applied twice.
func (s *server) gmRecordGameResult(ctx *gin.Context) {
	gameID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	game, err := s.dbClient.Vampire().GetGameByID(ctx, gameID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if game == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}
	if game.Status == "played" {
		ctx.JSON(http.StatusConflict, gin.H{"error": "this game's result is already recorded"})
		return
	}

	var body struct {
		FirstID        string   `json:"firstId"`
		SecondID       string   `json:"secondId"`
		ThirdID        string   `json:"thirdId"`
		ParticipantIDs []string `json:"participantIds"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	v := s.dbClient.Vampire()
	gmName := gmNameFromContext(ctx)
	placed := map[string]bool{}

	// resolve resolves a character id and applies its placement award.
	place := []struct {
		id  string
		pos int
	}{{body.FirstID, 1}, {body.SecondID, 2}, {body.ThirdID, 3}}
	placeIDs := [3]*uuid.UUID{}

	for i, p := range place {
		if p.id == "" {
			continue
		}
		cid, err := uuid.Parse(p.id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
			return
		}
		ch, err := v.GetCharacterByID(ctx, cid)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if ch == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "unknown character in results"})
			return
		}
		placed[cid.String()] = true
		placeIDs[i] = &cid

		// House Favor to the finisher's house.
		if ch.HouseID != nil {
			if err := v.AddHouseFavor(ctx, &models.VampireHouseFavorLedger{
				HouseID: *ch.HouseID,
				Delta:   gameHFByPlace[p.pos],
				Reason:  "Game: " + game.Name,
				GMName:  gmName,
				Source:  "game",
			}); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		// Blood Tokens to the finishing player, if the character has an active slot.
		if err := s.awardGameBT(ctx, cid, gameBTByPlace[p.pos], "Game: "+game.Name, gmName); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Participation: every other listed player earns +1 BT.
	for _, pid := range body.ParticipantIDs {
		if pid == "" || placed[pid] {
			continue
		}
		cid, err := uuid.Parse(pid)
		if err != nil {
			continue
		}
		if err := s.awardGameBT(ctx, cid, gameParticipationBT, "Game participation: "+game.Name, gmName); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := v.SetGameResult(ctx, gameID, placeIDs[0], placeIDs[1], placeIDs[2]); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logGM(ctx, "record_game_result", map[string]interface{}{"gameId": gameID.String(), "name": game.Name})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// awardGameBT credits a character's active player, or no-ops if the character has
// no active player slot (e.g. an NPC-run game entrant).
func (s *server) awardGameBT(ctx *gin.Context, characterID uuid.UUID, delta int, reason, gmName string) error {
	player, err := s.dbClient.Vampire().GetActivePlayerByCharacterID(ctx, characterID)
	if err != nil {
		return err
	}
	if player == nil {
		return nil
	}
	return s.dbClient.Vampire().AddBloodTokens(ctx, &models.VampireBloodTokenLog{
		PlayerID: player.ID,
		Delta:    delta,
		Reason:   reason,
		Source:   "game",
		GMName:   gmName,
	})
}

// GET /games — the player-facing game list: order, what's been played, and each
// played game's finishers.
func (s *server) getGames(ctx *gin.Context) {
	games, err := s.dbClient.Vampire().ListGames(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	byID, err := s.characterLookup(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"games": gamesResponse(games, byID)})
}
