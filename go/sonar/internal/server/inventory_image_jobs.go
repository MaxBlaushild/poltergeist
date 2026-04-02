package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

type queueMissingInventoryItemImagesResponse struct {
	TotalCount   int      `json:"totalCount"`
	MissingCount int      `json:"missingCount"`
	QueuedCount  int      `json:"queuedCount"`
	SkippedCount int      `json:"skippedCount"`
	FailedCount  int      `json:"failedCount"`
	Failures     []string `json:"failures,omitempty"`
}

func (s *server) queueMissingInventoryItemImages(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	items, err := s.dbClient.InventoryItem().FindAllInventoryItems(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := queueMissingInventoryItemImagesResponse{
		TotalCount: len(items),
		Failures:   make([]string, 0),
	}

	for _, item := range items {
		if strings.TrimSpace(item.ImageURL) != "" {
			response.SkippedCount++
			continue
		}

		response.MissingCount++
		status := strings.TrimSpace(strings.ToLower(item.ImageGenerationStatus))
		if status == models.InventoryImageGenerationStatusQueued || status == models.InventoryImageGenerationStatusInProgress {
			response.SkippedCount++
			continue
		}

		if err := s.dbClient.InventoryItem().UpdateInventoryItem(ctx, item.ID, map[string]interface{}{
			"image_generation_status": models.InventoryImageGenerationStatusQueued,
			"image_generation_error":  "",
		}); err != nil {
			response.FailedCount++
			response.Failures = append(response.Failures, fmt.Sprintf("%s (ID %d): failed to update inventory item: %v", item.Name, item.ID, err))
			continue
		}

		if err := s.enqueueInventoryItemImageGeneration(ctx, item.ID, item.Name, item.FlavorText, item.RarityTier); err != nil {
			response.FailedCount++
			response.Failures = append(response.Failures, fmt.Sprintf("%s (ID %d): %v", item.Name, item.ID, err))
			continue
		}
		response.QueuedCount++
	}

	ctx.JSON(http.StatusOK, response)
}
