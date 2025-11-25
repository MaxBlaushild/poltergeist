package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userHandle struct {
	db *gorm.DB
}

func (h *userHandle) LeaveParty(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("party_id", nil).Error
}

func (h *userHandle) Update(ctx context.Context, userID uuid.UUID, updates models.User) error {
	return h.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (h *userHandle) SetUsername(ctx context.Context, userID uuid.UUID, username string) error {
	return h.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("username", username).Error
}

func (h *userHandle) FindLikeByUsername(ctx context.Context, username string) ([]*models.User, error) {
	var users []*models.User
	if err := h.db.WithContext(ctx).Where("username LIKE ?", "%"+username+"%").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (h *userHandle) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := h.db.WithContext(ctx).Where("LOWER(username) = LOWER(?)", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (h *userHandle) Insert(ctx context.Context, name string, phoneNumber string, id *uuid.UUID, username *string, dateOfBirth *time.Time, gender *string, latitude *float64, longitude *float64, locationAddress *string, bio *string) (*models.User, error) {
	user := models.User{
		Name:            name,
		PhoneNumber:     phoneNumber,
		Username:        username,
		DateOfBirth:     dateOfBirth,
		Gender:          gender,
		Latitude:        latitude,
		Longitude:       longitude,
		LocationAddress: locationAddress,
		Bio:             bio,
	}

	if id != nil {
		user.ID = *id
	}

	if err := h.db.WithContext(ctx).Model(&models.User{}).Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *userHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := h.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *userHandle) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error) {
	var user models.User
	if err := h.db.WithContext(ctx).Where(&models.User{PhoneNumber: phoneNumber}).First(&user).Error; err != nil {
		return nil, err
	}

	if uuid.Nil == user.ID {
		return nil, gorm.ErrRecordNotFound
	}

	return &user, nil
}

func (h *userHandle) FindUsersByIDs(ctx context.Context, userIDs []uuid.UUID) ([]models.User, error) {
	var users []models.User

	if err := h.db.WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (h *userHandle) FindAll(ctx context.Context) ([]models.User, error) {
	var users []models.User

	if err := h.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (h *userHandle) Delete(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.User{}, userID).Error
}

func (h *userHandle) DeleteAll(ctx context.Context) error {
	return h.db.WithContext(ctx).Where("1 = 1").Delete(&models.User{}).Error
}

func (h *userHandle) UpdateProfilePictureUrl(ctx context.Context, userID uuid.UUID, url string) error {
	return h.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("profile_picture_url", url).Error
}

func (h *userHandle) UpdateHasSeenTutorial(ctx context.Context, userID uuid.UUID, hasSeenTutorial bool) error {
	return h.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("has_seen_tutorial", hasSeenTutorial).Error
}

func (h *userHandle) JoinParty(ctx context.Context, inviterID uuid.UUID, inviteeID uuid.UUID) error {
	inviter, err := h.FindByID(ctx, inviterID)
	if err != nil {
		return err
	}

	partyID := inviter.PartyID
	partyExisted := partyID != nil

	if partyID == nil {
		party := &models.Party{}
		if err := h.db.WithContext(ctx).Create(&party).Error; err != nil {
			return err
		}
		partyID = &party.ID
	}

	partyMembers := []models.User{}

	if err := h.db.WithContext(ctx).Where("party_id = ?", inviter.PartyID).Find(&partyMembers).Error; err != nil {
		return err
	}

	if len(partyMembers) >= MaxPartySize {
		return ErrMaxPartySizeReached
	}

	if err := h.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", inviteeID).Update("party_id", partyID).Error; err != nil {
		return err
	}

	if partyExisted {
		if err := h.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", inviterID).Update("party_id", partyID).Error; err != nil {
			return err
		}
	}

	return nil
}

func (h *userHandle) FindPartyMembers(ctx context.Context, userID uuid.UUID) ([]models.User, error) {
	var foundUsers []models.User
	var users []models.User

	user, err := h.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := h.db.WithContext(ctx).Where("party_id = ?", user.PartyID).Find(&foundUsers).Error; err != nil {
		return nil, err
	}

	for _, user := range foundUsers {
		if user.ID == userID {
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

func (h *userHandle) AddGold(ctx context.Context, userID uuid.UUID, amount int) error {
	if amount < 0 {
		// Disallow negative increments here; use Update if debit needed later
		return gorm.ErrInvalidData
	}
	return h.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		UpdateColumn("gold", gorm.Expr("gold + ?", amount)).Error
}

func (h *userHandle) SetGold(ctx context.Context, userID uuid.UUID, amount int) error {
	if amount < 0 {
		return gorm.ErrInvalidData
	}
	return h.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("gold", amount).Error
}

func (h *userHandle) SubtractGold(ctx context.Context, userID uuid.UUID, amount int) error {
	if amount < 0 {
		return gorm.ErrInvalidData
	}
	return h.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		UpdateColumn("gold", gorm.Expr("gold - ?", amount)).Error
}

func (h *userHandle) AddCredits(ctx context.Context, userID uuid.UUID, amount int) error {
	if amount < 0 {
		// Disallow negative increments here; use Update if debit needed later
		return gorm.ErrInvalidData
	}
	return h.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		UpdateColumn("credits", gorm.Expr("credits + ?", amount)).Error
}

func (h *userHandle) SetCredits(ctx context.Context, userID uuid.UUID, amount int) error {
	if amount < 0 {
		return gorm.ErrInvalidData
	}
	return h.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("credits", amount).Error
}

func (h *userHandle) SubtractCredits(ctx context.Context, userID uuid.UUID, amount int) error {
	if amount < 0 {
		return gorm.ErrInvalidData
	}
	return h.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		UpdateColumn("credits", gorm.Expr("credits - ?", amount)).Error
}
