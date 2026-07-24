package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type reefProductHandle struct {
	db *gorm.DB
}

func (h *reefProductHandle) FindBySlug(ctx context.Context, slug string) (*models.ReefProduct, error) {
	var product models.ReefProduct
	if err := h.db.WithContext(ctx).Where("slug = ? AND active = true", slug).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (h *reefProductHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ReefProduct, error) {
	var product models.ReefProduct
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (h *reefProductHandle) FindActive(ctx context.Context) ([]models.ReefProduct, error) {
	var products []models.ReefProduct
	if err := h.db.WithContext(ctx).Where("active = true").Order("kind, name").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

type reefProductVariantHandle struct {
	db *gorm.DB
}

func (h *reefProductVariantHandle) FindByProductID(ctx context.Context, productID uuid.UUID) ([]models.ReefProductVariant, error) {
	var variants []models.ReefProductVariant
	if err := h.db.WithContext(ctx).
		Where("product_id = ? AND active = true", productID).
		Order("price_cents ASC").
		Find(&variants).Error; err != nil {
		return nil, err
	}
	return variants, nil
}

func (h *reefProductVariantHandle) FindByProductAndKey(ctx context.Context, productID uuid.UUID, variantKey string) (*models.ReefProductVariant, error) {
	var variant models.ReefProductVariant
	if err := h.db.WithContext(ctx).
		Where("product_id = ? AND variant_key = ? AND active = true", productID, variantKey).
		First(&variant).Error; err != nil {
		return nil, err
	}
	return &variant, nil
}

type reefParameterSchemaHandle struct {
	db *gorm.DB
}

func (h *reefParameterSchemaHandle) FindActiveByProductID(ctx context.Context, productID uuid.UUID) (*models.ReefParameterSchema, error) {
	var schema models.ReefParameterSchema
	if err := h.db.WithContext(ctx).
		Where("product_id = ? AND active = true", productID).
		Order("version DESC").
		First(&schema).Error; err != nil {
		return nil, err
	}
	return &schema, nil
}

type reefTankProfileHandle struct {
	db *gorm.DB
}

// FindVerified returns only rows with a real source_url (R-3.4): unverified
// rows are research backlog and must never appear in configurator dropdowns.
func (h *reefTankProfileHandle) FindVerified(ctx context.Context) ([]models.ReefTankProfile, error) {
	var profiles []models.ReefTankProfile
	if err := h.db.WithContext(ctx).
		Where("verified = true").
		Order("manufacturer, model").
		Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

func (h *reefTankProfileHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ReefTankProfile, error) {
	var profile models.ReefTankProfile
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (h *reefTankProfileHandle) FindByManufacturerAndModel(ctx context.Context, manufacturer, model string) (*models.ReefTankProfile, error) {
	var profile models.ReefTankProfile
	if err := h.db.WithContext(ctx).
		Where("verified = true AND lower(manufacturer) = lower(?) AND lower(model) = lower(?)", manufacturer, model).
		First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}
