package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type itemWithPhoto struct {
	models.VampireItem
	HasPhoto bool `json:"hasPhoto"`
}

// GET /gm/items — the item catalog, for the assign dropdown, with a hasPhoto flag.
func (s *server) gmListItems(ctx *gin.Context) {
	v := s.dbClient.Vampire()
	items, err := v.ListItems(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	photoIDs, _ := v.ItemPhotoIDs(ctx)
	has := map[string]bool{}
	for _, id := range photoIDs {
		has[id.String()] = true
	}
	out := make([]itemWithPhoto, 0, len(items))
	for _, it := range items {
		out = append(out, itemWithPhoto{VampireItem: it, HasPhoto: has[it.ID.String()]})
	}
	ctx.JSON(http.StatusOK, gin.H{"items": out})
}

// GET /items/:id/photo — serve a catalog item's reference photo (no auth; not
// secret, and the catalog id isn't exposed to players — lets <img src> load it).
func (s *server) getItemPhoto(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}
	photo, err := s.dbClient.Vampire().GetItemPhoto(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if photo == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "no photo"})
		return
	}
	ctx.Header("Cache-Control", "private, max-age=3600")
	ctx.Data(http.StatusOK, photo.ContentType, photo.Data)
}

// POST /gm/items/:id/photo — set a catalog item's reference photo from a data URL.
func (s *server) gmSetItemPhoto(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}
	var body struct {
		DataUrl string `json:"dataUrl"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ct, data, err := decodeDataURL(body.DataUrl)
	if err != nil || len(data) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid image"})
		return
	}
	if len(data) > maxPhotoBytes {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "image too large"})
		return
	}
	if err := s.dbClient.Vampire().SetItemPhoto(ctx, id, ct, data); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "set_item_photo", map[string]interface{}{"id": id.String()})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /gm/items/:id/photo — remove a catalog item's reference photo.
func (s *server) gmDeleteItemPhoto(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}
	if err := s.dbClient.Vampire().DeleteItemPhoto(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "delete_item_photo", map[string]interface{}{"id": id.String()})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

type itemBody struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	Category        string `json:"category"`
	Description     string `json:"description"`
	Effect          string `json:"effect"`
	TargetsPlayer   bool   `json:"targetsPlayer"`
	HFEffect        int    `json:"hfEffect"`
	BTSelf          int    `json:"btSelf"`
	BTFromTarget    int    `json:"btFromTarget"`
	BTDeductTarget  int    `json:"btDeductTarget"`
	QuizBTPct       int    `json:"quizBtPct"`
	DoubleGameBT    bool   `json:"doubleGameBt"`
	Immune          bool   `json:"immune"`
	Reflect         bool   `json:"reflect"`
	StripResistance bool   `json:"stripResistance"`
}

func (b itemBody) toModel() *models.VampireItem {
	return &models.VampireItem{
		Code:            b.Code,
		Name:            strings.TrimSpace(b.Name),
		Category:        b.Category,
		Description:     b.Description,
		Effect:          b.Effect,
		TargetsPlayer:   b.TargetsPlayer,
		HFEffect:        b.HFEffect,
		BTSelf:          b.BTSelf,
		BTFromTarget:    b.BTFromTarget,
		BTDeductTarget:  b.BTDeductTarget,
		QuizBTPct:       b.QuizBTPct,
		DoubleGameBT:    b.DoubleGameBT,
		Immune:          b.Immune,
		Reflect:         b.Reflect,
		StripResistance: b.StripResistance,
	}
}

// POST /gm/items — create a new catalog item (GM-authored).
func (s *server) gmCreateItem(ctx *gin.Context) {
	var body itemBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	item := body.toModel()
	if err := s.dbClient.Vampire().CreateItem(ctx, item); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "create_item", map[string]interface{}{"name": item.Name})
	ctx.JSON(http.StatusOK, gin.H{"id": item.ID})
}

// PUT /gm/items/:id — edit an existing catalog item.
func (s *server) gmUpdateItem(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body itemBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if err := s.dbClient.Vampire().UpdateItem(ctx, id, body.toModel()); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "update_item", map[string]interface{}{"id": id.String(), "name": strings.TrimSpace(body.Name)})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /gm/items/:id — remove a catalog item (assignments cascade away).
func (s *server) gmDeleteItem(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := s.dbClient.Vampire().DeleteItem(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "delete_item", map[string]interface{}{"id": id.String()})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// GET /gm/player-items — every assignment, with owner + target names for display.
func (s *server) gmListPlayerItems(ctx *gin.Context) {
	v := s.dbClient.Vampire()
	pis, err := v.ListAllPlayerItems(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	name := s.playerNameLookup(ctx)
	out := make([]gin.H, 0, len(pis))
	for _, pi := range pis {
		row := gin.H{
			"id":         pi.ID,
			"playerId":   pi.PlayerID,
			"playerName": name(pi.PlayerID),
			"targetName": nil,
		}
		if pi.Item != nil {
			row["itemName"] = pi.Item.Name
			row["effect"] = pi.Item.Effect
			row["targetsPlayer"] = pi.Item.TargetsPlayer
		}
		if pi.TargetPlayerID != nil {
			row["targetName"] = name(*pi.TargetPlayerID)
		}
		out = append(out, row)
	}
	ctx.JSON(http.StatusOK, gin.H{"playerItems": out})
}

// POST /gm/player-items — assign an item to a player.
func (s *server) gmAssignItem(ctx *gin.Context) {
	var body struct {
		PlayerID string `json:"playerId"`
		ItemID   string `json:"itemId"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pid, err := uuid.Parse(body.PlayerID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}
	iid, err := uuid.Parse(body.ItemID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}
	pi, err := s.dbClient.Vampire().AssignItem(ctx, pid, iid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "assign_item", map[string]interface{}{"playerId": body.PlayerID, "itemId": body.ItemID})
	ctx.JSON(http.StatusOK, gin.H{"id": pi.ID})
}

// PUT /gm/player-items/:id/owner — transfer an owned item to a different player.
func (s *server) gmTransferPlayerItem(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body struct {
		PlayerID string `json:"playerId"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pid, err := uuid.Parse(body.PlayerID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}
	if err := s.dbClient.Vampire().TransferPlayerItem(ctx, id, pid); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "transfer_player_item", map[string]interface{}{"id": id.String(), "toPlayerId": body.PlayerID})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /gm/player-items/:id — remove an assignment.
func (s *server) gmRemovePlayerItem(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := s.dbClient.Vampire().DeletePlayerItem(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "remove_player_item", map[string]interface{}{"id": id.String()})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// playerNameLookup returns a func mapping a player id to a display name
// (character name, else guest label).
func (s *server) playerNameLookup(ctx context.Context) func(uuid.UUID) string {
	players, _ := s.dbClient.Vampire().ListPlayers(ctx)
	m := map[string]string{}
	for _, p := range players {
		n := p.GuestLabel
		if p.Character != nil && p.Character.Name != "" {
			n = p.Character.Name
		}
		m[p.ID.String()] = n
	}
	return func(id uuid.UUID) string { return m[id.String()] }
}

