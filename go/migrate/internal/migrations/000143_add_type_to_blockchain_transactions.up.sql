ALTER TABLE blockchain_transactions ADD COLUMN type TEXT;

CREATE INDEX idx_blockchain_transactions_type ON blockchain_transactions(type);
