package operator_commands

import (
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/witnesschain-com/diligencewatchtower-client/keystore"
	wc_common "github.com/witnesschain-com/operator-cli/common"
	"github.com/witnesschain-com/operator-cli/common/bindings/OperatorRegistry"
	operator_config "github.com/witnesschain-com/operator-cli/watchtower-operator/config"

	"github.com/urfave/cli/v2"
)

func DeRegisterWatchtowerCmd() *cli.Command {
	wc_common.ConfigPathFlag.Value = wc_common.DefaultOpConfig
	var deregisterWatchtowerCmd = &cli.Command{
		Name:  "deRegisterWatchtower",
		Usage: "De-register the watchtower",
		Flags: []cli.Flag{
			&wc_common.ConfigPathFlag,
		},
		Action: func(cCtx *cli.Context) error {
			config := operator_config.GetConfigFromContext(cCtx)
			if len(config.EthRPCUrl) != 0 {
				DeRegisterWatchtower(config)
			}
			if len(config.ProofSubmissionRPC) != 0 {
				config.EthRPCUrl = config.ProofSubmissionRPC
				DeRegisterWatchtower(config)
			}
			return nil
		},
	}
	return deregisterWatchtowerCmd
}

func DeRegisterWatchtower(config *operator_config.OperatorConfig) {
	var client *ethclient.Client
	client, config.ChainID = wc_common.ConnectToUrl(config.EthRPCUrl)

	operatorRegistry, err := OperatorRegistry.NewOperatorRegistry(wc_common.NetworkConfig[config.ChainID.String()].OperatorRegistryAddress, client)
	wc_common.CheckError(err, "Instantiating OperatorRegistry contract failed")

	vc := &keystore.VaultConfig{Address: config.OperatorAddress, ChainID: config.ChainID, PrivateKey: config.OperatorPrivateKey, Endpoint: config.Endpoint}
	operatorVault, err := keystore.SetupVault(vc)
	if err != nil {
		wc_common.CheckError(err, "unable to setup vault")
	}

	if !wc_common.IsOperatorWhitelisted(config.OperatorAddress, operatorRegistry) {
		fmt.Printf("Operator %s is not whitelisted\n", config.OperatorAddress.Hex())
		return
	}

	transactOpts := operatorVault.NewTransactOpts(config.ChainID)

	for _, watchtowerAddress := range config.WatchtowerAddresses {
		fmt.Println("Deregister watchtower: " + watchtowerAddress.Hex())
		if !wc_common.IsWatchtowerRegistered(watchtowerAddress, operatorRegistry) {
			fmt.Printf("Watchtower %s is already deRegistered\n", watchtowerAddress.Hex())
			continue
		}

		regTx, err := operatorRegistry.DeRegister(transactOpts, watchtowerAddress)
		wc_common.CheckError(err, "Registering watchtower as operator failed")
		fmt.Printf("Tx sent: %s/tx/%s\n", wc_common.NetworkConfig[config.ChainID.String()].BlockExplorer, regTx.Hash().Hex())
		wc_common.WaitForTransactionReceipt(client, regTx, config.TxReceiptTimeout)
	}
}
