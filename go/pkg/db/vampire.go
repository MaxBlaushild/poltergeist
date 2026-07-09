package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// HouseFavorStanding is a single row of the leaderboard (house + summed favor).
type HouseFavorStanding struct {
	HouseID   uuid.UUID `json:"houseId"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sortOrder"`
	Favor     float64   `json:"favor"`
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

// ListCharactersByHouse returns the playable members of a house.
func (h *vampireHandler) ListCharactersByHouse(ctx context.Context, houseID uuid.UUID) ([]models.VampireCharacter, error) {
	var chars []models.VampireCharacter
	if err := h.db.WithContext(ctx).
		Where("house_id = ? AND role_type = ?", houseID, "player").
		Order("name ASC").
		Find(&chars).Error; err != nil {
		return nil, err
	}
	return chars, nil
}

// ListHouseFavorLog returns a house's House Favor ledger, newest first.
func (h *vampireHandler) ListHouseFavorLog(ctx context.Context, houseID uuid.UUID) ([]models.VampireHouseFavorLedger, error) {
	var entries []models.VampireHouseFavorLedger
	if err := h.db.WithContext(ctx).
		Where("house_id = ?", houseID).
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

// UpdateCharacter patches a character's columns by id (used by the GM editor).
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

func (h *vampireHandler) ListCharacters(ctx context.Context) ([]models.VampireCharacter, error) {
	var chars []models.VampireCharacter
	if err := h.db.WithContext(ctx).Preload("House").Order("name ASC").Find(&chars).Error; err != nil {
		return nil, err
	}
	return chars, nil
}

func (h *vampireHandler) SetCharacterPassword(ctx context.Context, characterID uuid.UUID, password string) error {
	return h.db.WithContext(ctx).Model(&models.VampireCharacter{}).
		Where("id = ?", characterID).
		Updates(map[string]interface{}{"password": password, "updated_at": time.Now()}).Error
}

// GetActivePlayerByCharacterID returns the active player assigned to a character
// (the holder of the session token for that character's guest).
func (h *vampireHandler) GetActivePlayerByCharacterID(ctx context.Context, characterID uuid.UUID) (*models.VampirePlayer, error) {
	var p models.VampirePlayer
	if err := h.db.WithContext(ctx).
		Where("character_id = ? AND active = ?", characterID, true).
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

func (h *vampireHandler) CreatePlayer(ctx context.Context, p *models.VampirePlayer) error {
	return h.db.WithContext(ctx).Create(p).Error
}

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

func (h *vampireHandler) GetPlayerByID(ctx context.Context, id uuid.UUID) (*models.VampirePlayer, error) {
	var p models.VampirePlayer
	if err := h.db.WithContext(ctx).Preload("Character").First(&p, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (h *vampireHandler) ListPlayers(ctx context.Context) ([]models.VampirePlayer, error) {
	var players []models.VampirePlayer
	if err := h.db.WithContext(ctx).
		Preload("Character").
		Preload("Character.House").
		Order("guest_label ASC").
		Find(&players).Error; err != nil {
		return nil, err
	}
	return players, nil
}

func (h *vampireHandler) UpdatePlayerAssignment(ctx context.Context, id uuid.UUID, characterID *uuid.UUID, guestLabel string, active bool) error {
	return h.db.WithContext(ctx).Model(&models.VampirePlayer{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"character_id": characterID,
			"guest_label":  guestLabel,
			"active":       active,
			"updated_at":   time.Now(),
		}).Error
}

// ---- Mission submissions ----

// UpsertMissionSubmission records a player's answer and (re)sets status to submitted.
func (h *vampireHandler) UpsertMissionSubmission(ctx context.Context, playerID, missionID uuid.UUID, answer string) (*models.VampireMissionSubmission, error) {
	sub := models.VampireMissionSubmission{
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

func (h *vampireHandler) ListSubmissions(ctx context.Context, statusFilter string) ([]models.VampireMissionSubmission, error) {
	var subs []models.VampireMissionSubmission
	q := h.db.WithContext(ctx).Order("created_at ASC")
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

func (h *vampireHandler) ListSubmissionsDetailed(ctx context.Context, statusFilter string) ([]SubmissionDetail, error) {
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
		Order("s.created_at ASC")
	if statusFilter != "" {
		q = q.Where("s.status = ?", statusFilter)
	}
	if err := q.Scan(&details).Error; err != nil {
		return nil, err
	}
	return details, nil
}

func (h *vampireHandler) GetSubmissionByID(ctx context.Context, id uuid.UUID) (*models.VampireMissionSubmission, error) {
	var sub models.VampireMissionSubmission
	if err := h.db.WithContext(ctx).First(&sub, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

// ---- Submission photos ----

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

// ListPhotoRefs returns photo ids grouped by submission (no image data), for
// attaching to the player and GM submission views.
func (h *vampireHandler) ListPhotoRefs(ctx context.Context) ([]PhotoRef, error) {
	out := []PhotoRef{}
	if err := h.db.WithContext(ctx).
		Table("vampire_submission_photos").
		Select("id, submission_id").
		Order("created_at ASC").
		Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (h *vampireHandler) UpdateSubmissionStatus(ctx context.Context, id uuid.UUID, status string, awardedBT int, verifiedBy string) error {
	return h.db.WithContext(ctx).Model(&models.VampireMissionSubmission{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      status,
			"awarded_bt":  awardedBT,
			"verified_by": verifiedBy,
			"updated_at":  time.Now(),
		}).Error
}

// ---- House Favor ----

func (h *vampireHandler) AddHouseFavor(ctx context.Context, entry *models.VampireHouseFavorLedger) error {
	return h.db.WithContext(ctx).Create(entry).Error
}

func (h *vampireHandler) Leaderboard(ctx context.Context) ([]HouseFavorStanding, error) {
	var standings []HouseFavorStanding
	if err := h.db.WithContext(ctx).
		Table("vampire_houses h").
		Select("h.id AS house_id, h.name AS name, h.sort_order AS sort_order, COALESCE(SUM(l.delta), 0) AS favor").
		Joins("LEFT JOIN vampire_house_favor_ledger l ON l.house_id = h.id").
		Group("h.id, h.name, h.sort_order").
		Order("favor DESC, h.sort_order ASC").
		Scan(&standings).Error; err != nil {
		return nil, err
	}
	return standings, nil
}

// ---- Blood Tokens ----

func (h *vampireHandler) AddBloodTokens(ctx context.Context, entry *models.VampireBloodTokenLog) error {
	return h.db.WithContext(ctx).Create(entry).Error
}

func (h *vampireHandler) BloodTokenTotalsByPlayer(ctx context.Context) ([]BloodTokenTotal, error) {
	var totals []BloodTokenTotal
	if err := h.db.WithContext(ctx).
		Table("vampire_blood_token_log").
		Select("player_id, COALESCE(SUM(delta), 0) AS total").
		Group("player_id").
		Scan(&totals).Error; err != nil {
		return nil, err
	}
	return totals, nil
}

// BloodTokenTotalsBySource sums each player's BT for a single source (e.g. "game"),
// used by the tally engine to double game winnings.
func (h *vampireHandler) BloodTokenTotalsBySource(ctx context.Context, source string) ([]BloodTokenTotal, error) {
	var totals []BloodTokenTotal
	if err := h.db.WithContext(ctx).
		Table("vampire_blood_token_log").
		Select("player_id, COALESCE(SUM(delta), 0) AS total").
		Where("source = ?", source).
		Group("player_id").
		Scan(&totals).Error; err != nil {
		return nil, err
	}
	return totals, nil
}

// ---- Game state ----

func (h *vampireHandler) GetGameState(ctx context.Context) (*models.VampireGameState, error) {
	var state models.VampireGameState
	if err := h.db.WithContext(ctx).First(&state, "id = ?", 1).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Lazily create the singleton if the seed insert was missed.
			state = models.VampireGameState{ID: 1, CurrentAct: "pre_event"}
			if cerr := h.db.WithContext(ctx).Create(&state).Error; cerr != nil {
				return nil, cerr
			}
			return &state, nil
		}
		return nil, err
	}
	return &state, nil
}

func (h *vampireHandler) UpdateGameState(ctx context.Context, updates map[string]interface{}) (*models.VampireGameState, error) {
	updates["updated_at"] = time.Now()
	if err := h.db.WithContext(ctx).Model(&models.VampireGameState{}).Where("id = ?", 1).Updates(updates).Error; err != nil {
		return nil, err
	}
	return h.GetGameState(ctx)
}

// ---- Notifications ----

func (h *vampireHandler) CreateNotification(ctx context.Context, n *models.VampireNotification) error {
	return h.db.WithContext(ctx).Create(n).Error
}

func (h *vampireHandler) DeactivateAllNotifications(ctx context.Context) error {
	return h.db.WithContext(ctx).Model(&models.VampireNotification{}).
		Where("active = ?", true).
		Update("active", false).Error
}

// GetActiveNotificationForPlayer returns the most recent active notification
// that applies to this player: broadcast to all, to their house, or to them.
func (h *vampireHandler) GetActiveNotificationForPlayer(ctx context.Context, playerID uuid.UUID, houseID *uuid.UUID) (*models.VampireNotification, error) {
	q := h.db.WithContext(ctx).Where("active = ?", true)
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

func (h *vampireHandler) ListActiveNotifications(ctx context.Context) ([]models.VampireNotification, error) {
	var notifs []models.VampireNotification
	if err := h.db.WithContext(ctx).
		Where("active = ?", true).
		Order("created_at DESC").
		Find(&notifs).Error; err != nil {
		return nil, err
	}
	return notifs, nil
}

// ---- Quiz ----

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
	GuestLabel    string    `json:"guestLabel"`
	CharacterName string    `json:"characterName"`
	HouseName     string    `json:"houseName"`
	Ordinal       int       `json:"ordinal"`
	Prompt        string    `json:"prompt"`
	QuestionType  string    `json:"questionType"`
}

func (h *vampireHandler) ListQuizSubmissionsDetailed(ctx context.Context) ([]QuizSubmissionDetail, error) {
	out := []QuizSubmissionDetail{}
	if err := h.db.WithContext(ctx).
		Table("vampire_quiz_submissions s").
		Select(`s.id, s.player_id, s.question_id, s.answer, s.is_correct, s.ai_score,
			s.ai_rationale, s.awarded_bt, s.locked,
			q.part AS part,
			p.guest_label AS guest_label,
			COALESCE(c.name, '') AS character_name,
			COALESCE(h.name, '') AS house_name,
			q.ordinal AS ordinal, q.prompt AS prompt, q.question_type AS question_type`).
		Joins("JOIN vampire_players p ON p.id = s.player_id").
		Joins("JOIN vampire_quiz_questions q ON q.id = s.question_id").
		Joins("LEFT JOIN vampire_characters c ON c.id = p.character_id").
		Joins("LEFT JOIN vampire_houses h ON h.id = c.house_id").
		Order("q.part ASC, q.ordinal ASC, character_name ASC").
		Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
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

// GetPart1Question returns the single active Part 1 (open-end) question.
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

// Part2Answer is one player's answer to a Part 2 question, with their house —
// the raw material for the normalized per-house scoring.
type Part2Answer struct {
	PlayerID   uuid.UUID `json:"playerId"`
	HouseID    uuid.UUID `json:"houseId"`
	QuestionID uuid.UUID `json:"questionId"`
	Answer     string    `json:"answer"`
}

func (h *vampireHandler) ListPart2Answers(ctx context.Context) ([]Part2Answer, error) {
	out := []Part2Answer{}
	if err := h.db.WithContext(ctx).
		Table("vampire_quiz_submissions s").
		Select("s.player_id, c.house_id AS house_id, s.question_id, s.answer").
		Joins("JOIN vampire_quiz_questions q ON q.id = s.question_id AND q.part = 2").
		Joins("JOIN vampire_players p ON p.id = s.player_id").
		Joins("JOIN vampire_characters c ON c.id = p.character_id").
		Where("c.house_id IS NOT NULL").
		Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteHouseFavorBySource removes ledger entries of a given source (used to
// idempotently re-score Part 2).
func (h *vampireHandler) DeleteHouseFavorBySource(ctx context.Context, source string) error {
	return h.db.WithContext(ctx).Where("source = ?", source).Delete(&models.VampireHouseFavorLedger{}).Error
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

func (h *vampireHandler) UpsertQuizSubmission(ctx context.Context, playerID, questionID uuid.UUID, answer string, isCorrect *bool, locked bool) (*models.VampireQuizSubmission, error) {
	sub := models.VampireQuizSubmission{
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

func (h *vampireHandler) ListQuizSubmissions(ctx context.Context) ([]models.VampireQuizSubmission, error) {
	var subs []models.VampireQuizSubmission
	if err := h.db.WithContext(ctx).Order("created_at ASC").Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}

// ResetGameProgress wipes all play progress for a clean playtest run: mission
// submissions, both ledgers, quiz submissions, notifications, and the audit log,
// then resets the game state to a sealed pre-event. The roster (houses,
// characters, secrets, missions, players + their token assignments) and quiz
// questions are preserved.
func (h *vampireHandler) ResetGameProgress(ctx context.Context) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Snapshot the score ledgers before wiping them, so a reset is always
		// recoverable. Same transaction as the delete, so archive and wipe are
		// atomic — we never delete without a durable copy landing first.
		archives := map[string]string{
			"vampire_house_favor_ledger": "vampire_house_favor_ledger_archive",
			"vampire_blood_token_log":    "vampire_blood_token_log_archive",
		}
		for src, dst := range archives {
			if err := tx.Exec("INSERT INTO " + dst + " SELECT *, now() FROM " + src).Error; err != nil {
				return err
			}
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
			if err := tx.Exec("DELETE FROM " + table).Error; err != nil {
				return err
			}
		}
		return tx.Model(&models.VampireGameState{}).Where("id = ?", 1).Updates(map[string]interface{}{
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

// WipeCharactersAndRoster clears the roster and all character content so a seed
// run can rebuild from scratch (used for the --fresh re-seed). Deleting players
// cascades their submissions, blood-token log, and quiz answers; deleting
// characters cascades their secrets and missions. Score ledgers are archived
// first so the wipe is recoverable, and houses / game state are left intact.
func (h *vampireHandler) WipeCharactersAndRoster(ctx context.Context) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		archives := map[string]string{
			"vampire_house_favor_ledger": "vampire_house_favor_ledger_archive",
			"vampire_blood_token_log":    "vampire_blood_token_log_archive",
		}
		for src, dst := range archives {
			if err := tx.Exec("INSERT INTO " + dst + " SELECT *, now() FROM " + src).Error; err != nil {
				return err
			}
		}
		// Notifications first (may reference players), then players (cascades their
		// play data), then characters (cascades secrets + missions).
		stmts := []string{
			"DELETE FROM vampire_notifications",
			"DELETE FROM vampire_players",
			"DELETE FROM vampire_characters",
		}
		for _, s := range stmts {
			if err := tx.Exec(s).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ---- Physical games ----

func (h *vampireHandler) ListGames(ctx context.Context) ([]models.VampireGame, error) {
	var games []models.VampireGame
	if err := h.db.WithContext(ctx).Order("ordinal ASC, created_at ASC").Find(&games).Error; err != nil {
		return nil, err
	}
	return games, nil
}

func (h *vampireHandler) GetGameByID(ctx context.Context, id uuid.UUID) (*models.VampireGame, error) {
	var g models.VampireGame
	if err := h.db.WithContext(ctx).First(&g, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

// UpsertGame creates or updates a game by name (idempotent seeding).
func (h *vampireHandler) UpsertGame(ctx context.Context, ordinal int, name string) (*models.VampireGame, error) {
	game := models.VampireGame{Name: name, Ordinal: ordinal}
	if err := h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"ordinal": ordinal, "updated_at": time.Now()}),
		}).
		Create(&game).Error; err != nil {
		return nil, err
	}
	var out models.VampireGame
	if err := h.db.WithContext(ctx).First(&out, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (h *vampireHandler) SetGameResult(ctx context.Context, id uuid.UUID, first, second, third *uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.VampireGame{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":              "played",
			"first_character_id":  first,
			"second_character_id": second,
			"third_character_id":  third,
			"updated_at":          time.Now(),
		}).Error
}

func (h *vampireHandler) UpdateGame(ctx context.Context, id uuid.UUID, name string, ordinal int) error {
	return h.db.WithContext(ctx).Model(&models.VampireGame{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"name": name, "ordinal": ordinal, "updated_at": time.Now()}).Error
}

func (h *vampireHandler) DeleteGame(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.VampireGame{}, "id = ?", id).Error
}

// ClearGameResult resets a game to pending and drops its recorded finishers.
func (h *vampireHandler) ClearGameResult(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.VampireGame{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":              "pending",
			"first_character_id":  nil,
			"second_character_id": nil,
			"third_character_id":  nil,
			"updated_at":          time.Now(),
		}).Error
}

func (h *vampireHandler) DeleteGameAwards(ctx context.Context, gameName string) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"DELETE FROM vampire_house_favor_ledger WHERE source = 'game' AND reason = ?",
			"Game: "+gameName,
		).Error; err != nil {
			return err
		}
		return tx.Exec(
			"DELETE FROM vampire_blood_token_log WHERE source = 'game' AND reason IN (?, ?)",
			"Game: "+gameName, "Game participation: "+gameName,
		).Error
	})
}

// ---- Inventory ----

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
				"code", "description", "effect", "targets_player", "hf_effect",
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

func (h *vampireHandler) ListAllPlayerItems(ctx context.Context) ([]models.VampirePlayerItem, error) {
	var pis []models.VampirePlayerItem
	if err := h.db.WithContext(ctx).Preload("Item").Find(&pis).Error; err != nil {
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

func (h *vampireHandler) SetPlayerItemTarget(ctx context.Context, id uuid.UUID, targetPlayerID *uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.VampirePlayerItem{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"target_player_id": targetPlayerID, "updated_at": time.Now()}).Error
}

// ---- GM audit log ----

func (h *vampireHandler) LogGMAction(ctx context.Context, gmName, action string, payload []byte) error {
	entry := models.VampireGMActionLog{
		GMName:  gmName,
		Action:  action,
		Payload: payload,
	}
	return h.db.WithContext(ctx).Create(&entry).Error
}
