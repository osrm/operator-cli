package operator_commands

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/witnesschain-com/diligencewatchtower-client/keystore"
	wc_common "github.com/witnesschain-com/operator-cli/common"
	"github.com/witnesschain-com/operator-cli/common/bindings/OperatorRegistry"
	operator_config "github.com/witnesschain-com/operator-cli/config"

	"github.com/urfave/cli/v2"
)

func RegisterWatchtowerCmd() *cli.Command {
	wc_common.ConfigPathFlag.Value = wc_common.DefaultOpConfig
	var registerWatchtowerCmd = &cli.Command{
		Name:  "registerWatchtower",
		Usage: "Register a watchtower",
		Flags: []cli.Flag{
			&wc_common.ConfigPathFlag,
		},
		Action: func(cCtx *cli.Context) error {
			config := operator_config.GetConfigFromContext(cCtx)
			if len(config.EthRPCUrl) != 0 {
				// register on L1
				RegisterWatchtower(config)
			}

			if len(config.ProofSubmissionRPC) != 0 {
				// register on Proof submission chain
				config.EthRPCUrl = config.ProofSubmissionRPC
				RegisterWatchtower(config)
			}
			return nil
		},
	}
	return registerWatchtowerCmd
}

func RegisterWatchtower(config *operator_config.OperatorConfig) {
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

	if (wc_common.NetworkConfig[config.ChainID.String()].GasPrice == -1){
		transactOpts.GasPrice = big.NewInt(0)
	}

	expiry := wc_common.CalculateExpiry(client, config.ExpiryInDays)

	for i, watchtowerAddress := range config.WatchtowerAddresses {
		fmt.Println("watchtowerAddress: " + watchtowerAddress.Hex())

		var watchtowerPrivateKey *ecdsa.PrivateKey
		if len(config.WatchtowerPrivateKeys) != 0 {
			watchtowerPrivateKey = config.WatchtowerPrivateKeys[i]
		}

		vc := &keystore.VaultConfig{Address: watchtowerAddress, ChainID: config.ChainID, PrivateKey: watchtowerPrivateKey, Endpoint: config.Endpoint}
		watchtowerVault, err := keystore.SetupVault(vc)
		wc_common.CheckError(err, "unable to setup watchtower vault")

		if wc_common.IsWatchtowerRegistered(watchtowerAddress, operatorRegistry) {
			fmt.Printf("Watchtower %s is already registered\n", watchtowerAddress.Hex())
			continue
		}

		salt := wc_common.GenerateSalt()
		signedMessage := SignOperatorAddress(client, operatorRegistry, watchtowerVault, config.OperatorAddress, salt, expiry)
		regTx, err := operatorRegistry.RegisterWatchtowerAsOperator(transactOpts, watchtowerAddress, salt, expiry, signedMessage)
		wc_common.CheckError(err, "Registering watchtower as operator failed")
		fmt.Printf("Tx sent: %s/tx/%s\n", wc_common.NetworkConfig[config.ChainID.String()].BlockExplorer, regTx.Hash().Hex())
		wc_common.WaitForTransactionReceipt(client, regTx, config.TxReceiptTimeout)
	}
}
