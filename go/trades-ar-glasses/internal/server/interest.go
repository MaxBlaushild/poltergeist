package server

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

const (
	maxEmailLength     = 254
	maxTradeLength     = 80
	maxCrewSizeLength  = 30
	maxSourceLength    = 80
	maxUserAgentLength = 200
)

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type interestRequest struct {
	Email    string `json:"email"`
	Trade    string `json:"trade"`
	CrewSize string `json:"crewSize"`
	Source   string `json:"source"`
}

func (s *server) createInterestLead(ctx *gin.Context) {
	var req interestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body"})
		return
	}

	lead, err := buildLeadFromRequest(req, ctx.GetHeader("User-Agent"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := s.leadStore.CreateOrGetByEmail(ctx.Request.Context(), lead)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record interest"})
		return
	}

	status := http.StatusOK
	message := "email already recorded"
	if created {
		status = http.StatusCreated
		message = "interest recorded"
	}

	ctx.JSON(status, gin.H{
		"ok":      true,
		"created": created,
		"message": message,
		"lead": gin.H{
			"id":    lead.ID,
			"email": lead.Email,
		},
	})
}

func buildLeadFromRequest(req interestRequest, userAgent string) (*models.TradesARGlassesLead, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || len(email) > maxEmailLength || !emailPattern.MatchString(email) {
		return nil, errInvalidEmail
	}

	source := cleanOptionalString(req.Source, maxSourceLength)
	if source == "" {
		source = "landing-page"
	}

	return &models.TradesARGlassesLead{
		Email:     email,
		Trade:     cleanOptionalString(req.Trade, maxTradeLength),
		CrewSize:  cleanOptionalString(req.CrewSize, maxCrewSizeLength),
		Source:    source,
		UserAgent: cleanOptionalString(userAgent, maxUserAgentLength),
	}, nil
}

func cleanOptionalString(value string, maxLength int) string {
	cleaned := strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if len(cleaned) > maxLength {
		return cleaned[:maxLength]
	}
	return cleaned
}

type validationError string

func (e validationError) Error() string {
	return string(e)
}

const errInvalidEmail = validationError("valid email is required")
