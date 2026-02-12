CREATE TABLE insider_trades (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    external_id TEXT NOT NULL UNIQUE,
    market_id TEXT,
    market_name TEXT,
    outcome TEXT,
    side TEXT,
    price NUMERIC,
    size NUMERIC,
    notional NUMERIC,
    trader TEXT,
    trade_time TIMESTAMP,
    detected_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reason TEXT,
    raw JSONB
);

CREATE INDEX idx_insider_trades_detected_at ON insider_trades(detected_at DESC);
CREATE INDEX idx_insider_trades_market_id ON insider_trades(market_id);
