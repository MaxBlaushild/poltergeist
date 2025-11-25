package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/MaxBlaushild/poltergeist/pkg/billing"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *server) GetCredits(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Fetch updated user to get current credits
	updatedUser, err := s.dbClient.User().FindByID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"credits": updatedUser.Credits,
	})
}

func (s *server) PurchaseCredits(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var requestBody struct {
		AmountInDollars int `json:"amountInDollars" binding:"required"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate amount (minimum $1, maximum $1000)
	if requestBody.AmountInDollars < 1 || requestBody.AmountInDollars > 1000 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "amount must be between $1 and $1000",
		})
		return
	}

	// Convert dollars to cents
	amountInCents := int64(requestBody.AmountInDollars * 100)

	// Create success and cancel URLs
	// These should redirect back to the app
	successURL := "travelangels://credits/purchase/success?session_id={CHECKOUT_SESSION_ID}"
	cancelURL := "travelangels://credits/purchase/cancel"

	// Create callback URL for webhook - use baseURL from config, fallback to localhost for local dev
	baseURL := s.baseURL
	if baseURL == "" {
		baseURL = "http://localhost:8083"
		log.Printf("[WARNING] BASE_URL not configured, using localhost fallback")
	}
	paymentCompleteCallbackURL := fmt.Sprintf("%s/travel-angels/credits/webhook", baseURL)

	log.Printf("[PurchaseCredits] User %s purchasing %d credits ($%d)", user.ID, requestBody.AmountInDollars, requestBody.AmountInDollars)
	log.Printf("[PurchaseCredits] Callback URL: %s", paymentCompleteCallbackURL)

	// Create checkout session
	checkoutSession, err := s.billingClient.NewPaymentCheckoutSession(ctx, &billing.PaymentCheckoutSessionParams{
		SessionSuccessRedirectUrl:  successURL,
		SessionCancelRedirectUrl:   cancelURL,
		AmountInCents:              amountInCents,
		PaymentCompleteCallbackUrl: paymentCompleteCallbackURL,
		Metadata: map[string]string{
			"user_id":           user.ID.String(),
			"amount_in_dollars": strconv.Itoa(requestBody.AmountInDollars),
			"credits":           strconv.Itoa(requestBody.AmountInDollars), // 1 dollar = 1 credit
		},
	})
	if err != nil {
		log.Printf("[PurchaseCredits] Error creating checkout session: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Printf("[PurchaseCredits] Checkout session created: %s", checkoutSession.URL)

	ctx.JSON(http.StatusOK, gin.H{
		"checkoutUrl": checkoutSession.URL,
	})
}

func (s *server) HandleCreditsWebhook(ctx *gin.Context) {
	log.Printf("[HandleCreditsWebhook] Webhook received")

	var requestBody billing.OnPaymentComplete

	if err := ctx.Bind(&requestBody); err != nil {
		log.Printf("[HandleCreditsWebhook] Error binding request: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Printf("[HandleCreditsWebhook] Session ID: %s, Amount: %d cents, Metadata: %+v",
		requestBody.SessionID, requestBody.AmountInCents, requestBody.Metadata)

	// Extract user ID from metadata
	userIDStr, ok := requestBody.Metadata["user_id"]
	if !ok {
		log.Printf("[HandleCreditsWebhook] ERROR: user_id not found in metadata")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id not found in metadata",
		})
		return
	}

	log.Printf("[HandleCreditsWebhook] User ID from metadata: %s", userIDStr)

	// Parse user ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("[HandleCreditsWebhook] ERROR: invalid user_id format: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user_id",
		})
		return
	}

	// Extract credits amount from metadata (1 dollar = 1 credit)
	creditsStr, ok := requestBody.Metadata["credits"]
	if !ok {
		// Fallback: calculate from amount in cents (1 dollar = 1 credit)
		creditsStr = strconv.FormatInt(requestBody.AmountInCents/100, 10)
		log.Printf("[HandleCreditsWebhook] Credits not in metadata, calculated from amount: %s", creditsStr)
	} else {
		log.Printf("[HandleCreditsWebhook] Credits from metadata: %s", creditsStr)
	}

	credits, err := strconv.Atoi(creditsStr)
	if err != nil {
		log.Printf("[HandleCreditsWebhook] ERROR: invalid credits amount: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid credits amount",
		})
		return
	}

	log.Printf("[HandleCreditsWebhook] Adding %d credits to user %s", credits, userID)

	// Add credits to user
	if err := s.dbClient.User().AddCredits(ctx, userID, credits); err != nil {
		log.Printf("[HandleCreditsWebhook] ERROR: failed to add credits: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Fetch updated user to verify credits were added
	updatedUser, err := s.dbClient.User().FindByID(ctx, userID)
	if err != nil {
		log.Printf("[HandleCreditsWebhook] WARNING: failed to fetch updated user: %v", err)
	} else {
		log.Printf("[HandleCreditsWebhook] Success! User %s now has %d credits", userID, updatedUser.Credits)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "credits added successfully",
	})
}

func (s *server) AddCredits(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var requestBody struct {
		Amount int `json:"amount" binding:"required"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if requestBody.Amount < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "amount must be positive",
		})
		return
	}

	if err := s.dbClient.User().AddCredits(ctx, user.ID, requestBody.Amount); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Fetch updated user
	updatedUser, err := s.dbClient.User().FindByID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, updatedUser)
}

func (s *server) SubtractCredits(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var requestBody struct {
		Amount int `json:"amount" binding:"required"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if requestBody.Amount < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "amount must be positive",
		})
		return
	}

	// Check if user has enough credits
	currentUser, err := s.dbClient.User().FindByID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if currentUser.Credits < requestBody.Amount {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "insufficient credits",
		})
		return
	}

	if err := s.dbClient.User().SubtractCredits(ctx, user.ID, requestBody.Amount); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Fetch updated user
	updatedUser, err := s.dbClient.User().FindByID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, updatedUser)
}
