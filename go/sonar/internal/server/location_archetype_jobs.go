package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type generateLocationArchetypesRequest struct {
	Count int    `json:"count"`
	Salt  string `json:"salt"`
}

func (s *server) generateLocationArchetypes(ctx *gin.Context) {
	var requestBody generateLocationArchetypesRequest
	if err := ctx.ShouldBindJSON(&requestBody); err != nil && !errors.Is(err, io.EOF) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	count := requestBody.Count
	if count <= 0 {
		count = 50
	}
	if count > 100 {
		count = 100
	}

	salt := strings.TrimSpace(requestBody.Salt)
	if salt == "" {
		salt = uuid.NewString()
	}

	payload, err := json.Marshal(jobs.GenerateLocationArchetypesTaskPayload{
		Count: count,
		Salt:  salt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateLocationArchetypesTaskType, payload)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"queued": true,
		"count":  count,
		"salt":   salt,
	})
}
