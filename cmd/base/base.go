package base

import (
	"github.com/G7DAO/bifrost/bindings/L1StandardBridge"
	"github.com/G7DAO/bifrost/bindings/OptimismMintableERC20Factory"
	"github.com/spf13/cobra"
)

func CreateBaseCommand() *cobra.Command {
	baseCmd := &cobra.Command{
		Use:   "base",
		Short: "Base commands",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	l1StandardBridgeCmd := L1StandardBridge.CreateL1StandardBridgeCommand()
	optimismMintableERC20FactoryCmd := OptimismMintableERC20Factory.CreateOptimismMintableERC20FactoryCommand()

	baseCmd.AddCommand(l1StandardBridgeCmd, optimismMintableERC20FactoryCmd)

	return baseCmd
}
