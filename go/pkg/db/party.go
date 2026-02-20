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

func (h *partyHandle) CreateWithMembers(ctx context.Context, leaderID uuid.UUID, memberIDs []uuid.UUID) (*models.Party, error) {
	if leaderID == uuid.Nil {
		return nil, errors.New("leader ID is required")
	}

	uniqueIDs := map[uuid.UUID]struct{}{
		leaderID: {},
	}
	for _, id := range memberIDs {
		uniqueIDs[id] = struct{}{}
	}

	if len(uniqueIDs) > MaxPartySize {
		return nil, ErrMaxPartySizeReached
	}

	ids := make([]uuid.UUID, 0, len(uniqueIDs))
	for id := range uniqueIDs {
		ids = append(ids, id)
	}

	var createdParty *models.Party

	if err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var users []models.User
		if err := tx.Where("id IN ?", ids).Find(&users).Error; err != nil {
			return err
		}

		if len(users) != len(ids) {
			return errors.New("one or more users not found")
		}

		for _, user := range users {
			if user.PartyID != nil {
				return errors.New("user already in a party")
			}
		}

		party := &models.Party{
			LeaderID: leaderID,
		}
		if err := tx.Create(party).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.User{}).Where("id IN ?", ids).Update("party_id", party.ID).Error; err != nil {
			return err
		}

		createdParty = party
		return nil
	}); err != nil {
		return nil, err
	}

	return h.FindByID(ctx, createdParty.ID)
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

func (h *partyHandle) SetLeaderAdmin(ctx context.Context, partyID uuid.UUID, leaderID uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var party models.Party
		if err := tx.Where("id = ?", partyID).First(&party).Error; err != nil {
			return err
		}

		var user models.User
		if err := tx.Where("id = ?", leaderID).First(&user).Error; err != nil {
			return err
		}

		if user.PartyID == nil || *user.PartyID != partyID {
			return errors.New("leader must be a party member")
		}

		return tx.Model(&models.Party{}).Where("id = ?", partyID).Update("leader_id", leaderID).Error
	})
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

func (h *partyHandle) AddMember(ctx context.Context, partyID uuid.UUID, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var party models.Party
		if err := tx.Preload("Members").Where("id = ?", partyID).First(&party).Error; err != nil {
			return err
		}

		var user models.User
		if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
			return err
		}

		if user.PartyID != nil {
			if *user.PartyID == partyID {
				return nil
			}
			return errors.New("user already in a party")
		}

		if len(party.Members) >= MaxPartySize {
			return ErrMaxPartySizeReached
		}

		return tx.Model(&models.User{}).Where("id = ?", userID).Update("party_id", partyID).Error
	})
}

func (h *partyHandle) RemoveMember(ctx context.Context, partyID uuid.UUID, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var party models.Party
		if err := tx.Where("id = ?", partyID).First(&party).Error; err != nil {
			return err
		}

		if party.LeaderID == userID {
			return errors.New("cannot remove party leader")
		}

		var user models.User
		if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
			return err
		}

		if user.PartyID == nil || *user.PartyID != partyID {
			return errors.New("user not in party")
		}

		return tx.Model(&models.User{}).Where("id = ?", userID).Update("party_id", nil).Error
	})
}

func (h *partyHandle) Delete(ctx context.Context, partyID uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("party_id = ?", partyID).Update("party_id", nil).Error; err != nil {
			return err
		}

		return tx.Where("id = ?", partyID).Delete(&models.Party{}).Error
	})
}

func (h *partyHandle) FindAll(ctx context.Context) ([]models.Party, error) {
	var parties []models.Party
	if err := h.db.WithContext(ctx).
		Preload("Members").
		Preload("Leader").
		Order("created_at desc").
		Find(&parties).Error; err != nil {
		return nil, err
	}
	return parties, nil
}

func (h *partyHandle) FindByID(ctx context.Context, partyID uuid.UUID) (*models.Party, error) {
	var party models.Party
	if err := h.db.WithContext(ctx).
		Preload("Members").
		Preload("Leader").
		Where("id = ?", partyID).
		First(&party).Error; err != nil {
		return nil, err
	}
	return &party, nil
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
