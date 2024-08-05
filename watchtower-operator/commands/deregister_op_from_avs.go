package operator_commands

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/witnesschain-com/diligencewatchtower-client/keystore"
	wc_common "github.com/witnesschain-com/operator-cli/common"
	"github.com/witnesschain-com/operator-cli/common/bindings/AvsDirectory"
	"github.com/witnesschain-com/operator-cli/common/bindings/OperatorRegistry"
	"github.com/witnesschain-com/operator-cli/common/bindings/WitnessHub"
	operator_config "github.com/witnesschain-com/operator-cli/watchtower-operator/config"

	"github.com/urfave/cli/v2"
)

func DeRegisterOperatorFromAVSCmd() *cli.Command {
	wc_common.ConfigPathFlag.Value = wc_common.DefaultOpConfig
	var deregisterOperatorFromAVSCmd = &cli.Command{
		Name:  "deRegisterOperatorFromAVS",
		Usage: "De-register the operator from AVS",
		Flags: []cli.Flag{
			&wc_common.ConfigPathFlag,
		},
		Action: func(cCtx *cli.Context) error {
			config := operator_config.GetConfigFromContext(cCtx)
			if len(config.EthRPCUrl) != 0 {
				DeRegisterOperatorFromAVS(config)
			}

			return nil
		},
	}
	return deregisterOperatorFromAVSCmd
}

func DeRegisterOperatorFromAVS(config *operator_config.OperatorConfig) {
	var client *ethclient.Client
	client, config.ChainID = wc_common.ConnectToUrl(config.EthRPCUrl)

	if wc_common.NetworkConfig[config.ChainID.String()].WitnessHubAddress.Cmp(common.Address{0}) == 0 {
		fmt.Printf("Contract %v not found at %v\n. Please verify that witnesschain contract are deployed for this chain", wc_common.NetworkConfig[config.ChainID.String()].WitnessHubAddress, config.EthRPCUrl)
	}

	operatorRegistry, err := OperatorRegistry.NewOperatorRegistry(wc_common.NetworkConfig[config.ChainID.String()].OperatorRegistryAddress, client)
	wc_common.CheckError(err, "Instantiating OperatorRegistry contract failed")

	if !wc_common.IsOperatorWhitelisted(config.OperatorAddress, operatorRegistry) {
		fmt.Printf("Operator %s is not whitelisted\n", config.OperatorAddress.Hex())
		return
	}

	avsDirectory, err := AvsDirectory.NewAvsDirectory(wc_common.NetworkConfig[config.ChainID.String()].AVSDirectoryAddress, client)
	wc_common.CheckError(err, "Instantiating AvsDirectory contract failed")

	if !wc_common.IsOperatorRegistered(wc_common.NetworkConfig[config.ChainID.String()].WitnessHubAddress, config.OperatorAddress, avsDirectory) {
		fmt.Printf("Operator %s is already deregistered\n", config.OperatorAddress.Hex())
		return
	}

	witnessHub, err := WitnessHub.NewWitnessHub(wc_common.NetworkConfig[config.ChainID.String()].WitnessHubAddress, client)
	wc_common.CheckError(err, "Instantiating WitnessHub contract failed")

	vc := &keystore.VaultConfig{Address: config.OperatorAddress, PrivateKey: config.OperatorPrivateKey, Endpoint: config.Endpoint, ChainID: config.ChainID}
	operatorVault, err := keystore.SetupVault(vc)
	wc_common.CheckError(err, "unable to setup operator Vault: "+vc.Address.Hex())

	transactOpts := operatorVault.NewTransactOpts(config.ChainID)

	tx, err := witnessHub.DeregisterOperatorFromAVS(transactOpts, config.OperatorAddress)
	wc_common.CheckError(err, "deregistering operator to AVS failed")

	fmt.Printf("Tx sent: %s/tx/%s\n", wc_common.NetworkConfig[config.ChainID.String()].BlockExplorer, tx.Hash().Hex())

	wc_common.WaitForTransactionReceipt(client, tx, config.TxReceiptTimeout)
}
