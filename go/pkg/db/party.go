package db

import (
	"context"
	"errors"
	"math/rand"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type partyHandle struct {
	db *gorm.DB
}

func (h *partyHandle) Create(ctx context.Context) (*models.Party, error) {
	party := &models.Party{}

	if err := h.db.WithContext(ctx).Create(party).Error; err != nil {
		return nil, err
	}

	return party, nil
}

func (h *partyHandle) SetLeader(ctx context.Context, partyID uuid.UUID, leaderID uuid.UUID, userID uuid.UUID) error {
	party := &models.Party{}
	if err := h.db.WithContext(ctx).Where("id = ?", partyID).First(party).Error; err != nil {
		return err
	}

	if party.LeaderID != userID {
		return errors.New("user is not leader")
	}

	return h.db.WithContext(ctx).Model(&models.Party{}).Where("id = ?", partyID).Update("leader_id", leaderID).Error
}

func (h *partyHandle) LeaveParty(ctx context.Context, user *models.User) error {
	var party models.Party
	if err := h.db.WithContext(ctx).Where("id = ?", user.PartyID).First(&party).Error; err != nil {
		return err
	}

	if party.LeaderID == user.ID {
		partyMembers := []models.User{}
		if err := h.db.WithContext(ctx).Where("party_id = ?", user.PartyID).Find(&partyMembers).Error; err != nil {
			return err
		}

		var remainingMembers []models.User
		for _, member := range partyMembers {
			if member.ID != user.ID {
				remainingMembers = append(remainingMembers, member)
			}
		}

		if len(remainingMembers) == 0 {
			// Clear the user's party ID first
			if err := h.db.WithContext(ctx).Model(user).Update("party_id", nil).Error; err != nil {
				return err
			}
			// Now delete the party
			if err := h.db.WithContext(ctx).Where("id = ?", user.PartyID).Delete(&models.Party{}).Error; err != nil {
				return err
			}
			return nil
		}

		newLeader := remainingMembers[rand.Intn(len(remainingMembers))]
		if err := h.db.WithContext(ctx).Model(&models.Party{}).Where("id = ?", user.PartyID).Update("leader_id", newLeader.ID).Error; err != nil {
			return err
		}
	}

	// Clear the user's party ID when leaving
	return h.db.WithContext(ctx).Model(user).Update("party_id", nil).Error
}

func (h *partyHandle) FindUsersParty(ctx context.Context, partyID uuid.UUID) (*models.Party, error) {
	var party *models.Party
	if err := h.db.Preload("Members").Preload("Leader").WithContext(ctx).Where("id = ?", partyID).Find(&party).Error; err != nil {
		return nil, err
	}
	return party, nil
}

func (h *partyHandle) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	// Delete parties where the user is the leader
	return h.db.WithContext(ctx).Where("leader_id = ?", userID).Delete(&models.Party{}).Error
}
