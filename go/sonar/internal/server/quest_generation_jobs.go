package server

import (
	"context"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *server) hydrateQuestGenerationJobs(
	ctx context.Context,
	jobsList []*models.QuestGenerationJob,
) error {
	questIDs := make([]uuid.UUID, 0)
	seen := map[uuid.UUID]struct{}{}
	for _, job := range jobsList {
		if job == nil {
			continue
		}
		for _, idStr := range job.QuestIDs {
			questID, err := uuid.Parse(idStr)
			if err != nil {
				continue
			}
			if _, ok := seen[questID]; ok {
				continue
			}
			seen[questID] = struct{}{}
			questIDs = append(questIDs, questID)
		}
	}

	questsByID := map[uuid.UUID]models.Quest{}
	if len(questIDs) > 0 {
		quests, err := s.dbClient.Quest().FindByIDs(ctx, questIDs)
		if err != nil {
			return err
		}
		for _, quest := range quests {
			questsByID[quest.ID] = quest
		}
	}

	for _, job := range jobsList {
		if job == nil || len(job.QuestIDs) == 0 {
			continue
		}
		job.Quests = make([]models.Quest, 0, len(job.QuestIDs))
		for _, idStr := range job.QuestIDs {
			questID, err := uuid.Parse(idStr)
			if err != nil {
				continue
			}
			if quest, ok := questsByID[questID]; ok {
				job.Quests = append(job.Quests, quest)
			}
		}
	}

	return nil
}

func (s *server) getQuestGenerationJob(ctx *gin.Context) {
	id := ctx.Param("id")
	jobID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest generation job ID"})
		return
	}

	job, err := s.dbClient.QuestGenerationJob().FindByID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "quest generation job not found"})
		return
	}

	if err := s.hydrateQuestGenerationJobs(ctx, []*models.QuestGenerationJob{job}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, job)
}
