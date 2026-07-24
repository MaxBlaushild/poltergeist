package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

type operatorMetricsResponse struct {
	SinceDays int `json:"sinceDays"`

	// R-9.2: "configurator-to-purchase conversion"
	ConfiguratorOpened int64   `json:"configuratorOpened"`
	PurchaseCompleted  int64   `json:"purchaseCompleted"`
	ConversionRate     float64 `json:"conversionRate"`

	// "validation rejection rate by rule"
	ValidRejectedTotal int64            `json:"validatedTotal"`
	RejectedTotal      int64            `json:"rejectedTotal"`
	RejectionRate      float64          `json:"rejectionRate"`
	RejectionsByRule   map[string]int64 `json:"rejectionsByRule"`

	// "mean landed COGS per order"
	PaidOrderCount  int64   `json:"paidOrderCount"`
	CogsRecordedFor int64   `json:"cogsRecordedForOrders"`
	MeanCogsCents   float64 `json:"meanCogsCents"`

	// "reprint rate"
	OrdersReprinted int64   `json:"ordersReprinted"`
	ReprintRate     float64 `json:"reprintRate"`

	// "CAC when ad spend is entered manually"
	AdSpendCents int64   `json:"adSpendCents"`
	CACCents     float64 `json:"cacCents"`
}

// GET /api/reef/operator/metrics (R-9.2). This is the one place all four
// go/no-go numbers for the vertical live — configurator-to-purchase
// conversion, validation rejection rate by rule, mean landed COGS, and CAC.
// Query params: days (window, default 30), adSpendCents (manual entry, R-9.2).
func (s *server) getOperatorMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	days := 30
	if raw := c.Query("days"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			days = parsed
		}
	}
	since := time.Now().AddDate(0, 0, -days)

	var adSpendCents int64
	if raw := c.Query("adSpendCents"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil && parsed >= 0 {
			adSpendCents = parsed
		}
	}

	opened, err := s.deps.DbClient.ReefEvent().CountByType(ctx, models.ReefEventConfiguratorOpened, since)
	if err != nil {
		internalError(c, "count configurator_opened events", err)
		return
	}
	purchased, err := s.deps.DbClient.ReefEvent().CountByType(ctx, models.ReefEventPurchaseCompleted, since)
	if err != nil {
		internalError(c, "count purchase_completed events", err)
		return
	}

	validCount, err := s.deps.DbClient.ReefConfiguration().CountByStatusSince(ctx, models.ReefConfigurationStatusValid, since)
	if err != nil {
		internalError(c, "count valid configurations", err)
		return
	}
	rejectedCount, err := s.deps.DbClient.ReefConfiguration().CountByStatusSince(ctx, models.ReefConfigurationStatusRejected, since)
	if err != nil {
		internalError(c, "count rejected configurations", err)
		return
	}
	rejectionRows, err := s.deps.DbClient.ReefEvent().CountRejectionsByRule(ctx, since)
	if err != nil {
		internalError(c, "count rejections by rule", err)
		return
	}

	cogsStats, err := s.deps.DbClient.ReefOrder().CogsStatsSince(ctx, since)
	if err != nil {
		internalError(c, "compute cogs stats", err)
		return
	}

	resp := operatorMetricsResponse{
		SinceDays:          days,
		ConfiguratorOpened: opened,
		PurchaseCompleted:  purchased,
		ConversionRate:     ratio(purchased, opened),

		ValidRejectedTotal: validCount + rejectedCount,
		RejectedTotal:      rejectedCount,
		RejectionRate:      ratio(rejectedCount, validCount+rejectedCount),
		RejectionsByRule:   map[string]int64{},

		PaidOrderCount:  cogsStats.OrderCount,
		CogsRecordedFor: cogsStats.CogsRecorded,
		MeanCogsCents:   cogsStats.MeanCogsCents,

		OrdersReprinted: cogsStats.OrdersReprinted,
		ReprintRate:     ratio(cogsStats.OrdersReprinted, cogsStats.OrderCount),

		AdSpendCents: adSpendCents,
		CACCents:     costPerAcquisition(adSpendCents, purchased),
	}
	for _, row := range rejectionRows {
		resp.RejectionsByRule[row.Rule] = row.Count
	}

	c.JSON(http.StatusOK, resp)
}

func ratio(numerator, denominator int64) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func costPerAcquisition(adSpendCents, purchases int64) float64 {
	if purchases == 0 {
		return 0
	}
	return float64(adSpendCents) / float64(purchases)
}
