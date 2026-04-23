package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type monsterTemplateDraftZoneKindCueRule struct {
	monsterKeywords []string
	zoneCues        []string
}

var monsterTemplateDraftZoneKindCueRules = []monsterTemplateDraftZoneKindCueRule{
	{monsterKeywords: []string{"forest", "wood", "briar", "thorn", "moss", "wolf", "bear"}, zoneCues: []string{"forest", "wood", "grove", "wild", "jungle"}},
	{monsterKeywords: []string{"swamp", "bog", "mire", "ooze", "frog", "toad", "marsh"}, zoneCues: []string{"swamp", "bog", "marsh", "wetland", "mire"}},
	{monsterKeywords: []string{"desert", "dune", "sand", "scorpion", "sun", "vulture"}, zoneCues: []string{"desert", "dune", "sand", "waste", "arid"}},
	{monsterKeywords: []string{"mountain", "peak", "stone", "cliff", "goat", "eagle"}, zoneCues: []string{"mountain", "peak", "highland", "cliff", "rock"}},
	{monsterKeywords: []string{"cave", "cavern", "mine", "burrow", "tunnel", "deep"}, zoneCues: []string{"cave", "cavern", "underground", "mine", "tunnel"}},
	{monsterKeywords: []string{"crypt", "grave", "tomb", "bone", "undead", "wraith", "specter", "spectre"}, zoneCues: []string{"crypt", "grave", "tomb", "cemetery", "catacomb", "ruin"}},
	{monsterKeywords: []string{"coast", "reef", "shore", "tide", "sea", "ocean", "kraken"}, zoneCues: []string{"coast", "reef", "shore", "sea", "ocean", "water"}},
	{monsterKeywords: []string{"river", "lake", "torrent", "flood"}, zoneCues: []string{"river", "lake", "water", "wetland"}},
	{monsterKeywords: []string{"ice", "frost", "snow", "glacier", "winter"}, zoneCues: []string{"ice", "snow", "tundra", "glacier", "winter"}},
	{monsterKeywords: []string{"fire", "ember", "cinder", "magma", "lava", "ash", "volcan"}, zoneCues: []string{"volcan", "lava", "magma", "ash", "fire"}},
	{monsterKeywords: []string{"city", "clockwork", "guard", "bandit", "assassin"}, zoneCues: []string{"city", "urban", "street", "ruin", "fort"}},
}

func monsterTemplateSuggestionJobToBulkStatus(
	job *models.MonsterTemplateSuggestionJob,
) jobs.MonsterTemplateBulkStatus {
	status := jobs.MonsterTemplateBulkStatus{}
	if job == nil {
		return status
	}
	status.JobID = job.ID
	status.Status = job.Status
	status.Source = strings.TrimSpace(job.Source)
	status.MonsterType = string(models.NormalizeMonsterTemplateType(string(job.MonsterType)))
	if job.GenreID != uuid.Nil {
		status.GenreID = job.GenreID.String()
	}
	status.ZoneKind = models.NormalizeZoneKind(job.ZoneKind)
	status.YeetIt = job.YeetIt
	status.TotalCount = job.Count
	status.CreatedCount = job.CreatedCount
	if job.ErrorMessage != nil {
		status.Error = strings.TrimSpace(*job.ErrorMessage)
	}
	queuedAt := job.CreatedAt.UTC()
	status.QueuedAt = &queuedAt
	status.UpdatedAt = job.UpdatedAt.UTC()
	if job.Status == models.MonsterTemplateSuggestionJobStatusCompleted {
		completedAt := job.UpdatedAt.UTC()
		status.CompletedAt = &completedAt
	}
	if job.Status == models.MonsterTemplateSuggestionJobStatusInProgress {
		startedAt := job.UpdatedAt.UTC()
		status.StartedAt = &startedAt
	}
	return status
}

func (s *server) createMonsterTemplateSuggestionJobRecord(
	ctx *gin.Context,
	requestBody bulkGenerateMonsterTemplatesRequest,
) (*models.MonsterTemplateSuggestionJob, int, error) {
	if requestBody.Count < 1 || requestBody.Count > 100 {
		return nil, http.StatusBadRequest, fmt.Errorf("count must be between 1 and 100")
	}

	monsterType := models.NormalizeMonsterTemplateType(requestBody.MonsterType)
	genre, err := s.resolveMonsterGenre(ctx, requestBody.GenreID)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	zoneKind, err := s.resolveOptionalZoneKind(ctx, requestBody.ZoneKind)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	zoneKinds, err := s.dbClient.ZoneKind().FindAll(ctx)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	existingTemplates, err := s.dbClient.MonsterTemplate().FindAll(ctx)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	usedNames := make(map[string]struct{}, len(existingTemplates)+requestBody.Count)
	existingNames := make([]string, 0, len(existingTemplates))
	for _, template := range existingTemplates {
		normalized := strings.ToLower(strings.TrimSpace(template.Name))
		if normalized != "" {
			usedNames[normalized] = struct{}{}
		}
		name := strings.TrimSpace(template.Name)
		if name != "" {
			existingNames = append(existingNames, name)
		}
	}

	now := time.Now().UTC()
	job := &models.MonsterTemplateSuggestionJob{
		ID:           uuid.New(),
		CreatedAt:    now,
		UpdatedAt:    now,
		Status:       models.MonsterTemplateSuggestionJobStatusInProgress,
		MonsterType:  monsterType,
		GenreID:      genre.ID,
		Genre:        genre,
		ZoneKind:     models.NormalizeZoneKind(requestBody.ZoneKind),
		YeetIt:       requestBody.YeetIt,
		Source:       "seed_generated",
		Count:        requestBody.Count,
		CreatedCount: 0,
	}
	if zoneKind != nil {
		job.ZoneKind = models.NormalizeZoneKind(zoneKind.Slug)
	}
	if err := s.dbClient.MonsterTemplateSuggestionJob().Create(ctx, job); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	failJob := func(jobErr error, statusCode int) (*models.MonsterTemplateSuggestionJob, int, error) {
		msg := jobErr.Error()
		job.Status = models.MonsterTemplateSuggestionJobStatusFailed
		job.ErrorMessage = &msg
		job.UpdatedAt = time.Now().UTC()
		_ = s.dbClient.MonsterTemplateSuggestionJob().Update(ctx, job)
		return nil, statusCode, jobErr
	}

	templateSpecs, source, err := s.buildBulkMonsterTemplateSpecs(
		requestBody.Count,
		usedNames,
		existingNames,
		monsterType,
		genre,
		zoneKind,
		zoneKinds,
	)
	if err != nil {
		return failJob(err, http.StatusInternalServerError)
	}

	job.Source = source
	for _, spec := range templateSpecs {
		payload := buildMonsterTemplateSuggestionPayload(spec, genre, zoneKinds, zoneKind)
		if job.YeetIt {
			template := monsterTemplateFromSuggestionPayload(&models.MonsterTemplateSuggestionDraft{
				MonsterType: models.NormalizeMonsterTemplateType(payload.MonsterType),
				GenreID:     payload.GenreID,
				Genre:       genre,
				ZoneKind:    models.NormalizeZoneKind(payload.ZoneKind),
				Payload:     models.MonsterTemplateSuggestionPayloadValue(payload),
			})
			if err := s.dbClient.MonsterTemplate().Create(ctx, template); err != nil {
				return failJob(err, http.StatusInternalServerError)
			}
		} else {
			draft := &models.MonsterTemplateSuggestionDraft{
				ID:          uuid.New(),
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
				JobID:       job.ID,
				Status:      models.MonsterTemplateSuggestionDraftStatusSuggested,
				MonsterType: models.NormalizeMonsterTemplateType(payload.MonsterType),
				GenreID:     payload.GenreID,
				Genre:       genre,
				ZoneKind:    models.NormalizeZoneKind(payload.ZoneKind),
				Name:        strings.TrimSpace(payload.Name),
				Description: strings.TrimSpace(payload.Description),
				Payload:     models.MonsterTemplateSuggestionPayloadValue(payload),
			}
			if err := s.dbClient.MonsterTemplateSuggestionDraft().Create(ctx, draft); err != nil {
				return failJob(err, http.StatusInternalServerError)
			}
		}
		job.CreatedCount++
	}

	job.Status = models.MonsterTemplateSuggestionJobStatusCompleted
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now().UTC()
	if err := s.dbClient.MonsterTemplateSuggestionJob().Update(ctx, job); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return job, http.StatusAccepted, nil
}

func (s *server) getMonsterTemplateSuggestionJobs(ctx *gin.Context) {
	limit := 20
	if limitParam := strings.TrimSpace(ctx.Query("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	jobsList, err := s.dbClient.MonsterTemplateSuggestionJob().FindRecent(ctx, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]jobs.MonsterTemplateBulkStatus, 0, len(jobsList))
	for _, job := range jobsList {
		jobCopy := job
		response = append(response, monsterTemplateSuggestionJobToBulkStatus(&jobCopy))
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) getMonsterTemplateSuggestionDrafts(ctx *gin.Context) {
	jobID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster template suggestion job ID"})
		return
	}
	drafts, err := s.dbClient.MonsterTemplateSuggestionDraft().FindByJobID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, drafts)
}

func (s *server) deleteMonsterTemplateSuggestionDraft(ctx *gin.Context) {
	draftID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster template suggestion draft ID"})
		return
	}
	draft, err := s.dbClient.MonsterTemplateSuggestionDraft().FindByID(ctx, draftID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if draft == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "monster template suggestion draft not found"})
		return
	}
	if draft.MonsterTemplateID != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "converted drafts cannot be deleted"})
		return
	}
	if err := s.dbClient.MonsterTemplateSuggestionDraft().Delete(ctx, draftID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "monster template suggestion draft deleted"})
}

func (s *server) convertMonsterTemplateSuggestionDraft(ctx *gin.Context) {
	draftID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster template suggestion draft ID"})
		return
	}
	draft, err := s.dbClient.MonsterTemplateSuggestionDraft().FindByID(ctx, draftID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if draft == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "monster template suggestion draft not found"})
		return
	}
	if draft.MonsterTemplateID != nil && *draft.MonsterTemplateID != uuid.Nil {
		existing, findErr := s.dbClient.MonsterTemplate().FindByID(ctx, *draft.MonsterTemplateID)
		if findErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": findErr.Error()})
			return
		}
		ctx.JSON(http.StatusOK, existing)
		return
	}

	template := monsterTemplateFromSuggestionPayload(draft)
	if err := s.dbClient.MonsterTemplate().Create(ctx, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	convertedAt := time.Now().UTC()
	draft.Status = models.MonsterTemplateSuggestionDraftStatusConverted
	draft.MonsterTemplateID = &template.ID
	draft.ConvertedAt = &convertedAt
	draft.UpdatedAt = convertedAt
	if err := s.dbClient.MonsterTemplateSuggestionDraft().Update(ctx, draft); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created, err := s.dbClient.MonsterTemplate().FindByID(ctx, template.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, created)
}

func buildMonsterTemplateSuggestionPayload(
	spec jobs.MonsterTemplateCreationSpec,
	genre *models.ZoneGenre,
	zoneKinds []models.ZoneKind,
	preferredZoneKind *models.ZoneKind,
) models.MonsterTemplateSuggestionPayload {
	sanitized := sanitizeMonsterTemplateSpec(spec)
	payload := models.MonsterTemplateSuggestionPayload{
		MonsterType:      string(models.NormalizeMonsterTemplateType(sanitized.MonsterType)),
		ZoneKind:         normalizeMonsterTemplateSuggestionZoneKind(sanitized.ZoneKind, zoneKinds, preferredZoneKind),
		Name:             strings.TrimSpace(sanitized.Name),
		Description:      strings.TrimSpace(sanitized.Description),
		BaseStrength:     sanitized.BaseStrength,
		BaseDexterity:    sanitized.BaseDexterity,
		BaseConstitution: sanitized.BaseConstitution,
		BaseIntelligence: sanitized.BaseIntelligence,
		BaseWisdom:       sanitized.BaseWisdom,
		BaseCharisma:     sanitized.BaseCharisma,
	}
	if genre != nil {
		payload.GenreID = genre.ID
	}
	applyMonsterTemplateSuggestionAffinities(&payload)
	if payload.ZoneKind == "" {
		payload.ZoneKind = deriveMonsterTemplateSuggestionZoneKind(payload, zoneKinds)
	}
	return payload
}

func monsterTemplateFromSuggestionPayload(
	draft *models.MonsterTemplateSuggestionDraft,
) *models.MonsterTemplate {
	payload := models.MonsterTemplateSuggestionPayload(draft.Payload)
	monsterType := models.NormalizeMonsterTemplateType(payload.MonsterType)
	var genre *models.ZoneGenre
	if draft != nil && draft.MonsterType != "" {
		monsterType = models.NormalizeMonsterTemplateType(string(draft.MonsterType))
	}
	zoneKind := models.NormalizeZoneKind(payload.ZoneKind)
	if draft != nil && strings.TrimSpace(draft.ZoneKind) != "" {
		zoneKind = models.NormalizeZoneKind(draft.ZoneKind)
	}
	genreID := payload.GenreID
	if draft != nil && draft.GenreID != uuid.Nil {
		genreID = draft.GenreID
		genre = draft.Genre
	}
	return &models.MonsterTemplate{
		MonsterType:                   monsterType,
		ZoneKind:                      zoneKind,
		GenreID:                       genreID,
		Genre:                         genre,
		Name:                          strings.TrimSpace(payload.Name),
		Description:                   strings.TrimSpace(payload.Description),
		BaseStrength:                  payload.BaseStrength,
		BaseDexterity:                 payload.BaseDexterity,
		BaseConstitution:              payload.BaseConstitution,
		BaseIntelligence:              payload.BaseIntelligence,
		BaseWisdom:                    payload.BaseWisdom,
		BaseCharisma:                  payload.BaseCharisma,
		PhysicalDamageBonusPercent:    payload.PhysicalDamageBonusPercent,
		PiercingDamageBonusPercent:    payload.PiercingDamageBonusPercent,
		SlashingDamageBonusPercent:    payload.SlashingDamageBonusPercent,
		BludgeoningDamageBonusPercent: payload.BludgeoningDamageBonusPercent,
		FireDamageBonusPercent:        payload.FireDamageBonusPercent,
		IceDamageBonusPercent:         payload.IceDamageBonusPercent,
		LightningDamageBonusPercent:   payload.LightningDamageBonusPercent,
		PoisonDamageBonusPercent:      payload.PoisonDamageBonusPercent,
		ArcaneDamageBonusPercent:      payload.ArcaneDamageBonusPercent,
		HolyDamageBonusPercent:        payload.HolyDamageBonusPercent,
		ShadowDamageBonusPercent:      payload.ShadowDamageBonusPercent,
		PhysicalResistancePercent:     payload.PhysicalResistancePercent,
		PiercingResistancePercent:     payload.PiercingResistancePercent,
		SlashingResistancePercent:     payload.SlashingResistancePercent,
		BludgeoningResistancePercent:  payload.BludgeoningResistancePercent,
		FireResistancePercent:         payload.FireResistancePercent,
		IceResistancePercent:          payload.IceResistancePercent,
		LightningResistancePercent:    payload.LightningResistancePercent,
		PoisonResistancePercent:       payload.PoisonResistancePercent,
		ArcaneResistancePercent:       payload.ArcaneResistancePercent,
		HolyResistancePercent:         payload.HolyResistancePercent,
		ShadowResistancePercent:       payload.ShadowResistancePercent,
		ImageGenerationStatus:         models.MonsterTemplateImageGenerationStatusNone,
	}
}

func normalizeMonsterTemplateSuggestionZoneKind(
	raw string,
	zoneKinds []models.ZoneKind,
	preferred *models.ZoneKind,
) string {
	normalized := models.NormalizeZoneKind(raw)
	if normalized != "" {
		for _, zoneKind := range zoneKinds {
			if models.NormalizeZoneKind(zoneKind.Slug) == normalized {
				return normalized
			}
		}
	}
	if preferred != nil {
		return models.NormalizeZoneKind(preferred.Slug)
	}
	if len(zoneKinds) == 1 {
		return models.NormalizeZoneKind(zoneKinds[0].Slug)
	}
	return ""
}

func deriveMonsterTemplateSuggestionZoneKind(
	payload models.MonsterTemplateSuggestionPayload,
	zoneKinds []models.ZoneKind,
) string {
	existing := models.NormalizeZoneKind(payload.ZoneKind)
	if existing != "" {
		return existing
	}
	if len(zoneKinds) == 0 {
		return ""
	}
	if len(zoneKinds) == 1 {
		return models.NormalizeZoneKind(zoneKinds[0].Slug)
	}

	text := strings.ToLower(strings.TrimSpace(payload.Name + " " + payload.Description))
	bestSlug := ""
	bestScore := 0
	for _, zoneKind := range zoneKinds {
		slug := models.NormalizeZoneKind(zoneKind.Slug)
		if slug == "" {
			continue
		}
		haystack := strings.ToLower(strings.TrimSpace(zoneKind.Name + " " + zoneKind.Description + " " + slug))
		score := 0
		for _, rule := range monsterTemplateDraftZoneKindCueRules {
			if !containsMonsterTemplateSuggestionKeyword(text, rule.monsterKeywords) {
				continue
			}
			if containsMonsterTemplateSuggestionKeyword(haystack, rule.zoneCues) {
				score += 3
			}
		}
		for _, token := range strings.FieldsFunc(text, func(char rune) bool {
			return (char < 'a' || char > 'z') && (char < '0' || char > '9')
		}) {
			if len(token) < 4 {
				continue
			}
			if strings.Contains(haystack, token) {
				score++
			}
		}
		if score > bestScore || (score == bestScore && score > 0 && slug < bestSlug) {
			bestScore = score
			bestSlug = slug
		}
	}
	return bestSlug
}

func applyMonsterTemplateSuggestionAffinities(
	payload *models.MonsterTemplateSuggestionPayload,
) {
	if payload == nil {
		return
	}

	text := strings.ToLower(strings.TrimSpace(payload.Name + " " + payload.Description))
	addDamage := func(kind string, value int) {
		switch kind {
		case "piercing":
			payload.PiercingDamageBonusPercent = clampMonsterTemplateSuggestionAffinity(value, -25, 60)
		case "slashing":
			payload.SlashingDamageBonusPercent = clampMonsterTemplateSuggestionAffinity(value, -25, 60)
		case "bludgeoning":
			payload.BludgeoningDamageBonusPercent = clampMonsterTemplateSuggestionAffinity(value, -25, 60)
		case "fire":
			payload.FireDamageBonusPercent = clampMonsterTemplateSuggestionAffinity(value, -25, 60)
		case "ice":
			payload.IceDamageBonusPercent = clampMonsterTemplateSuggestionAffinity(value, -25, 60)
		case "lightning":
			payload.LightningDamageBonusPercent = clampMonsterTemplateSuggestionAffinity(value, -25, 60)
		case "poison":
			payload.PoisonDamageBonusPercent = clampMonsterTemplateSuggestionAffinity(value, -25, 60)
		case "arcane":
			payload.ArcaneDamageBonusPercent = clampMonsterTemplateSuggestionAffinity(value, -25, 60)
		case "holy":
			payload.HolyDamageBonusPercent = clampMonsterTemplateSuggestionAffinity(value, -25, 60)
		case "shadow":
			payload.ShadowDamageBonusPercent = clampMonsterTemplateSuggestionAffinity(value, -25, 60)
		default:
			payload.PhysicalDamageBonusPercent = clampMonsterTemplateSuggestionAffinity(value, -25, 60)
		}
	}
	addResistance := func(kind string, value int) {
		switch kind {
		case "piercing":
			payload.PiercingResistancePercent = clampMonsterTemplateSuggestionAffinity(value, -50, 60)
		case "slashing":
			payload.SlashingResistancePercent = clampMonsterTemplateSuggestionAffinity(value, -50, 60)
		case "bludgeoning":
			payload.BludgeoningResistancePercent = clampMonsterTemplateSuggestionAffinity(value, -50, 60)
		case "fire":
			payload.FireResistancePercent = clampMonsterTemplateSuggestionAffinity(value, -50, 60)
		case "ice":
			payload.IceResistancePercent = clampMonsterTemplateSuggestionAffinity(value, -50, 60)
		case "lightning":
			payload.LightningResistancePercent = clampMonsterTemplateSuggestionAffinity(value, -50, 60)
		case "poison":
			payload.PoisonResistancePercent = clampMonsterTemplateSuggestionAffinity(value, -50, 60)
		case "arcane":
			payload.ArcaneResistancePercent = clampMonsterTemplateSuggestionAffinity(value, -50, 60)
		case "holy":
			payload.HolyResistancePercent = clampMonsterTemplateSuggestionAffinity(value, -50, 60)
		case "shadow":
			payload.ShadowResistancePercent = clampMonsterTemplateSuggestionAffinity(value, -50, 60)
		default:
			payload.PhysicalResistancePercent = clampMonsterTemplateSuggestionAffinity(value, -50, 60)
		}
	}

	type keywordRule struct {
		keywords      []string
		damage        []string
		resist        []string
		vulnerableTo  []string
		damageDelta   int
		resistDelta   int
		vulnerablePct int
	}

	rules := []keywordRule{
		{keywords: []string{"fire", "ember", "cinder", "flame", "inferno", "lava", "ash"}, damage: []string{"fire"}, resist: []string{"fire"}, vulnerableTo: []string{"ice"}, damageDelta: 40, resistDelta: 50, vulnerablePct: -25},
		{keywords: []string{"frost", "ice", "glacier", "winter", "rime"}, damage: []string{"ice"}, resist: []string{"ice"}, vulnerableTo: []string{"fire"}, damageDelta: 40, resistDelta: 50, vulnerablePct: -25},
		{keywords: []string{"storm", "lightning", "thunder", "spark", "volt"}, damage: []string{"lightning"}, resist: []string{"lightning"}, damageDelta: 40, resistDelta: 45},
		{keywords: []string{"venom", "poison", "toxic", "corrosive", "ooze", "acid"}, damage: []string{"poison"}, resist: []string{"poison"}, vulnerableTo: []string{"fire"}, damageDelta: 35, resistDelta: 50, vulnerablePct: -15},
		{keywords: []string{"arcane", "mage", "sorcer", "wizard", "psionic", "mind", "eldritch"}, damage: []string{"arcane"}, resist: []string{"arcane"}, damageDelta: 35, resistDelta: 30},
		{keywords: []string{"holy", "radiant", "sun", "angel", "saint"}, damage: []string{"holy"}, resist: []string{"holy"}, vulnerableTo: []string{"shadow"}, damageDelta: 35, resistDelta: 35, vulnerablePct: -20},
		{keywords: []string{"shadow", "umbral", "void", "necrot", "undead", "wraith", "ghost", "specter", "spectre", "shade"}, damage: []string{"shadow"}, resist: []string{"shadow"}, vulnerableTo: []string{"holy"}, damageDelta: 35, resistDelta: 45, vulnerablePct: -30},
		{keywords: []string{"skeleton", "zombie", "bone", "grave"}, resist: []string{"piercing"}, vulnerableTo: []string{"bludgeoning"}, resistDelta: 20, vulnerablePct: -25},
		{keywords: []string{"stone", "iron", "armored", "armoured", "golem"}, resist: []string{"slashing", "piercing"}, vulnerableTo: []string{"bludgeoning"}, resistDelta: 20, vulnerablePct: -20},
	}

	for _, rule := range rules {
		if !containsMonsterTemplateSuggestionKeyword(text, rule.keywords) {
			continue
		}
		for _, affinity := range rule.damage {
			addDamage(affinity, rule.damageDelta)
		}
		for _, affinity := range rule.resist {
			addResistance(affinity, rule.resistDelta)
		}
		for _, affinity := range rule.vulnerableTo {
			addResistance(affinity, rule.vulnerablePct)
		}
	}

	physicalScore := payload.BaseStrength + payload.BaseDexterity + payload.BaseConstitution
	mentalScore := payload.BaseIntelligence + payload.BaseWisdom + payload.BaseCharisma
	if physicalScore >= mentalScore+6 {
		if payload.BaseDexterity >= payload.BaseStrength+3 {
			addDamage("piercing", 25)
			addResistance("piercing", 10)
		} else if payload.BaseStrength >= payload.BaseDexterity+3 {
			addDamage("bludgeoning", 25)
			addResistance("physical", 10)
		} else {
			addDamage("slashing", 20)
			addResistance("physical", 10)
		}
	}
	if mentalScore >= physicalScore+6 &&
		payload.FireDamageBonusPercent <= 0 &&
		payload.IceDamageBonusPercent <= 0 &&
		payload.LightningDamageBonusPercent <= 0 &&
		payload.PoisonDamageBonusPercent <= 0 &&
		payload.ArcaneDamageBonusPercent <= 0 &&
		payload.HolyDamageBonusPercent <= 0 &&
		payload.ShadowDamageBonusPercent <= 0 {
		addDamage("arcane", 25)
		addResistance("arcane", 15)
	}
	if payload.BaseConstitution >= 15 {
		addResistance("physical", 10)
	}
}

func containsMonsterTemplateSuggestionKeyword(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func clampMonsterTemplateSuggestionAffinity(value int, minValue int, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
