package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MaxBlaushild/fount-of-erebos/internal/config"
	"github.com/MaxBlaushild/fount-of-erebos/internal/open_ai"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	cfg, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		panic(err)
	}

	openApiClient := open_ai.NewClient(open_ai.ClientConfig{
		ApiKey: cfg.Secret.OpenAIKey,
	})

	// grokClient := grok.NewClient(grok.ClientConfig{
	// 	ApiKey: cfg.Secret.GrokApiKey,
	// })

	router.GET("/", func(c *gin.Context) {
		c.String(200, "Goodbye, World!")
	})

	router.POST("/consult", func(ctx *gin.Context) {
		var consultQuestion deep_priest.Question

		if err := ctx.Bind(&consultQuestion); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "question must be included",
			})
			return
		}

		answer, err := openApiClient.GetAnswer(ctx, consultQuestion.Question)
		if err != nil {
			fmt.Println(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "something went wrong",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"answer": answer})
	})

	router.POST("/consultWithImage", func(ctx *gin.Context) {
		var judgeSubmission deep_priest.QuestionWithImage

		if err := ctx.Bind(&judgeSubmission); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "question must be included"})
			return
		}

		answer, err := openApiClient.GetAnswerWithImage(ctx, judgeSubmission.Question, judgeSubmission.Image)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"answer": answer})
	})

	router.POST("/generateImage", func(ctx *gin.Context) {
		log.Println("Received request to generate image")
		var generateImageRequest deep_priest.GenerateImageRequest

		if err := ctx.Bind(&generateImageRequest); err != nil {
			log.Printf("Error binding request: %v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "question must be included"})
			return
		}

		log.Printf("Generating image with Grok. Prompt: %s", generateImageRequest.Prompt)
		imageUrl, err := openApiClient.GenerateImage(ctx, generateImageRequest)
		if err != nil {
			log.Printf("Error generating image: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"imageUrl": imageUrl})
	})

	router.POST("/editImage", func(ctx *gin.Context) {
		var editImageRequest deep_priest.EditImageRequest

		if err := ctx.Bind(&editImageRequest); err != nil {
			log.Printf("Error binding request: %v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "question must be included"})
			return
		}

		imageUrl, err := openApiClient.EditImage(ctx, editImageRequest)
		if err != nil {
			log.Printf("Error editing image: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"imageUrl": imageUrl})
	})

	router.Run(":8081")
}
