package processors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type ShuffleZoneSeedChallengeProcessor struct {
	dbClient db.DbClient
}

func NewShuffleZoneSeedChallengeProcessor(dbClient db.DbClient) ShuffleZoneSeedChallengeProcessor {
	log.Println("Initializing ShuffleZoneSeedChallengeProcessor")
	return ShuffleZoneSeedChallengeProcessor{
		dbClient: dbClient,
	}
}

func (p *ShuffleZoneSeedChallengeProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing shuffle zone seed challenge task: %v", task.Type())

	var payload jobs.ShuffleZoneSeedChallengeTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if payload.QuestDraftID == nil && payload.MainQuestNodeDraftID == nil {
		return fmt.Errorf("missing shuffle target")
	}
	if payload.QuestDraftID != nil && payload.MainQuestNodeDraftID != nil {
		return fmt.Errorf("shuffle target must be either questDraftId or mainQuestNodeDraftId")
	}

	job, err := p.dbClient.ZoneSeedJob().FindByID(ctx, payload.JobID)
	if err != nil {
		return err
	}
	if job == nil {
		return nil
	}

	if job.Status == models.ZoneSeedStatusInProgress || job.Status == models.ZoneSeedStatusApplying || job.Status == models.ZoneSeedStatusApplied {
		return nil
	}

	if !hasZoneSeedDraftContent(job.Draft) {
		return fmt.Errorf("zone seed draft is empty")
	}

	if !setShuffleStatusInDraft(&job.Draft, payload, models.ZoneSeedChallengeShuffleStatusInProgress, nil) {
		return fmt.Errorf("shuffle target not found in zone seed draft")
	}
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ZoneSeedJob().Update(ctx, job); err != nil {
		return err
	}

	placeByID := make(map[string]models.ZoneSeedPointOfInterestDraft, len(job.Draft.PointsOfInterest))
	for _, poi := range job.Draft.PointsOfInterest {
		placeID := strings.TrimSpace(poi.PlaceID)
		if placeID == "" {
			continue
		}
		placeByID[placeID] = poi
	}

	changed := false
	if payload.QuestDraftID != nil {
		changed = shuffleQuestDraftChallenge(job, placeByID, *payload.QuestDraftID)
	}
	if payload.MainQuestNodeDraftID != nil {
		changed = shuffleMainQuestNodeChallenge(job, placeByID, *payload.MainQuestNodeDraftID)
	}
	if !changed {
		msg := "shuffle target not found in zone seed draft"
		_ = setShuffleStatusInDraft(&job.Draft, payload, models.ZoneSeedChallengeShuffleStatusFailed, &msg)
		job.UpdatedAt = time.Now()
		_ = p.dbClient.ZoneSeedJob().Update(ctx, job)
		return errors.New(msg)
	}

	_ = setShuffleStatusInDraft(&job.Draft, payload, models.ZoneSeedChallengeShuffleStatusCompleted, nil)
	if job.Status == models.ZoneSeedStatusFailed {
		job.Status = models.ZoneSeedStatusAwaitingApproval
	}
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	return p.dbClient.ZoneSeedJob().Update(ctx, job)
}

func setShuffleStatusInDraft(
	draft *models.ZoneSeedDraft,
	payload jobs.ShuffleZoneSeedChallengeTaskPayload,
	status string,
	errMsg *string,
) bool {
	if payload.QuestDraftID != nil {
		return draft.SetQuestChallengeShuffleStatus(*payload.QuestDraftID, status, errMsg)
	}
	if payload.MainQuestNodeDraftID != nil {
		return draft.SetMainQuestNodeChallengeShuffleStatus(*payload.MainQuestNodeDraftID, status, errMsg)
	}
	return false
}

func shuffleQuestDraftChallenge(
	job *models.ZoneSeedJob,
	placeByID map[string]models.ZoneSeedPointOfInterestDraft,
	questDraftID uuid.UUID,
) bool {
	for idx := range job.Draft.Quests {
		if job.Draft.Quests[idx].DraftID != questDraftID {
			continue
		}
		place := resolveDraftChallengePlace(job.Draft.Quests[idx].PlaceID, placeByID)
		challenge := regenerateZoneSeedChallengeMetadata(zoneSeedDraftPOIToGooglePlace(place))
		job.Draft.Quests[idx].ChallengeQuestion = challenge.Question
		job.Draft.Quests[idx].ChallengeDifficulty = challenge.Difficulty
		return true
	}
	return false
}

func shuffleMainQuestNodeChallenge(
	job *models.ZoneSeedJob,
	placeByID map[string]models.ZoneSeedPointOfInterestDraft,
	nodeDraftID uuid.UUID,
) bool {
	for mainIdx := range job.Draft.MainQuests {
		for nodeIdx := range job.Draft.MainQuests[mainIdx].Nodes {
			if job.Draft.MainQuests[mainIdx].Nodes[nodeIdx].DraftID != nodeDraftID {
				continue
			}
			place := resolveDraftChallengePlace(job.Draft.MainQuests[mainIdx].Nodes[nodeIdx].PlaceID, placeByID)
			challenge := regenerateZoneSeedChallengeMetadata(zoneSeedDraftPOIToGooglePlace(place))
			job.Draft.MainQuests[mainIdx].Nodes[nodeIdx].ChallengeQuestion = challenge.Question
			job.Draft.MainQuests[mainIdx].Nodes[nodeIdx].ChallengeDifficulty = challenge.Difficulty
			return true
		}
	}
	return false
}

func resolveDraftChallengePlace(
	placeID string,
	placeByID map[string]models.ZoneSeedPointOfInterestDraft,
) (place models.ZoneSeedPointOfInterestDraft) {
	if poi, ok := placeByID[strings.TrimSpace(placeID)]; ok {
		return poi
	}
	place.PlaceID = strings.TrimSpace(placeID)
	place.Name = "this location"
	return place
}

func hasZoneSeedDraftContent(d models.ZoneSeedDraft) bool {
	if strings.TrimSpace(d.FantasyName) != "" {
		return true
	}
	if strings.TrimSpace(d.ZoneDescription) != "" {
		return true
	}
	if len(d.PointsOfInterest) > 0 {
		return true
	}
	if len(d.Characters) > 0 {
		return true
	}
	if len(d.Quests) > 0 {
		return true
	}
	if len(d.MainQuests) > 0 {
		return true
	}
	return false
}
