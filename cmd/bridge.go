package bifrost

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/G7DAO/bifrost/bindings/ERC20Inbox"
	"github.com/G7DAO/bifrost/bindings/NodeInterface"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func GetNativeTokenBridgeCalldata(key *keystore.Key, l1Client *ethclient.Client, l2Client *ethclient.Client, to common.Address, l2CallValue *big.Int, l2Calldata []byte) ([]byte, error) {
	l1BaseFee, l1BaseFeeErr := l1Client.SuggestGasPrice(context.Background())
	if l1BaseFeeErr != nil {
		return nil, l1BaseFeeErr
	}

	l2BaseFee, l2BaseFeeErr := l2Client.SuggestGasPrice(context.Background())
	if l2BaseFeeErr != nil {
		return nil, l2BaseFeeErr
	}

	senderDeposit := big.NewInt(0).Add(l2CallValue, ONE_ETHER)
	gasLimit, gasLimitErr := CalculateRetryableGasLimit(l2Client, key.Address, senderDeposit, to, l2CallValue, key.Address, key.Address, l2Calldata)
	if gasLimitErr != nil {
		return nil, gasLimitErr
	}

	maxSubmissionCost, maxSubmissionCostErr := CalculateRetryableSubmissionFee(l2Calldata, l1BaseFee)
	if maxSubmissionCostErr != nil {
		return nil, maxSubmissionCostErr
	}

	inboxAbi, inboxAbiErr := abi.JSON(strings.NewReader(ERC20Inbox.ERC20InboxABI))
	if inboxAbiErr != nil {
		return nil, inboxAbiErr
	}

	parsedGasLimit := big.NewInt(0).SetUint64(gasLimit)
	executionCost := big.NewInt(0).Mul(parsedGasLimit, l2BaseFee)
	tokenTotalFeeAmount := big.NewInt(0).Add(maxSubmissionCost, executionCost)
	tokenTotalFeeAmount.Add(tokenTotalFeeAmount, l2CallValue)

	// function createRetryableTicket(address to, uint256 l2CallValue, uint256 maxSubmissionCost, address excessFeeRefundAddress, address callValueRefundAddress, uint256 gasLimit, uint256 maxFeePerGas, uint256 tokenTotalFeeAmount, bytes calldata data) external;
	createRetryableTicketData, createRetryableTicketDataErr := inboxAbi.Pack("createRetryableTicket", to, l2CallValue, maxSubmissionCost, key.Address, key.Address, parsedGasLimit, l2BaseFee, tokenTotalFeeAmount, l2Calldata)
	if createRetryableTicketDataErr != nil {
		fmt.Fprintln(os.Stderr, createRetryableTicketDataErr.Error())
		return nil, createRetryableTicketDataErr
	}
	return createRetryableTicketData, nil
}

func NativeTokenBridgeCall(inboxAddress common.Address, keyFile string, password string, l1Rpc string, l2Rpc string, to common.Address, l2CallValue *big.Int, l2Calldata []byte) (*types.Transaction, error) {
	l1Client, l1ClientErr := ethclient.DialContext(context.Background(), l1Rpc)
	if l1ClientErr != nil {
		return nil, l1ClientErr
	}

	l2Client, l2ClientErr := ethclient.DialContext(context.Background(), l2Rpc)
	if l2ClientErr != nil {
		return nil, l2ClientErr
	}

	key, keyErr := NodeInterface.KeyFromFile(keyFile, password)
	if keyErr != nil {
		return nil, keyErr
	}

	createRetryableTicketData, createRetryableTicketDataErr := GetNativeTokenBridgeCalldata(key, l1Client, l2Client, to, l2CallValue, l2Calldata)
	if createRetryableTicketDataErr != nil {
		return nil, createRetryableTicketDataErr
	}

	fmt.Println("Sending transaction...")
	transaction, transactionErr := SendTransaction(l1Client, key, password, createRetryableTicketData, inboxAddress.Hex(), big.NewInt(0))
	if transactionErr != nil {
		fmt.Fprintln(os.Stderr, transactionErr.Error())
		return nil, transactionErr
	}
	fmt.Println("Transaction sent! Transaction hash:", transaction.Hash().Hex())

	fmt.Println("Waiting for transaction to be mined...")
	_, receiptErr := bind.WaitMined(context.Background(), l1Client, transaction)
	if receiptErr != nil {
		fmt.Fprintln(os.Stderr, receiptErr.Error())
		return nil, receiptErr
	}
	fmt.Println("Transaction mined!")

	return transaction, nil
}

func NativeTokenBridgePropose(inboxAddress common.Address, keyFile string, password string, l1Rpc string, l2Rpc string, to common.Address, l2CallValue *big.Int, l2Calldata []byte, safeAddress common.Address, safeApi string, safeOperation uint8, safeNonce *big.Int) error {
	l1Client, l1ClientErr := ethclient.DialContext(context.Background(), l1Rpc)
	if l1ClientErr != nil {
		return l1ClientErr
	}

	l2Client, l2ClientErr := ethclient.DialContext(context.Background(), l2Rpc)
	if l2ClientErr != nil {
		return l2ClientErr
	}

	key, keyErr := NodeInterface.KeyFromFile(keyFile, password)
	if keyErr != nil {
		return keyErr
	}

	createRetryableTicketData, createRetryableTicketDataErr := GetNativeTokenBridgeCalldata(key, l1Client, l2Client, to, l2CallValue, l2Calldata)
	if createRetryableTicketDataErr != nil {
		return createRetryableTicketDataErr
	}

	return CreateSafeProposal(l1Client, key, safeAddress, inboxAddress, createRetryableTicketData, big.NewInt(0), safeApi, OperationType(safeOperation), safeNonce)
}
