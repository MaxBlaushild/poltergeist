package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type documentHandler struct {
	db *gorm.DB
}

func (h *documentHandler) Create(
	ctx context.Context,
	document *models.Document,
	existingTagIDs []uuid.UUID,
	newTagTexts []string,
) (*models.Document, error) {
	// Start transaction
	tx := h.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Set document ID and timestamps if not set
	if document.ID == uuid.Nil {
		document.ID = uuid.New()
	}
	if document.CreatedAt.IsZero() {
		document.CreatedAt = time.Now()
	}
	if document.UpdatedAt.IsZero() {
		document.UpdatedAt = time.Now()
	}

	// Create the document
	if err := tx.Create(document).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Get DocumentTag handler with transaction
	documentTagHandler := &documentTagHandler{db: tx}

	// Collect all tags to associate
	var tagsToAssociate []models.DocumentTag

	// Handle existing tag IDs
	if len(existingTagIDs) > 0 {
		existingTags, err := documentTagHandler.FindByIDs(ctx, existingTagIDs)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		tagsToAssociate = append(tagsToAssociate, existingTags...)
	}

	// Handle new tag texts
	if len(newTagTexts) > 0 {
		for _, tagText := range newTagTexts {
			if tagText == "" {
				continue
			}
			newTag, err := documentTagHandler.FindOrCreateByText(ctx, tagText)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
			tagsToAssociate = append(tagsToAssociate, *newTag)
		}
	}

	// Associate tags with document using GORM association
	if len(tagsToAssociate) > 0 {
		if err := tx.Model(document).Association("DocumentTags").Append(tagsToAssociate); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Reload document with tags and locations preloaded
	var createdDocument models.Document
	if err := h.db.WithContext(ctx).
		Preload("DocumentTags").
		Preload("DocumentLocations").
		Where("id = ?", document.ID).
		First(&createdDocument).Error; err != nil {
		return nil, err
	}

	return &createdDocument, nil
}

func (h *documentHandler) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.Document, error) {
	var documents []models.Document
	if err := h.db.WithContext(ctx).
		Preload("DocumentTags").
		Preload("DocumentLocations").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&documents).Error; err != nil {
		return nil, err
	}
	return documents, nil
}

func (h *documentHandler) FindByUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]models.Document, error) {
	if len(userIDs) == 0 {
		return []models.Document{}, nil
	}
	var documents []models.Document
	if err := h.db.WithContext(ctx).
		Preload("DocumentTags").
		Preload("DocumentLocations").
		Preload("User").
		Where("user_id IN ?", userIDs).
		Order("created_at DESC").
		Find(&documents).Error; err != nil {
		return nil, err
	}
	return documents, nil
}

func (h *documentHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	var document models.Document
	if err := h.db.WithContext(ctx).
		Preload("DocumentTags").
		Preload("DocumentLocations").
		Where("id = ?", id).
		First(&document).Error; err != nil {
		return nil, err
	}
	return &document, nil
}

func (h *documentHandler) Update(ctx context.Context, document *models.Document) error {
	document.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).Save(document).Error
}

func (h *documentHandler) UpdateTags(ctx context.Context, documentID uuid.UUID, existingTagIDs []uuid.UUID, newTagTexts []string) error {
	// Start transaction
	tx := h.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get DocumentTag handler with transaction
	documentTagHandler := &documentTagHandler{db: tx}

	// Collect all tags to associate
	var tagsToAssociate []models.DocumentTag

	// Handle existing tag IDs
	if len(existingTagIDs) > 0 {
		existingTags, err := documentTagHandler.FindByIDs(ctx, existingTagIDs)
		if err != nil {
			tx.Rollback()
			return err
		}
		tagsToAssociate = append(tagsToAssociate, existingTags...)
	}

	// Handle new tag texts
	if len(newTagTexts) > 0 {
		for _, tagText := range newTagTexts {
			if tagText == "" {
				continue
			}
			newTag, err := documentTagHandler.FindOrCreateByText(ctx, tagText)
			if err != nil {
				tx.Rollback()
				return err
			}
			tagsToAssociate = append(tagsToAssociate, *newTag)
		}
	}

	// Replace document tags using GORM association
	document := &models.Document{ID: documentID}
	if err := tx.Model(document).Association("DocumentTags").Replace(tagsToAssociate); err != nil {
		tx.Rollback()
		return err
	}

	// Update document's updated_at timestamp
	if err := tx.Model(document).Update("updated_at", time.Now()).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (h *documentHandler) Delete(ctx context.Context, id uuid.UUID) error {
	// Start transaction to handle cascade deletes
	tx := h.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Delete associations first (GORM many-to-many)
	if err := tx.Model(&models.Document{ID: id}).Association("DocumentTags").Clear(); err != nil {
		tx.Rollback()
		return err
	}

	// Delete the document
	if err := tx.Delete(&models.Document{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
