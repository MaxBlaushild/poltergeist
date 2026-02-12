package processors

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/polymarket"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	"github.com/hibiken/asynq"
	"gorm.io/datatypes"
)

type MonitorPolymarketTradesProcessor struct {
	dbClient         db.DbClient
	polymarketClient polymarket.Client
	texterClient     texter.Client
	alertTo          string
	alertFrom        string
	notionalThreshold float64
	sizeThreshold     float64
	tradesLimit       int
}

func NewMonitorPolymarketTradesProcessor(
	dbClient db.DbClient,
	polymarketClient polymarket.Client,
	texterClient texter.Client,
	alertTo string,
	alertFrom string,
	notionalThreshold float64,
	sizeThreshold float64,
	tradesLimit int,
) *MonitorPolymarketTradesProcessor {
	log.Println("Initializing MonitorPolymarketTradesProcessor")
	return &MonitorPolymarketTradesProcessor{
		dbClient:          dbClient,
		polymarketClient: polymarketClient,
		texterClient:     texterClient,
		alertTo:          alertTo,
		alertFrom:        alertFrom,
		notionalThreshold: notionalThreshold,
		sizeThreshold:     sizeThreshold,
		tradesLimit:       tradesLimit,
	}
}

func (p *MonitorPolymarketTradesProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing polymarket trades monitor task: %v", task.Type())
	return p.monitor(ctx)
}

func (p *MonitorPolymarketTradesProcessor) monitor(ctx context.Context) error {
	if p.polymarketClient == nil {
		log.Printf("Polymarket client not configured; skipping")
		return nil
	}

	latest, err := p.dbClient.InsiderTrade().LatestTradeTime(ctx)
	if err != nil {
		log.Printf("Failed to fetch latest trade time: %v", err)
		return err
	}

	var since *time.Time
	if latest != nil {
		since = latest
	} else {
		fallback := time.Now().Add(-10 * time.Minute)
		since = &fallback
	}

	trades, err := p.polymarketClient.ListTrades(ctx, since, p.tradesLimit)
	if err != nil {
		log.Printf("Failed to fetch polymarket trades: %v", err)
		return err
	}

	if len(trades) == 0 {
		log.Printf("No polymarket trades returned")
		return nil
	}

	log.Printf("Fetched %d polymarket trades", len(trades))

	for _, trade := range trades {
		if trade.ExternalID == "" {
			continue
		}
		tradeTime := trade.TradeTime
		if tradeTime.IsZero() {
			tradeTime = time.Now().UTC()
		}
		if trade.Size == 0 && trade.Price == 0 {
			// Not enough info to score, skip for now.
			continue
		}

		notional := trade.Size * trade.Price
		reason := p.detectReason(trade.Size, notional)
		if reason == "" {
			continue
		}

		insiderTrade := &models.InsiderTrade{
			ExternalID: trade.ExternalID,
			MarketID:   trade.MarketID,
			MarketName: trade.MarketName,
			Outcome:    trade.Outcome,
			Side:       trade.Side,
			Price:      trade.Price,
			Size:       trade.Size,
			Notional:   notional,
			Trader:     trade.Trader,
			TradeTime:  tradeTime,
			DetectedAt: time.Now().UTC(),
			Reason:     reason,
			Raw:        datatypes.JSON(trade.Raw),
		}

		inserted, err := p.dbClient.InsiderTrade().Upsert(ctx, insiderTrade)
		if err != nil {
			log.Printf("Failed to persist suspicious trade %s: %v", trade.ExternalID, err)
			continue
		}
		if !inserted {
			continue
		}

		p.sendAlert(ctx, insiderTrade)
	}

	return nil
}

func (p *MonitorPolymarketTradesProcessor) detectReason(size float64, notional float64) string {
	if p.sizeThreshold > 0 && size >= p.sizeThreshold {
		return fmt.Sprintf("size %.2f >= %.2f", size, p.sizeThreshold)
	}
	if p.notionalThreshold > 0 && notional >= p.notionalThreshold {
		return fmt.Sprintf("notional %.2f >= %.2f", notional, p.notionalThreshold)
	}
	return ""
}

func (p *MonitorPolymarketTradesProcessor) sendAlert(ctx context.Context, trade *models.InsiderTrade) {
	if p.texterClient == nil {
		return
	}
	if p.alertTo == "" || p.alertFrom == "" {
		log.Printf("Alert phone numbers not configured; skipping SMS")
		return
	}

	body := fmt.Sprintf(
		"Suspicious Polymarket trade detected. Market: %s (%s). Side: %s. Outcome: %s. Size: %.2f. Price: %.4f. Reason: %s",
		trade.MarketName,
		trade.MarketID,
		trade.Side,
		trade.Outcome,
		trade.Size,
		trade.Price,
		trade.Reason,
	)

	if err := p.texterClient.Text(ctx, &texter.Text{
		Body:     body,
		From:     p.alertFrom,
		To:       p.alertTo,
		TextType: "polymarket-suspicious-trade",
	}); err != nil {
		log.Printf("Failed to send SMS alert: %v", err)
	}
}
