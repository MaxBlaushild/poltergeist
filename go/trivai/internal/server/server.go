package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/email"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/MaxBlaushild/poltergeist/trivai/internal/trivai"
	"github.com/gin-gonic/gin"
)

type Server struct {
	dbClient     db.DbClient
	emailClient  email.EmailClient
	triviaClient trivai.TrivaiClient
	texterClient texter.TexterClient
}

func NewServer(
	dbClient db.DbClient,
	emailClient email.EmailClient,
	trivaiClient trivai.TrivaiClient,
	texterClient texter.TexterClient,
) Server {
	r := gin.Default()

	s := Server{
		dbClient:     dbClient,
		emailClient:  emailClient,
		triviaClient: trivaiClient,
		texterClient: texterClient,
	}

	r.POST("/trivai/matches", s.createMatch)
	r.POST("/trivai/question_sets/:questionSetID/answer", s.submitAnswers)
	r.GET("/trivai/questions", s.getQuestions)
	r.GET("/trivai/users/:userId", s.getUser)

	r.POST("/")
	r.POST("/trivai/receive-sms", s.receiveSms)
	r.POST("/trivai/how_many_questions/subscribe", s.subscribeToHowManyQuestions)
	r.GET("/trivai/how_many_questions/current", s.getCurrentQuestion)
	r.POST("/trivai/how_many_questions/grade", s.gradeQuestion)
	r.GET("/trivai/how_many_questions", s.getHowManyQuestions)
	r.GET("/trivai/how_many_questions/answer", s.getHowManyQuestionAnswer)
	r.POST("/trivai/how_many_questions", s.generateNewHowManyQuestion)
	r.POST("/trivai/how_many_questions/:id/validate", s.markHowManyQuestionValid)

	r.Run(":8082")

	return s
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
		UserID:            &user.ID,
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
	userId := c.Param("userId")

	uint64Val, err := strconv.ParseUint(userId, 10, 64) // Base 10, BitSize 64
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := s.dbClient.User().FindByID(c, uint(uint64Val))
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
		UserId      string `json:"userId" binding:"required"`
		PhoneNumber string `json:"phoneNumber" binding:"required"`
	}

	if err := ctx.Bind(&subscribeRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "shit subscribe request",
		})
		return
	}

	user, err := s.dbClient.User().Insert(ctx, subscribeRequest.UserId, subscribeRequest.PhoneNumber)
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
	ephemeralUserID := c.Query("ephemeralUserId")

	if len(stringQuestionID) == 0 {
		c.JSON(400, gin.H{
			"error": "question id is required",
		})
		return
	}

	questionId64, err := strconv.ParseUint(stringQuestionID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid question id",
		})
		return
	}

	if len(stringUserID) != 0 && len(ephemeralUserID) != 0 {
		c.JSON(400, gin.H{
			"error": "only allowed one of user id and emphemeral user id",
		})
		return
	}

	if len(stringUserID) == 0 && len(ephemeralUserID) == 0 {
		c.JSON(400, gin.H{
			"error": "one of user id and emphemeral user id required",
		})
		return
	}

	if len(stringUserID) != 0 {
		userId64, err := strconv.ParseUint(stringUserID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid id",
			})
			return
		}

		answer, err := s.dbClient.HowManyAnswer().FindByQuestionIDAndUserID(c, uint(questionId64), uint(userId64))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, answer)
		return
	}

	answer, err := s.dbClient.HowManyAnswer().FindByQuestionIDAndEphemeralUserID(c, uint(questionId64), ephemeralUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, answer)
}

func (s *Server) gradeQuestion(ctx *gin.Context) {
	var gradeRequest struct {
		Guess           int     `json:"guess" binding:"required"`
		Id              uint    `json:"id" binding:"required"`
		UserId          *uint   `json:"userId"`
		EphemeralUserId *string `json:"ephemeralUserId"`
	}

	if err := ctx.Bind(&gradeRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "shit grade request",
		})
		return
	}

	question, err := s.dbClient.HowManyQuestion().FindById(ctx, gradeRequest.Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	correctness, offBy := s.triviaClient.GradeHowManyQuestion(ctx, gradeRequest.Guess, question.HowMany)

	answer := models.HowManyAnswer{
		UserID:            gradeRequest.UserId,
		HowManyQuestionID: gradeRequest.Id,
		Correctness:       correctness,
		OffBy:             offBy,
		Answer:            question.HowMany,
		Guess:             gradeRequest.Guess,
		EphemeralUserID:   gradeRequest.EphemeralUserId,
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

	if err := s.dbClient.HowManyQuestion().MarkValid(ctx, id); err != nil {
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

	fmt.Println(howManyQuestion.Explanation)

	question, err := s.dbClient.HowManyQuestion().Insert(ctx, howManyQuestion.Text, howManyQuestion.Explanation, howManyQuestion.HowMany)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, question)
}

func (s *Server) getQuestions(ctx *gin.Context) {
	questions, err := s.dbClient.Question().GetAllQuestions(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "something went wrong",
		})
		return
	}

	ctx.JSON(http.StatusOK, questions)
}

func (s *Server) getUserSubmissionForQuestionSet(c *gin.Context) {
	userID := c.Param("userID")
	intUserID, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid user id",
		})
		return
	}

	questionSetID := c.Param("questionSetID")
	intQuestionSetID, err := strconv.Atoi(questionSetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid question set id",
		})
		return
	}

	userSubmission, err := s.dbClient.UserSubmission().FindByUserAndQuestionSetID(c, uint(intUserID), uint(intQuestionSetID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "no submission for user yet",
		})
		return
	}

	c.JSON(http.StatusOK, userSubmission)
}

func (s *Server) getCurrentMatch(c *gin.Context) {
	userID := c.Param("userID")
	intUserID, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid user id",
		})
		return
	}
	match, err := s.dbClient.Match().GetCurrentMatchForUser(c, uint(intUserID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "no current match for user",
		})
		return
	}

	c.JSON(http.StatusOK, match)
}

func (s *Server) createMatch(c *gin.Context) {
	var createMatchRequest struct {
		HomeUserID uint `json:"home_user_id" binding:"required"`
		AwayUserID uint `json:"away_user_id" binding:"required"`
	}

	if err := c.Bind(&createMatchRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "your request sucks ass",
		})
		return
	}

	questions, err := s.triviaClient.GenerateQuestions(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "trivai shit the bed",
		})
		return
	}

	var qs []models.Question
	for _, q := range questions {
		qs = append(qs, models.Question{
			Prompt: q.Prompt,
			Category: models.Category{
				Title: q.Category,
			},
			Answer: q.Answer,
		})
	}

	questionSet, err := s.dbClient.QuestionSet().Insert(c, qs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	homeUser, err := s.dbClient.User().FindByID(c, createMatchRequest.HomeUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "home user not found",
		})
		return
	}

	awayUser, err := s.dbClient.User().FindByID(c, createMatchRequest.AwayUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "away user not found",
		})
		return
	}

	match := models.Match{
		Home:        *homeUser,
		Away:        *awayUser,
		QuestionSet: *questionSet,
	}

	if err := s.dbClient.Match().Insert(c, &match); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"match_id": match.ID})
}

func (s *Server) submitAnswers(c *gin.Context) {
	var submitAnswersRequest struct {
		UserID  uint                `json:"userId" binding:"required"`
		Answers []models.UserAnswer `json:"answers" binding:"required"`
	}

	questionSetID := c.Param("questionSetID")
	intQuestionSetID, err := strconv.Atoi(questionSetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid question set id",
		})
		return
	}

	if err := c.Bind(&submitAnswersRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	questions, err := s.dbClient.Question().FindByQuestionSetID(c, uint(intQuestionSetID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	var trivaiQuestions []trivai.Question
	var answers []string

	for _, userAnswer := range submitAnswersRequest.Answers {
		answers = append(answers, userAnswer.Answer)

		var question *models.Question
		for i, q := range questions {
			if q.ID == userAnswer.QuestionID {
				question = &questions[i]
			}
		}

		if question == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "submitted answer for ineligible question",
			})
			return
		}

		trivaiQuestions = append(trivaiQuestions, trivai.Question{
			Prompt: question.Prompt,
		})
	}

	grades, err := s.triviaClient.GradeUserSubmission(c, trivaiQuestions, answers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	for i := range submitAnswersRequest.Answers {
		submitAnswersRequest.Answers[i].Correct = grades[i]
		submitAnswersRequest.Answers[i].UserID = submitAnswersRequest.UserID
	}

	submission, err := s.dbClient.UserSubmission().Insert(
		c,
		uint(intQuestionSetID),
		submitAnswersRequest.UserID,
		submitAnswersRequest.Answers,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, submission)
}
