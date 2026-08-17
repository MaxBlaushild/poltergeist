package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ConflictError is returned by a business-rule guard that blocked an action
// (e.g. removing the Host, un-including a character that's still actively
// assigned). Server handlers type-assert this to return 409 with the message
// as-is.
type ConflictError struct{ Message string }

func (e *ConflictError) Error() string { return e.Message }

// ---- Instances ("Toasts" in user-facing copy) ----

// CreateInstance also records the instance's selected subplots (zero or
// many, in addition to its one required mysteryID) in the same
// transaction — see vampire_instance_subplots.
func (h *vampireHandler) CreateInstance(ctx context.Context, name string, createdBy *uuid.UUID, mysteryID uuid.UUID, subplotIDs []uuid.UUID) (*models.VampireInstance, error) {
	var inst models.VampireInstance
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		inst = models.VampireInstance{Name: name, Status: "active", CreatedBy: createdBy, MysteryID: mysteryID}
		if err := tx.Create(&inst).Error; err != nil {
			return err
		}
		if len(subplotIDs) == 0 {
			return nil
		}
		rows := make([]models.VampireInstanceSubplot, 0, len(subplotIDs))
		for _, sid := range subplotIDs {
			rows = append(rows, models.VampireInstanceSubplot{InstanceID: inst.ID, MysteryID: sid})
		}
		return tx.Create(&rows).Error
	})
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

func (h *vampireHandler) GetInstanceByID(ctx context.Context, id uuid.UUID) (*models.VampireInstance, error) {
	var inst models.VampireInstance
	if err := h.db.WithContext(ctx).First(&inst, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &inst, nil
}

// ListSubplotIDsForInstance returns the subplots selected for one instance
// — used by withPlayer/withInstanceAdmin to cache alongside the instance's
// main mystery ID (see subplotIDsContextKey), and by createInstance's own
// validation.
func (h *vampireHandler) ListSubplotIDsForInstance(ctx context.Context, instanceID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	if err := h.db.WithContext(ctx).Model(&models.VampireInstanceSubplot{}).
		Where("instance_id = ?", instanceID).
		Pluck("mystery_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// ListInstancesForUser returns every instance the user Hosts or Co-Hosts
// ("My Toasts"), newest first.
func (h *vampireHandler) ListInstancesForUser(ctx context.Context, userID uuid.UUID) ([]models.VampireInstance, error) {
	var instances []models.VampireInstance
	if err := h.db.WithContext(ctx).
		Table("vampire_instances i").
		Select("i.*").
		Joins("JOIN vampire_instance_admins a ON a.instance_id = i.id AND a.user_id = ?", userID).
		Order("i.created_at DESC").
		Find(&instances).Error; err != nil {
		return nil, err
	}
	return instances, nil
}

// SeedInstanceLibrary includes everything currently in the shared content
// library (characters, items, quiz questions) for a newly created instance.
// Idempotent — safe to call again later (e.g. to backfill library content
// added after the instance was created).
// SeedInstanceLibrary includes every shared item for a newly created
// instance. Characters no longer get a join row here — which characters
// are "in" is implicit via invites — and quiz questions no longer either:
// an instance's quiz is simply every vampire_quiz_questions row whose
// mystery_id matches the instance's own (see MYSTERY_REQUIREMENTS.md), not
// a per-instance toggle.
func (h *vampireHandler) SeedInstanceLibrary(ctx context.Context, instanceID uuid.UUID) error {
	return h.db.WithContext(ctx).Exec(`
		INSERT INTO vampire_instance_items (instance_id, item_id, included)
		SELECT ?, i.id, TRUE
		FROM vampire_items i
		ON CONFLICT (instance_id, item_id) DO NOTHING`, instanceID).Error
}

// ---- Administrators ("Host" for role=owner, "Co-Host" for role=admin) ----

func (h *vampireHandler) AddInstanceAdmin(ctx context.Context, instanceID, userID uuid.UUID, role string) (*models.VampireInstanceAdmin, error) {
	admin := models.VampireInstanceAdmin{InstanceID: instanceID, UserID: userID, Role: role}
	if err := h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "instance_id"}, {Name: "user_id"}},
			DoNothing: true,
		}).
		Create(&admin).Error; err != nil {
		return nil, err
	}
	var out models.VampireInstanceAdmin
	if err := h.db.WithContext(ctx).First(&out, "instance_id = ? AND user_id = ?", instanceID, userID).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

// ListInstanceAdmins returns an instance's administrators, Host first.
func (h *vampireHandler) ListInstanceAdmins(ctx context.Context, instanceID uuid.UUID) ([]models.VampireInstanceAdmin, error) {
	var admins []models.VampireInstanceAdmin
	if err := h.db.WithContext(ctx).
		Preload("User").
		Where("instance_id = ?", instanceID).
		Order("(role = 'owner') DESC, created_at ASC").
		Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}

func (h *vampireHandler) GetInstanceAdmin(ctx context.Context, instanceID, userID uuid.UUID) (*models.VampireInstanceAdmin, error) {
	var admin models.VampireInstanceAdmin
	if err := h.db.WithContext(ctx).First(&admin, "instance_id = ? AND user_id = ?", instanceID, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

// RemoveInstanceAdmin removes a Co-Host. The Host (role=owner) cannot be
// removed this way — only via TransferInstanceOwnership — so an instance
// can never end up with zero administrators. Returns a *ConflictError if the
// target is the Host.
func (h *vampireHandler) RemoveInstanceAdmin(ctx context.Context, instanceID, userID uuid.UUID) error {
	admin, err := h.GetInstanceAdmin(ctx, instanceID, userID)
	if err != nil {
		return err
	}
	if admin == nil {
		return nil
	}
	if admin.Role == models.InstanceAdminRoleOwner {
		return &ConflictError{Message: "can't remove the Host directly — make someone else the Host first"}
	}
	return h.db.WithContext(ctx).
		Where("instance_id = ? AND user_id = ?", instanceID, userID).
		Delete(&models.VampireInstanceAdmin{}).Error
}

// TransferInstanceOwnership makes toUserID the new Host and demotes
// fromUserID to Co-Host. toUserID must already be a Co-Host on the instance.
func (h *vampireHandler) TransferInstanceOwnership(ctx context.Context, instanceID, fromUserID, toUserID uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current models.VampireInstanceAdmin
		if err := tx.First(&current, "instance_id = ? AND user_id = ?", instanceID, fromUserID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return &ConflictError{Message: "only the current Host can hand off hosting"}
			}
			return err
		}
		if current.Role != models.InstanceAdminRoleOwner {
			return &ConflictError{Message: "only the current Host can hand off hosting"}
		}
		var target models.VampireInstanceAdmin
		if err := tx.First(&target, "instance_id = ? AND user_id = ?", instanceID, toUserID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return &ConflictError{Message: "the new Host must already be a Co-Host on this Toast"}
			}
			return err
		}
		// Demote-then-promote, in that order, so the one-owner-per-instance
		// partial unique index is never briefly violated by two owner rows.
		if err := tx.Model(&models.VampireInstanceAdmin{}).
			Where("instance_id = ? AND user_id = ?", instanceID, fromUserID).
			Update("role", models.InstanceAdminRoleAdmin).Error; err != nil {
			return err
		}
		return tx.Model(&models.VampireInstanceAdmin{}).
			Where("instance_id = ? AND user_id = ?", instanceID, toUserID).
			Update("role", models.InstanceAdminRoleOwner).Error
	})
}

// ---- Admin invites ("Invite a Co-Host") ----

func (h *vampireHandler) CreateInstanceAdminInvite(ctx context.Context, instanceID uuid.UUID, email string, invitedBy uuid.UUID, token string, expiresAt time.Time) (*models.VampireInstanceAdminInvite, error) {
	inv := models.VampireInstanceAdminInvite{
		InstanceID: instanceID,
		Email:      email,
		InvitedBy:  invitedBy,
		Token:      token,
		ExpiresAt:  expiresAt,
	}
	if err := h.db.WithContext(ctx).Create(&inv).Error; err != nil {
		return nil, err
	}
	return &inv, nil
}

func (h *vampireHandler) GetInstanceAdminInviteByToken(ctx context.Context, token string) (*models.VampireInstanceAdminInvite, error) {
	var inv models.VampireInstanceAdminInvite
	if err := h.db.WithContext(ctx).First(&inv, "token = ?", token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &inv, nil
}

// ListPendingInstanceAdminInvites returns unexpired, unaccepted invites for
// an instance's Co-Hosts tab.
func (h *vampireHandler) ListPendingInstanceAdminInvites(ctx context.Context, instanceID uuid.UUID) ([]models.VampireInstanceAdminInvite, error) {
	var invites []models.VampireInstanceAdminInvite
	if err := h.db.WithContext(ctx).
		Where("instance_id = ? AND accepted_at IS NULL AND expires_at > ?", instanceID, time.Now()).
		Order("created_at DESC").
		Find(&invites).Error; err != nil {
		return nil, err
	}
	return invites, nil
}

// DeleteInstanceAdminInvite revokes a pending invite.
func (h *vampireHandler) DeleteInstanceAdminInvite(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.VampireInstanceAdminInvite{}, "id = ?", id).Error
}

// AcceptInstanceAdminInvite marks the invite accepted and grants the
// accepting user Co-Host on its instance. Returns the instance id.
func (h *vampireHandler) AcceptInstanceAdminInvite(ctx context.Context, token string, userID uuid.UUID) (uuid.UUID, error) {
	var instanceID uuid.UUID
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inv models.VampireInstanceAdminInvite
		if err := tx.First(&inv, "token = ?", token).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return &ConflictError{Message: "invite not found"}
			}
			return err
		}
		if inv.AcceptedAt != nil {
			return &ConflictError{Message: "this invite has already been used"}
		}
		if time.Now().After(inv.ExpiresAt) {
			return &ConflictError{Message: "this invite has expired"}
		}
		now := time.Now()
		if err := tx.Model(&inv).Update("accepted_at", now).Error; err != nil {
			return err
		}
		admin := models.VampireInstanceAdmin{InstanceID: inv.InstanceID, UserID: userID, Role: models.InstanceAdminRoleAdmin}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "instance_id"}, {Name: "user_id"}},
			DoNothing: true,
		}).Create(&admin).Error; err != nil {
			return err
		}
		instanceID = inv.InstanceID
		return nil
	})
	return instanceID, err
}

// ---- Content library: characters ----

// GetInstanceCharacter returns the raw per-instance row (currently just the
// portrait) for one character.
func (h *vampireHandler) GetInstanceCharacter(ctx context.Context, instanceID, characterID uuid.UUID) (*models.VampireInstanceCharacter, error) {
	var ic models.VampireInstanceCharacter
	if err := h.db.WithContext(ctx).First(&ic, "instance_id = ? AND character_id = ?", instanceID, characterID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &ic, nil
}

// SetInstanceCharacterImageURL sets a character's per-instance portrait —
// the one character-level edit a Host/Co-Host can make directly (bios,
// secrets, missions are shared content; see the /admin editor). Upserts, so
// it also works for a character with no join row yet.
func (h *vampireHandler) SetInstanceCharacterImageURL(ctx context.Context, instanceID, characterID uuid.UUID, url string) error {
	return h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "instance_id"}, {Name: "character_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"image_url": url, "updated_at": time.Now()}),
		}).
		Create(&models.VampireInstanceCharacter{InstanceID: instanceID, CharacterID: characterID, ImageURL: url}).Error
}

// ---- Content library: items ----

type LibraryItem struct {
	ID              uuid.UUID `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Category        string    `json:"category"`
	Description     string    `json:"description"`
	Effect          string    `json:"effect"`
	TargetsPlayer   bool      `json:"targetsPlayer"`
	HFEffect        int       `json:"hfEffect"`
	BTSelf          int       `json:"btSelf"`
	BTFromTarget    int       `json:"btFromTarget"`
	BTDeductTarget  int       `json:"btDeductTarget"`
	QuizBTPct       int       `json:"quizBtPct"`
	DoubleGameBT    bool      `json:"doubleGameBt"`
	Immune          bool      `json:"immune"`
	Reflect         bool      `json:"reflect"`
	StripResistance bool      `json:"stripResistance"`
	Included        bool      `json:"included"`
}

func (h *vampireHandler) ListLibraryItems(ctx context.Context, instanceID uuid.UUID) ([]LibraryItem, error) {
	out := []LibraryItem{}
	if err := h.db.WithContext(ctx).
		Table("vampire_items i").
		Select(`i.id, i.code, i.name, i.category, i.description, i.effect, i.targets_player,
			i.hf_effect, i.bt_self, i.bt_from_target, i.bt_deduct_target, i.quiz_bt_pct,
			i.double_game_bt, i.immune, i.reflect, i.strip_resistance,
			COALESCE(ii.included, false) AS included`).
		Joins("LEFT JOIN vampire_instance_items ii ON ii.item_id = i.id AND ii.instance_id = ?", instanceID).
		Order("i.name ASC").
		Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// ListIncludedItems returns the catalog items included in this instance.
func (h *vampireHandler) ListIncludedItems(ctx context.Context, instanceID uuid.UUID) ([]models.VampireItem, error) {
	var items []models.VampireItem
	if err := h.db.WithContext(ctx).
		Joins("JOIN vampire_instance_items ii ON ii.item_id = vampire_items.id AND ii.instance_id = ? AND ii.included", instanceID).
		Order("vampire_items.name ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// SetItemIncluded toggles an item in or out of an instance's roster.
// Un-including an item currently assigned to a player is blocked
// (*ConflictError).
func (h *vampireHandler) SetItemIncluded(ctx context.Context, instanceID, itemID uuid.UUID, included bool) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if !included {
			var count int64
			if err := tx.Table("vampire_player_items pi").
				Joins("JOIN vampire_players p ON p.id = pi.player_id").
				Where("p.instance_id = ? AND pi.item_id = ?", instanceID, itemID).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return &ConflictError{Message: "can't remove this item — it's currently assigned to a player"}
			}
		}
		res := tx.Model(&models.VampireInstanceItem{}).
			Where("instance_id = ? AND item_id = ?", instanceID, itemID).
			Updates(map[string]interface{}{"included": included, "updated_at": time.Now()})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return tx.Create(&models.VampireInstanceItem{InstanceID: instanceID, ItemID: itemID, Included: included}).Error
		}
		return nil
	})
}

// ---- Quiz questions, scoped to a mystery (see MYSTERY_REQUIREMENTS.md) ----
// Replaces the old per-instance include/exclude toggle entirely: an
// instance's quiz is simply every question whose MysteryID matches the
// instance's own.

// GetPart1QuestionForMystery returns a mystery's single active Part 1
// (open-end) question.
func (h *vampireHandler) GetPart1QuestionForMystery(ctx context.Context, mysteryID uuid.UUID) (*models.VampireQuizQuestion, error) {
	var qq models.VampireQuizQuestion
	if err := h.db.WithContext(ctx).
		Where("mystery_id = ? AND part = ? AND active = ?", mysteryID, 1, true).
		Order("ordinal ASC").First(&qq).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &qq, nil
}

// ListQuizQuestionsByMysteryAndPart returns a mystery's questions for one part.
func (h *vampireHandler) ListQuizQuestionsByMysteryAndPart(ctx context.Context, mysteryID uuid.UUID, part int, activeOnly bool) ([]models.VampireQuizQuestion, error) {
	var qs []models.VampireQuizQuestion
	q := h.db.WithContext(ctx).
		Where("mystery_id = ? AND part = ?", mysteryID, part).
		Order("ordinal ASC")
	if activeOnly {
		q = q.Where("active = ?", true)
	}
	if err := q.Find(&qs).Error; err != nil {
		return nil, err
	}
	return qs, nil
}

// ReplaceQuizQuestionsForMystery removes a mystery's existing quiz questions
// and inserts the new set — the mystery-scoped equivalent of the old
// wholesale ReplaceQuizQuestions.
func (h *vampireHandler) ReplaceQuizQuestionsForMystery(ctx context.Context, mysteryID uuid.UUID, questions []models.VampireQuizQuestion) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("mystery_id = ?", mysteryID).Delete(&models.VampireQuizQuestion{}).Error; err != nil {
			return err
		}
		for i := range questions {
			questions[i].MysteryID = mysteryID
		}
		if len(questions) == 0 {
			return nil
		}
		return tx.Create(&questions).Error
	})
}
