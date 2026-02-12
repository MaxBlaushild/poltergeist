package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client interface {
	ListTrades(ctx context.Context, since *time.Time, limit int) ([]Trade, error)
}

type ClientConfig struct {
	BaseURL    string
	TradesPath string
	TradesURL  string
	APIKey     string
	Timeout    time.Duration
}

type client struct {
	baseURL    string
	tradesPath string
	tradesURL  string
	apiKey     string
	httpClient *http.Client
}

type Trade struct {
	ExternalID string
	MarketID   string
	MarketName string
	Outcome    string
	Side       string
	Price      float64
	Size       float64
	Trader     string
	TradeTime  time.Time
	Raw        json.RawMessage
}

func NewClient(cfg ClientConfig) Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	tradesPath := strings.TrimSpace(cfg.TradesPath)
	if tradesPath == "" {
		tradesPath = "/trades"
	}
	return &client{
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		tradesPath: tradesPath,
		tradesURL:  strings.TrimSpace(cfg.TradesURL),
		apiKey:     strings.TrimSpace(cfg.APIKey),
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *client) ListTrades(ctx context.Context, since *time.Time, limit int) ([]Trade, error) {
	reqURL, err := c.buildTradesURL(since, limit)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("polymarket trades request failed: %s", resp.Status)
	}

	trades, err := parseTrades(body)
	if err != nil {
		return nil, err
	}
	return trades, nil
}

func (c *client) buildTradesURL(since *time.Time, limit int) (string, error) {
	var base string
	if c.tradesURL != "" {
		base = c.tradesURL
	} else if c.baseURL != "" {
		base = c.baseURL + c.tradesPath
	} else {
		return "", fmt.Errorf("polymarket trades url not configured")
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	q := parsed.Query()
	if since != nil {
		q.Set("since", since.Format(time.RFC3339))
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	parsed.RawQuery = q.Encode()
	return parsed.String(), nil
}

func parseTrades(payload []byte) ([]Trade, error) {
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}

	var tradeObjs []map[string]any
	switch v := raw.(type) {
	case []any:
		tradeObjs = make([]map[string]any, 0, len(v))
		for _, item := range v {
			if obj, ok := item.(map[string]any); ok {
				tradeObjs = append(tradeObjs, obj)
			}
		}
	case map[string]any:
		if data, ok := v["data"]; ok {
			if arr, ok := data.([]any); ok {
				for _, item := range arr {
					if obj, ok := item.(map[string]any); ok {
						tradeObjs = append(tradeObjs, obj)
					}
				}
			}
		}
	default:
		return nil, fmt.Errorf("unexpected trades payload format")
	}

	trades := make([]Trade, 0, len(tradeObjs))
	for _, obj := range tradeObjs {
		rawBytes, _ := json.Marshal(obj)
		trade := Trade{
			ExternalID: firstString(obj, "id", "trade_id", "tradeId"),
			MarketID:   firstString(obj, "market_id", "marketId", "market"),
			MarketName: firstString(obj, "market_name", "marketName", "title"),
			Outcome:    firstString(obj, "outcome", "side_outcome", "outcomeName"),
			Side:       firstString(obj, "side", "direction"),
			Price:      firstFloat(obj, "price", "avg_price", "execution_price"),
			Size:       firstFloat(obj, "size", "amount", "quantity", "shares"),
			Trader:     firstString(obj, "trader", "trader_address", "maker", "taker"),
			TradeTime:  firstTime(obj, "timestamp", "created_at", "createdAt", "time"),
			Raw:        rawBytes,
		}
		trades = append(trades, trade)
	}

	return trades, nil
}

func firstString(obj map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := obj[key]; ok {
			switch v := value.(type) {
			case string:
				if v != "" {
					return v
				}
			case fmt.Stringer:
				return v.String()
			}
		}
	}
	return ""
}

func firstFloat(obj map[string]any, keys ...string) float64 {
	for _, key := range keys {
		if value, ok := obj[key]; ok {
			switch v := value.(type) {
			case float64:
				return v
			case float32:
				return float64(v)
			case int:
				return float64(v)
			case int64:
				return float64(v)
			case json.Number:
				if parsed, err := v.Float64(); err == nil {
					return parsed
				}
			case string:
				if parsed, err := strconv.ParseFloat(v, 64); err == nil {
					return parsed
				}
			}
		}
	}
	return 0
}

func firstTime(obj map[string]any, keys ...string) time.Time {
	for _, key := range keys {
		if value, ok := obj[key]; ok {
			switch v := value.(type) {
			case string:
				if t, err := time.Parse(time.RFC3339, v); err == nil {
					return t
				}
			case float64:
				return timeFromUnix(v)
			case int64:
				return time.Unix(v, 0).UTC()
			case json.Number:
				if parsed, err := v.Int64(); err == nil {
					return time.Unix(parsed, 0).UTC()
				}
			}
		}
	}
	return time.Time{}
}

func timeFromUnix(value float64) time.Time {
	if value > 1e12 {
		return time.UnixMilli(int64(value)).UTC()
	}
	return time.Unix(int64(value), 0).UTC()
}
