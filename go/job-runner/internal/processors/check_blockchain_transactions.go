package processors

import (
	"context"
	"fmt"
	"log"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/ethereum"
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
