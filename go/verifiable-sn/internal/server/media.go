package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *server) GetPresignedUploadUrl(ctx *gin.Context) {
	_, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var requestBody struct {
		Bucket      string `json:"bucket" binding:"required"`
		Key         string `json:"key" binding:"required"`
		ContentType string `json:"contentType"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var url string
	if requestBody.ContentType != "" {
		url, err = s.awsClient.GeneratePresignedUploadURLWithContentType(requestBody.Bucket, requestBody.Key, requestBody.ContentType, time.Hour)
	} else {
		url, err = s.awsClient.GeneratePresignedUploadURL(requestBody.Bucket, requestBody.Key, time.Hour)
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}
