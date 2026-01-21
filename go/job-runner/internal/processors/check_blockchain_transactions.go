package processors

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/ethereum"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hibiken/asynq"
)

type CheckBlockchainTransactionsProcessor struct {
	dbClient       db.DbClient
	ethereumClient ethereum.EthereumClient
}

func NewCheckBlockchainTransactionsProcessor(dbClient db.DbClient, ethereumClient ethereum.EthereumClient) CheckBlockchainTransactionsProcessor {
	log.Println("Initializing CheckBlockchainTransactionsProcessor")
	return CheckBlockchainTransactionsProcessor{
		dbClient:       dbClient,
		ethereumClient: ethereumClient,
	}
}

func (p *CheckBlockchainTransactionsProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing check blockchain transactions task: %v", task.Type())

	return p.checkBlockchainTransactions(ctx)
}

func (p *CheckBlockchainTransactionsProcessor) checkBlockchainTransactions(ctx context.Context) error {
	log.Printf("Checking pending blockchain transactions")

	// Find all pending transactions
	pendingTxs, err := p.dbClient.BlockchainTransaction().FindPending(ctx)
	if err != nil {
		log.Printf("Failed to find pending transactions: %v", err)
		return fmt.Errorf("failed to find pending transactions: %w", err)
	}

	log.Printf("Found %d pending transactions", len(pendingTxs))

	// Check status of each pending transaction
	for _, tx := range pendingTxs {
		if tx.TxHash == nil {
			log.Printf("Transaction %s has no tx hash, skipping", tx.ID)
			continue
		}

		txHash := common.HexToHash(*tx.TxHash)
		status, err := p.ethereumClient.GetTransactionStatus(ctx, txHash)
		if err != nil {
			log.Printf("Failed to get transaction status for %s: %v", *tx.TxHash, err)
			continue
		}

		if status.Status == "confirmed" {
			log.Printf("Transaction %s confirmed in block %d", *tx.TxHash, *status.BlockNumber)
			err = p.dbClient.BlockchainTransaction().UpdateStatus(ctx, tx.ID, "confirmed", status.BlockNumber, status.ConfirmedAt)
			if err != nil {
				log.Printf("Failed to update transaction status: %v", err)
			}

			// Check if this is a registerCertificate transaction by function selector or type
			isRegisterCert := false
			if tx.Type != nil && *tx.Type == "registerCertificate" {
				isRegisterCert = true
			} else if tx.Data != nil {
				// Check function selector: registerCertificate(bytes32,string,string)
				// Function selector is keccak256("registerCertificate(bytes32,string,string)")[:4]
				// = 0x8f4ffcb1 (pre-computed)
				registerCertSelector := "8f4ffcb1"
				dataHex := strings.TrimPrefix(*tx.Data, "0x")
				if len(dataHex) >= 8 && strings.HasPrefix(strings.ToLower(dataHex), registerCertSelector) {
					isRegisterCert = true
				}
			}

			if isRegisterCert {
				err = p.activateCertificateFromTransaction(ctx, &tx)
				if err != nil {
					log.Printf("Failed to activate certificate from transaction %s: %v", tx.ID, err)
					// Don't return error - transaction is confirmed, certificate activation can be retried
				}
			}
		} else if status.Status == "failed" {
			log.Printf("Transaction %s failed", *tx.TxHash)
			err = p.dbClient.BlockchainTransaction().UpdateStatus(ctx, tx.ID, "failed", status.BlockNumber, status.ConfirmedAt)
			if err != nil {
				log.Printf("Failed to update transaction status: %v", err)
			}
		}
		// If still pending, leave it as is
	}

	// Find and mark expired transactions
	expiredTxs, err := p.dbClient.BlockchainTransaction().FindPendingExpired(ctx)
	if err != nil {
		log.Printf("Failed to find expired transactions: %v", err)
		return fmt.Errorf("failed to find expired transactions: %w", err)
	}

	log.Printf("Found %d expired transactions", len(expiredTxs))

	for _, tx := range expiredTxs {
		log.Printf("Marking transaction %s as expired", tx.ID)
		err = p.dbClient.BlockchainTransaction().UpdateStatus(ctx, tx.ID, "expired", nil, nil)
		if err != nil {
			log.Printf("Failed to mark transaction as expired: %v", err)
		}
	}

	log.Printf("Finished checking blockchain transactions")
	return nil
}

// activateCertificateFromTransaction extracts the fingerprint from a registerCertificate transaction
// and activates the corresponding certificate
func (p *CheckBlockchainTransactionsProcessor) activateCertificateFromTransaction(ctx context.Context, tx *models.BlockchainTransaction) error {
	if tx.Data == nil {
		return fmt.Errorf("transaction has no data")
	}

	// Remove "0x" prefix if present
	dataHex := strings.TrimPrefix(*tx.Data, "0x")
	data, err := hex.DecodeString(dataHex)
	if err != nil {
		return fmt.Errorf("failed to decode transaction data: %w", err)
	}

	// The function selector is the first 4 bytes, followed by the encoded parameters
	// For registerCertificate(bytes32,string,string), the fingerprint is the first parameter
	// ABI encoding: function selector (4 bytes) + bytes32 (32 bytes) + offset to string data + ...
	if len(data) < 4+32 {
		return fmt.Errorf("transaction data too short")
	}

	// Extract fingerprint (bytes32) - it's at offset 4 (after function selector)
	var fingerprint [32]byte
	copy(fingerprint[:], data[4:36])

	// Find certificate by fingerprint
	cert, err := p.dbClient.UserCertificate().FindByFingerprint(ctx, fingerprint[:])
	if err != nil {
		return fmt.Errorf("failed to find certificate by fingerprint: %w", err)
	}

	if cert == nil {
		return fmt.Errorf("certificate not found for fingerprint")
	}

	// Activate the certificate
	err = p.dbClient.UserCertificate().UpdateActive(ctx, cert.UserID, true)
	if err != nil {
		return fmt.Errorf("failed to activate certificate: %w", err)
	}

	log.Printf("Activated certificate for user %s (fingerprint: %x)", cert.UserID, fingerprint)
	return nil
}
