package bifrost

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

func CreateBifrostCommand() *cobra.Command {
	bifrostCmd := &cobra.Command{
		Use:   "bifrost",
		Short: "Bifrost CLI",
		Long:  `Bifrost CLI`,
	}

	bifrostCmd.AddCommand(CreateArbitrumCommand())

	return bifrostCmd
}

func CreateArbitrumCommand() *cobra.Command {
	var keyFile, password, l1Rpc, l2Rpc, inboxRaw, toRaw, l2CallValueRaw, l2CalldataRaw, safeAddressRaw, safeApi, safeNonceRaw string
	var inboxAddress, to, safeAddress common.Address
	var l2CallValue *big.Int
	var l2Calldata []byte
	var safeOperation uint8
	var safeNonce *big.Int

	arbitrumCmd := &cobra.Command{
		Use:   "l1-to-l2",
		Short: "Bridge tokens from L1 to L2",
		Long:  `Bridge tokens from L1 to L2 with a single transaction and arbitrary calldata`,

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !common.IsHexAddress(inboxRaw) {
				return errors.New("invalid inbox address")
			}
			inboxAddress = common.HexToAddress(inboxRaw)

			if !common.IsHexAddress(toRaw) {
				return errors.New("invalid recipient address")
			}
			to = common.HexToAddress(toRaw)

			l2CallValue = new(big.Int)
			if l2CallValueRaw != "" {
				_, ok := l2CallValue.SetString(l2CallValueRaw, 10)
				if !ok {
					return errors.New("invalid L2 call value")
				}
			} else {
				fmt.Println("No L2 call value provided, defaulting to 0")
				l2CallValue.SetInt64(0)
			}

			if l2CalldataRaw != "" {
				var err error
				l2Calldata, err = hex.DecodeString(l2CalldataRaw)
				if err != nil {
					return err
				}
			}

			if keyFile == "" {
				return errors.New("keyfile is required")
			}

			if safeAddressRaw != "" {
				if !common.IsHexAddress(safeAddressRaw) {
					return fmt.Errorf("--safe is not a valid Ethereum address")
				} else {
					safeAddress = common.HexToAddress(safeAddressRaw)
				}

				if safeApi == "" {
					client, clientErr := ethclient.DialContext(context.Background(), l1Rpc)
					if clientErr != nil {
						return clientErr
					}

					chainID, chainIDErr := client.ChainID(context.Background())
					if chainIDErr != nil {
						return chainIDErr
					}
					safeApi = "https://safe-client.safe.global/v1/chains/" + chainID.String() + "/transactions/" + safeAddress.Hex() + "/propose"
					fmt.Println("--safe-api not specified, using default (", safeApi, ")")
				}

				if OperationType(safeOperation).String() == "Unknown" {
					return fmt.Errorf("--safe-operation must be 0 (Call) or 1 (DelegateCall)")
				}

				if safeNonceRaw != "" {
					safeNonce = new(big.Int)
					_, ok := safeNonce.SetString(safeNonceRaw, 0)
					if !ok {
						return fmt.Errorf("--safe-nonce is not a valid big integer")
					}
				} else {
					fmt.Println("--safe-nonce not specified, fetching from Safe")
					safeNonce = big.NewInt(0)
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Bridging to", to.Hex())
			if safeAddressRaw != "" {
				err := NativeTokenBridgePropose(inboxAddress, keyFile, password, l1Rpc, l2Rpc, to, l2CallValue, l2Calldata, safeAddress, safeApi, safeOperation, safeNonce)
				if err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
					return err
				}
			} else {
				transaction, transactionErr := NativeTokenBridgeCall(inboxAddress, keyFile, password, l1Rpc, l2Rpc, to, l2CallValue, l2Calldata)
				if transactionErr != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), transactionErr.Error())
					return transactionErr
				}

				fmt.Println("Transaction sent:", transaction.Hash().Hex())
			}

			return nil
		},
	}

	arbitrumCmd.Flags().StringVar(&password, "password", "", "Password to encrypt accounts with")
	arbitrumCmd.Flags().StringVar(&keyFile, "keyfile", "", "Keyfile to sign transaction with")
	arbitrumCmd.Flags().StringVar(&l1Rpc, "l1-rpc", "", "L1 RPC URL")
	arbitrumCmd.Flags().StringVar(&l2Rpc, "l2-rpc", "", "L2 RPC URL")
	arbitrumCmd.Flags().StringVar(&inboxRaw, "inbox", "", "Inbox address")
	arbitrumCmd.Flags().StringVar(&toRaw, "to", "", "Recipient or contract address")
	arbitrumCmd.Flags().StringVar(&l2CallValueRaw, "amount", "", "L2 call value")
	arbitrumCmd.Flags().StringVar(&l2CalldataRaw, "l2-calldata", "", "Calldata to send")
	arbitrumCmd.Flags().StringVar(&safeAddressRaw, "safe", "", "Address of the Safe contract")
	arbitrumCmd.Flags().StringVar(&safeApi, "safe-api", "", "Safe API for the Safe Transaction Service (optional)")
	arbitrumCmd.Flags().Uint8Var(&safeOperation, "safe-operation", 0, "Safe operation type: 0 (Call) or 1 (DelegateCall)")
	arbitrumCmd.Flags().StringVar(&safeNonceRaw, "safe-nonce", "", "Safe nonce")

	return arbitrumCmd
}
