package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type partyInviteHandle struct {
	db *gorm.DB
}

func (h *partyInviteHandle) Create(ctx context.Context, inviter *models.User, inviteeID uuid.UUID) (*models.PartyInvite, error) {
	partyMembers := []models.User{}
	if inviter.PartyID != nil {
		if err := h.db.WithContext(ctx).Where("party_id = ?", inviter.PartyID).Find(&partyMembers).Error; err != nil {
			return nil, err
		}
	}

	if len(partyMembers) >= MaxPartySize {
		return nil, errors.New("party is full")
	}

	if inviter.PartyID == nil {
		party := &models.Party{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			LeaderID:  inviter.ID,
		}
		if err := h.db.WithContext(ctx).Create(&party).Error; err != nil {
			return nil, err
		}
	}

	partyInvite := &models.PartyInvite{
		InviterID: inviter.ID,
		InviteeID: inviteeID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ID:        uuid.New(),
	}

	if err := h.db.WithContext(ctx).Create(partyInvite).Error; err != nil {
		return nil, err
	}

	return partyInvite, nil
}

func (h *partyInviteHandle) FindAllInvites(ctx context.Context, userID uuid.UUID) ([]models.PartyInvite, error) {
	var partyInvites []models.PartyInvite
	if err := h.db.WithContext(ctx).Preload("Invitee").Preload("Inviter").Where("invitee_id = ? OR inviter_id = ?", userID, userID).Find(&partyInvites).Error; err != nil {
		return nil, err
	}

	return partyInvites, nil
}

func (h *partyInviteHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.PartyInvite, error) {
	var partyInvite models.PartyInvite
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&partyInvite).Error; err != nil {
		return nil, err
	}

	return &partyInvite, nil
}

func (h *partyInviteHandle) Accept(ctx context.Context, id uuid.UUID, user *models.User) (*models.PartyInvite, error) {
	var partyInvite models.PartyInvite
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&partyInvite).Error; err != nil {
		return nil, err
	}

	if partyInvite.InviteeID != user.ID {
		return nil, errors.New("not invitee")
	}

	if user.PartyID != nil {
		return nil, errors.New("already in a party")
	}

	var inviter *models.User
	if err := h.db.WithContext(ctx).Where("id = ?", partyInvite.InviterID).First(&inviter).Error; err != nil {
		return nil, err
	}

	partyID := inviter.PartyID

	if partyID == nil {
		party := &models.Party{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			LeaderID:  inviter.ID,
		}
		if err := h.db.WithContext(ctx).Create(&party).Error; err != nil {
			return nil, err
		}
		partyID = &party.ID

		if err := h.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", inviter.ID).Update("party_id", partyID).Error; err != nil {
			return nil, err
		}
	}

	partyMembers := []models.User{}
	if inviter.PartyID != nil {
		if err := h.db.WithContext(ctx).Where("party_id = ?", inviter.PartyID).Find(&partyMembers).Error; err != nil {
			return nil, err
		}
	}

	if len(partyMembers) >= MaxPartySize {
		return nil, errors.New("party is full")
	}

	if err := h.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", user.ID).Update("party_id", partyID).Error; err != nil {
		return nil, err
	}

	if err := h.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", inviter.ID).Update("party_id", partyID).Error; err != nil {
		return nil, err
	}

	if err := h.db.WithContext(ctx).Model(&models.PartyInvite{}).Where("invitee_id = ?", user.ID).Delete(&models.PartyInvite{}).Error; err != nil {
		return nil, err
	}

	if err := h.db.WithContext(ctx).Model(&models.PartyInvite{}).Where("inviter_id = ?", user.ID).Delete(&models.PartyInvite{}).Error; err != nil {
		return nil, err
	}

	return &partyInvite, nil
}

func (h *partyInviteHandle) Reject(ctx context.Context, id uuid.UUID, user *models.User) error {
	var partyInvite models.PartyInvite
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&partyInvite).Error; err != nil {
		return err
	}

	if partyInvite.InviteeID != user.ID && partyInvite.InviterID != user.ID {
		return errors.New("not invitee or inviter")
	}

	if err := h.db.WithContext(ctx).Model(&models.PartyInvite{}).Where("id = ?", id).Delete(&models.PartyInvite{}).Error; err != nil {
		return err
	}

	return nil
}
