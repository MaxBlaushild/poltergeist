package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type blockchainTransactionHandle struct {
	db *gorm.DB
}

func (h *blockchainTransactionHandle) Create(ctx context.Context, tx *models.BlockchainTransaction) (*models.BlockchainTransaction, error) {
	tx.CreatedAt = time.Now()
	tx.UpdatedAt = time.Now()
	if tx.ExpiresAt.IsZero() {
		tx.ExpiresAt = time.Now().Add(24 * time.Hour)
	}
	if err := h.db.WithContext(ctx).Create(tx).Error; err != nil {
		return nil, err
	}
	return tx, nil
}

func (h *blockchainTransactionHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.BlockchainTransaction, error) {
	var tx models.BlockchainTransaction
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&tx).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

func (h *blockchainTransactionHandle) FindByTxHash(ctx context.Context, txHash string) (*models.BlockchainTransaction, error) {
	var tx models.BlockchainTransaction
	if err := h.db.WithContext(ctx).Where("tx_hash = ?", txHash).First(&tx).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

func (h *blockchainTransactionHandle) FindPending(ctx context.Context) ([]models.BlockchainTransaction, error) {
	var txs []models.BlockchainTransaction
	now := time.Now()
	if err := h.db.WithContext(ctx).
		Where("status = ? AND expires_at > ?", "pending", now).
		Order("created_at ASC").
		Find(&txs).Error; err != nil {
		return nil, err
	}
	return txs, nil
}

func (h *blockchainTransactionHandle) FindPendingExpired(ctx context.Context) ([]models.BlockchainTransaction, error) {
	var txs []models.BlockchainTransaction
	now := time.Now()
	if err := h.db.WithContext(ctx).
		Where("status = ? AND expires_at <= ?", "pending", now).
		Find(&txs).Error; err != nil {
		return nil, err
	}
	return txs, nil
}

func (h *blockchainTransactionHandle) UpdateStatus(ctx context.Context, id uuid.UUID, status string, blockNumber *uint64, confirmedAt *time.Time) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if blockNumber != nil {
		updates["block_number"] = *blockNumber
	}
	if confirmedAt != nil {
		updates["confirmed_at"] = *confirmedAt
	}
	return h.db.WithContext(ctx).Model(&models.BlockchainTransaction{}).Where("id = ?", id).Updates(updates).Error
}

func (h *blockchainTransactionHandle) GetNextNonce(ctx context.Context, chainID int64, fromAddress string) (uint64, error) {
	var maxNonce sql.NullInt64
	err := h.db.WithContext(ctx).
		Model(&models.BlockchainTransaction{}).
		Where("chain_id = ? AND from_address = ?", chainID, fromAddress).
		Select("COALESCE(MAX(nonce), -1)").
		Scan(&maxNonce).Error

	if err != nil {
		return 0, err
	}

	if !maxNonce.Valid || maxNonce.Int64 < 0 {
		return 0, nil
	}

	return uint64(maxNonce.Int64) + 1, nil
}
