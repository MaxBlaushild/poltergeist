package server

import (
	"encoding/hex"
	"math/big"
	"net/http"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

func (s *server) CreateTransaction(ctx *gin.Context) {
	var requestBody struct {
		To       *string `json:"to"`
		Value    string  `json:"value" binding:"required"`
		Data     *string `json:"data"`
		GasLimit *uint64 `json:"gasLimit"`
		GasPrice *string `json:"gasPrice"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Parse value
	value, ok := new(big.Int).SetString(requestBody.Value, 10)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid value format",
		})
		return
	}

	// Parse data if provided
	var data []byte
	if requestBody.Data != nil {
		var err error
		data, err = hex.DecodeString((*requestBody.Data)[2:]) // Remove "0x" prefix
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid data format: " + err.Error(),
			})
			return
		}
	}

	// Parse gas price if provided
	var gasPrice *big.Int
	if requestBody.GasPrice != nil {
		var ok bool
		gasPrice, ok = new(big.Int).SetString(*requestBody.GasPrice, 10)
		if !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid gasPrice format",
			})
			return
		}
	}

	// Parse to address if provided
	var toAddress *common.Address
	if requestBody.To != nil {
		addr := common.HexToAddress(*requestBody.To)
		toAddress = &addr
	}

	// Calculate gas limit
	var gasLimit uint64
	if requestBody.GasLimit != nil {
		// Use provided gas limit
		gasLimit = *requestBody.GasLimit
	} else if len(data) > 0 && toAddress != nil {
		// For contract calls with data, estimate gas
		estimatedGas, err := s.ethereumClient.EstimateGas(ctx, toAddress, value, data)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to estimate gas: " + err.Error(),
			})
			return
		}
		gasLimit = estimatedGas
	} else {
		// Default gas limit for simple ETH transfers
		gasLimit = uint64(21000)
	}

	// Get next nonce
	fromAddress := s.ethereumClient.GetAddress()
	chainID := s.chainID

	dbNonce, err := s.dbClient.BlockchainTransaction().GetNextNonce(ctx, chainID, fromAddress.Hex())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get nonce from database: " + err.Error(),
		})
		return
	}

	rpcNonce, err := s.ethereumClient.GetPendingNonce(ctx, fromAddress)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get nonce from RPC: " + err.Error(),
		})
		return
	}

	// Use the higher of the two + 1
	nonce := dbNonce
	if rpcNonce > dbNonce {
		nonce = rpcNonce
	}

	// Send transaction
	txHash, err := s.ethereumClient.SendTransaction(ctx, toAddress, value, data, gasLimit, gasPrice, nonce)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to send transaction: " + err.Error(),
		})
		return
	}

	// Create transaction record
	txHashStr := txHash.Hex()
	toAddressStr := ""
	if toAddress != nil {
		toAddressStr = toAddress.Hex()
	}
	dataStr := ""
	if len(data) > 0 {
		dataStr = "0x" + hex.EncodeToString(data)
	}
	gasPriceStr := ""
	if gasPrice != nil {
		gasPriceStr = gasPrice.String()
	}

	blockchainTx := &models.BlockchainTransaction{
		ChainID:     chainID,
		FromAddress: fromAddress.Hex(),
		ToAddress:   &toAddressStr,
		Value:       value.String(),
		Data:        &dataStr,
		GasLimit:    &gasLimit,
		GasPrice:    &gasPriceStr,
		Nonce:       nonce,
		TxHash:      &txHashStr,
		Status:      "pending",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	createdTx, err := s.dbClient.BlockchainTransaction().Create(ctx, blockchainTx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to save transaction: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"id":     createdTx.ID,
		"txHash": txHashStr,
		"status": "pending",
		"nonce":  nonce,
	})
}
