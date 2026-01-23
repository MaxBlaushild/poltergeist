package db

import (
	"context"
	"database/sql"
	"encoding/hex"
	"strings"
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

// FindByCertificateFingerprint finds the blockchain transaction that registered a certificate
// by extracting the fingerprint from transaction data and matching it
func (h *blockchainTransactionHandle) FindByCertificateFingerprint(ctx context.Context, fingerprint []byte) (*models.BlockchainTransaction, error) {
	// Query all transactions with type "registerCertificate"
	var txs []models.BlockchainTransaction
	registerCertType := string(models.RegisterCertificateType)
	if err := h.db.WithContext(ctx).
		Where("type = ?", registerCertType).
		Order("block_number DESC NULLS LAST, created_at DESC").
		Find(&txs).Error; err != nil {
		return nil, err
	}

	// Iterate through transactions and extract fingerprint from data
	for _, tx := range txs {
		if tx.Data == nil {
			continue
		}

		// Remove "0x" prefix if present
		dataHex := strings.TrimPrefix(*tx.Data, "0x")
		data, err := hex.DecodeString(dataHex)
		if err != nil {
			continue // Skip invalid hex data
		}

		// The function selector is the first 4 bytes, followed by the encoded parameters
		// For registerCertificate(bytes32,string,string), the fingerprint is the first parameter
		// ABI encoding: function selector (4 bytes) + bytes32 (32 bytes) + offset to string data + ...
		if len(data) < 4+32 {
			continue // Skip if data is too short
		}

		// Extract fingerprint (bytes32) - it's at offset 4 (after function selector)
		var extractedFingerprint [32]byte
		copy(extractedFingerprint[:], data[4:36])

		// Compare with provided fingerprint
		if len(fingerprint) == 32 {
			match := true
			for i := 0; i < 32; i++ {
				if extractedFingerprint[i] != fingerprint[i] {
					match = false
					break
				}
			}
			if match {
				return &tx, nil
			}
		} else {
			// Handle case where fingerprint might be shorter (shouldn't happen, but be safe)
			if len(fingerprint) <= 32 {
				match := true
				for i := 0; i < len(fingerprint); i++ {
					if extractedFingerprint[i] != fingerprint[i] {
						match = false
						break
					}
				}
				if match {
					return &tx, nil
				}
			}
		}
	}

	// No matching transaction found
	return nil, nil
}

// FindByManifestHash finds the blockchain transaction that anchored a manifest
// by extracting the manifest hash from transaction data and matching it
func (h *blockchainTransactionHandle) FindByManifestHash(ctx context.Context, manifestHash []byte) (*models.BlockchainTransaction, error) {
	// Query all transactions with type "anchorManifest"
	var txs []models.BlockchainTransaction
	anchorManifestType := string(models.AnchorManifestType)
	if err := h.db.WithContext(ctx).
		Where("type = ?", anchorManifestType).
		Order("block_number DESC NULLS LAST, created_at DESC").
		Find(&txs).Error; err != nil {
		return nil, err
	}

	// Iterate through transactions and extract manifest hash from data
	for _, tx := range txs {
		if tx.Data == nil {
			continue
		}

		// Remove "0x" prefix if present
		dataHex := strings.TrimPrefix(*tx.Data, "0x")
		data, err := hex.DecodeString(dataHex)
		if err != nil {
			continue // Skip invalid hex data
		}

		// The function selector is the first 4 bytes, followed by the encoded parameters
		// For anchorManifest(bytes32,string,string,bytes32), the manifest hash is the first parameter
		// ABI encoding: function selector (4 bytes) + bytes32 (32 bytes) + offset to string data + ...
		if len(data) < 4+32 {
			continue // Skip if data is too short
		}

		// Extract manifest hash (bytes32) - it's at offset 4 (after function selector)
		var extractedManifestHash [32]byte
		copy(extractedManifestHash[:], data[4:36])

		// Compare with provided manifest hash
		if len(manifestHash) == 32 {
			match := true
			for i := 0; i < 32; i++ {
				if extractedManifestHash[i] != manifestHash[i] {
					match = false
					break
				}
			}
			if match {
				return &tx, nil
			}
		} else {
			// Handle case where manifest hash might be shorter (shouldn't happen, but be safe)
			if len(manifestHash) <= 32 {
				match := true
				for i := 0; i < len(manifestHash); i++ {
					if extractedManifestHash[i] != manifestHash[i] {
						match = false
						break
					}
				}
				if match {
					return &tx, nil
				}
			}
		}
	}

	// No matching transaction found
	return nil, nil
}
