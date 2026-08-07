package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/datatypes"
)

// The shared content library editor — characters, houses, items, and quiz
// questions are global, read by every instance, so editing them here is
// restricted to super users (see withSuperUser) rather than any instance's
// Host/Co-Host. Per-instance concerns that used to live on these same
// endpoints — a character's sigil, portrait, and the real guest name playing
// them — stay on the per-instance gm routes (gm_players.go, gm_content.go).

// ---- Houses ----

// GET /admin/houses
func (s *server) adminListHouses(ctx *gin.Context) {
	houses, err := s.dbClient.Vampire().ListHouses(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"houses": houses})
}

// PUT /admin/houses/:id — edit a house's tagline.
func (s *server) adminUpdateHouse(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid house id"})
		return
	}
	var body struct {
		Tagline string `json:"tagline"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Vampire().UpdateHouseTagline(ctx, id, body.Tagline); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logSuperUser(ctx, "update_house", map[string]interface{}{"houseId": id.String()})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---- Characters ----

// GET /admin/characters — the full shared roster (every instance's superset).
func (s *server) adminListCharacters(ctx *gin.Context) {
	chars, err := s.dbClient.Vampire().ListCharacters(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(chars))
	for _, c := range chars {
		tags := []string(c.Tags)
		if tags == nil {
			tags = []string{}
		}
		row := gin.H{
			"id":         c.ID,
			"name":       c.Name,
			"title":      c.Title,
			"roleType":   c.RoleType,
			"isOptional": c.IsOptional,
			"houseId":    c.HouseID,
			"tags":       tags,
		}
		if c.House != nil {
			row["house"] = c.House.Name
		}
		out = append(out, row)
	}
	ctx.JSON(http.StatusOK, gin.H{"characters": out})
}

// GET /admin/characters/:id — a character's full editable content: bio,
// missions. No sigil/portrait/player name — those are per-instance (see GET
// /gm/characters/:id instead). No secrets either — those are mystery-scoped
// now (see MYSTERY_REQUIREMENTS.md) and edited from the Mysteries tab's
// per-character secrets editor instead, not here.
func (s *server) adminGetCharacter(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	c, err := s.dbClient.Vampire().GetCharacterByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if c == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	missions := make([]gin.H, 0, len(c.Missions))
	for _, m := range c.Missions {
		missions = append(missions, gin.H{
			"ordinal":      m.Ordinal,
			"tier":         m.Tier,
			"rewardBt":     m.RewardBT,
			"prompt":       m.Prompt,
			"answerFormat": m.AnswerFormat,
		})
	}

	tags := []string(c.Tags)
	if tags == nil {
		tags = []string{}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id":                   c.ID,
		"name":                 c.Name,
		"title":                c.Title,
		"roleType":             c.RoleType,
		"isOptional":           c.IsOptional,
		"houseId":              c.HouseID,
		"preEventInfo":         c.PreEventInfo,
		"postAct1Context":      c.PostAct1Context,
		"tags":                 tags,
		"tagsGenerationStatus": c.TagsGenerationStatus,
		"tagsGenerationError":  c.TagsGenerationError,
		"missions":             missions,
	})
}

// POST /admin/characters/:id/generate-tags — enqueue an LLM job that reads
// this character's full content (bio, secrets, missions) and proposes
// personality/trait tags, overwriting the current tag list when it
// finishes. Same fire-and-poll shape as quiz grading: this call only
// enqueues and flips the status to "queued"; the Tags field and status
// update once the job-runner worker finishes (see
// GenerateCharacterTagsProcessor in job-runner).
func (s *server) adminGenerateCharacterTags(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	if s.asyncClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "background jobs aren't configured (REDIS_URL unset)"})
		return
	}
	v := s.dbClient.Vampire()
	c, err := v.GetCharacterByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if c == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	payload, err := json.Marshal(jobs.GenerateCharacterTagsTaskPayload{CharacterID: id})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := s.asyncClient.Enqueue(
		asynq.NewTask(jobs.GenerateCharacterTagsTaskType, payload),
		asynq.Queue("grading"),
		asynq.MaxRetry(2),
	); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = v.SetCharacterTagsStatus(ctx, id, models.CharacterTagsStatusQueued, "")
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// PUT /admin/characters/:id — save the shared character editor: core
// fields and missions (missions replaced wholesale). Secrets are edited
// from the Mysteries tab instead (mystery-scoped, not here — see
// MYSTERY_REQUIREMENTS.md). Characters are only ever created by the seed
// importer (from the master packet PDF); this only edits existing ones.
func (s *server) adminUpdateCharacter(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}

	var body struct {
		Name            string   `json:"name"`
		Title           string   `json:"title"`
		RoleType        string   `json:"roleType"`
		HouseID         *string  `json:"houseId"`
		PreEventInfo    string   `json:"preEventInfo"`
		PostAct1Context string   `json:"postAct1Context"`
		Tags            []string `json:"tags"`
		Missions        []struct {
			Tier         string `json:"tier"`
			RewardBt     int    `json:"rewardBt"`
			Prompt       string `json:"prompt"`
			AnswerFormat string `json:"answerFormat"`
		} `json:"missions"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	tags := make([]string, 0, len(body.Tags))
	for _, t := range body.Tags {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}

	v := s.dbClient.Vampire()
	fields := map[string]interface{}{
		"name":              body.Name,
		"title":             body.Title,
		"role_type":         body.RoleType,
		"pre_event_info":    body.PreEventInfo,
		"post_act1_context": body.PostAct1Context,
		"tags":              models.StringArray(tags),
	}
	if body.HouseID != nil {
		if *body.HouseID == "" {
			fields["house_id"] = nil
		} else {
			hid, err := uuid.Parse(*body.HouseID)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid house id"})
				return
			}
			fields["house_id"] = hid
		}
	}
	if err := v.UpdateCharacter(ctx, id, fields); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	missions := make([]models.VampireMission, 0, len(body.Missions))
	for _, m := range body.Missions {
		if strings.TrimSpace(m.Prompt) == "" {
			continue
		}
		tier := m.Tier
		if tier == "" {
			tier = "easy"
		}
		missions = append(missions, models.VampireMission{
			Ordinal:      len(missions) + 1,
			Tier:         tier,
			RewardBT:     m.RewardBt,
			Prompt:       m.Prompt,
			AnswerFormat: m.AnswerFormat,
		})
	}
	if err := v.ReplaceMissions(ctx, id, missions); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logSuperUser(ctx, "update_character", map[string]interface{}{"characterId": id.String(), "name": body.Name})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---- Items ----

// GET /admin/items — the full catalog (not filtered to any instance's
// inclusion — see GET /gm/items for that).
func (s *server) adminListItems(ctx *gin.Context) {
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

// POST /admin/items — create a new catalog item.
func (s *server) adminCreateItem(ctx *gin.Context) {
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
	s.logSuperUser(ctx, "create_item", map[string]interface{}{"name": item.Name})
	ctx.JSON(http.StatusOK, gin.H{"id": item.ID})
}

// PUT /admin/items/:id — edit an existing catalog item.
func (s *server) adminUpdateItem(ctx *gin.Context) {
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
	s.logSuperUser(ctx, "update_item", map[string]interface{}{"id": id.String(), "name": strings.TrimSpace(body.Name)})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /admin/items/:id
func (s *server) adminDeleteItem(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := s.dbClient.Vampire().DeleteItem(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logSuperUser(ctx, "delete_item", map[string]interface{}{"id": id.String()})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /admin/items/:id/photo
func (s *server) adminSetItemPhoto(ctx *gin.Context) {
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
	s.logSuperUser(ctx, "set_item_photo", map[string]interface{}{"id": id.String()})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /admin/items/:id/photo
func (s *server) adminDeleteItemPhoto(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}
	if err := s.dbClient.Vampire().DeleteItemPhoto(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logSuperUser(ctx, "delete_item_photo", map[string]interface{}{"id": id.String()})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---- Quiz questions, scoped to one mystery ----

// GET /admin/mysteries/:id/quiz — the editable quiz for one mystery: Part 1
// open-end prompt/rubric and Part 2 multiple-choice questions, including
// the answer key (why this is super-user-only, not just editing: Co-Hosts
// shouldn't see spoilers for a story they may also be playing in another
// instance).
func (s *server) adminGetMysteryQuiz(ctx *gin.Context) {
	mysteryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid mystery id"})
		return
	}
	v := s.dbClient.Vampire()

	part1 := gin.H{"prompt": "", "rubric": "", "maxBt": 6}
	if p1, _ := v.GetPart1QuestionForMystery(ctx, mysteryID); p1 != nil {
		part1 = gin.H{"prompt": p1.Prompt, "rubric": p1.Rubric, "maxBt": p1.MaxBT}
	}

	p2qs, err := v.ListQuizQuestionsByMysteryAndPart(ctx, mysteryID, 2, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	part2 := make([]gin.H, 0, len(p2qs))
	for _, q := range p2qs {
		if q.QuestionType != "multiple_choice" {
			continue
		}
		var opts []string
		_ = json.Unmarshal(q.Options, &opts)
		part2 = append(part2, gin.H{
			"ordinal":       q.Ordinal,
			"prompt":        q.Prompt,
			"options":       opts,
			"correctAnswer": q.CorrectAnswer,
			"hfValue":       q.HFValue,
			"tier":          q.Tier,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"part1": part1, "part2": part2})
}

// PUT /admin/mysteries/:id/quiz — replace this mystery's quiz from the
// editor. Rebuilds Part 1 and the Part 2 multiple-choice set, and
// preserves any numeric questions (the Blood-Tokens-on-hand question) so
// editing MC doesn't drop them. Replacing the question set clears existing
// quiz answers in every instance running this mystery, so this is a
// pre-quiz, whole-mystery operation.
func (s *server) adminUpdateMysteryQuiz(ctx *gin.Context) {
	mysteryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid mystery id"})
		return
	}
	var body struct {
		Part1 struct {
			Prompt string `json:"prompt"`
			Rubric string `json:"rubric"`
			MaxBt  int    `json:"maxBt"`
		} `json:"part1"`
		Part2 []struct {
			Prompt        string   `json:"prompt"`
			Options       []string `json:"options"`
			CorrectAnswer string   `json:"correctAnswer"`
			HFValue       float64  `json:"hfValue"`
			Tier          string   `json:"tier"`
		} `json:"part2"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	v := s.dbClient.Vampire()
	questions := make([]models.VampireQuizQuestion, 0, len(body.Part2)+2)

	maxBT := body.Part1.MaxBt
	if maxBT <= 0 {
		maxBT = 6
	}
	questions = append(questions, models.VampireQuizQuestion{
		Part:         1,
		Ordinal:      0,
		Prompt:       body.Part1.Prompt,
		QuestionType: "open",
		Rubric:       body.Part1.Rubric,
		MaxBT:        maxBT,
		Active:       true,
	})

	ord := 1
	for _, q := range body.Part2 {
		opts, _ := json.Marshal(q.Options)
		if len(q.Options) == 0 {
			opts = []byte("[]")
		}
		questions = append(questions, models.VampireQuizQuestion{
			Part:          2,
			Ordinal:       ord,
			Prompt:        q.Prompt,
			QuestionType:  "multiple_choice",
			Options:       opts,
			CorrectAnswer: q.CorrectAnswer,
			HFValue:       q.HFValue,
			Tier:          q.Tier,
			Active:        true,
		})
		ord++
	}

	// Preserve numeric questions (e.g. Blood Tokens on hand), appended after MC.
	if existing, err := v.ListQuizQuestionsByMysteryAndPart(ctx, mysteryID, 2, true); err == nil {
		for _, q := range existing {
			if q.QuestionType != "multiple_choice" {
				questions = append(questions, models.VampireQuizQuestion{
					Part:         2,
					Ordinal:      ord,
					Prompt:       q.Prompt,
					QuestionType: q.QuestionType,
					Options:      datatypes.JSON("[]"),
					Tier:         q.Tier,
					Active:       true,
				})
				ord++
			}
		}
	}

	if err := v.ReplaceQuizQuestionsForMystery(ctx, mysteryID, questions); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logSuperUser(ctx, "update_mystery_quiz", map[string]interface{}{"mysteryId": mysteryID.String(), "part2Count": len(body.Part2)})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---- Super users ----

// GET /admin/super-users
func (s *server) adminListSuperUsers(ctx *gin.Context) {
	rows, err := s.dbClient.Vampire().ListSuperUsers(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		row := gin.H{"userId": r.UserID}
		if r.User != nil {
			row["name"] = r.User.Name
			row["email"] = r.User.Email
		}
		out = append(out, row)
	}
	ctx.JSON(http.StatusOK, gin.H{"superUsers": out})
}

// POST /admin/super-users { email } — grant another signed-up user
// shared-library edit access. Any super user can grant another, same as
// Co-Host invites — there's no separate "super-super-user" tier.
func (s *server) adminAddSuperUser(ctx *gin.Context) {
	var body struct {
		Email string `json:"email"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	email := strings.TrimSpace(body.Email)
	if email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "an email is required"})
		return
	}
	target, err := s.dbClient.User().FindByEmail(ctx, email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if target == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "no account with that email — they need to sign up first"})
		return
	}
	actor := userFromContext(ctx)
	if err := s.dbClient.Vampire().AddSuperUser(ctx, target.ID, &actor.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logSuperUser(ctx, "add_super_user", map[string]interface{}{"email": email})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /admin/super-users/:userId
func (s *server) adminRemoveSuperUser(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("userId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := s.dbClient.Vampire().RemoveSuperUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logSuperUser(ctx, "remove_super_user", map[string]interface{}{"userId": userID.String()})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}
