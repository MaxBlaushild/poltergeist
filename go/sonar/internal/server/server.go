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
	"github.com/MaxBlaushild/poltergeist/sonar/internal/config"
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
	config       *config.Config
}

type Server interface {
	ListenAndServe(port string)
}

func NewServer(authClient auth.Client, texterClient texter.Client, dbClient db.DbClient, config *config.Config) Server {
	return &server{authClient: authClient, texterClient: texterClient, dbClient: dbClient, config: config}
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
	r.GET("/sonar/userProfiles", middleware.WithAuthentication(s.authClient, s.getUserProfiles))
	r.POST("/sonar/surveys/:id/submissions", middleware.WithAuthentication(s.authClient, s.submitSurveyAnswer))
	r.POST("/sonar/categories", middleware.WithAuthentication(s.authClient, s.createCategory))
	r.POST("/sonar/activities", middleware.WithAuthentication(s.authClient, s.createActivity))
	r.DELETE("/sonar/categories/:id", middleware.WithAuthentication(s.authClient, s.deleteCategory))
	r.DELETE("/sonar/activities/:id", middleware.WithAuthentication(s.authClient, s.deleteActivity))
	r.GET("/sonar/teams", middleware.WithAuthentication(s.authClient, s.getTeams))
	r.POST("/sonar/pointsOfInterest", middleware.WithAuthentication(s.authClient, s.createPointOfInterest))
	r.GET("/sonar/pointsOfInterest", middleware.WithAuthentication(s.authClient, s.getPointsOfInterest))
	r.POST("/sonar/crystal/unlock", middleware.WithAuthentication(s.authClient, s.createCrystalUnlock))
	r.POST("/sonar/crystal/capture", middleware.WithAuthentication(s.authClient, s.createCrystalCapture))
	r.POST("/sonar/neighbors", middleware.WithAuthentication(s.authClient, s.createNeighbor))
	r.GET("/sonar/neighbors", middleware.WithAuthentication(s.authClient, s.getNeighbors))
	r.POST("/sonar/matches/:id/start", middleware.WithAuthentication(s.authClient, s.startMatch))
	r.POST("/sonar/matches/:id/end", middleware.WithAuthentication(s.authClient, s.endMatch))
	r.POST("/sonar/matches", middleware.WithAuthentication(s.authClient, s.createMatch))
	r.GET("/sonar/matches/:id", middleware.WithAuthentication(s.authClient, s.getMatch))
	r.POST("/sonar/matches/teams/newTeam", middleware.WithAuthentication(s.authClient, s.createTeamForMatch))
	r.POST("/sonar/matches/teams/addUser", middleware.WithAuthentication(s.authClient, s.addUserToTeam))
	r.GET("/sonar/pointsOfInterest/group/:id", s.getPointsOfInterestByGroup)
	r.POST("/sonar/pointsOfInterest/group", middleware.WithAuthentication(s.authClient, s.createPointOfInterestGroup))
	r.GET("/sonar/pointsOfInterest/groups", s.getPointsOfInterestGroups)
	r.GET("/sonar/matches/current", middleware.WithAuthentication(s.authClient, s.getCurrentMatch))

	r.Run(":8042")
}

func (s *server) getCurrentMatch(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	match, err := s.dbClient.Match().FindCurrentMatchForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if match == nil || match.EndedAt != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "no current match",
		})
		return
	}

	ctx.JSON(http.StatusOK, match)
}

func (s *server) getPointsOfInterestGroups(ctx *gin.Context) {
	groups, err := s.dbClient.PointOfInterestGroup().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, groups)
}

func (s *server) createPointOfInterestGroup(ctx *gin.Context) {
	var requestBody struct {
		PointOfInterestIDs []uuid.UUID `binding:"required" json:"pointOfInterestIDs"`
		Name               string      `binding:"required" json:"name"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	group, err := s.dbClient.PointOfInterestGroup().Create(ctx, requestBody.PointOfInterestIDs, requestBody.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, group)
}

func (s *server) getPointsOfInterestByGroup(ctx *gin.Context) {
	groupID := ctx.Param("id")
	if groupID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "group ID is required",
		})
		return
	}

	uuidGroupID, err := uuid.Parse(groupID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid group ID",
		})
		return
	}

	pointsOfInterest, err := s.dbClient.PointOfInterest().FindByGroupID(ctx, uuidGroupID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, pointsOfInterest)
}

func (s *server) getUserProfiles(ctx *gin.Context) {
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

	var sonarUsers []*models.SonarUser
	for _, submission := range submissions {
		sonarUser, err := s.dbClient.SonarUser().FindOrCreateSonarUser(ctx, user.ID, submission.UserID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		sonarUsers = append(sonarUsers, sonarUser)
	}

	ctx.JSON(http.StatusOK, sonarUsers)
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

	ctx.JSON(200, gin.H{
		"user":  authenticateResponse.User,
		"token": authenticateResponse.Token,
	})
}

func (s *server) createNeighbor(c *gin.Context) {
	var neighbor models.NeighboringPointsOfInterest

	if err := c.Bind(&neighbor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "shit neighbor create request",
		})
		return
	}

	if err := s.dbClient.NeighboringPointsOfInterest().Create(c, neighbor.PointOfInterestOneID, neighbor.PointOfInterestTwoID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "everythings ok",
	})
}

func (s *server) getNeighbors(c *gin.Context) {
	neighbors, err := s.dbClient.NeighboringPointsOfInterest().FindAll(c)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, neighbors)
}

func (s *server) getTeams(c *gin.Context) {
	teams, err := s.dbClient.Team().GetAll(c)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var userIDs []uuid.UUID
	for _, team := range teams {
		for _, userTeam := range team.UserTeams {
			userIDs = append(userIDs, userTeam.UserID)
		}
	}

	payload := gin.H{
		"teams": teams,
	}

	if len(teams) > 0 {
		users, err := s.authClient.GetUsers(c, userIDs)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		payload["users"] = users
	}

	c.JSON(200, payload)
}

func (s *server) getPointsOfInterest(c *gin.Context) {
	stringTeamID := c.Param("teamID")
	pointOfInterests, err := s.dbClient.PointOfInterest().FindAll(c)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	teamID, err := uuid.Parse(stringTeamID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	pointOfInterestTeams, err := s.dbClient.PointOfInterestTeam().FindByTeamID(c, teamID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	for i, pointOfInterest := range pointOfInterests {
		found := false
		for _, pointOfInterestTeam := range pointOfInterestTeams {
			if pointOfInterestTeam.PointOfInterestID == pointOfInterest.ID {
				found = true
			}
		}

		if !found {
			pointOfInterests[i].AttuneChallenge = ""
			pointOfInterests[i].CaptureChallenge = ""
		} else {
			pointOfInterests[i].Clue = ""
		}
	}

	c.JSON(200, pointOfInterests)
}

func (s *server) createMatch(c *gin.Context) {
	var createMatchRequest struct {
		PointsOfInterestIDs []uuid.UUID `json:"pointsOfInterestIds" binding:"required"`
	}

	if err := c.Bind(&createMatchRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := s.getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	match, err := s.dbClient.Match().Create(c, user.ID, createMatchRequest.PointsOfInterestIDs)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, match)
}

func (s *server) createTeamForMatch(c *gin.Context) {
	var createTeamForMatchRequest struct {
		UserID  string `json:"userIds" binding:"required"`
		Name    string `json:"name" binding:"required"`
		MatchID string `json:"matchId" binding:"required"`
	}

	matchID, err := uuid.Parse(createTeamForMatchRequest.MatchID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(createTeamForMatchRequest.UserID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	team, err := s.dbClient.Team().Create(c, []uuid.UUID{userID}, createTeamForMatchRequest.Name, matchID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, team)
}

func (s *server) addUserToTeam(c *gin.Context) {
	var addUserToTeamRequest struct {
		UserID string `json:"userId" binding:"required"`
		TeamID string `json:"teamId" binding:"required"`
	}

	if err := c.Bind(&addUserToTeamRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(addUserToTeamRequest.UserID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	teamID, err := uuid.Parse(addUserToTeamRequest.TeamID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Team().AddUserToTeam(c, teamID, userID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "done"})
}

func (s *server) getMatch(c *gin.Context) {
	matchID := c.Param("id")

	uuidMatchID, err := uuid.Parse(matchID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	match, err := s.dbClient.Match().FindByID(c, uuidMatchID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, match)
}

func (s *server) startMatch(c *gin.Context) {
	matchID := c.Param("id")

	uuidMatchID, err := uuid.Parse(matchID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Match().StartMatch(c, uuidMatchID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "done"})
}

func (s *server) endMatch(c *gin.Context) {
	matchID := c.Param("id")

	uuidMatchID, err := uuid.Parse(matchID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Match().EndMatch(c, uuidMatchID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "done"})
}

func (s *server) createPointOfInterest(c *gin.Context) {
	var createPointOfInterestRequest struct {
		Name             string `json:"name" binding:"required"`
		Clue             string `json:"clue" binding:"required"`
		CaptureChallenge string `json:"captureChallenge" binding:"required"`
		AttuneChallenge  string `json:"attuneChallenge" binding:"required"`
		Lat              string `json:"lat" binding:"required"`
		Lng              string `json:"lng" binding:"required"`
	}

	if err := c.Bind(&createPointOfInterestRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := s.dbClient.PointOfInterest().Create(c, models.PointOfInterest{
		Name:             createPointOfInterestRequest.Name,
		Clue:             createPointOfInterestRequest.Clue,
		AttuneChallenge:  createPointOfInterestRequest.AttuneChallenge,
		CaptureChallenge: createPointOfInterestRequest.CaptureChallenge,
		Lat:              createPointOfInterestRequest.Lat,
		Lng:              createPointOfInterestRequest.Lng,
	}); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "everything cool",
	})
}

func (s *server) createCrystalUnlock(c *gin.Context) {
	var crystalUnlockRequest struct {
		TeamID    string `json:"teamId" binding:"required"`
		CrystalID string `json:"crystalId" binding:"required"`
	}

	if err := c.Bind(&crystalUnlockRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "shit crystal unlock request",
		})
		return
	}

	crystalID, err := uuid.Parse(crystalUnlockRequest.CrystalID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "shit crystal id",
		})
		return
	}

	teamID, err := uuid.Parse(crystalUnlockRequest.TeamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "shit team id",
		})
		return
	}

	if err := s.dbClient.PointOfInterest().Unlock(c, crystalID, teamID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "everything cool",
	})
}

func (s *server) createCrystalCapture(c *gin.Context) {
	var captureCrystalRequest struct {
		PointOfInterestID string `json:"pointOfInterestId" binding:"required"`
		TeamID            string `json:"teamId" binding:"required"`
		Attune            bool   `json:"attune"`
	}

	if err := c.Bind(&captureCrystalRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	pointOfInterestID, err := uuid.Parse(captureCrystalRequest.PointOfInterestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "shit crystal id",
		})
		return
	}

	teamID, err := uuid.Parse(captureCrystalRequest.TeamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "shit team id",
		})
		return
	}

	if err := s.dbClient.PointOfInterest().Capture(c, pointOfInterestID, teamID, captureCrystalRequest.Attune); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	/// TEXT EVERYBODY!!!!!!!!!!!

	teams, err := s.dbClient.Team().GetAll(c)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var userIDs []uuid.UUID
	for _, team := range teams {
		for _, userTeam := range team.UserTeams {
			userIDs = append(userIDs, userTeam.UserID)
		}
	}

	users, err := s.authClient.GetUsers(c, userIDs)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var capturedOrAttuned string
	if captureCrystalRequest.Attune {
		capturedOrAttuned = "attuned"
	} else {
		capturedOrAttuned = "captured"
	}

	var capturingTeam models.Team
	for _, team := range teams {
		if team.ID == teamID {
			capturingTeam = team
		}
	}

	pointOfInterest, err := s.dbClient.PointOfInterest().FindByID(c, pointOfInterestID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	for _, user := range users {
		s.texterClient.Text(c, &texter.Text{
			To:   user.PhoneNumber,
			From: s.config.Public.PhoneNumber,
			Body: fmt.Sprintf("%s team has %s %s.", capturingTeam.Name, capturedOrAttuned, pointOfInterest.Name),
		})
	}

	c.JSON(200, gin.H{"messgae": "done!"})
}
