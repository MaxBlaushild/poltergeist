package models

import (
	"time"

	"github.com/google/uuid"
)

type BlockchainTransactionType string

const (
	RegisterCertificateType BlockchainTransactionType = "registerCertificate"
	AnchorManifestType      BlockchainTransactionType = "anchorManifest"
)

type BlockchainTransactionStatus string

const (
	PendingStatus   BlockchainTransactionStatus = "pending"
	ConfirmedStatus BlockchainTransactionStatus = "confirmed"
	FailedStatus    BlockchainTransactionStatus = "failed"
	ExpiredStatus   BlockchainTransactionStatus = "expired"
)

type BlockchainTransaction struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt   time.Time  `gorm:"not null" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"not null" json:"updatedAt"`
	ChainID     int64      `gorm:"type:bigint;not null;index:idx_chain_from_nonce" json:"chainId"`
	FromAddress string     `gorm:"type:text;not null;index:idx_chain_from_nonce" json:"fromAddress"`
	ToAddress   *string    `gorm:"type:text" json:"toAddress,omitempty"`
	Value       string     `gorm:"type:text;not null" json:"value"`
	Data        *string    `gorm:"type:text" json:"data,omitempty"`
	GasLimit    *uint64    `gorm:"type:bigint" json:"gasLimit,omitempty"`
	GasPrice    *string    `gorm:"type:text" json:"gasPrice,omitempty"`
	Nonce       uint64     `gorm:"type:bigint;not null;index:idx_chain_from_nonce" json:"nonce"`
	TxHash      *string    `gorm:"type:text;index" json:"txHash,omitempty"`
	Status      string     `gorm:"type:text;not null;default:'pending';index" json:"status"`
	Type        *string    `gorm:"type:text;index" json:"type,omitempty"`
	BlockNumber *uint64    `gorm:"type:bigint" json:"blockNumber,omitempty"`
	ConfirmedAt *time.Time `gorm:"type:timestamp" json:"confirmedAt,omitempty"`
	ExpiresAt   time.Time  `gorm:"type:timestamp;not null;index" json:"expiresAt"`
}
