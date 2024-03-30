package server

import (
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/billing"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/email"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/MaxBlaushild/poltergeist/trivai/internal/config"
	"github.com/MaxBlaushild/poltergeist/trivai/internal/trivai"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Server struct {
	dbClient      db.DbClient
	emailClient   email.EmailClient
	triviaClient  trivai.TrivaiClient
	texterClient  texter.Client
	billingClient billing.Client
	cfg           config.Config
	authClient    auth.Client
}

func NewServer(
	dbClient db.DbClient,
	emailClient email.EmailClient,
	trivaiClient trivai.TrivaiClient,
	texterClient texter.Client,
	billingClient billing.Client,
	cfg config.Config,
	authClient auth.Client,
) Server {
	r := gin.Default()

	s := Server{
		dbClient:      dbClient,
		emailClient:   emailClient,
		triviaClient:  trivaiClient,
		texterClient:  texterClient,
		billingClient: billingClient,
		cfg:           cfg,
		authClient:    authClient,
	}

	r.GET("/trivai/subscriptions/:userID", s.getSubscription)
	r.GET("/trivai/users/:userId", s.getUser)
	r.GET("/trivai/users/:userId/subscribe", s.getSubscriptionLink)

	r.POST("/")
	r.POST("/trivai/receive-sms", s.receiveSms)
	r.POST("/trivai/how_many_questions/subscribe", s.subscribeToHowManyQuestions)
	r.GET("/trivai/how_many_questions/current", s.getCurrentQuestion)
	r.POST("/trivai/how_many_questions/grade", s.gradeQuestion)
	r.GET("/trivai/how_many_questions", s.getHowManyQuestions)
	r.GET("/trivai/how_many_questions/answer", s.getHowManyQuestionAnswer)
	r.POST("/trivai/how_many_questions", s.generateNewHowManyQuestion)
	r.POST("/trivai/how_many_questions/:id/validate", s.markHowManyQuestionValid)
	r.POST("/trivai/begin-checkout", s.beginCheckout)
	r.POST("/trivai/finish-checkout", s.finishCheckout)
	r.POST("/trivai/subscriptions/cancel", s.cancelSubscription)
	r.POST("/trivai/subscriptions/delete", s.deleteSubscription)
	r.POST("/trivai/register", s.register)
	r.POST("/trivai/login", s.login)

	r.Run(":8082")

	return s
}

func (s *Server) cancelSubscription(ctx *gin.Context) {
	var cancelSubscriptionRequest struct {
		UserID string `json:"userId" binding:"required"`
	}

	if err := ctx.Bind(&cancelSubscriptionRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(cancelSubscriptionRequest.UserID)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	howManySubscription, err := s.dbClient.HowManySubscription().FindByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "no subscription found for user id",
		})
		return
	}

	if howManySubscription.StripeID == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "subscription still in trial mode",
		})
		return
	}

	if _, err := s.billingClient.CancelSubscription(ctx, &billing.CancelSubscriptionParams{
		StripeID: *howManySubscription.StripeID,
	}); err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "subscription cancellation in progress",
	})
}

func (s *Server) deleteSubscription(ctx *gin.Context) {
	var onUnsubscribe billing.OnSubscriptionDelete

	if err := ctx.Bind(&onUnsubscribe); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := s.dbClient.HowManySubscription().DeleteByStripeID(ctx, onUnsubscribe.SubscriptionID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "cool beans!",
	})
}

func (s *Server) getSubscription(ctx *gin.Context) {
	userID := ctx.Param("userID")

	uuidUserID, err := uuid.Parse(userID)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	subscription, err := s.dbClient.HowManySubscription().FindByUserID(ctx, uuidUserID)
	if err != nil {
		ctx.JSON(404, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(200, subscription)
}

func (s *Server) register(ctx *gin.Context) {
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

	subscription, err := s.dbClient.HowManySubscription().Insert(ctx, authenticateResponse.User.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := s.texterClient.Text(ctx, &texter.Text{
		Body:     "Welcome to Guess How Many! New question every day at noon EST. Text CANCEL at any point to cancel your subscription.",
		From:     s.cfg.Secret.GuessHowManyPhoneNumber,
		To:       authenticateResponse.User.PhoneNumber,
		TextType: "guess-how-many-welcome-email",
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	ctx.JSON(200, gin.H{
		"user":         authenticateResponse.User,
		"subscription": subscription,
		"token":        authenticateResponse.Token,
	})
}

func (s *Server) login(ctx *gin.Context) {
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

	subscription, err := s.dbClient.HowManySubscription().FindByUserID(ctx, authenticateResponse.User.ID)
	if err == nil {
		payload["subscription"] = subscription
	} else {
		subscription, err := s.dbClient.HowManySubscription().Insert(ctx, authenticateResponse.User.ID)
		if err == nil {
			payload["subscription"] = subscription
		}
	}

	ctx.JSON(200, payload)
}

func (s *Server) finishCheckout(ctx *gin.Context) {
	var onSubscribe billing.OnSubscribe

	if err := ctx.Bind(&onSubscribe); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	userID, ok := onSubscribe.Metadata["user_id"]
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "user id is required to subscribe",
		})
		return
	}

	uuidUserID, err := uuid.Parse(userID)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := s.dbClient.HowManySubscription().SetSubscribed(ctx, uuidUserID, onSubscribe.SubscriptionID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "cool beans!",
	})
}

func (s *Server) handleSubscriptionLinkRedirect(ctx *gin.Context, userID string) {
	if len(userID) == 0 {
		ctx.JSON(400, gin.H{
			"error": "user id required",
		})
		return
	}

	uuidUserID, err := uuid.Parse(userID)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := s.dbClient.User().FindByID(ctx, uuidUserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	session, err := s.billingClient.NewCheckoutSession(ctx, &billing.CheckoutSessionParams{
		PlanID:                        s.cfg.Public.GuessHowManyPlanID,
		SessionSuccessRedirectUrl:     s.cfg.Public.GuessHowManySubscribeSuccessUrl,
		SessionCancelRedirectUrl:      s.cfg.Public.GuessHowManySubscribeCancelUrl,
		SubscriptionCreateCallbackUrl: "http://localhost:8082/trivai/finish-checkout",
		SubscriptionCancelCallbackUrl: "http://localhost:8082/trivai/subscriptions/delete",
		Metadata: map[string]string{
			"user_id": user.ID.String(),
		},
	})
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.Redirect(http.StatusSeeOther, session.URL)
}

func (s *Server) getSubscriptionLink(ctx *gin.Context) {
	userID := ctx.Param("userId")

	s.handleSubscriptionLinkRedirect(ctx, userID)

}

func (s *Server) beginCheckout(ctx *gin.Context) {
	userID := ctx.PostForm("userId")

	s.handleSubscriptionLinkRedirect(ctx, userID)
}

func (s *Server) receiveSms(ctx *gin.Context) {
	var smsRequest struct {
		From string `json:"from" binding:"required"`
		Body string `json:"body" binding:"required"`
	}

	if err := ctx.Bind(&smsRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "shit sms receive request",
		})
		return
	}

	question, err := s.dbClient.HowManyQuestion().FindTodaysQuestion(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	user, err := s.dbClient.User().FindByPhoneNumber(ctx, smsRequest.From)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "no user found for that phone number: " + smsRequest.From,
		})
		return
	}

	if strings.Contains(strings.ToLower(smsRequest.Body), "cancel") {
		howManySubscription, err := s.dbClient.HowManySubscription().FindByUserID(ctx, user.ID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "no subscription found for that user id: " + user.ID.String(),
			})
			return
		}

		if howManySubscription.StripeID != nil && howManySubscription.Subscribed {
			if _, err := s.billingClient.CancelSubscription(ctx, &billing.CancelSubscriptionParams{
				StripeID: *howManySubscription.StripeID,
			}); err != nil {
				ctx.JSON(500, gin.H{
					"error": err.Error(),
				})
				return
			}
		}

		ctx.JSON(200, gin.H{
			"body": "You have been sucessfully unsubscribed. Cya later!",
			"to":   user.PhoneNumber,
		})
		return
	}

	guess, err := util.ParseNumber(util.LongestNumericSubstring(smsRequest.Body))
	if err != nil {
		ctx.JSON(200, gin.H{
			"body": "We're sorry. We weren't able to understand that. Please try again.",
			"to":   user.PhoneNumber,
		})
		return
	}

	correctness, offBy := s.triviaClient.GradeHowManyQuestion(ctx, guess, question.HowMany)

	answer := models.HowManyAnswer{
		UserID:            user.ID,
		HowManyQuestionID: question.ID,
		Correctness:       correctness,
		OffBy:             offBy,
		Answer:            question.HowMany,
		Guess:             guess,
	}

	if _, err := s.dbClient.HowManyAnswer().Insert(ctx, &answer); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"body": util.GetShareMessage(guess, question.HowMany, correctness, question.Explanation),
		"to":   user.PhoneNumber,
	})
}

func (s *Server) getUser(c *gin.Context) {
	userID := c.Param("userId")

	uuidUserID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := s.dbClient.User().FindByID(c, uuidUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "no user found",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (s *Server) subscribeToHowManyQuestions(ctx *gin.Context) {
	var subscribeRequest struct {
		UserID      string `json:"userId" binding:"required"`
		PhoneNumber string `json:"phoneNumber" binding:"required"`
	}

	if err := ctx.Bind(&subscribeRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "shit subscribe request",
		})
		return
	}

	user, err := s.dbClient.User().Insert(ctx, subscribeRequest.UserID, subscribeRequest.PhoneNumber, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (s *Server) getHowManyQuestionAnswer(c *gin.Context) {
	stringQuestionID := c.Query("questionId")
	stringUserID := c.Query("userId")

	questionID, err := uuid.Parse(stringQuestionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid question id",
		})
		return
	}

	userID, err := uuid.Parse(stringUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid question id",
		})
		return
	}

	answer, err := s.dbClient.HowManyAnswer().FindByQuestionIDAndUserID(c, questionID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, answer)
}

func (s *Server) gradeQuestion(ctx *gin.Context) {
	var gradeRequest struct {
		Guess  int    `json:"guess" binding:"required"`
		Id     string `json:"id" binding:"required"`
		UserID string `json:"userId"`
	}

	if err := ctx.Bind(&gradeRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "shit grade request",
		})
		return
	}

	userID, err := uuid.Parse(gradeRequest.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "shit user id",
		})
		return
	}

	id, err := uuid.Parse(gradeRequest.Id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "shit id",
		})
		return
	}

	question, err := s.dbClient.HowManyQuestion().FindById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	correctness, offBy := s.triviaClient.GradeHowManyQuestion(ctx, gradeRequest.Guess, question.HowMany)

	answer := models.HowManyAnswer{
		UserID:            userID,
		HowManyQuestionID: id,
		Correctness:       correctness,
		OffBy:             offBy,
		Answer:            question.HowMany,
		Guess:             gradeRequest.Guess,
	}

	if _, err := s.dbClient.HowManyAnswer().Insert(ctx, &answer); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, answer)
}

func (s *Server) getCurrentQuestion(ctx *gin.Context) {
	question, err := s.dbClient.HowManyQuestion().FindTodaysQuestion(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	question.HowMany = 0

	ctx.JSON(http.StatusOK, question)
}

func (s *Server) markHowManyQuestionValid(ctx *gin.Context) {
	id := ctx.Param("id")

	questionID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "shit user id",
		})
		return
	}

	if err := s.dbClient.HowManyQuestion().MarkValid(ctx, questionID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "You're all set!"})
}

func (s *Server) getHowManyQuestions(ctx *gin.Context) {
	questions, err := s.dbClient.HowManyQuestion().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, questions)
}

func (s *Server) generateNewHowManyQuestion(ctx *gin.Context) {
	var promptString string

	var generateQuestionRequest struct {
		PrompSeed string `json:"promptSeed" binding:"required"`
	}

	if err := ctx.Bind(&generateQuestionRequest); err != nil {
		promptString = ""
	} else {
		promptString = generateQuestionRequest.PrompSeed
	}

	howManyQuestion, err := s.triviaClient.GenerateNewHowManyQuestion(ctx, promptString)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	question, err := s.dbClient.HowManyQuestion().Insert(ctx, howManyQuestion.Text, howManyQuestion.Explanation, howManyQuestion.HowMany, 0, "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, question)
}
