package bifrost

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Source: https://github.com/OffchainLabs/nitro-contracts/blob/main/src/bridge/Inbox.sol#L323
func CalculateRetryableSubmissionFee(calldata []byte, baseFee *big.Int) (*big.Int, error) {
	multiplier := big.NewInt(int64(1400 + 6*len(calldata)))
	submissionFee := multiplier.Mul(multiplier, baseFee)
	increasedSubmissionFee := PercentIncrease(submissionFee, DEFAULT_SUBMISSION_FEE_PERCENT_INCREASE)

	return increasedSubmissionFee, nil
}

func CalculateRequiredEth(gasParams RetryableGasParams, teleportationType TeleportationType) (*big.Int, *big.Int) {
	l1l2FeeTokenBridgeGasCost := big.NewInt(0).Mul(gasParams.L2GasPriceBid, big.NewInt(int64(gasParams.L1l2FeeTokenBridgeGasLimit)))
	l1l2FeeTokenBridgeTotalCost := big.NewInt(0).Add(gasParams.L1l2FeeTokenBridgeMaxSubmissionCost, l1l2FeeTokenBridgeGasCost)

	l1l2TokenBridgeGasCost := big.NewInt(0).Mul(gasParams.L2GasPriceBid, big.NewInt(int64(gasParams.L1l2TokenBridgeGasLimit)))
	l1l2TokenBridgeTotalCost := big.NewInt(0).Add(gasParams.L1l2TokenBridgeMaxSubmissionCost, l1l2TokenBridgeGasCost)

	l2ForwarderFactoryGasCost := big.NewInt(0).Mul(gasParams.L2GasPriceBid, big.NewInt(int64(gasParams.L2ForwarderFactoryGasLimit)))
	l2ForwarderFactoryTotalCost := big.NewInt(0).Add(gasParams.L2ForwarderFactoryMaxSubmissionCost, l2ForwarderFactoryGasCost)

	l2l3TokenBridgeGasCost := big.NewInt(0).Mul(gasParams.L3GasPriceBid, big.NewInt(int64(gasParams.L2l3TokenBridgeGasLimit)))
	l2l3TokenBridgeTotalCost := big.NewInt(0).Add(gasParams.L2l3TokenBridgeMaxSubmissionCost, l2l3TokenBridgeGasCost)

	// all teleportation types require at least these 2 retryables to L2
	requiredEth := big.NewInt(0).Add(l2ForwarderFactoryTotalCost, l1l2TokenBridgeTotalCost)
	requiredFeeToken := big.NewInt(0)

	// in addition to the above ETH amount, more fee token and/or ETH is required depending on the teleportation type
	if teleportationType == Standard {
		// standard type requires 1 retryable to L3 paid for in ETH
		requiredEth.Add(requiredEth, l2l3TokenBridgeTotalCost)
	} else if teleportationType == OnlyCustomFee {
		// only custom fee type requires 1 retryable to L3 paid for in fee token
		requiredFeeToken = l2l3TokenBridgeTotalCost
	} else if l2l3TokenBridgeTotalCost.Cmp(big.NewInt(0)) > 0 {
		// non-fee token to custom fee type requires:
		// 1 retryable to L2 paid for in ETH
		// 1 retryable to L3 paid for in fee token
		requiredEth.Add(requiredEth, l1l2FeeTokenBridgeTotalCost)
		requiredFeeToken = l2l3TokenBridgeTotalCost
	}

	return requiredEth, requiredFeeToken
}

// Source: https://github.com/OffchainLabs/nitro/blob/057bf836fcf719e803b0486914bc957134f691fd/arbos/util/util.go#L204
func RemapL1Address(l1Addr common.Address) common.Address {
	AddressAliasOffset, success := new(big.Int).SetString("0x1111000000000000000000000000000000001111", 0)
	if !success {
		panic("Error initializing AddressAliasOffset")
	}

	sumBytes := new(big.Int).Add(new(big.Int).SetBytes(l1Addr.Bytes()), AddressAliasOffset).Bytes()
	if len(sumBytes) > 20 {
		sumBytes = sumBytes[len(sumBytes)-20:]
	}
	return common.BytesToAddress(sumBytes)
}

func PercentIncrease(value *big.Int, percentage *big.Int) *big.Int {
	multipliedValue := big.NewInt(0).Mul(value, percentage)
	increase := big.NewInt(0).Div(multipliedValue, big.NewInt(100))
	return big.NewInt(0).Add(value, increase)
}
