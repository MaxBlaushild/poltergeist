package main

import (
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

	router.GET("/", func(c *gin.Context) {
		c.String(200, "Goodbye, World!")
	})

	router.POST("/consult", func(ctx *gin.Context) {
		var consultQuestion deep_priest.QuestionWithImage

		if err := ctx.Bind(&consultQuestion); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "question must be included",
			})
			return
		}

		answer, err := openApiClient.GetAnswer(ctx, consultQuestion.Question)
		if err != nil {
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
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"answer": answer})
	})

	router.Run(":8081")
}
