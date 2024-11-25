package bifrost

import (
	"context"
	"math/big"
	"strings"

	"github.com/G7DAO/bifrost/bindings/ArbitrumL1OrbitCustomGateway"
	"github.com/G7DAO/bifrost/bindings/ArbitrumL1OrbitGatewayRouter"
	"github.com/G7DAO/bifrost/bindings/NodeInterface"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func GetL1l2TokenBridgeGasParams(l1Client *ethclient.Client, l2Client *ethclient.Client, l1BaseFee *big.Int, key *keystore.Key, teleportParams *TeleportParams, teleporterAddress common.Address, l2ForwarderAddress common.Address) (uint64, *big.Int, error) {
	router, routerErr := ArbitrumL1OrbitGatewayRouter.NewL1OrbitGatewayRouter(teleportParams.L1l2Router, l1Client)
	if routerErr != nil {
		return uint64(0), nil, routerErr
	}

	gatewayAddress, gatewayAddressErr := router.GetGateway(nil, teleportParams.L1Token)
	if gatewayAddressErr != nil {
		return uint64(0), nil, gatewayAddressErr
	}

	gateway, gatewayErr := ArbitrumL1OrbitCustomGateway.NewL1OrbitCustomGateway(gatewayAddress, l1Client)
	if gatewayErr != nil {
		return uint64(0), nil, gatewayErr
	}

	outboundCalldata, outboundCalldataErr := gateway.GetOutboundCalldata(nil, teleportParams.L1Token, teleporterAddress, l2ForwarderAddress, teleportParams.Amount, []byte{})
	if outboundCalldataErr != nil {
		return uint64(0), nil, outboundCalldataErr
	}

	l1l2TokenBridgeMaxSubmissionCost, l1l2TokenBridgeMaxSubmissionCostErr := CalculateRetryableSubmissionFee(outboundCalldata, l1BaseFee)
	if l1l2TokenBridgeMaxSubmissionCostErr != nil {
		return uint64(0), nil, l1l2TokenBridgeMaxSubmissionCostErr
	}

	// Source: https://github.com/OffchainLabs/arbitrum-sdk/blob/0da65020438fc3e46728ea182f1b4dcf04e3cb7f/src/lib/message/L1ToL2MessageGasEstimator.ts#L154
	senderDeposit := big.NewInt(0).Add(teleportParams.Amount, ONE_ETHER)
	counterpartGatewayAddress, counterpartGatewayAddressErr := gateway.CounterpartGateway(nil)
	if counterpartGatewayAddressErr != nil {
		return uint64(0), nil, counterpartGatewayAddressErr
	}

	l1l2TokenBridgeGasLimit, l1l2TokenBridgeGasLimitErr := CalculateRetryableGasLimit(l2Client, gatewayAddress, senderDeposit, counterpartGatewayAddress, big.NewInt(0), l2ForwarderAddress, RemapL1Address(teleporterAddress), outboundCalldata)
	if l1l2TokenBridgeGasLimitErr != nil {
		return uint64(0), nil, l1l2TokenBridgeGasLimitErr
	}

	return l1l2TokenBridgeGasLimit, l1l2TokenBridgeMaxSubmissionCost, nil
}

func GetL1l2FeeTokenBridgeGasParams(l1Client *ethclient.Client, l2Client *ethclient.Client, l1BaseFee *big.Int, key *keystore.Key, teleportParams *TeleportParams, l1l2RouterAddress *common.Address, teleporterAddress common.Address, l2ForwarderAddress common.Address) (uint64, *big.Int, error) {
	router, routerErr := ArbitrumL1OrbitGatewayRouter.NewL1OrbitGatewayRouter(teleportParams.L1l2Router, l1Client)
	if routerErr != nil {
		return uint64(0), nil, routerErr
	}

	gatewayAddress, gatewayAddressErr := router.GetGateway(nil, teleportParams.L1Token)
	if gatewayAddressErr != nil {
		return uint64(0), nil, gatewayAddressErr
	}

	gateway, gatewayErr := ArbitrumL1OrbitCustomGateway.NewL1OrbitCustomGateway(gatewayAddress, l1Client)
	if gatewayErr != nil {
		return uint64(0), nil, gatewayErr
	}

	feeAmount := big.NewInt(0).Mul(big.NewInt(int64(teleportParams.GasParams.L2l3TokenBridgeGasLimit)), teleportParams.GasParams.L3GasPriceBid)
	outboundCalldata, outboundCalldataErr := gateway.GetOutboundCalldata(nil, teleportParams.L3FeeTokenL1Addr, teleporterAddress, l2ForwarderAddress, feeAmount, []byte{})
	if outboundCalldataErr != nil {
		return uint64(0), nil, outboundCalldataErr
	}

	l1l2FeeTokenBridgeMaxSubmissionCost, l1l2FeeTokenBridgeMaxSubmissionCostErr := CalculateRetryableSubmissionFee(outboundCalldata, l1BaseFee)
	if l1l2FeeTokenBridgeMaxSubmissionCostErr != nil {
		return uint64(0), nil, l1l2FeeTokenBridgeMaxSubmissionCostErr
	}

	// Source: https://github.com/OffchainLabs/arbitrum-sdk/blob/0da65020438fc3e46728ea182f1b4dcf04e3cb7f/src/lib/message/L1ToL2MessageGasEstimator.ts#L154
	senderDeposit := big.NewInt(0).Add(teleportParams.Amount, ONE_ETHER)
	counterpartGatewayAddress := RemapL1Address(gatewayAddress)

	l1l2FeeTokenBridgeGasLimit, l1l2FeeTokenBridgeGasLimitErr := CalculateRetryableGasLimit(l2Client, gatewayAddress, senderDeposit, counterpartGatewayAddress, big.NewInt(0), counterpartGatewayAddress, RemapL1Address(key.Address), outboundCalldata)
	if l1l2FeeTokenBridgeGasLimitErr != nil {
		return uint64(0), nil, l1l2FeeTokenBridgeGasLimitErr
	}

	return l1l2FeeTokenBridgeGasLimit, l1l2FeeTokenBridgeMaxSubmissionCost, nil
}

// Source: https://github.com/OffchainLabs/nitro-contracts/blob/main/src/node-interface/NodeInterface.sol#L25
func CalculateRetryableGasLimit(client *ethclient.Client, sender common.Address, deposit *big.Int, to common.Address, l2CallValue *big.Int, excessFeeRefundAddress common.Address, callValueRefundAddress common.Address, calldata []byte) (uint64, error) {
	nodeInterfaceAbi, nodeInterfaceAbiErr := abi.JSON(strings.NewReader(NodeInterface.NodeInterfaceABI))
	if nodeInterfaceAbiErr != nil {
		return uint64(0), nodeInterfaceAbiErr
	}

	// Source: https://github.com/OffchainLabs/arbitrum-sdk/blob/0da65020438fc3e46728ea182f1b4dcf04e3cb7f/src/lib/message/L1ToL2MessageGasEstimator.ts#L154
	senderDeposit := big.NewInt(0).Add(l2CallValue, ONE_ETHER)

	retryableTicketCalldata, retryableTicketCalldataErr := nodeInterfaceAbi.Pack("estimateRetryableTicket", sender, senderDeposit, to, l2CallValue, excessFeeRefundAddress, callValueRefundAddress, calldata)
	if retryableTicketCalldataErr != nil {
		return uint64(0), retryableTicketCalldataErr
	}

	retryableTicketCallMsg := ethereum.CallMsg{
		From:  sender,
		To:    &NODE_INTERFACE_ADDRESS,
		Value: nil,
		Data:  retryableTicketCalldata,
	}

	retryableTicketGasLimit, retryableTicketGasLimitErr := client.EstimateGas(context.Background(), retryableTicketCallMsg)
	if retryableTicketGasLimitErr != nil {
		return uint64(0), retryableTicketGasLimitErr
	}

	retryableTicketGasLimit = PercentIncrease(big.NewInt(int64(retryableTicketGasLimit)), DEFAULT_GAS_LIMIT_PERCENT_INCREASE).Uint64()

	return retryableTicketGasLimit, nil
}
