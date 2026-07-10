package server

import (
	"encoding/json"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// House Favor per place, applied automatically to the finishers' house. Blood
// Tokens (1st +5, 2nd +3, 3rd +2, participants +1) are handed out in person by
// the GM, so they are not awarded here.
var gameHFByPlace = map[int]float64{1: 5, 2: 3, 3: 2}

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

// winnersJSON resolves a JSON array of character id strings into winner objects
// (name + house), preserving order and skipping ids that no longer resolve.
func winnersJSON(raw datatypes.JSON, byID map[string]models.VampireCharacter) []gin.H {
	out := []gin.H{}
	if len(raw) == 0 {
		return out
	}
	var ids []string
	if err := json.Unmarshal(raw, &ids); err != nil {
		return out
	}
	for _, id := range ids {
		c, ok := byID[id]
		if !ok {
			continue
		}
		row := gin.H{"characterId": c.ID, "characterName": c.Name}
		if c.House != nil {
			row["house"] = c.House.Name
		}
		out = append(out, row)
	}
	return out
}

func gamesResponse(games []models.VampireGame, byID map[string]models.VampireCharacter) []gin.H {
	out := make([]gin.H, 0, len(games))
	for _, g := range games {
		out = append(out, gin.H{
			"id":           g.ID,
			"ordinal":      g.Ordinal,
			"name":         g.Name,
			"status":       g.Status,
			"first":        winnersJSON(g.FirstCharacterIDs, byID),
			"second":       winnersJSON(g.SecondCharacterIDs, byID),
			"third":        winnersJSON(g.ThirdCharacterIDs, byID),
			"startMinutes": g.StartMinutes,
			"endMinutes":   g.EndMinutes,
			"location":     g.Location,
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
		FirstIDs  []string `json:"firstIds"`
		SecondIDs []string `json:"secondIds"`
		ThirdIDs  []string `json:"thirdIds"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	v := s.dbClient.Vampire()
	gmName := gmNameFromContext(ctx)
	placed := map[string]bool{}

	places := []struct {
		ids []string
		pos int
	}{{body.FirstIDs, 1}, {body.SecondIDs, 2}, {body.ThirdIDs, 3}}
	placeIDs := [3][]uuid.UUID{}
	placeHouse := [3]*uuid.UUID{} // the single house that shares each place

	// Validate each place first: everyone sharing a place must be from the same
	// house (House Favor is awarded once, to that house). No Blood Tokens here —
	// the GM hands those out in person.
	for i, place := range places {
		houseSet := false
		for _, idStr := range place.ids {
			if idStr == "" {
				continue
			}
			cid, err := uuid.Parse(idStr)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
				return
			}
			if placed[cid.String()] {
				continue // same character listed in two places — count once
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
			if !houseSet {
				placeHouse[i] = ch.HouseID
				houseSet = true
			} else if !sameHouse(placeHouse[i], ch.HouseID) {
				ctx.JSON(http.StatusConflict, gin.H{"error": "everyone sharing a place must be from the same house"})
				return
			}
			placed[cid.String()] = true
			placeIDs[i] = append(placeIDs[i], cid)
		}
	}

	// Each place must be a different house from the others.
	seenHouse := map[string]bool{}
	for i := range places {
		if len(placeIDs[i]) == 0 || placeHouse[i] == nil {
			continue
		}
		h := placeHouse[i].String()
		if seenHouse[h] {
			ctx.JSON(http.StatusConflict, gin.H{"error": "each place must be won by a different house"})
			return
		}
		seenHouse[h] = true
	}

	// Award each place's House Favor once, to its house.
	for i, place := range places {
		if len(placeIDs[i]) == 0 || placeHouse[i] == nil {
			continue
		}
		if err := v.AddHouseFavor(ctx, &models.VampireHouseFavorLedger{
			HouseID: *placeHouse[i],
			Delta:   gameHFByPlace[place.pos],
			Reason:  "Game: " + game.Name,
			GMName:  gmName,
			Source:  "game",
		}); err != nil {
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

// PUT /gm/games/:id/schedule — set (or clear) a game's time slot and location.
// Times are minutes-of-day (6pm = 1080, midnight = 1440); nil = unscheduled.
func (s *server) gmSetGameSchedule(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	var body struct {
		StartMinutes *int   `json:"startMinutes"`
		EndMinutes   *int   `json:"endMinutes"`
		Location     string `json:"location"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if (body.StartMinutes == nil) != (body.EndMinutes == nil) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start and end must both be set or both cleared"})
		return
	}
	if body.StartMinutes != nil {
		if *body.StartMinutes < 0 || *body.EndMinutes > 1440 || *body.EndMinutes <= *body.StartMinutes {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid time range"})
			return
		}
	}
	if err := s.dbClient.Vampire().SetGameSchedule(ctx, id, body.StartMinutes, body.EndMinutes, body.Location); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "set_game_schedule", map[string]interface{}{"gameId": id.String()})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// sameHouse reports whether two (nullable) house ids refer to the same house.
// Two "no house" finishers count as the same; a house and a no-house do not.
func sameHouse(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// PUT /gm/games/:id — rename / reorder a game. A recorded game can't be renamed
// (its award ledger entries are matched by name); clear its result first.
func (s *server) gmUpdateGame(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	game, err := s.dbClient.Vampire().GetGameByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if game == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}
	var body struct {
		Name    string `json:"name"`
		Ordinal int    `json:"ordinal"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil || body.Name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "a game name is required"})
		return
	}
	if game.Status == "played" && body.Name != game.Name {
		ctx.JSON(http.StatusConflict, gin.H{"error": "clear this game's result before renaming it"})
		return
	}
	if err := s.dbClient.Vampire().UpdateGame(ctx, id, body.Name, body.Ordinal); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "update_game", map[string]interface{}{"gameId": id.String(), "name": body.Name})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /gm/games/:id — remove a game, reversing its awards first if recorded.
func (s *server) gmDeleteGame(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	v := s.dbClient.Vampire()
	game, err := v.GetGameByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if game == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}
	if game.Status == "played" {
		if err := v.DeleteGameAwards(ctx, game.Name); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if err := v.DeleteGame(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "delete_game", map[string]interface{}{"gameId": id.String(), "name": game.Name})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /gm/games/:id/clear — undo a recorded result: reverse the Blood Token /
// House Favor awards and reset the game to pending so it can be re-recorded.
func (s *server) gmClearGameResult(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	v := s.dbClient.Vampire()
	game, err := v.GetGameByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if game == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}
	if game.Status == "played" {
		if err := v.DeleteGameAwards(ctx, game.Name); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if err := v.ClearGameResult(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "clear_game_result", map[string]interface{}{"gameId": id.String(), "name": game.Name})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
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
