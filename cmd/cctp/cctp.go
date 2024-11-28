package cctp

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/G7DAO/bifrost/bindings/TokenMessenger"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

type ChainDomain uint32

const (
	ChainDomainEthereum = 0
	ChainDomainArbitrum = 3
)

// String returns the string representation of the ChainDomain
func (o ChainDomain) String() string {
	switch o {
	case ChainDomainEthereum:
		return "Ethereum"
	case ChainDomainArbitrum:
		return "Arbitrum"
	default:
		return "Unknown"
	}
}

func cctpBridge(key *keystore.Key, client *ethclient.Client, recipient common.Address, domain uint32, amount *big.Int, token common.Address, contractAddress common.Address) error {
	mintRecipient, err := ParseMintRecipientFrom20BytesTo32Bytes(recipient)
	if err != nil {
		return err
	}

	abi, err := TokenMessenger.TokenMessengerMetaData.GetAbi()
	if err != nil {
		return err
	}

	packed, err := abi.Pack("depositForBurn", amount, domain, mintRecipient, token)
	if err != nil {
		return err
	}

	tx, err := SendTransaction(client, key, packed, contractAddress.Hex(), big.NewInt(0))
	if err != nil {
		return err
	}

	fmt.Println("Transaction hash:", tx.Hash().Hex())
	return nil
}

// CCTP uses 32 bytes addresses, while EVEM uses 20 bytes addresses
// const mintRecipient = utils.hexlify(utils.zeroPad(recipient, 32)) as Address
func ParseMintRecipientFrom20BytesTo32Bytes(recipient common.Address) ([32]byte, error) {
	var mintRecipient [32]byte

	// Validate input length
	if len(recipient) != 20 {
		return mintRecipient, fmt.Errorf("invalid recipient length: expected 20 bytes, got %d", len(recipient))
	}

	// Copy the 20 bytes to the end of the 32-byte array, leaving the first 12 bytes as zeros
	// This is equivalent to zero-padding on the left
	copy(mintRecipient[32-20:], recipient.Bytes())

	return mintRecipient, nil
}

func CreateCctpCommand() *cobra.Command {
	var keyFile, password, rpc, recipientRaw, amountRaw, tokenRaw, contractRaw string
	var domain uint32
	var token, recipient, contract common.Address
	var amount *big.Int

	cctpCmd := &cobra.Command{
		Use:   "cctp",
		Short: "Bifrost for CCTP cross-chain messaging protocol",
		Long:  `Bifrost for CCTP cross-chain messaging protocol`,

		PreRunE: func(cmd *cobra.Command, args []string) error {

			if !common.IsHexAddress(recipientRaw) {
				return errors.New("invalid recipient address")
			}
			recipient = common.HexToAddress(recipientRaw)

			if ChainDomain(domain).String() == "Unknown" {
				return errors.New("invalid domain")
			}

			if amountRaw == "" {
				return errors.New("amount is required")
			} else {
				var ok bool
				amount, ok = new(big.Int).SetString(amountRaw, 10)
				if !ok {
					return errors.New("invalid amount")
				}
			}
			if tokenRaw == "" {
				return errors.New("token is required")
			} else {
				token = common.HexToAddress(tokenRaw)
			}
			if contractRaw == "" {
				return errors.New("contract is required")
			} else {
				contract = common.HexToAddress(contractRaw)
			}

			if keyFile == "" {
				return errors.New("keyfile is required")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Bridging to", recipientRaw)

			key, err := TokenMessenger.KeyFromFile(keyFile, password)
			if err != nil {
				return err
			}

			client, err := ethclient.Dial(rpc)
			if err != nil {
				return err
			}

			return cctpBridge(key, client, recipient, domain, amount, token, contract)
		},
	}

	cctpCmd.Flags().StringVar(&password, "password", "", "Password to encrypt accounts with")
	cctpCmd.Flags().StringVar(&keyFile, "keyfile", "", "Keyfile to sign transaction with")
	cctpCmd.Flags().StringVar(&rpc, "rpc", "", "RPC URL")
	cctpCmd.Flags().StringVar(&recipientRaw, "recipient", "", "Recipient address in the destination domain")
	cctpCmd.Flags().Uint32Var(&domain, "domain", 0, "Destination domain")
	cctpCmd.Flags().StringVar(&amountRaw, "amount", "", "Amount of tokens to send")
	cctpCmd.Flags().StringVar(&tokenRaw, "token", "", "Token to send")
	cctpCmd.Flags().StringVar(&contractRaw, "contract", "", "Contract to send tokens from")
	return cctpCmd
}
