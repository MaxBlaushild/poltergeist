package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	"github.com/gin-gonic/gin"
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

	r.POST("/trivai/register", s.register)
	r.POST("/trivai/login", s.login)

	r.Run(":8042")
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
	// 	Body:     "Welcome to Guess How Many! New question every day at noon EST. Text CANCEL at any point to cancel your subscription.",
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
