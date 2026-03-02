package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// GenerateMonsterTemplatesBulkProcessor creates monster templates in the background.
type GenerateMonsterTemplatesBulkProcessor struct {
	dbClient    db.DbClient
	redisClient *redis.Client
}

type monsterTemplateAbilityPool struct {
	ordered []models.Spell
	byName  map[string]models.Spell
}

func NewGenerateMonsterTemplatesBulkProcessor(dbClient db.DbClient, redisClient *redis.Client) GenerateMonsterTemplatesBulkProcessor {
	log.Println("Initializing GenerateMonsterTemplatesBulkProcessor")
	return GenerateMonsterTemplatesBulkProcessor{
		dbClient:    dbClient,
		redisClient: redisClient,
	}
}

func (p *GenerateMonsterTemplatesBulkProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate monster templates bulk task: %v", task.Type())

	var payload jobs.GenerateMonsterTemplatesBulkTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal generate monster templates bulk payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if payload.JobID == uuid.Nil {
		return fmt.Errorf("missing job ID")
	}

	statusKey := jobs.MonsterTemplateBulkStatusKey(payload.JobID)
	now := time.Now().UTC()
	status := jobs.MonsterTemplateBulkStatus{
		JobID:        payload.JobID,
		Status:       jobs.MonsterTemplateBulkStatusInProgress,
		Source:       strings.TrimSpace(payload.Source),
		TotalCount:   payload.TotalCount,
		CreatedCount: 0,
		StartedAt:    &now,
		UpdatedAt:    now,
	}
	if status.TotalCount <= 0 {
		status.TotalCount = len(payload.Templates)
	}
	if status.Source == "" {
		status.Source = "seed_generated"
	}
	p.setStatus(ctx, statusKey, status)

	if len(payload.Templates) == 0 {
		err := fmt.Errorf("no monster templates provided for bulk generation")
		p.markFailed(ctx, statusKey, status, err)
		return err
	}

	abilityPool, err := p.loadAbilityPool(ctx)
	if err != nil {
		p.markFailed(ctx, statusKey, status, err)
		return fmt.Errorf("failed to load abilities for template assignment: %w", err)
	}

	for index, spec := range payload.Templates {
		emptyError := ""
		template := &models.MonsterTemplate{
			Name:                  strings.TrimSpace(spec.Name),
			Description:           strings.TrimSpace(spec.Description),
			BaseStrength:          spec.BaseStrength,
			BaseDexterity:         spec.BaseDexterity,
			BaseConstitution:      spec.BaseConstitution,
			BaseIntelligence:      spec.BaseIntelligence,
			BaseWisdom:            spec.BaseWisdom,
			BaseCharisma:          spec.BaseCharisma,
			ImageGenerationStatus: models.MonsterTemplateImageGenerationStatusNone,
			ImageGenerationError:  &emptyError,
		}
		if template.Name == "" {
			template.Name = fmt.Sprintf("Monster Template %d", index+1)
		}

		if err := p.dbClient.MonsterTemplate().Create(ctx, template); err != nil {
			p.markFailed(ctx, statusKey, status, err)
			return fmt.Errorf("failed to create monster template %d/%d: %w", index+1, len(payload.Templates), err)
		}
		if err := p.assignAbilitiesToTemplate(ctx, template, abilityPool); err != nil {
			p.markFailed(ctx, statusKey, status, err)
			return fmt.Errorf("failed to assign abilities to monster template %d/%d: %w", index+1, len(payload.Templates), err)
		}

		status.CreatedCount = index + 1
		status.UpdatedAt = time.Now().UTC()
		p.setStatus(ctx, statusKey, status)
	}

	completedAt := time.Now().UTC()
	status.Status = jobs.MonsterTemplateBulkStatusCompleted
	status.CompletedAt = &completedAt
	status.UpdatedAt = completedAt
	p.setStatus(ctx, statusKey, status)

	return nil
}

func (p *GenerateMonsterTemplatesBulkProcessor) loadAbilityPool(ctx context.Context) (*monsterTemplateAbilityPool, error) {
	existing, err := p.dbClient.Spell().FindAll(ctx)
	if err != nil {
		return nil, err
	}

	pool := &monsterTemplateAbilityPool{
		ordered: make([]models.Spell, 0, len(existing)),
		byName:  make(map[string]models.Spell, len(existing)),
	}
	for _, spell := range existing {
		if spell.ID == uuid.Nil {
			continue
		}
		pool.ordered = append(pool.ordered, spell)
		normalized := normalizeAbilityName(spell.Name)
		if normalized == "" {
			continue
		}
		if _, exists := pool.byName[normalized]; exists {
			continue
		}
		pool.byName[normalized] = spell
	}

	return pool, nil
}

func normalizeAbilityName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func preferredAbilityCountForTemplate(template *models.MonsterTemplate) int {
	if template == nil {
		return 1
	}
	seed := template.BaseStrength + template.BaseDexterity + template.BaseConstitution + template.BaseIntelligence + template.BaseWisdom + template.BaseCharisma
	if seed < 0 {
		seed = -seed
	}
	return 1 + (seed % 3)
}

func (p *GenerateMonsterTemplatesBulkProcessor) assignAbilitiesToTemplate(
	ctx context.Context,
	template *models.MonsterTemplate,
	pool *monsterTemplateAbilityPool,
) error {
	if template == nil || pool == nil {
		return nil
	}

	targetCount := preferredAbilityCountForTemplate(template)
	selected := make([]models.Spell, 0, targetCount)
	selectedIDs := map[uuid.UUID]struct{}{}

	// Reuse existing abilities first.
	for _, ability := range randomAbilitySelection(pool.ordered, targetCount) {
		if ability.ID == uuid.Nil {
			continue
		}
		if _, exists := selectedIDs[ability.ID]; exists {
			continue
		}
		selected = append(selected, ability)
		selectedIDs[ability.ID] = struct{}{}
		if len(selected) >= targetCount {
			break
		}
	}

	for len(selected) < targetCount {
		created, err := p.findOrCreateAbilityForTemplate(ctx, template, len(selected), pool)
		if err != nil {
			return err
		}
		if created == nil || created.ID == uuid.Nil {
			return fmt.Errorf("created ability is invalid")
		}
		if _, exists := selectedIDs[created.ID]; exists {
			continue
		}
		selected = append(selected, *created)
		selectedIDs[created.ID] = struct{}{}
	}

	links := make([]models.MonsterTemplateSpell, 0, len(selected))
	for _, ability := range selected {
		links = append(links, models.MonsterTemplateSpell{
			SpellID: ability.ID,
		})
	}
	return p.dbClient.MonsterTemplate().ReplaceSpells(ctx, template.ID, links)
}

func randomAbilitySelection(all []models.Spell, count int) []models.Spell {
	if len(all) == 0 || count <= 0 {
		return []models.Spell{}
	}
	order := rand.Perm(len(all))
	selected := make([]models.Spell, 0, minInt(count, len(all)))
	for _, idx := range order {
		selected = append(selected, all[idx])
		if len(selected) >= count {
			break
		}
	}
	return selected
}

func (p *GenerateMonsterTemplatesBulkProcessor) findOrCreateAbilityForTemplate(
	ctx context.Context,
	template *models.MonsterTemplate,
	slot int,
	pool *monsterTemplateAbilityPool,
) (*models.Spell, error) {
	if template == nil || pool == nil {
		return nil, fmt.Errorf("template or ability pool missing")
	}
	abilityType := preferredAbilityTypeForTemplate(template, slot)
	baseName := strings.TrimSpace(template.Name)
	if baseName == "" {
		baseName = "Monster"
	}

	candidates := generatedAbilityNameCandidates(baseName, abilityType)
	for _, candidate := range candidates {
		normalized := normalizeAbilityName(candidate)
		if normalized == "" {
			continue
		}
		if existing, ok := pool.byName[normalized]; ok {
			copy := existing
			return &copy, nil
		}
	}

	name := nextUniqueGeneratedAbilityName(candidates[0], pool.byName)
	description, seededEffectText, schoolOfMagic, manaCost := generatedAbilityDetails(template, abilityType)
	spec := jobs.SpellCreationSpec{
		Name:          name,
		Description:   description,
		EffectText:    seededEffectText,
		SchoolOfMagic: schoolOfMagic,
		ManaCost:      manaCost,
		AbilityType:   string(abilityType),
	}
	effects := inferGeneratedAbilityEffects(spec, abilityType, manaCost)
	effectText := buildGeneratedAbilityEffectText(effects, abilityType)
	emptyError := ""
	spell := &models.Spell{
		Name:                  name,
		Description:           description,
		AbilityType:           abilityType,
		EffectText:            effectText,
		SchoolOfMagic:         schoolOfMagic,
		ManaCost:              manaCost,
		Effects:               effects,
		ImageGenerationStatus: models.SpellImageGenerationStatusNone,
		ImageGenerationError:  &emptyError,
	}
	if err := p.dbClient.Spell().Create(ctx, spell); err != nil {
		return nil, err
	}

	pool.ordered = append(pool.ordered, *spell)
	pool.byName[normalizeAbilityName(spell.Name)] = *spell
	return spell, nil
}

func preferredAbilityTypeForTemplate(template *models.MonsterTemplate, slot int) models.SpellAbilityType {
	if template == nil {
		return models.SpellAbilityTypeSpell
	}
	mental := template.BaseIntelligence + template.BaseWisdom
	physical := template.BaseStrength + template.BaseDexterity
	if mental >= physical {
		if slot%3 == 2 {
			return models.SpellAbilityTypeTechnique
		}
		return models.SpellAbilityTypeSpell
	}
	if slot == 0 {
		return models.SpellAbilityTypeTechnique
	}
	if slot%2 == 0 {
		return models.SpellAbilityTypeSpell
	}
	return models.SpellAbilityTypeTechnique
}

func generatedAbilityNameCandidates(baseName string, abilityType models.SpellAbilityType) []string {
	base := strings.TrimSpace(baseName)
	if base == "" {
		base = "Monster"
	}
	if abilityType == models.SpellAbilityTypeTechnique {
		return []string{
			fmt.Sprintf("%s Assault", base),
			fmt.Sprintf("%s Pounce", base),
			fmt.Sprintf("%s Riposte", base),
		}
	}
	return []string{
		fmt.Sprintf("%s Hex", base),
		fmt.Sprintf("%s Burst", base),
		fmt.Sprintf("%s Volley", base),
	}
}

func nextUniqueGeneratedAbilityName(base string, existing map[string]models.Spell) string {
	trimmed := strings.TrimSpace(base)
	if trimmed == "" {
		trimmed = "Monster Ability"
	}
	candidate := trimmed
	suffix := 2
	for {
		if _, exists := existing[normalizeAbilityName(candidate)]; !exists {
			return candidate
		}
		candidate = fmt.Sprintf("%s %d", trimmed, suffix)
		suffix++
	}
}

func generatedAbilityDetails(
	template *models.MonsterTemplate,
	abilityType models.SpellAbilityType,
) (description string, effectText string, schoolOfMagic string, manaCost int) {
	templateName := "monster"
	templateDescription := ""
	if template != nil {
		if strings.TrimSpace(template.Name) != "" {
			templateName = strings.TrimSpace(template.Name)
		}
		templateDescription = strings.TrimSpace(template.Description)
	}

	if abilityType == models.SpellAbilityTypeTechnique {
		description = fmt.Sprintf("A combat maneuver used by %s to pressure opponents in melee.", templateName)
		if templateDescription != "" {
			description = fmt.Sprintf("%s %s", description, templateDescription)
		}
		effectText = "A disciplined strike that exploits openings."
		schoolOfMagic = "Martial"
		manaCost = 0
		return description, effectText, schoolOfMagic, manaCost
	}

	description = fmt.Sprintf("A signature magical attack channeled by %s.", templateName)
	if templateDescription != "" {
		description = fmt.Sprintf("%s %s", description, templateDescription)
	}
	effectText = "A focused magical surge that punishes vulnerable targets."
	schoolOfMagic = "Arcane"
	manaCost = 8
	if template != nil {
		manaCost = 5 + ((template.BaseIntelligence + template.BaseWisdom) / 3)
	}
	if manaCost < 0 {
		manaCost = 0
	}
	return description, effectText, schoolOfMagic, manaCost
}

func (p *GenerateMonsterTemplatesBulkProcessor) markFailed(ctx context.Context, statusKey string, status jobs.MonsterTemplateBulkStatus, cause error) {
	if cause != nil {
		status.Error = cause.Error()
	}
	completedAt := time.Now().UTC()
	status.Status = jobs.MonsterTemplateBulkStatusFailed
	status.CompletedAt = &completedAt
	status.UpdatedAt = completedAt
	p.setStatus(ctx, statusKey, status)
}

func (p *GenerateMonsterTemplatesBulkProcessor) setStatus(ctx context.Context, statusKey string, status jobs.MonsterTemplateBulkStatus) {
	if p.redisClient == nil || strings.TrimSpace(statusKey) == "" {
		return
	}
	payload, err := json.Marshal(status)
	if err != nil {
		log.Printf("Failed to marshal monster template bulk status: %v", err)
		return
	}
	if err := p.redisClient.Set(ctx, statusKey, payload, jobs.MonsterTemplateBulkStatusTTL).Err(); err != nil {
		log.Printf("Failed to write monster template bulk status: %v", err)
	}
}
