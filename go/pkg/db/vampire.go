package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// jsonUUIDs marshals a slice of ids into a JSON array for the game placement columns.
func jsonUUIDs(ids []uuid.UUID) datatypes.JSON {
	strs := make([]string, 0, len(ids))
	for _, x := range ids {
		strs = append(strs, x.String())
	}
	b, _ := json.Marshal(strs)
	return datatypes.JSON(b)
}

// HouseFavorStanding is a single row of the leaderboard (house + summed favor).
type HouseFavorStanding struct {
	HouseID   uuid.UUID `json:"houseId"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sortOrder"`
	Favor     float64   `json:"favor"`     // ledger total, excluding item effects
	ItemFavor float64   `json:"itemFavor"` // live overlay from owned items (not in ledger)
}

// HouseFavorSourceTotal is one house's House Favor from a single ledger source.
type HouseFavorSourceTotal struct {
	HouseID uuid.UUID `json:"houseId"`
	Source  string    `json:"source"`
	Total   float64   `json:"total"`
}

// HouseFavorBySource sums each house's ledger House Favor grouped by source
// (excluding "item", which is a live overlay computed from current ownership),
// scoped to one instance.
func (h *vampireHandler) HouseFavorBySource(ctx context.Context, instanceID uuid.UUID) ([]HouseFavorSourceTotal, error) {
	var out []HouseFavorSourceTotal
	if err := h.db.WithContext(ctx).
		Table("vampire_house_favor_ledger").
		Select("house_id, source, COALESCE(SUM(delta), 0) AS total").
		Where("instance_id = ? AND source <> 'item'", instanceID).
		Group("house_id, source").
		Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// BloodTokenTotal is a player's summed blood tokens (for resolution / reference).
type BloodTokenTotal struct {
	PlayerID uuid.UUID `json:"playerId"`
	Total    int       `json:"total"`
}

type vampireHandler struct {
	db *gorm.DB
}

// ---- Houses ----
//
// Houses are part of the shared global content library, same as characters
// and items — they are not toggled per instance (a house is implicitly
// "in" an instance if any of its characters are included there). Authoring
// (UpsertHouse/UpdateHouseTagline) stays an ops-only, global operation.

func (h *vampireHandler) ListHouses(ctx context.Context) ([]models.VampireHouse, error) {
	var houses []models.VampireHouse
	if err := h.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&houses).Error; err != nil {
		return nil, err
	}
	return houses, nil
}

func (h *vampireHandler) GetHouseByID(ctx context.Context, id uuid.UUID) (*models.VampireHouse, error) {
	var house models.VampireHouse
	if err := h.db.WithContext(ctx).First(&house, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &house, nil
}

// ListCharactersByHouse returns the playable members of a house, filtered to
// those included in the given instance.
func (h *vampireHandler) ListCharactersByHouse(ctx context.Context, instanceID, houseID uuid.UUID) ([]models.VampireCharacter, error) {
	var chars []models.VampireCharacter
	if err := h.db.WithContext(ctx).
		Table("vampire_characters c").
		Select("c.*").
		Joins("JOIN vampire_instance_characters ic ON ic.character_id = c.id AND ic.instance_id = ? AND ic.included", instanceID).
		Where("c.house_id = ? AND c.role_type = ?", houseID, "player").
		Order("c.name ASC").
		Find(&chars).Error; err != nil {
		return nil, err
	}
	return chars, nil
}

// ListHouseFavorLog returns a house's House Favor ledger for one instance, newest first.
func (h *vampireHandler) ListHouseFavorLog(ctx context.Context, instanceID, houseID uuid.UUID) ([]models.VampireHouseFavorLedger, error) {
	var entries []models.VampireHouseFavorLedger
	if err := h.db.WithContext(ctx).
		Where("instance_id = ? AND house_id = ?", instanceID, houseID).
		Order("created_at DESC").
		Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func (h *vampireHandler) UpsertHouse(ctx context.Context, name string, sortOrder int, tagline string) (*models.VampireHouse, error) {
	house := models.VampireHouse{Name: name, SortOrder: sortOrder, Tagline: tagline}
	if err := h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"sort_order": sortOrder, "tagline": tagline, "updated_at": time.Now()}),
		}).
		Create(&house).Error; err != nil {
		return nil, err
	}
	var out models.VampireHouse
	if err := h.db.WithContext(ctx).First(&out, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateHouseTagline sets a house's tagline by id (GM editor).
func (h *vampireHandler) UpdateHouseTagline(ctx context.Context, id uuid.UUID, tagline string) error {
	return h.db.WithContext(ctx).Model(&models.VampireHouse{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"tagline": tagline, "updated_at": time.Now()}).Error
}

// ---- Characters ----
//
// The character roster (name, title, house, story text, secrets, missions)
// is shared global content, authored once and reused by every instance.
// Authoring methods (Upsert/Update/Replace*) stay global/ops-only. What is
// per-instance — inclusion, sigil, portrait — lives in
// vampire_instance.go's library-inclusion methods.

func (h *vampireHandler) UpsertCharacter(ctx context.Context, c *models.VampireCharacter) (*models.VampireCharacter, error) {
	if err := h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"title", "house_id", "role_type", "is_optional",
				"pre_event_info", "post_act1_context", "updated_at",
			}),
		}).
		Create(c).Error; err != nil {
		return nil, err
	}
	return h.GetCharacterByName(ctx, c.Name)
}

// UpdateCharacter patches a character's columns by id (used by the GM content editor).
func (h *vampireHandler) UpdateCharacter(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	fields["updated_at"] = time.Now()
	return h.db.WithContext(ctx).Model(&models.VampireCharacter{}).Where("id = ?", id).Updates(fields).Error
}

func (h *vampireHandler) GetCharacterByName(ctx context.Context, name string) (*models.VampireCharacter, error) {
	var c models.VampireCharacter
	if err := h.db.WithContext(ctx).First(&c, "name = ?", name).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (h *vampireHandler) GetCharacterByID(ctx context.Context, id uuid.UUID) (*models.VampireCharacter, error) {
	var c models.VampireCharacter
	if err := h.db.WithContext(ctx).
		Preload("House").
		Preload("Secrets", func(db *gorm.DB) *gorm.DB { return db.Order("ordinal ASC") }).
		Preload("Missions", func(db *gorm.DB) *gorm.DB { return db.Order("ordinal ASC") }).
		First(&c, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// ListCharacters returns the full global roster (every character in the
// library, regardless of any instance). Used by content authoring and by
// the library-inclusion editor, not by player- or GM-facing instance views —
// see ListIncludedCharacters in vampire_instance.go for those.
func (h *vampireHandler) ListCharacters(ctx context.Context) ([]models.VampireCharacter, error) {
	var chars []models.VampireCharacter
	if err := h.db.WithContext(ctx).Preload("House").Order("name ASC").Find(&chars).Error; err != nil {
		return nil, err
	}
	return chars, nil
}

// GetActivePlayerByCharacterID returns the active player assigned to a
// character within one instance (the holder of the session token for that
// character's guest in that instance).
func (h *vampireHandler) GetActivePlayerByCharacterID(ctx context.Context, instanceID, characterID uuid.UUID) (*models.VampirePlayer, error) {
	var p models.VampirePlayer
	if err := h.db.WithContext(ctx).
		Where("instance_id = ? AND character_id = ? AND active = ?", instanceID, characterID, true).
		First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// ReplaceSecrets removes existing secrets for a character and inserts the new set.
func (h *vampireHandler) ReplaceSecrets(ctx context.Context, characterID uuid.UUID, secrets []models.VampireSecret) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("character_id = ?", characterID).Delete(&models.VampireSecret{}).Error; err != nil {
			return err
		}
		for i := range secrets {
			secrets[i].CharacterID = characterID
		}
		if len(secrets) == 0 {
			return nil
		}
		return tx.Create(&secrets).Error
	})
}

// ReplaceMissions removes existing missions for a character and inserts the new set.
func (h *vampireHandler) ReplaceMissions(ctx context.Context, characterID uuid.UUID, missions []models.VampireMission) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("character_id = ?", characterID).Delete(&models.VampireMission{}).Error; err != nil {
			return err
		}
		for i := range missions {
			missions[i].CharacterID = characterID
		}
		if len(missions) == 0 {
			return nil
		}
		return tx.Create(&missions).Error
	})
}

func (h *vampireHandler) GetMissionByID(ctx context.Context, id uuid.UUID) (*models.VampireMission, error) {
	var m models.VampireMission
	if err := h.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// ---- Players ----

// CreatePlayer inserts a player slot. p.InstanceID must be set by the caller.
func (h *vampireHandler) CreatePlayer(ctx context.Context, p *models.VampirePlayer) error {
	return h.db.WithContext(ctx).Create(p).Error
}

// GetPlayerByToken looks a player up by their opaque token. Tokens are
// globally unique (see MULTI_TENANT_REQUIREMENTS.md), so this is
// deliberately not instance-scoped — callers must check the returned
// player's InstanceID against the instance in the request URL themselves
// (see withPlayer in the server) as defense in depth.
func (h *vampireHandler) GetPlayerByToken(ctx context.Context, token string) (*models.VampirePlayer, error) {
	var p models.VampirePlayer
	if err := h.db.WithContext(ctx).
		Preload("Character").
		First(&p, "token = ?", token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (h *vampireHandler) GetPlayerByID(ctx context.Context, instanceID, id uuid.UUID) (*models.VampirePlayer, error) {
	var p models.VampirePlayer
	if err := h.db.WithContext(ctx).
		Preload("Character").
		First(&p, "instance_id = ? AND id = ?", instanceID, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (h *vampireHandler) ListPlayers(ctx context.Context, instanceID uuid.UUID) ([]models.VampirePlayer, error) {
	var players []models.VampirePlayer
	if err := h.db.WithContext(ctx).
		Preload("Character").
		Preload("Character.House").
		Where("instance_id = ?", instanceID).
		Order("guest_label ASC").
		Find(&players).Error; err != nil {
		return nil, err
	}
	return players, nil
}

func (h *vampireHandler) UpdatePlayerAssignment(ctx context.Context, instanceID, id uuid.UUID, characterID *uuid.UUID, guestLabel string, active bool) error {
	return h.db.WithContext(ctx).Model(&models.VampirePlayer{}).
		Where("instance_id = ? AND id = ?", instanceID, id).
		Updates(map[string]interface{}{
			"character_id": characterID,
			"guest_label":  guestLabel,
			"active":       active,
			"updated_at":   time.Now(),
		}).Error
}

// ---- Mission submissions ----

// UpsertMissionSubmission records a player's answer and (re)sets status to submitted.
func (h *vampireHandler) UpsertMissionSubmission(ctx context.Context, instanceID, playerID, missionID uuid.UUID, answer string) (*models.VampireMissionSubmission, error) {
	sub := models.VampireMissionSubmission{
		InstanceID:   instanceID,
		PlayerID:     playerID,
		MissionID:    missionID,
		Status:       "submitted",
		PlayerAnswer: answer,
	}
	if err := h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "player_id"}, {Name: "mission_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"player_answer": answer,
				"status":        "submitted",
				"updated_at":    time.Now(),
			}),
		}).
		Create(&sub).Error; err != nil {
		return nil, err
	}
	var out models.VampireMissionSubmission
	if err := h.db.WithContext(ctx).First(&out, "player_id = ? AND mission_id = ?", playerID, missionID).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (h *vampireHandler) ListSubmissionsForPlayer(ctx context.Context, playerID uuid.UUID) ([]models.VampireMissionSubmission, error) {
	var subs []models.VampireMissionSubmission
	if err := h.db.WithContext(ctx).Where("player_id = ?", playerID).Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}

func (h *vampireHandler) ListSubmissions(ctx context.Context, instanceID uuid.UUID, statusFilter string) ([]models.VampireMissionSubmission, error) {
	var subs []models.VampireMissionSubmission
	q := h.db.WithContext(ctx).Where("instance_id = ?", instanceID).Order("created_at ASC")
	if statusFilter != "" {
		q = q.Where("status = ?", statusFilter)
	}
	if err := q.Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}

// SubmissionDetail is a mission submission enriched with the player, character,
// house, and mission context the GM needs to adjudicate.
type SubmissionDetail struct {
	ID                  uuid.UUID `json:"id"`
	PlayerID            uuid.UUID `json:"playerId"`
	MissionID           uuid.UUID `json:"missionId"`
	Status              string    `json:"status"`
	PlayerAnswer        string    `json:"playerAnswer"`
	AwardedBT           int       `json:"awardedBt"`
	VerifiedBy          string    `json:"verifiedBy"`
	CreatedAt           time.Time `json:"createdAt"`
	GuestLabel          string    `json:"guestLabel"`
	CharacterName       string    `json:"characterName"`
	HouseName           string    `json:"houseName"`
	MissionTier         string    `json:"missionTier"`
	MissionPrompt       string    `json:"missionPrompt"`
	MissionAnswerFormat string    `json:"missionAnswerFormat"`
	RewardBT            int       `json:"rewardBt"`
}

func (h *vampireHandler) ListSubmissionsDetailed(ctx context.Context, instanceID uuid.UUID, statusFilter string) ([]SubmissionDetail, error) {
	details := []SubmissionDetail{}
	q := h.db.WithContext(ctx).
		Table("vampire_mission_submissions s").
		Select(`s.id, s.player_id, s.mission_id, s.status, s.player_answer,
			s.awarded_bt, s.verified_by, s.created_at,
			p.guest_label AS guest_label,
			c.name AS character_name,
			COALESCE(h.name, '') AS house_name,
			m.tier AS mission_tier, m.prompt AS mission_prompt,
			m.answer_format AS mission_answer_format, m.reward_bt AS reward_bt`).
		Joins("JOIN vampire_players p ON p.id = s.player_id").
		Joins("JOIN vampire_missions m ON m.id = s.mission_id").
		Joins("LEFT JOIN vampire_characters c ON c.id = p.character_id").
		Joins("LEFT JOIN vampire_houses h ON h.id = c.house_id").
		Where("s.instance_id = ?", instanceID).
		Order("s.created_at ASC")
	if statusFilter != "" {
		q = q.Where("s.status = ?", statusFilter)
	}
	if err := q.Scan(&details).Error; err != nil {
		return nil, err
	}
	return details, nil
}

func (h *vampireHandler) GetSubmissionByID(ctx context.Context, instanceID, id uuid.UUID) (*models.VampireMissionSubmission, error) {
	var sub models.VampireMissionSubmission
	if err := h.db.WithContext(ctx).First(&sub, "instance_id = ? AND id = ?", instanceID, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

// ---- Submission photos ----
//
// Not directly instance-scoped (no instance_id column) — a photo belongs to
// a submission, which belongs to a player, which belongs to an instance.
// Callers that need the scoping check (e.g. the GM approve/reject flow)
// resolve the submission through GetSubmissionByID first.

func (h *vampireHandler) AddSubmissionPhoto(ctx context.Context, submissionID uuid.UUID, contentType string, data []byte) (uuid.UUID, error) {
	photo := models.VampireSubmissionPhoto{
		SubmissionID: submissionID,
		ContentType:  contentType,
		Data:         data,
	}
	if err := h.db.WithContext(ctx).Create(&photo).Error; err != nil {
		return uuid.Nil, err
	}
	return photo.ID, nil
}

func (h *vampireHandler) DeletePhotosForSubmission(ctx context.Context, submissionID uuid.UUID) error {
	return h.db.WithContext(ctx).
		Where("submission_id = ?", submissionID).
		Delete(&models.VampireSubmissionPhoto{}).Error
}

func (h *vampireHandler) GetPhoto(ctx context.Context, id uuid.UUID) (*models.VampireSubmissionPhoto, error) {
	var photo models.VampireSubmissionPhoto
	if err := h.db.WithContext(ctx).First(&photo, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &photo, nil
}

// PhotoRef is a lightweight (id, submission) pair — never carries the bytes.
type PhotoRef struct {
	ID           uuid.UUID `json:"id"`
	SubmissionID uuid.UUID `json:"submissionId"`
}

// ListPhotoRefs returns photo ids grouped by submission (no image data) for
// one instance's submissions, for attaching to the player and GM views.
func (h *vampireHandler) ListPhotoRefs(ctx context.Context, instanceID uuid.UUID) ([]PhotoRef, error) {
	out := []PhotoRef{}
	if err := h.db.WithContext(ctx).
		Table("vampire_submission_photos ph").
		Select("ph.id, ph.submission_id").
		Joins("JOIN vampire_mission_submissions s ON s.id = ph.submission_id").
		Where("s.instance_id = ?", instanceID).
		Order("ph.created_at ASC").
		Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (h *vampireHandler) UpdateSubmissionStatus(ctx context.Context, instanceID, id uuid.UUID, status string, awardedBT int, verifiedBy string) error {
	return h.db.WithContext(ctx).Model(&models.VampireMissionSubmission{}).
		Where("instance_id = ? AND id = ?", instanceID, id).
		Updates(map[string]interface{}{
			"status":      status,
			"awarded_bt":  awardedBT,
			"verified_by": verifiedBy,
			"updated_at":  time.Now(),
		}).Error
}

// ---- House Favor ----

// AddHouseFavor inserts a ledger entry. entry.InstanceID must be set by the caller.
func (h *vampireHandler) AddHouseFavor(ctx context.Context, entry *models.VampireHouseFavorLedger) error {
	return h.db.WithContext(ctx).Create(entry).Error
}

func (h *vampireHandler) Leaderboard(ctx context.Context, instanceID uuid.UUID) ([]HouseFavorStanding, error) {
	var standings []HouseFavorStanding
	if err := h.db.WithContext(ctx).
		Table("vampire_houses h").
		Select("h.id AS house_id, h.name AS name, h.sort_order AS sort_order, COALESCE(SUM(l.delta), 0) AS favor").
		// Item House Favor is shown as a live "+X" overlay (computed from current
		// item ownership), so it is deliberately excluded from the ledger base.
		Joins("LEFT JOIN vampire_house_favor_ledger l ON l.house_id = h.id AND l.source <> 'item' AND l.instance_id = ?", instanceID).
		Group("h.id, h.name, h.sort_order").
		Order("favor DESC, h.sort_order ASC").
		Scan(&standings).Error; err != nil {
		return nil, err
	}
	return standings, nil
}

// ---- Blood Tokens ----

// AddBloodTokens inserts a log entry. entry.InstanceID must be set by the caller.
func (h *vampireHandler) AddBloodTokens(ctx context.Context, entry *models.VampireBloodTokenLog) error {
	return h.db.WithContext(ctx).Create(entry).Error
}

func (h *vampireHandler) BloodTokenTotalsByPlayer(ctx context.Context, instanceID uuid.UUID) ([]BloodTokenTotal, error) {
	var totals []BloodTokenTotal
	if err := h.db.WithContext(ctx).
		Table("vampire_blood_token_log").
		Select("player_id, COALESCE(SUM(delta), 0) AS total").
		Where("instance_id = ?", instanceID).
		Group("player_id").
		Scan(&totals).Error; err != nil {
		return nil, err
	}
	return totals, nil
}

// BloodTokenTotalsBySource sums each player's BT for a single source (e.g. "game"),
// used by the tally engine to double game winnings.
func (h *vampireHandler) BloodTokenTotalsBySource(ctx context.Context, instanceID uuid.UUID, source string) ([]BloodTokenTotal, error) {
	var totals []BloodTokenTotal
	if err := h.db.WithContext(ctx).
		Table("vampire_blood_token_log").
		Select("player_id, COALESCE(SUM(delta), 0) AS total").
		Where("instance_id = ? AND source = ?", instanceID, source).
		Group("player_id").
		Scan(&totals).Error; err != nil {
		return nil, err
	}
	return totals, nil
}

// ---- Game state ----
//
// One row per instance (instance_id is the primary key) since 000455 —
// previously a singleton row with id always 1.

func (h *vampireHandler) GetGameState(ctx context.Context, instanceID uuid.UUID) (*models.VampireGameState, error) {
	var state models.VampireGameState
	if err := h.db.WithContext(ctx).First(&state, "instance_id = ?", instanceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Lazily create the row if it wasn't seeded when the instance was created.
			state = models.VampireGameState{InstanceID: instanceID, CurrentAct: "pre_event"}
			if cerr := h.db.WithContext(ctx).Create(&state).Error; cerr != nil {
				return nil, cerr
			}
			return &state, nil
		}
		return nil, err
	}
	return &state, nil
}

func (h *vampireHandler) UpdateGameState(ctx context.Context, instanceID uuid.UUID, updates map[string]interface{}) (*models.VampireGameState, error) {
	updates["updated_at"] = time.Now()
	if err := h.db.WithContext(ctx).Model(&models.VampireGameState{}).Where("instance_id = ?", instanceID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return h.GetGameState(ctx, instanceID)
}

// ---- Notifications ----

// CreateNotification inserts a notification. n.InstanceID must be set by the caller.
func (h *vampireHandler) CreateNotification(ctx context.Context, n *models.VampireNotification) error {
	return h.db.WithContext(ctx).Create(n).Error
}

func (h *vampireHandler) DeactivateAllNotifications(ctx context.Context, instanceID uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.VampireNotification{}).
		Where("instance_id = ? AND active = ?", instanceID, true).
		Update("active", false).Error
}

// GetActiveNotificationForPlayer returns the most recent active notification
// that applies to this player: broadcast to all, to their house, or to them.
func (h *vampireHandler) GetActiveNotificationForPlayer(ctx context.Context, instanceID, playerID uuid.UUID, houseID *uuid.UUID) (*models.VampireNotification, error) {
	q := h.db.WithContext(ctx).Where("instance_id = ? AND active = ?", instanceID, true)
	if houseID != nil {
		q = q.Where(
			"scope = 'all' OR (scope = 'house' AND target_id = ?) OR (scope = 'player' AND target_id = ?)",
			*houseID, playerID,
		)
	} else {
		q = q.Where("scope = 'all' OR (scope = 'player' AND target_id = ?)", playerID)
	}
	var n models.VampireNotification
	if err := q.Order("created_at DESC").First(&n).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

func (h *vampireHandler) ListActiveNotifications(ctx context.Context, instanceID uuid.UUID) ([]models.VampireNotification, error) {
	var notifs []models.VampireNotification
	if err := h.db.WithContext(ctx).
		Where("instance_id = ? AND active = ?", instanceID, true).
		Order("created_at DESC").
		Find(&notifs).Error; err != nil {
		return nil, err
	}
	return notifs, nil
}

// ---- Quiz ----
//
// Questions are shared global content, same as characters/items — authoring
// (ReplaceQuizQuestions, the seed importer) stays global/ops-only. Which
// questions are live for a given instance is the included flag in
// vampire_instance_quiz_questions (see ListIncludedQuizQuestions* in
// vampire_instance.go). Submissions are instance-scoped.

func (h *vampireHandler) ListQuizQuestions(ctx context.Context, activeOnly bool) ([]models.VampireQuizQuestion, error) {
	var qs []models.VampireQuizQuestion
	q := h.db.WithContext(ctx).Order("ordinal ASC")
	if activeOnly {
		q = q.Where("active = ?", true)
	}
	if err := q.Find(&qs).Error; err != nil {
		return nil, err
	}
	return qs, nil
}

// ReplaceQuizQuestions wholesale-replaces the authored quiz (used by the seed
// importer). Authored before the event, so wiping/reloading is safe.
func (h *vampireHandler) ReplaceQuizQuestions(ctx context.Context, questions []models.VampireQuizQuestion) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM vampire_quiz_questions").Error; err != nil {
			return err
		}
		if len(questions) == 0 {
			return nil
		}
		return tx.Create(&questions).Error
	})
}

// QuizSubmissionDetail is a quiz answer enriched with player and question context
// for GM review (used to adjudicate open-ended answers).
type QuizSubmissionDetail struct {
	ID            uuid.UUID `json:"id"`
	PlayerID      uuid.UUID `json:"playerId"`
	QuestionID    uuid.UUID `json:"questionId"`
	Part          int       `json:"part"`
	Answer        string    `json:"answer"`
	IsCorrect     *bool     `json:"isCorrect"`
	AIScore       *float64  `json:"aiScore"`
	AIRationale   string    `json:"aiRationale"`
	AwardedBT     int       `json:"awardedBt"`
	Locked        bool      `json:"locked"`
	GradeStatus   string    `json:"gradeStatus"`
	GradeError    string    `json:"gradeError"`
	GradeAttempts int       `json:"gradeAttempts"`
	GuestLabel    string    `json:"guestLabel"`
	CharacterName string    `json:"characterName"`
	HouseName     string    `json:"houseName"`
	Ordinal       int       `json:"ordinal"`
	Prompt        string    `json:"prompt"`
	QuestionType  string    `json:"questionType"`
}

func (h *vampireHandler) ListQuizSubmissionsDetailed(ctx context.Context, instanceID uuid.UUID) ([]QuizSubmissionDetail, error) {
	out := []QuizSubmissionDetail{}
	if err := h.db.WithContext(ctx).
		Table("vampire_quiz_submissions s").
		Select(`s.id, s.player_id, s.question_id, s.answer, s.is_correct, s.ai_score,
			s.ai_rationale, s.awarded_bt, s.locked,
			s.grade_status, s.grade_error, s.grade_attempts,
			q.part AS part,
			p.guest_label AS guest_label,
			COALESCE(c.name, '') AS character_name,
			COALESCE(h.name, '') AS house_name,
			q.ordinal AS ordinal, q.prompt AS prompt, q.question_type AS question_type`).
		Joins("JOIN vampire_players p ON p.id = s.player_id").
		Joins("JOIN vampire_quiz_questions q ON q.id = s.question_id").
		Joins("LEFT JOIN vampire_characters c ON c.id = p.character_id").
		Joins("LEFT JOIN vampire_houses h ON h.id = c.house_id").
		Where("s.instance_id = ?", instanceID).
		Order("q.part ASC, q.ordinal ASC, character_name ASC").
		Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// GetPart1Question returns the library's single active Part 1 (open-end)
// question, unscoped — used by the content editor (which edits the shared
// library, not one instance's play state). See GetIncludedPart1Question for
// the play-time, instance-scoped equivalent.
func (h *vampireHandler) GetPart1Question(ctx context.Context) (*models.VampireQuizQuestion, error) {
	var qq models.VampireQuizQuestion
	if err := h.db.WithContext(ctx).
		Where("part = ? AND active = ?", 1, true).
		Order("ordinal ASC").First(&qq).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &qq, nil
}

func (h *vampireHandler) ListQuizQuestionsByPart(ctx context.Context, part int, activeOnly bool) ([]models.VampireQuizQuestion, error) {
	var qs []models.VampireQuizQuestion
	q := h.db.WithContext(ctx).Where("part = ?", part).Order("ordinal ASC")
	if activeOnly {
		q = q.Where("active = ?", true)
	}
	if err := q.Find(&qs).Error; err != nil {
		return nil, err
	}
	return qs, nil
}

func (h *vampireHandler) UpdateQuizSubmissionGrade(ctx context.Context, id uuid.UUID, aiScore *float64, awardedBT int) error {
	return h.db.WithContext(ctx).Model(&models.VampireQuizSubmission{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"ai_score":   aiScore,
			"awarded_bt": awardedBT,
			"updated_at": time.Now(),
		}).Error
}

// SetQuizSubmissionRationale stores the one-line AI note for a Part 1 grade.
func (h *vampireHandler) SetQuizSubmissionRationale(ctx context.Context, id uuid.UUID, rationale string) error {
	return h.db.WithContext(ctx).Model(&models.VampireQuizSubmission{}).
		Where("id = ?", id).
		Update("ai_rationale", rationale).Error
}

// SetQuizGradeStatus records a grading state transition (queued / graded / failed)
// along with any error message.
func (h *vampireHandler) SetQuizGradeStatus(ctx context.Context, id uuid.UUID, status, errMsg string) error {
	return h.db.WithContext(ctx).Model(&models.VampireQuizSubmission{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"grade_status": status,
			"grade_error":  errMsg,
			"updated_at":   time.Now(),
		}).Error
}

// MarkQuizGradeStarted flips a submission into the "grading" state, stamping the
// start time and bumping the attempt count (for stuck detection / diagnostics).
func (h *vampireHandler) MarkQuizGradeStarted(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return h.db.WithContext(ctx).Model(&models.VampireQuizSubmission{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"grade_status":     models.QuizGradeStatusGrading,
			"grade_started_at": now,
			"grade_attempts":   gorm.Expr("grade_attempts + 1"),
			"updated_at":       now,
		}).Error
}

// SetCharacterTagsStatus records an AI tag-generation state transition
// (queued / generating / generated / failed) along with any error message —
// the character equivalent of SetQuizGradeStatus above.
func (h *vampireHandler) SetCharacterTagsStatus(ctx context.Context, id uuid.UUID, status, errMsg string) error {
	return h.db.WithContext(ctx).Model(&models.VampireCharacter{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"tags_generation_status": status,
			"tags_generation_error":  errMsg,
			"updated_at":             time.Now(),
		}).Error
}

// Part2Answer is one player's answer to a Part 2 question, with their house —
// the raw material for the normalized per-house scoring.
type Part2Answer struct {
	PlayerID   uuid.UUID `json:"playerId"`
	HouseID    uuid.UUID `json:"houseId"`
	QuestionID uuid.UUID `json:"questionId"`
	Answer     string    `json:"answer"`
}

func (h *vampireHandler) ListPart2Answers(ctx context.Context, instanceID uuid.UUID) ([]Part2Answer, error) {
	out := []Part2Answer{}
	if err := h.db.WithContext(ctx).
		Table("vampire_quiz_submissions s").
		Select("s.player_id, c.house_id AS house_id, s.question_id, s.answer").
		Joins("JOIN vampire_quiz_questions q ON q.id = s.question_id AND q.part = 2").
		Joins("JOIN vampire_players p ON p.id = s.player_id").
		Joins("JOIN vampire_characters c ON c.id = p.character_id").
		Where("s.instance_id = ? AND c.house_id IS NOT NULL", instanceID).
		Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteHouseFavorBySource removes one instance's ledger entries of a given
// source (used to idempotently re-score Part 2).
func (h *vampireHandler) DeleteHouseFavorBySource(ctx context.Context, instanceID uuid.UUID, source string) error {
	return h.db.WithContext(ctx).
		Where("instance_id = ? AND source = ?", instanceID, source).
		Delete(&models.VampireHouseFavorLedger{}).Error
}

// DeleteBloodTokensBySourceForPlayer removes a player's BT entries of a given
// source (used to idempotently re-grade Part 1).
func (h *vampireHandler) DeleteBloodTokensBySourceForPlayer(ctx context.Context, playerID uuid.UUID, source string) error {
	return h.db.WithContext(ctx).
		Where("player_id = ? AND source = ?", playerID, source).
		Delete(&models.VampireBloodTokenLog{}).Error
}

func (h *vampireHandler) GetQuizQuestionByID(ctx context.Context, id uuid.UUID) (*models.VampireQuizQuestion, error) {
	var qq models.VampireQuizQuestion
	if err := h.db.WithContext(ctx).First(&qq, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &qq, nil
}

func (h *vampireHandler) UpsertQuizSubmission(ctx context.Context, instanceID, playerID, questionID uuid.UUID, answer string, isCorrect *bool, locked bool) (*models.VampireQuizSubmission, error) {
	sub := models.VampireQuizSubmission{
		InstanceID: instanceID,
		PlayerID:   playerID,
		QuestionID: questionID,
		Answer:     answer,
		IsCorrect:  isCorrect,
		Locked:     locked,
	}
	if err := h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "player_id"}, {Name: "question_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"answer":     answer,
				"is_correct": isCorrect,
				"locked":     locked,
				"updated_at": time.Now(),
			}),
		}).
		Create(&sub).Error; err != nil {
		return nil, err
	}
	var out models.VampireQuizSubmission
	if err := h.db.WithContext(ctx).First(&out, "player_id = ? AND question_id = ?", playerID, questionID).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (h *vampireHandler) ListQuizSubmissionsForPlayer(ctx context.Context, playerID uuid.UUID) ([]models.VampireQuizSubmission, error) {
	var subs []models.VampireQuizSubmission
	if err := h.db.WithContext(ctx).Where("player_id = ?", playerID).Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}

func (h *vampireHandler) ListQuizSubmissions(ctx context.Context, instanceID uuid.UUID) ([]models.VampireQuizSubmission, error) {
	var subs []models.VampireQuizSubmission
	if err := h.db.WithContext(ctx).Where("instance_id = ?", instanceID).Order("created_at ASC").Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}

// ResetGameProgress wipes all play progress for one instance's clean playtest
// run: mission submissions, both ledgers, quiz submissions, notifications,
// and the audit log, then resets that instance's game state to a sealed
// pre-event. The roster (houses, characters, secrets, missions), the
// instance's player slots + token assignments, and quiz questions are
// preserved.
func (h *vampireHandler) ResetGameProgress(ctx context.Context, instanceID uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Snapshot this instance's score ledgers before wiping them, so a reset
		// is always recoverable. Same transaction as the delete, so archive and
		// wipe are atomic — we never delete without a durable copy landing
		// first. Explicit column lists (not SELECT *) so the live and archive
		// tables' column orders never have to match.
		if err := tx.Exec(`
			INSERT INTO vampire_house_favor_ledger_archive
				(id, created_at, instance_id, house_id, delta, reason, gm_name, source, archived_at)
			SELECT id, created_at, instance_id, house_id, delta, reason, gm_name, source, now()
			FROM vampire_house_favor_ledger WHERE instance_id = ?`, instanceID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			INSERT INTO vampire_blood_token_log_archive
				(id, created_at, instance_id, player_id, delta, reason, source, gm_name, archived_at)
			SELECT id, created_at, instance_id, player_id, delta, reason, source, gm_name, now()
			FROM vampire_blood_token_log WHERE instance_id = ?`, instanceID).Error; err != nil {
			return err
		}

		tables := []string{
			"vampire_mission_submissions",
			"vampire_house_favor_ledger",
			"vampire_blood_token_log",
			"vampire_quiz_submissions",
			"vampire_notifications",
			"vampire_gm_action_log",
		}
		for _, table := range tables {
			if err := tx.Exec("DELETE FROM "+table+" WHERE instance_id = ?", instanceID).Error; err != nil {
				return err
			}
		}
		return tx.Model(&models.VampireGameState{}).Where("instance_id = ?", instanceID).Updates(map[string]interface{}{
			"current_act":            "pre_event",
			"content_unlocked":       false,
			"quiz_part1_open":        false,
			"quiz_part2_open":        false,
			"quiz_part1_opened_at":   nil,
			"active_notification_id": nil,
			"updated_at":             time.Now(),
		}).Error
	})
}

// WipeCharactersAndRoster clears the global character library (and cascades
// their secrets/missions) so a seed run can rebuild it from scratch (used
// for the --fresh re-seed). Score ledgers are archived first so the wipe is
// recoverable, and houses are left intact.
//
// Multi-tenant caveat: characters are now shared library content, referenced
// by every instance's vampire_instance_characters rows (ON DELETE CASCADE).
// Wiping the library therefore drops every instance's inclusion/sigil/
// portrait rows for the deleted characters too — this is a global content
// operation, not a per-instance one, and re-seeding immediately after
// re-creates the characters but NOT each instance's inclusion rows (those
// only get seeded when an instance is created). Only run --fresh when you
// mean to affect every instance built on this library.
func (h *vampireHandler) WipeCharactersAndRoster(ctx context.Context) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			INSERT INTO vampire_house_favor_ledger_archive
				(id, created_at, instance_id, house_id, delta, reason, gm_name, source, archived_at)
			SELECT id, created_at, instance_id, house_id, delta, reason, gm_name, source, now()
			FROM vampire_house_favor_ledger`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			INSERT INTO vampire_blood_token_log_archive
				(id, created_at, instance_id, player_id, delta, reason, source, gm_name, archived_at)
			SELECT id, created_at, instance_id, player_id, delta, reason, source, gm_name, now()
			FROM vampire_blood_token_log`).Error; err != nil {
			return err
		}
		// Characters only — secrets/missions cascade via FK. Players are now
		// per-instance data and are deliberately left alone by a content wipe.
		return tx.Exec("DELETE FROM vampire_characters").Error
	})
}

// ---- Physical games ----

func (h *vampireHandler) ListGames(ctx context.Context, instanceID uuid.UUID) ([]models.VampireGame, error) {
	var games []models.VampireGame
	if err := h.db.WithContext(ctx).
		Where("instance_id = ?", instanceID).
		Order("ordinal ASC, created_at ASC").
		Find(&games).Error; err != nil {
		return nil, err
	}
	return games, nil
}

func (h *vampireHandler) GetGameByID(ctx context.Context, instanceID, id uuid.UUID) (*models.VampireGame, error) {
	var g models.VampireGame
	if err := h.db.WithContext(ctx).First(&g, "instance_id = ? AND id = ?", instanceID, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

// UpsertGame creates or updates a game by (instance, name) — idempotent seeding.
func (h *vampireHandler) UpsertGame(ctx context.Context, instanceID uuid.UUID, ordinal int, name string) (*models.VampireGame, error) {
	game := models.VampireGame{InstanceID: instanceID, Name: name, Ordinal: ordinal}
	if err := h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "instance_id"}, {Name: "name"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"ordinal": ordinal, "updated_at": time.Now()}),
		}).
		Create(&game).Error; err != nil {
		return nil, err
	}
	var out models.VampireGame
	if err := h.db.WithContext(ctx).First(&out, "instance_id = ? AND name = ?", instanceID, name).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (h *vampireHandler) SetGameResult(ctx context.Context, instanceID, id uuid.UUID, first, second, third []uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.VampireGame{}).
		Where("instance_id = ? AND id = ?", instanceID, id).
		Updates(map[string]interface{}{
			"status":               "played",
			"first_character_ids":  jsonUUIDs(first),
			"second_character_ids": jsonUUIDs(second),
			"third_character_ids":  jsonUUIDs(third),
			"updated_at":           time.Now(),
		}).Error
}

// SetGameSchedule sets (or clears, with nil times) a game's slot, location,
// assigned GM, and GM-only run notes.
func (h *vampireHandler) SetGameSchedule(ctx context.Context, instanceID, id uuid.UUID, start, end *int, location, assignedGM, runNotes string) error {
	return h.db.WithContext(ctx).Model(&models.VampireGame{}).
		Where("instance_id = ? AND id = ?", instanceID, id).
		Updates(map[string]interface{}{
			"start_minutes": start,
			"end_minutes":   end,
			"location":      location,
			"assigned_gm":   assignedGM,
			"run_notes":     runNotes,
			"updated_at":    time.Now(),
		}).Error
}

func (h *vampireHandler) UpdateGame(ctx context.Context, instanceID, id uuid.UUID, name string, ordinal int) error {
	return h.db.WithContext(ctx).Model(&models.VampireGame{}).
		Where("instance_id = ? AND id = ?", instanceID, id).
		Updates(map[string]interface{}{"name": name, "ordinal": ordinal, "updated_at": time.Now()}).Error
}

func (h *vampireHandler) DeleteGame(ctx context.Context, instanceID, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.VampireGame{}, "instance_id = ? AND id = ?", instanceID, id).Error
}

// ClearGameResult resets a game to pending and drops its recorded finishers.
func (h *vampireHandler) ClearGameResult(ctx context.Context, instanceID, id uuid.UUID) error {
	empty := datatypes.JSON([]byte("[]"))
	return h.db.WithContext(ctx).Model(&models.VampireGame{}).
		Where("instance_id = ? AND id = ?", instanceID, id).
		Updates(map[string]interface{}{
			"status":               "pending",
			"first_character_ids":  empty,
			"second_character_ids": empty,
			"third_character_ids":  empty,
			"updated_at":           time.Now(),
		}).Error
}

func (h *vampireHandler) DeleteGameAwards(ctx context.Context, instanceID uuid.UUID, gameName string) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"DELETE FROM vampire_house_favor_ledger WHERE instance_id = ? AND source = 'game' AND reason = ?",
			instanceID, "Game: "+gameName,
		).Error; err != nil {
			return err
		}
		return tx.Exec(
			"DELETE FROM vampire_blood_token_log WHERE instance_id = ? AND source = 'game' AND reason IN (?, ?)",
			instanceID, "Game: "+gameName, "Game participation: "+gameName,
		).Error
	})
}

// ---- Inventory (item catalog) ----
//
// The catalog is shared global content, same as characters — authoring
// stays global/ops-only. Per-instance inclusion lives in
// vampire_instance_items (see vampire_instance.go). Player *ownership* of
// an item (VampirePlayerItem) is scoped implicitly through the owning
// player's instance.

func (h *vampireHandler) ListItems(ctx context.Context) ([]models.VampireItem, error) {
	var items []models.VampireItem
	if err := h.db.WithContext(ctx).Order("name ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// UpsertItem creates or updates an item by name (idempotent seeding).
func (h *vampireHandler) UpsertItem(ctx context.Context, item *models.VampireItem) error {
	return h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"code", "category", "description", "effect", "targets_player", "hf_effect",
				"bt_self", "bt_from_target", "bt_deduct_target", "quiz_bt_pct",
				"double_game_bt", "immune", "reflect", "strip_resistance", "updated_at",
			}),
		}).
		Create(item).Error
}

// CreateItem inserts a new catalog item (GM-authored). Fails on duplicate name.
func (h *vampireHandler) CreateItem(ctx context.Context, item *models.VampireItem) error {
	return h.db.WithContext(ctx).Create(item).Error
}

// UpdateItem edits every mutable field of a catalog item by id. Uses a map so
// booleans cleared to false are written (a struct update would skip zero values).
func (h *vampireHandler) UpdateItem(ctx context.Context, id uuid.UUID, item *models.VampireItem) error {
	return h.db.WithContext(ctx).Model(&models.VampireItem{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"code":             item.Code,
			"name":             item.Name,
			"description":      item.Description,
			"effect":           item.Effect,
			"targets_player":   item.TargetsPlayer,
			"hf_effect":        item.HFEffect,
			"bt_self":          item.BTSelf,
			"bt_from_target":   item.BTFromTarget,
			"bt_deduct_target": item.BTDeductTarget,
			"quiz_bt_pct":      item.QuizBTPct,
			"double_game_bt":   item.DoubleGameBT,
			"immune":           item.Immune,
			"reflect":          item.Reflect,
			"strip_resistance": item.StripResistance,
			"updated_at":       time.Now(),
		}).Error
}

// DeleteItem removes a catalog item; its player assignments cascade away.
func (h *vampireHandler) DeleteItem(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.VampireItem{}, "id = ?", id).Error
}

// SetItemPhoto stores (or replaces) a catalog item's reference photo.
func (h *vampireHandler) SetItemPhoto(ctx context.Context, itemID uuid.UUID, contentType string, data []byte) error {
	photo := models.VampireItemPhoto{ItemID: itemID, ContentType: contentType, Data: data, UpdatedAt: time.Now()}
	return h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "item_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"content_type", "data", "updated_at"}),
		}).
		Create(&photo).Error
}

// GetItemPhoto returns a catalog item's photo, or nil if none.
func (h *vampireHandler) GetItemPhoto(ctx context.Context, itemID uuid.UUID) (*models.VampireItemPhoto, error) {
	var photo models.VampireItemPhoto
	if err := h.db.WithContext(ctx).First(&photo, "item_id = ?", itemID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &photo, nil
}

// DeleteItemPhoto removes a catalog item's photo.
func (h *vampireHandler) DeleteItemPhoto(ctx context.Context, itemID uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.VampireItemPhoto{}, "item_id = ?", itemID).Error
}

// ItemPhotoIDs returns the item ids that currently have a photo (no bytes fetched).
func (h *vampireHandler) ItemPhotoIDs(ctx context.Context) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	if err := h.db.WithContext(ctx).
		Model(&models.VampireItemPhoto{}).
		Pluck("item_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func (h *vampireHandler) ListPlayerItems(ctx context.Context, playerID uuid.UUID) ([]models.VampirePlayerItem, error) {
	var pis []models.VampirePlayerItem
	if err := h.db.WithContext(ctx).
		Preload("Item").
		Where("player_id = ?", playerID).
		Order("created_at ASC").
		Find(&pis).Error; err != nil {
		return nil, err
	}
	return pis, nil
}

// ListAllPlayerItems returns every item assignment for players in one instance.
func (h *vampireHandler) ListAllPlayerItems(ctx context.Context, instanceID uuid.UUID) ([]models.VampirePlayerItem, error) {
	var pis []models.VampirePlayerItem
	if err := h.db.WithContext(ctx).
		Preload("Item").
		Joins("JOIN vampire_players p ON p.id = vampire_player_items.player_id").
		Where("p.instance_id = ?", instanceID).
		Find(&pis).Error; err != nil {
		return nil, err
	}
	return pis, nil
}

func (h *vampireHandler) AssignItem(ctx context.Context, playerID, itemID uuid.UUID) (*models.VampirePlayerItem, error) {
	pi := models.VampirePlayerItem{PlayerID: playerID, ItemID: itemID}
	if err := h.db.WithContext(ctx).Create(&pi).Error; err != nil {
		return nil, err
	}
	return &pi, nil
}

func (h *vampireHandler) DeletePlayerItem(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.VampirePlayerItem{}, "id = ?", id).Error
}

// TransferPlayerItem moves an owned item to a different player, clearing any
// target the previous owner had chosen (the new owner picks their own).
func (h *vampireHandler) TransferPlayerItem(ctx context.Context, id, newPlayerID uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.VampirePlayerItem{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"player_id":        newPlayerID,
			"target_player_id": nil,
		}).Error
}

func (h *vampireHandler) SetPlayerItemTarget(ctx context.Context, id uuid.UUID, targetPlayerID *uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.VampirePlayerItem{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"target_player_id": targetPlayerID, "updated_at": time.Now()}).Error
}

// ---- GM audit log ----

// LogGMAction records an admin action for one instance.
func (h *vampireHandler) LogGMAction(ctx context.Context, instanceID uuid.UUID, gmName, action string, payload []byte) error {
	entry := models.VampireGMActionLog{
		InstanceID: instanceID,
		GMName:     gmName,
		Action:     action,
		Payload:    payload,
	}
	return h.db.WithContext(ctx).Create(&entry).Error
}
