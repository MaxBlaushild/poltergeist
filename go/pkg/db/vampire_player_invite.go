package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreatePlayerInvite invites a specific real person (by phone number) to
// join this instance as a player. Character-agnostic — which character
// they'll play is chosen by them, after accepting, from the instance's
// curated pool (see VampireInstanceCharacterPool and ClaimCharacterForPlayer).
func (h *vampireHandler) CreatePlayerInvite(ctx context.Context, instanceID uuid.UUID, guestName, phoneNumber string, invitedBy uuid.UUID, token string) (*models.VampirePlayerInvite, error) {
	inv := &models.VampirePlayerInvite{
		InstanceID:  instanceID,
		GuestName:   guestName,
		PhoneNumber: phoneNumber,
		Token:       token,
		Status:      models.PlayerInviteStatusPending,
		InvitedBy:   &invitedBy,
	}
	if err := h.db.WithContext(ctx).Create(inv).Error; err != nil {
		return nil, err
	}
	return inv, nil
}

// ListPlayerInvites returns every invite (any status) for an instance,
// newest first — for the Invites tab.
func (h *vampireHandler) ListPlayerInvites(ctx context.Context, instanceID uuid.UUID) ([]models.VampirePlayerInvite, error) {
	var invites []models.VampirePlayerInvite
	if err := h.db.WithContext(ctx).
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
	if err := h.db.WithContext(ctx).First(&inv, "token = ?", token).Error; err != nil {
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
// for the accepting (real, signed-in) account — with no character yet; see
// ClaimCharacterForPlayer for the self-select step that follows. Returns a
// *ConflictError if the invite isn't pending, or the user already holds a
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

		player := &models.VampirePlayer{
			InstanceID: inv.InstanceID,
			UserID:     &userID,
			GuestLabel: inv.GuestName,
			Active:     true,
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
