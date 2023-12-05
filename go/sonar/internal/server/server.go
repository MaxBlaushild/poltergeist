package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/middleware"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type server struct {
	authClient   auth.Client
	texterClient texter.Client
	dbClient     db.DbClient
}

type Server interface {
	ListenAndServe(port string)
}

func NewServer(authClient auth.Client, texterClient texter.Client, dbClient db.DbClient) Server {
	return &server{authClient: authClient, texterClient: texterClient, dbClient: dbClient}
}

func (s *server) ListenAndServe(port string) {
	r := gin.Default()

	r.POST("/sonar/register", s.register)
	r.POST("/sonar/login", s.login)
	r.GET("/sonar/surveys", middleware.WithAuthentication(s.authClient, s.getSurverys))

	r.Run(":8042")
}

func (s *server) getSurverys(ctx *gin.Context) {
	user, ok := ctx.Get("user")

	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "no user found in request",
		})
		return
	}

	surveys, err := s.dbClient.SonarSurvey().GetSurveys(ctx, user.(*models.User).ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.Wrap(err, "survey fetch error").Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, surveys)
}

func (s *server) login(ctx *gin.Context) {
	var requestBody auth.LoginByTextRequest

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	authenticateResponse, err := s.authClient.LoginByText(ctx, &requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	payload := gin.H{
		"user":  authenticateResponse.User,
		"token": authenticateResponse.Token,
	}

	ctx.JSON(200, payload)
}

func (s *server) register(ctx *gin.Context) {
	var requestBody auth.RegisterByTextRequest

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	authenticateResponse, err := s.authClient.RegisterByText(ctx, &requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// if err := s.texterClient.Text(ctx, &texter.Text{
	// 	Body:     "Welcome to Guess How Many! New question every day at noon EST.",
	// 	From:     s.cfg.Secret.GuessHowManyPhoneNumber,
	// 	To:       authenticateResponse.User.PhoneNumber,
	// 	TextType: "guess-how-many-welcome-email",
	// }); err != nil {
	// 	ctx.JSON(http.StatusInternalServerError, gin.H{
	// 		"error": err.Error(),
	// 	})
	// }

	ctx.JSON(200, gin.H{
		"user":  authenticateResponse.User,
		"token": authenticateResponse.Token,
	})
}
