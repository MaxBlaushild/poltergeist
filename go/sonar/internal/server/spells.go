package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type spellEffectPayload struct {
	Type             string                         `json:"type"`
	Amount           *int                           `json:"amount"`
	StatusesToApply  []scenarioFailureStatusPayload `json:"statusesToApply"`
	StatusesToRemove []string                       `json:"statusesToRemove"`
	EffectData       map[string]interface{}         `json:"effectData"`
}

type spellUpsertRequest struct {
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	IconURL       string               `json:"iconUrl"`
	EffectText    string               `json:"effectText"`
	SchoolOfMagic string               `json:"schoolOfMagic"`
	ManaCost      int                  `json:"manaCost"`
	Effects       []spellEffectPayload `json:"effects"`
}

func normalizeSpellStatusNames(values []string) models.StringArray {
	out := models.StringArray{}
	seen := map[string]bool{}
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		key := strings.ToLower(normalized)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, normalized)
	}
	return out
}

func (s *server) parseSpellEffects(input []spellEffectPayload) (models.SpellEffects, error) {
	effects := models.SpellEffects{}
	for index, effectPayload := range input {
		effectType := models.SpellEffectType(strings.TrimSpace(strings.ToLower(effectPayload.Type)))
		if effectType == "" {
			return nil, fmt.Errorf("effects[%d].type is required", index)
		}

		amount := 0
		if effectPayload.Amount != nil {
			amount = *effectPayload.Amount
		}

		statusesToApply, err := parseScenarioFailureStatusTemplates(
			effectPayload.StatusesToApply,
			fmt.Sprintf("effects[%d].statusesToApply", index),
		)
		if err != nil {
			return nil, err
		}
		statusesToRemove := normalizeSpellStatusNames(effectPayload.StatusesToRemove)

		switch effectType {
		case models.SpellEffectTypeDealDamage,
			models.SpellEffectTypeRestoreLifePartyMember,
			models.SpellEffectTypeRestoreLifeAllParty:
			if amount <= 0 {
				return nil, fmt.Errorf("effects[%d].amount must be greater than 0", index)
			}
		case models.SpellEffectTypeApplyBeneficialStatus:
			if len(statusesToApply) == 0 {
				return nil, fmt.Errorf("effects[%d].statusesToApply is required", index)
			}
		case models.SpellEffectTypeRemoveDetrimental:
			if len(statusesToRemove) == 0 {
				return nil, fmt.Errorf("effects[%d].statusesToRemove is required", index)
			}
		default:
			// Allow new effect types without backend changes.
		}

		effects = append(effects, models.SpellEffect{
			Type:             effectType,
			Amount:           amount,
			StatusesToApply:  statusesToApply,
			StatusesToRemove: statusesToRemove,
			EffectData:       effectPayload.EffectData,
		})
	}
	if effects == nil {
		return models.SpellEffects{}, nil
	}
	return effects, nil
}

func (s *server) parseSpellUpsertRequest(body spellUpsertRequest) (*models.Spell, error) {
	name := strings.TrimSpace(body.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if body.ManaCost < 0 {
		return nil, fmt.Errorf("manaCost must be zero or greater")
	}
	schoolOfMagic := strings.TrimSpace(body.SchoolOfMagic)
	if schoolOfMagic == "" {
		return nil, fmt.Errorf("schoolOfMagic is required")
	}

	effects, err := s.parseSpellEffects(body.Effects)
	if err != nil {
		return nil, err
	}

	return &models.Spell{
		Name:          name,
		Description:   strings.TrimSpace(body.Description),
		IconURL:       strings.TrimSpace(body.IconURL),
		EffectText:    strings.TrimSpace(body.EffectText),
		SchoolOfMagic: schoolOfMagic,
		ManaCost:      body.ManaCost,
		Effects:       effects,
	}, nil
}

func (s *server) getSpells(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	spells, err := s.dbClient.Spell().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, spells)
}

func (s *server) getSpell(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	spellID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid spell ID"})
		return
	}

	spell, err := s.dbClient.Spell().FindByID(ctx, spellID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "spell not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, spell)
}

func (s *server) createSpell(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody spellUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	spell, err := s.parseSpellUpsertRequest(requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Spell().Create(ctx, spell); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created, err := s.dbClient.Spell().FindByID(ctx, spell.ID)
	if err != nil {
		ctx.JSON(http.StatusCreated, spell)
		return
	}
	ctx.JSON(http.StatusCreated, created)
}

func (s *server) updateSpell(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	spellID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid spell ID"})
		return
	}

	if _, err := s.dbClient.Spell().FindByID(ctx, spellID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "spell not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody spellUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	spell, err := s.parseSpellUpsertRequest(requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Spell().Update(ctx, spellID, map[string]interface{}{
		"name":            spell.Name,
		"description":     spell.Description,
		"icon_url":        spell.IconURL,
		"effect_text":     spell.EffectText,
		"school_of_magic": spell.SchoolOfMagic,
		"mana_cost":       spell.ManaCost,
		"effects":         spell.Effects,
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := s.dbClient.Spell().FindByID(ctx, spellID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"id": spellID})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) deleteSpell(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	spellID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid spell ID"})
		return
	}

	if _, err := s.dbClient.Spell().FindByID(ctx, spellID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "spell not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Spell().Delete(ctx, spellID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "spell deleted successfully"})
}

func (s *server) getCurrentUserSpells(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userSpells, err := s.dbClient.UserSpell().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, userSpells)
}
