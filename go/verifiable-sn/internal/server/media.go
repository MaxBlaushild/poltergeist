package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *server) GetPresignedUploadUrl(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var requestBody struct {
		Bucket string `json:"bucket" binding:"required"`
		Key    string `json:"key" binding:"required"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	url, err := s.awsClient.GeneratePresignedUploadURL(requestBody.Bucket, requestBody.Key, time.Hour)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Prevent unused variable warning
	_ = user

	ctx.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}

