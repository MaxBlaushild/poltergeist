package server

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/middleware"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	ErrNotAuthenticated = errors.New("no authenticated user found")
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
	r.GET("/sonar/surveys/submissions", middleware.WithAuthentication(s.authClient, s.getSurveySubmissions))
	r.POST("/sonar/surveys", middleware.WithAuthentication(s.authClient, s.newSurvey))
	r.GET("/sonar/activities", middleware.WithAuthentication(s.authClient, s.getActivities))
	r.GET("/sonar/categories", middleware.WithAuthentication(s.authClient, s.getCategories))
	r.GET("/sonar/surveys/:id", middleware.WithAuthentication(s.authClient, s.getSurvey))
	r.GET("sonar/surveys/:id/submissions", middleware.WithAuthentication(s.authClient, s.getSubmissionForSurvey))
	r.GET("/sonar/submissions/:id", middleware.WithAuthentication(s.authClient, s.getSubmission))
	r.GET("/sonar/whoami", middleware.WithAuthentication(s.authClient, s.whoami))
	r.POST("/sonar/surveys/:id/submissions", middleware.WithAuthentication(s.authClient, s.submitSurveyAnswer))
	r.POST("/sonar/categories", middleware.WithAuthentication(s.authClient, s.createCategory))
	r.POST("/sonar/activities", middleware.WithAuthentication(s.authClient, s.createActivity))
	r.DELETE("/sonar/categories/:id", middleware.WithAuthentication(s.authClient, s.deleteCategory))
	r.DELETE("/sonar/activities/:id", middleware.WithAuthentication(s.authClient, s.deleteActivity))

	r.Run(":8042")
}

func (s *server) deleteCategory(ctx *gin.Context) {
	categoryID := ctx.Param("id")
	if categoryID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "category ID is required",
		})
		return
	}

	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid category ID",
		})
		return
	}

	if err := s.dbClient.SonarCategory().DeleteCategory(ctx, categoryUUID, user.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "category deleted successfully",
	})
}

func (s *server) deleteActivity(ctx *gin.Context) {
	activityID := ctx.Param("id")
	if activityID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "activity ID is required",
		})
		return
	}
}

func (s *server) createCategory(ctx *gin.Context) {
	var requestBody struct {
		Title string `binding:"required" json:"name"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	category, err := s.dbClient.SonarCategory().CreateCategory(ctx, models.SonarCategory{
		Title:  requestBody.Title,
		UserID: &user.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, category)
}

func (s *server) createActivity(ctx *gin.Context) {
	var requestBody struct {
		Title      string    `binding:"required" json:"title"`
		CategoryID uuid.UUID `binding:"required" json:"categoryId"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	activity, err := s.dbClient.SonarActivity().CreateActivity(ctx, models.SonarActivity{
		Title:           requestBody.Title,
		SonarCategoryID: requestBody.CategoryID,
		UserID:          &user.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, activity)
}

func (s *server) whoami(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (s *server) getSubmission(ctx *gin.Context) {
	submissionID := ctx.Param("id")
	if submissionID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "submission ID is required",
		})
		return
	}

	submissionUUID, err := uuid.Parse(submissionID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid survey ID",
		})
		return
	}

	submission, err := s.dbClient.SonarSurveySubmission().GetSubmissionByID(ctx, submissionUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch submission",
		})
		return
	}

	ctx.JSON(http.StatusOK, submission)
}

func (s *server) getSubmissionForSurvey(ctx *gin.Context) {
	surveyID := ctx.Param("id")
	if surveyID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "survey ID is required",
		})
		return
	}

	surveyUUID, err := uuid.Parse(surveyID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid survey ID",
		})
		return
	}

	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	submission, err := s.dbClient.SonarSurveySubmission().GetUserSubmissionForSurvey(ctx, user.ID, surveyUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch submission for survey",
		})
		return
	}

	ctx.JSON(http.StatusOK, submission)

}

func (s *server) getSurvey(ctx *gin.Context) {
	surveyID := ctx.Param("id")
	if surveyID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "survey ID is required",
		})
		return
	}

	surveyUUID, err := uuid.Parse(surveyID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid survey ID",
		})
		return
	}

	survey, err := s.dbClient.SonarSurvey().GetSurveyByID(ctx, surveyUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, survey)
}

func (s *server) getCategories(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	categories, err := s.dbClient.SonarCategory().GetAllCategoriesWithActivities(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	userSpecificCategories, err := s.dbClient.SonarCategory().GetCategoriesByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	allCategories := append(categories, userSpecificCategories...)
	sort.Slice(allCategories, func(i, j int) bool {
		return allCategories[i].Title < allCategories[j].Title
	})

	ctx.JSON(http.StatusOK, allCategories)
}

func (s *server) getActivities(ctx *gin.Context) {
	activities, err := s.dbClient.SonarActivity().GetAllActivities(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, activities)
}

func (s *server) getSurveySubmissions(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	submissions, err := s.dbClient.SonarSurveySubmission().GetAllSubmissionsForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, submissions)
}

func (s *server) submitSurveyAnswer(ctx *gin.Context) {
	sID := ctx.Param("id")
	if sID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "survey ID is required",
		})
		return
	}
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var requestBody struct {
		ActivityIDs []string `json:"activityIds"`
		Downs       []bool   `json:"downs"`
	}

	if err := ctx.BindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	surveyID, err := uuid.Parse(sID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid survey ID",
		})
		return
	}

	activityUUIDs := make([]uuid.UUID, len(requestBody.ActivityIDs))
	for i, activityID := range requestBody.ActivityIDs {
		activityUUID, err := uuid.Parse(activityID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("invalid activity ID at index %d", i),
			})
			return
		}
		activityUUIDs[i] = activityUUID
	}

	submission, err := s.dbClient.SonarSurveySubmission().CreateSubmission(ctx, surveyID, user.ID, activityUUIDs, requestBody.Downs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "survey submission created successfully",
		"submission": submission,
	})
}

func (s *server) getAuthenticatedUser(ctx *gin.Context) (*models.User, error) {
	u, ok := ctx.Get("user")
	if !ok {
		return nil, ErrNotAuthenticated
	}

	user, ok := u.(*models.User)
	if !ok {
		return nil, ErrNotAuthenticated
	}

	return user, nil
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

func (s *server) newSurvey(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var newSurveyRequest struct {
		ActivityIDs []uuid.UUID `binding:"required" json:"activityIds"`
		Name        string      `binding:"required" json:"name"`
	}

	if err := ctx.Bind(&newSurveyRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	survey, err := s.dbClient.SonarSurvey().CreateSurvey(ctx, user.ID, newSurveyRequest.Name, newSurveyRequest.ActivityIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, survey)
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
