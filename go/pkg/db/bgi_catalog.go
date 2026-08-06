package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type bgiGameHandle struct {
	db *gorm.DB
}

func (h *bgiGameHandle) FindBySlug(ctx context.Context, slug string) (*models.BgiGame, error) {
	var game models.BgiGame
	if err := h.db.WithContext(ctx).Where("slug = ? AND active = true", slug).First(&game).Error; err != nil {
		return nil, err
	}
	return &game, nil
}

func (h *bgiGameHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.BgiGame, error) {
	var game models.BgiGame
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&game).Error; err != nil {
		return nil, err
	}
	return &game, nil
}

func (h *bgiGameHandle) FindActive(ctx context.Context) ([]models.BgiGame, error) {
	var games []models.BgiGame
	if err := h.db.WithContext(ctx).Where("active = true").Order("name").Find(&games).Error; err != nil {
		return nil, err
	}
	return games, nil
}

type bgiExpansionHandle struct {
	db *gorm.DB
}

func (h *bgiExpansionHandle) FindByGameID(ctx context.Context, gameID uuid.UUID) ([]models.BgiExpansion, error) {
	var expansions []models.BgiExpansion
	if err := h.db.WithContext(ctx).
		Where("game_id = ? AND active = true", gameID).
		Order("name").
		Find(&expansions).Error; err != nil {
		return nil, err
	}
	return expansions, nil
}

type bgiBoxProfileHandle struct {
	db *gorm.DB
}

// FindByGameID intentionally does NOT filter on verified — unlike
// ReefTankProfile.FindVerified, unverified rows must still surface in the
// UI with an honest caveat rather than be hidden (see the seed migration's
// own comment on why active/verified are separate concerns for bgi).
func (h *bgiBoxProfileHandle) FindByGameID(ctx context.Context, gameID uuid.UUID) ([]models.BgiBoxProfile, error) {
	var profiles []models.BgiBoxProfile
	if err := h.db.WithContext(ctx).
		Where("game_id = ?", gameID).
		Order("source, label").
		Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

func (h *bgiBoxProfileHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.BgiBoxProfile, error) {
	var profile models.BgiBoxProfile
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

type bgiSleeveProfileHandle struct {
	db *gorm.DB
}

// FindAll deliberately returns every row regardless of verified, same
// reasoning as bgiBoxProfileHandle.FindByGameID above.
func (h *bgiSleeveProfileHandle) FindAll(ctx context.Context) ([]models.BgiSleeveProfile, error) {
	var profiles []models.BgiSleeveProfile
	if err := h.db.WithContext(ctx).Order("base_card_thickness_mm, sleeve_material_thickness_mm").Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

func (h *bgiSleeveProfileHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.BgiSleeveProfile, error) {
	var profile models.BgiSleeveProfile
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

type bgiComponentManifestHandle struct {
	db *gorm.DB
}

// FindByGameID returns the base-game manifest rows (expansion_id IS NULL)
// plus any rows for the given expansion IDs — v1 only ever calls this with
// an empty expansionIDs slice (base game only), but the shape supports
// R-2.4's eventual expansion toggling without a signature change.
func (h *bgiComponentManifestHandle) FindByGameID(ctx context.Context, gameID uuid.UUID, expansionIDs []uuid.UUID) ([]models.BgiComponentManifest, error) {
	var manifests []models.BgiComponentManifest
	q := h.db.WithContext(ctx).Where("game_id = ?", gameID)
	if len(expansionIDs) > 0 {
		q = q.Where("expansion_id IS NULL OR expansion_id IN ?", expansionIDs)
	} else {
		q = q.Where("expansion_id IS NULL")
	}
	if err := q.Order("component_type").Find(&manifests).Error; err != nil {
		return nil, err
	}
	return manifests, nil
}

type bgiTrayTemplateHandle struct {
	db *gorm.DB
}

func (h *bgiTrayTemplateHandle) FindBySlug(ctx context.Context, slug string) (*models.BgiTrayTemplate, error) {
	var tmpl models.BgiTrayTemplate
	if err := h.db.WithContext(ctx).Where("slug = ? AND active = true", slug).First(&tmpl).Error; err != nil {
		return nil, err
	}
	return &tmpl, nil
}

func (h *bgiTrayTemplateHandle) FindByComponentType(ctx context.Context, componentType string) (*models.BgiTrayTemplate, error) {
	var tmpl models.BgiTrayTemplate
	err := h.db.WithContext(ctx).Where("component_type = ? AND active = true", componentType).First(&tmpl).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tmpl, nil
}

type bgiProductHandle struct {
	db *gorm.DB
}

func (h *bgiProductHandle) FindBySlug(ctx context.Context, slug string) (*models.BgiProduct, error) {
	var product models.BgiProduct
	if err := h.db.WithContext(ctx).Where("slug = ? AND active = true", slug).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (h *bgiProductHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.BgiProduct, error) {
	var product models.BgiProduct
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (h *bgiProductHandle) FindByGameID(ctx context.Context, gameID uuid.UUID) (*models.BgiProduct, error) {
	var product models.BgiProduct
	if err := h.db.WithContext(ctx).Where("game_id = ? AND active = true", gameID).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (h *bgiProductHandle) FindActive(ctx context.Context) ([]models.BgiProduct, error) {
	var products []models.BgiProduct
	if err := h.db.WithContext(ctx).Where("active = true").Order("name").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

type bgiParameterSchemaHandle struct {
	db *gorm.DB
}

func (h *bgiParameterSchemaHandle) FindActiveByProductID(ctx context.Context, productID uuid.UUID) (*models.BgiParameterSchema, error) {
	var schema models.BgiParameterSchema
	if err := h.db.WithContext(ctx).
		Where("product_id = ? AND active = true", productID).
		Order("version DESC").
		First(&schema).Error; err != nil {
		return nil, err
	}
	return &schema, nil
}
