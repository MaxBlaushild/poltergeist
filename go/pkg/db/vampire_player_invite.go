package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// CreatePlayerInvite invites a specific real person (by phone number) to
// play a specific character in one instance. Returns a *ConflictError if
// the character already has an accepted player, or already has another
// pending invite (the partial unique index on (instance_id, character_id)
// WHERE status = 'pending' enforces the latter at the DB level; this method
// translates that into the same typed error the rest of the app uses).
func (h *vampireHandler) CreatePlayerInvite(ctx context.Context, instanceID, characterID uuid.UUID, guestName, phoneNumber string, invitedBy uuid.UUID, token string) (*models.VampirePlayerInvite, error) {
	var out *models.VampirePlayerInvite
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.VampirePlayer{}).
			Where("instance_id = ? AND character_id = ? AND active = ?", instanceID, characterID, true).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return &ConflictError{Message: "this character is already played by someone in this Toast"}
		}

		inv := &models.VampirePlayerInvite{
			InstanceID:  instanceID,
			CharacterID: characterID,
			GuestName:   guestName,
			PhoneNumber: phoneNumber,
			Token:       token,
			Status:      models.PlayerInviteStatusPending,
			InvitedBy:   &invitedBy,
		}
		if err := tx.Create(inv).Error; err != nil {
			return err
		}
		out = inv
		return nil
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, &ConflictError{Message: "this character already has a pending invite — revoke it first"}
		}
		return nil, err
	}
	return out, nil
}

// ListPlayerInvites returns every invite (any status) for an instance,
// newest first, with the invited character preloaded — for the Invites tab.
func (h *vampireHandler) ListPlayerInvites(ctx context.Context, instanceID uuid.UUID) ([]models.VampirePlayerInvite, error) {
	var invites []models.VampirePlayerInvite
	if err := h.db.WithContext(ctx).
		Preload("Character").
		Preload("Character.House").
		Where("instance_id = ?", instanceID).
		Order("created_at DESC").
		Find(&invites).Error; err != nil {
		return nil, err
	}
	return invites, nil
}

// GetPlayerInviteByToken looks up an invite for the public RSVP page — no
// instance scoping needed, the token alone identifies it.
func (h *vampireHandler) GetPlayerInviteByToken(ctx context.Context, token string) (*models.VampirePlayerInvite, error) {
	var inv models.VampirePlayerInvite
	if err := h.db.WithContext(ctx).
		Preload("Character").
		Preload("Character.House").
		First(&inv, "token = ?", token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &inv, nil
}

// DeletePlayerInvite revokes a pending invite (or clears a stale
// accepted/declined one from the list).
func (h *vampireHandler) DeletePlayerInvite(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.VampirePlayerInvite{}, "id = ?", id).Error
}

// DeclinePlayerInvite marks a pending invite declined. No account required
// — anyone holding the link (the token itself is the credential) can
// decline without signing up first. Returns a *ConflictError if the invite
// isn't pending (already responded to, or unknown).
func (h *vampireHandler) DeclinePlayerInvite(ctx context.Context, token string) error {
	now := time.Now()
	res := h.db.WithContext(ctx).Model(&models.VampirePlayerInvite{}).
		Where("token = ? AND status = ?", token, models.PlayerInviteStatusPending).
		Updates(map[string]interface{}{
			"status":       models.PlayerInviteStatusDeclined,
			"responded_at": now,
			"updated_at":   now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return &ConflictError{Message: "this invite is no longer pending"}
	}
	return nil
}

// AcceptPlayerInvite marks the invite accepted and creates the player row
// for the accepting (real, signed-in) account. Returns a *ConflictError if
// the invite isn't pending, or the user already holds a different
// character in this instance.
func (h *vampireHandler) AcceptPlayerInvite(ctx context.Context, token string, userID uuid.UUID) (*models.VampirePlayer, error) {
	var out *models.VampirePlayer
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inv models.VampirePlayerInvite
		if err := tx.First(&inv, "token = ?", token).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return &ConflictError{Message: "invite not found"}
			}
			return err
		}
		if inv.Status != models.PlayerInviteStatusPending {
			return &ConflictError{Message: "this invite is no longer pending"}
		}

		var existing int64
		if err := tx.Model(&models.VampirePlayer{}).
			Where("instance_id = ? AND user_id = ?", inv.InstanceID, userID).
			Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return &ConflictError{Message: "you already have a character in this Toast"}
		}

		now := time.Now()
		if err := tx.Model(&inv).Updates(map[string]interface{}{
			"status":           models.PlayerInviteStatusAccepted,
			"accepted_user_id": userID,
			"responded_at":     now,
			"updated_at":       now,
		}).Error; err != nil {
			return err
		}

		cid := inv.CharacterID
		player := &models.VampirePlayer{
			InstanceID:  inv.InstanceID,
			UserID:      &userID,
			CharacterID: &cid,
			GuestLabel:  inv.GuestName,
			Active:      true,
		}
		if err := tx.Create(player).Error; err != nil {
			return err
		}
		out = player
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetPlayerByUserAndInstance resolves a signed-in account's character
// assignment in one instance — the player-auth equivalent of
// GetInstanceAdmin, used by withPlayer.
func (h *vampireHandler) GetPlayerByUserAndInstance(ctx context.Context, instanceID, userID uuid.UUID) (*models.VampirePlayer, error) {
	var p models.VampirePlayer
	if err := h.db.WithContext(ctx).
		Preload("Character").
		First(&p, "instance_id = ? AND user_id = ?", instanceID, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// ListPlayerInstancesForUser is the player-side counterpart to
// ListInstancesForUser — every Toast a signed-in account holds an active
// character in (as opposed to administers). Used by the "My Toasts"
// dashboard to fold a player's Toasts in alongside the ones they Host/
// Co-Host, with the character preloaded for the dashboard's teaser card.
func (h *vampireHandler) ListPlayerInstancesForUser(ctx context.Context, userID uuid.UUID) ([]models.VampirePlayer, error) {
	var players []models.VampirePlayer
	if err := h.db.WithContext(ctx).
		Preload("Character.House").
		Where("user_id = ? AND active = ?", userID, true).
		Find(&players).Error; err != nil {
		return nil, err
	}
	return players, nil
}

// isUniqueViolation reports whether err is a Postgres unique_violation
// (23505) — used to translate the one-pending-invite-per-character partial
// unique index into a friendly *ConflictError.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
