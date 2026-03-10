package processors

import (
	"context"
	"log"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const standaloneRecurringBatchSize = 50

type ProcessRecurringStandaloneContentProcessor struct {
	dbClient db.DbClient
}

func NewProcessRecurringStandaloneContentProcessor(
	dbClient db.DbClient,
) ProcessRecurringStandaloneContentProcessor {
	log.Println("Initializing ProcessRecurringStandaloneContentProcessor")
	return ProcessRecurringStandaloneContentProcessor{dbClient: dbClient}
}

func (p *ProcessRecurringStandaloneContentProcessor) ProcessTask(
	ctx context.Context,
	task *asynq.Task,
) error {
	log.Printf("Processing recurring standalone content task: %v", task.Type())

	now := time.Now()
	if err := p.processScenarios(ctx, now); err != nil {
		return err
	}
	if err := p.processChallenges(ctx, now); err != nil {
		return err
	}
	if err := p.processMonsterEncounters(ctx, now); err != nil {
		return err
	}

	return nil
}

func (p *ProcessRecurringStandaloneContentProcessor) processScenarios(
	ctx context.Context,
	now time.Time,
) error {
	due, err := p.dbClient.Scenario().FindDueRecurring(ctx, now, standaloneRecurringBatchSize)
	if err != nil {
		log.Printf("Failed to find recurring scenarios: %v", err)
		return err
	}
	if len(due) == 0 {
		return nil
	}

	log.Printf("Found %d recurring scenarios due", len(due))
	for _, scenario := range due {
		if err := p.processScenario(ctx, scenario.ID, now); err != nil {
			log.Printf("Failed to process recurring scenario %s: %v", scenario.ID, err)
		}
	}
	return nil
}

func (p *ProcessRecurringStandaloneContentProcessor) processChallenges(
	ctx context.Context,
	now time.Time,
) error {
	due, err := p.dbClient.Challenge().FindDueRecurring(ctx, now, standaloneRecurringBatchSize)
	if err != nil {
		log.Printf("Failed to find recurring challenges: %v", err)
		return err
	}
	if len(due) == 0 {
		return nil
	}

	log.Printf("Found %d recurring challenges due", len(due))
	for _, challenge := range due {
		if err := p.processChallenge(ctx, challenge.ID, now); err != nil {
			log.Printf("Failed to process recurring challenge %s: %v", challenge.ID, err)
		}
	}
	return nil
}

func (p *ProcessRecurringStandaloneContentProcessor) processMonsterEncounters(
	ctx context.Context,
	now time.Time,
) error {
	due, err := p.dbClient.MonsterEncounter().FindDueRecurring(ctx, now, standaloneRecurringBatchSize)
	if err != nil {
		log.Printf("Failed to find recurring monster encounters: %v", err)
		return err
	}
	if len(due) == 0 {
		return nil
	}

	log.Printf("Found %d recurring monster encounters due", len(due))
	for _, encounter := range due {
		if err := p.processMonsterEncounter(ctx, encounter.ID, now); err != nil {
			log.Printf("Failed to process recurring monster encounter %s: %v", encounter.ID, err)
		}
	}
	return nil
}

func (p *ProcessRecurringStandaloneContentProcessor) processScenario(
	ctx context.Context,
	scenarioID uuid.UUID,
	now time.Time,
) error {
	scenario, err := p.dbClient.Scenario().FindByID(ctx, scenarioID)
	if err != nil {
		return err
	}
	if scenario == nil || scenario.RecurrenceFrequency == nil || scenario.NextRecurrenceAt == nil {
		return nil
	}
	if scenario.NextRecurrenceAt.After(now) {
		return nil
	}

	frequency := models.NormalizeQuestRecurrenceFrequency(*scenario.RecurrenceFrequency)
	if !models.IsValidQuestRecurrenceFrequency(frequency) {
		log.Printf("Recurring scenario %s has invalid frequency %q", scenario.ID, frequency)
		return nil
	}

	recurringID := scenario.RecurringScenarioID
	if recurringID == nil {
		newID := uuid.New()
		recurringID = &newID
	}

	nextAt, ok := advanceStandaloneRecurrence(now, scenario.NextRecurrenceAt, frequency)
	if !ok {
		log.Printf("Recurring scenario %s has unsupported frequency %q", scenario.ID, frequency)
		return nil
	}

	newScenario := &models.Scenario{
		ID:                        uuid.New(),
		CreatedAt:                 now,
		UpdatedAt:                 now,
		ZoneID:                    scenario.ZoneID,
		PointOfInterestID:         scenario.PointOfInterestID,
		Latitude:                  scenario.Latitude,
		Longitude:                 scenario.Longitude,
		Prompt:                    scenario.Prompt,
		ImageURL:                  scenario.ImageURL,
		ThumbnailURL:              scenario.ThumbnailURL,
		ScaleWithUserLevel:        scenario.ScaleWithUserLevel,
		RecurringScenarioID:       recurringID,
		RecurrenceFrequency:       &frequency,
		NextRecurrenceAt:          &nextAt,
		RewardMode:                scenario.RewardMode,
		RandomRewardSize:          scenario.RandomRewardSize,
		Difficulty:                scenario.Difficulty,
		RewardExperience:          scenario.RewardExperience,
		RewardGold:                scenario.RewardGold,
		OpenEnded:                 scenario.OpenEnded,
		FailurePenaltyMode:        scenario.FailurePenaltyMode,
		FailureHealthDrainType:    scenario.FailureHealthDrainType,
		FailureHealthDrainValue:   scenario.FailureHealthDrainValue,
		FailureManaDrainType:      scenario.FailureManaDrainType,
		FailureManaDrainValue:     scenario.FailureManaDrainValue,
		FailureStatuses:           scenario.FailureStatuses,
		SuccessRewardMode:         scenario.SuccessRewardMode,
		SuccessHealthRestoreType:  scenario.SuccessHealthRestoreType,
		SuccessHealthRestoreValue: scenario.SuccessHealthRestoreValue,
		SuccessManaRestoreType:    scenario.SuccessManaRestoreType,
		SuccessManaRestoreValue:   scenario.SuccessManaRestoreValue,
		SuccessStatuses:           scenario.SuccessStatuses,
	}

	if err := p.dbClient.Scenario().Create(ctx, newScenario); err != nil {
		return err
	}
	if err := p.dbClient.Scenario().ReplaceOptions(ctx, newScenario.ID, scenario.Options); err != nil {
		return err
	}
	if err := p.dbClient.Scenario().ReplaceItemRewards(ctx, newScenario.ID, scenario.ItemRewards); err != nil {
		return err
	}
	if err := p.dbClient.Scenario().ReplaceSpellRewards(ctx, newScenario.ID, scenario.SpellRewards); err != nil {
		return err
	}

	scenario.RecurringScenarioID = recurringID
	scenario.RecurrenceFrequency = nil
	scenario.NextRecurrenceAt = nil
	scenario.RetiredAt = &now
	scenario.UpdatedAt = now
	if err := p.dbClient.Scenario().Update(ctx, scenario.ID, scenario); err != nil {
		log.Printf("Failed to clear recurrence fields for scenario %s: %v", scenario.ID, err)
	}

	log.Printf("Recurring scenario %s recreated as %s", scenario.ID, newScenario.ID)
	return nil
}

func (p *ProcessRecurringStandaloneContentProcessor) processChallenge(
	ctx context.Context,
	challengeID uuid.UUID,
	now time.Time,
) error {
	challenge, err := p.dbClient.Challenge().FindByID(ctx, challengeID)
	if err != nil {
		return err
	}
	if challenge == nil || challenge.RecurrenceFrequency == nil || challenge.NextRecurrenceAt == nil {
		return nil
	}
	if challenge.NextRecurrenceAt.After(now) {
		return nil
	}

	frequency := models.NormalizeQuestRecurrenceFrequency(*challenge.RecurrenceFrequency)
	if !models.IsValidQuestRecurrenceFrequency(frequency) {
		log.Printf("Recurring challenge %s has invalid frequency %q", challenge.ID, frequency)
		return nil
	}

	recurringID := challenge.RecurringChallengeID
	if recurringID == nil {
		newID := uuid.New()
		recurringID = &newID
	}

	nextAt, ok := advanceStandaloneRecurrence(now, challenge.NextRecurrenceAt, frequency)
	if !ok {
		log.Printf("Recurring challenge %s has unsupported frequency %q", challenge.ID, frequency)
		return nil
	}

	newChallenge := &models.Challenge{
		ID:                   uuid.New(),
		CreatedAt:            now,
		UpdatedAt:            now,
		ZoneID:               challenge.ZoneID,
		PointOfInterestID:    challenge.PointOfInterestID,
		Latitude:             challenge.Latitude,
		Longitude:            challenge.Longitude,
		Question:             challenge.Question,
		Description:          challenge.Description,
		ImageURL:             challenge.ImageURL,
		ThumbnailURL:         challenge.ThumbnailURL,
		ScaleWithUserLevel:   challenge.ScaleWithUserLevel,
		RecurringChallengeID: recurringID,
		RecurrenceFrequency:  &frequency,
		NextRecurrenceAt:     &nextAt,
		RewardMode:           challenge.RewardMode,
		RandomRewardSize:     challenge.RandomRewardSize,
		RewardExperience:     challenge.RewardExperience,
		Reward:               challenge.Reward,
		InventoryItemID:      challenge.InventoryItemID,
		SubmissionType:       challenge.SubmissionType,
		Difficulty:           challenge.Difficulty,
		StatTags:             challenge.StatTags,
		Proficiency:          challenge.Proficiency,
	}
	if err := p.dbClient.Challenge().Create(ctx, newChallenge); err != nil {
		return err
	}

	challenge.RecurringChallengeID = recurringID
	challenge.RecurrenceFrequency = nil
	challenge.NextRecurrenceAt = nil
	challenge.RetiredAt = &now
	challenge.UpdatedAt = now
	if err := p.dbClient.Challenge().Update(ctx, challenge.ID, challenge); err != nil {
		log.Printf("Failed to clear recurrence fields for challenge %s: %v", challenge.ID, err)
	}

	log.Printf("Recurring challenge %s recreated as %s", challenge.ID, newChallenge.ID)
	return nil
}

func (p *ProcessRecurringStandaloneContentProcessor) processMonsterEncounter(
	ctx context.Context,
	encounterID uuid.UUID,
	now time.Time,
) error {
	encounter, err := p.dbClient.MonsterEncounter().FindByID(ctx, encounterID)
	if err != nil {
		return err
	}
	if encounter == nil || encounter.RecurrenceFrequency == nil || encounter.NextRecurrenceAt == nil {
		return nil
	}
	if encounter.NextRecurrenceAt.After(now) {
		return nil
	}

	frequency := models.NormalizeQuestRecurrenceFrequency(*encounter.RecurrenceFrequency)
	if !models.IsValidQuestRecurrenceFrequency(frequency) {
		log.Printf("Recurring monster encounter %s has invalid frequency %q", encounter.ID, frequency)
		return nil
	}

	recurringID := encounter.RecurringMonsterEncounterID
	if recurringID == nil {
		newID := uuid.New()
		recurringID = &newID
	}

	nextAt, ok := advanceStandaloneRecurrence(now, encounter.NextRecurrenceAt, frequency)
	if !ok {
		log.Printf("Recurring monster encounter %s has unsupported frequency %q", encounter.ID, frequency)
		return nil
	}

	newEncounter := &models.MonsterEncounter{
		ID:                          uuid.New(),
		CreatedAt:                   now,
		UpdatedAt:                   now,
		Name:                        encounter.Name,
		Description:                 encounter.Description,
		ImageURL:                    encounter.ImageURL,
		ThumbnailURL:                encounter.ThumbnailURL,
		ScaleWithUserLevel:          encounter.ScaleWithUserLevel,
		RecurringMonsterEncounterID: recurringID,
		RecurrenceFrequency:         &frequency,
		NextRecurrenceAt:            &nextAt,
		ZoneID:                      encounter.ZoneID,
		Latitude:                    encounter.Latitude,
		Longitude:                   encounter.Longitude,
	}
	if err := p.dbClient.MonsterEncounter().Create(ctx, newEncounter); err != nil {
		return err
	}

	members := make([]models.MonsterEncounterMember, 0, len(encounter.Members))
	for _, member := range encounter.Members {
		if member.MonsterID == uuid.Nil {
			continue
		}
		members = append(members, models.MonsterEncounterMember{
			MonsterID: member.MonsterID,
			Slot:      member.Slot,
		})
	}
	if err := p.dbClient.MonsterEncounter().ReplaceMembers(ctx, newEncounter.ID, members); err != nil {
		return err
	}

	encounter.RecurringMonsterEncounterID = recurringID
	encounter.RecurrenceFrequency = nil
	encounter.NextRecurrenceAt = nil
	encounter.RetiredAt = &now
	encounter.UpdatedAt = now
	if err := p.dbClient.MonsterEncounter().Update(ctx, encounter.ID, encounter); err != nil {
		log.Printf("Failed to clear recurrence fields for monster encounter %s: %v", encounter.ID, err)
	}

	log.Printf("Recurring monster encounter %s recreated as %s", encounter.ID, newEncounter.ID)
	return nil
}

func advanceStandaloneRecurrence(
	now time.Time,
	scheduled *time.Time,
	frequency string,
) (time.Time, bool) {
	base := now
	if scheduled != nil && !scheduled.IsZero() {
		base = *scheduled
	}
	next, ok := models.NextQuestRecurrenceAt(base, frequency)
	if !ok {
		return time.Time{}, false
	}
	for !next.After(now) {
		next, ok = models.NextQuestRecurrenceAt(next, frequency)
		if !ok {
			return time.Time{}, false
		}
	}
	return next, true
}
