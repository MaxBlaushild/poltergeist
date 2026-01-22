package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type TransactionStatus struct {
	Status      string     // "pending", "confirmed", "failed"
	BlockNumber *uint64    // Block number if confirmed
	ConfirmedAt *time.Time // Confirmation time if confirmed
}

type EthereumClient interface {
	SendTransaction(ctx context.Context, to *common.Address, value *big.Int, data []byte, gasLimit uint64, gasPrice *big.Int, nonce uint64) (common.Hash, error)
	EstimateGas(ctx context.Context, to *common.Address, value *big.Int, data []byte) (uint64, error)
	GetTransactionStatus(ctx context.Context, txHash common.Hash) (*TransactionStatus, error)
	GetPendingNonce(ctx context.Context, address common.Address) (uint64, error)
	GetAddress() common.Address
}

type client struct {
	ethClient  *ethclient.Client
	privateKey *ecdsa.PrivateKey
	chainID    *big.Int
	address    common.Address
}

func NewClient(rpcURL string, privateKeyHex string, chainID int64) (EthereumClient, error) {
	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	var privateKey *ecdsa.PrivateKey
	var address common.Address

	if privateKeyHex != "" {
		// Strip "0x" prefix if present
		privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
		key, err := crypto.HexToECDSA(privateKeyHex)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		privateKey = key

		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("failed to get public key from private key")
		}

		address = crypto.PubkeyToAddress(*publicKeyECDSA)
	}

	return &client{
		ethClient:  ethClient,
		privateKey: privateKey,
		chainID:    big.NewInt(chainID),
		address:    address,
	}, nil
}

// NewReadOnlyClient creates a read-only Ethereum client (no private key needed)
func NewReadOnlyClient(rpcURL string, chainID int64) (EthereumClient, error) {
	return NewClient(rpcURL, "", chainID)
}

func (c *client) GetAddress() common.Address {
	return c.address
}

func (c *client) SendTransaction(ctx context.Context, to *common.Address, value *big.Int, data []byte, gasLimit uint64, gasPrice *big.Int, nonce uint64) (common.Hash, error) {
	if c.privateKey == nil {
		return common.Hash{}, fmt.Errorf("private key not set, cannot send transactions")
	}

	var txData types.TxData

	if gasPrice == nil {
		// Get suggested gas price
		suggestedGasPrice, err := c.ethClient.SuggestGasPrice(ctx)
		if err != nil {
			return common.Hash{}, err
		}
		gasPrice = suggestedGasPrice
	}

	if to == nil {
		// Contract creation transaction
		txData = &types.LegacyTx{
			Nonce:    nonce,
			GasPrice: gasPrice,
			Gas:      gasLimit,
			Value:    value,
			Data:     data,
		}
	} else {
		txData = &types.LegacyTx{
			Nonce:    nonce,
			To:       to,
			GasPrice: gasPrice,
			Gas:      gasLimit,
			Value:    value,
			Data:     data,
		}
	}

	tx := types.NewTx(txData)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(c.chainID), c.privateKey)
	if err != nil {
		return common.Hash{}, err
	}

	err = c.ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return common.Hash{}, err
	}

	return signedTx.Hash(), nil
}

func (c *client) GetTransactionStatus(ctx context.Context, txHash common.Hash) (*TransactionStatus, error) {
	_, isPending, err := c.ethClient.TransactionByHash(ctx, txHash)
	if err != nil {
		// Transaction might not exist yet
		return &TransactionStatus{
			Status: "pending",
		}, nil
	}

	if isPending {
		return &TransactionStatus{
			Status: "pending",
		}, nil
	}

	// Transaction is in a block, check receipt
	receipt, err := c.ethClient.TransactionReceipt(ctx, txHash)
	if err != nil {
		return &TransactionStatus{
			Status: "pending",
		}, nil
	}

	confirmedAt := time.Now()
	if receipt.BlockNumber != nil {
		// Try to get block timestamp
		block, err := c.ethClient.BlockByNumber(ctx, receipt.BlockNumber)
		if err == nil && block != nil && block.Time() > 0 {
			confirmedAt = time.Unix(int64(block.Time()), 0)
		}
	}

	blockNumber := receipt.BlockNumber.Uint64()

	if receipt.Status == 0 {
		return &TransactionStatus{
			Status:      "failed",
			BlockNumber: &blockNumber,
			ConfirmedAt: &confirmedAt,
		}, nil
	}

	return &TransactionStatus{
		Status:      "confirmed",
		BlockNumber: &blockNumber,
		ConfirmedAt: &confirmedAt,
	}, nil
}

func (c *client) EstimateGas(ctx context.Context, to *common.Address, value *big.Int, data []byte) (uint64, error) {
	msg := ethereum.CallMsg{
		From:  c.address, // Set From address so the RPC can properly simulate the transaction
		To:    to,
		Value: value,
		Data:  data,
	}

	gasLimit, err := c.ethClient.EstimateGas(ctx, msg)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Add a 20% buffer to the estimated gas to account for variability
	gasLimit = gasLimit + (gasLimit * 20 / 100)

	return gasLimit, nil
}

func (c *client) GetPendingNonce(ctx context.Context, address common.Address) (uint64, error) {
	nonce, err := c.ethClient.PendingNonceAt(ctx, address)
	if err != nil {
		return 0, err
	}
	return nonce, nil
}
