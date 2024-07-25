# WitnessChain Operator CLI

## Description
watchtower-operator is a command-line interface (CLI) tool for interacting with some functionalities provided by the WatchTower(EigenLayer AVS) contracts . You can refer to the [How to use the config files](#how-to-use-the_config-files) section to understand how to use the config files.

## Installation
You can get the watchtower-operator cli prebuilt, or build from source


### Prebuilt
   
You can run the following command in your terminal and follow instructions provided by the script to use the cli
```
curl -sSfL witnesschain.com/install-operator-cli | bash
```

Installation instructions for building from source is available in 
[docs/install.md](docs/install.md).


## How to use
Once you have the watchtower-operator installed, you can directly use the exectable -

```
watchtower-operator command [command options] [arguments...]
```
Note: In case you haven't exported the path for watchtower-operator executable, you can start the cli by `./watchtower-operator`

## Commands available
| Command | Description |
|----------|----------|
|keys | Used to store private keys in an encrypted format |
|registerWatchtower | Used to register watch tower |
|deRegisterWatchtower | Used to deregister watch tower |
|registerOperatorToAVS | Used to notify EigenLayer that an operator is registered to the AVS |
|deRegisterOperatorFromAVS | Used to notify EigenLayer that an operator is de-registered from the AVS |

## How to use the config files
The structure and details in the config file might differ based on the functionality you are using. Config files related to operator have been explained below. Enter the following details in the config file. Changing the key names in the json file will lead to misbehavior

### Operator config file

#### for command - registerOperatorToAVS, deRegisterOperatorFromAVS
Default file: config/operator-config.template.json (reference file)

| Field | Description |
|----------|----------|
|watchtower_private_keys | Private keys of the watchtowers (use this field if you want to enter raw keys)|
|watchtower_encrypted_keys | Encrypted private keys of the watchtowers (use this field if you want to enter encrypted key names)|
|operator_private_key | Private key of the operator(on which the actions will be performed) (use this field if you want to enter raw key)|
|operator_encrypted_key | Encrypted private key of the operator(on which the actions will be performed) (use this field if you want to enter raw key)|
|encrypted_key_type | The type of encryption used for the keys (valid values = w3secretkeys/gocryptfs) |
|eth_rpc_url | The RPC URL where you want to perform the transactions |
|gas_limit | The gas limit you want to set while sending the transactions (Default value = 1000000). No need to add in the config unless you want to overwrite the default values.  |
|tx_receipt_timeout| Timeout in seconds for waiting of tx receipts (Default value = 300). No need to add in the config unless you want to overwrite the default values. |
|expiry| Expiry in days after which the operator signature becomes invalid (Default value = 1). No need to add in the config unless you want to overwrite the default values. |

## How to use the encrypted keys

To store and use the keys in an encrypted format, use the `watchtower_encrypted_keys` and `operator_encrypted_key` fields in the config. If they have same values entered into it, the CLI tool will use the keys stored in the encrypted format in the given key name.

For `w3secretkeys`, the default path is `~/.witnesschain/cli/.w3secretkeys`. These are similar to EigenLayer ECDSA keys. Read more about [it](https://docs.eigenlayer.xyz/eigenlayer/operator-guides/operator-installation#create-and-list-keys)
For `gocryptfs`, the default path is `~/.witnesschain/cli/.gocryptfs/.encrypted_keys`. Read more about [gocrytpfs encryption](https://github.com/rfjakob/gocryptfs)

### Sub-commands
The `keys` command has 4 sub-commands. Let's look at them one by one -
The flags which can be used with the key commands are

| Flag | Alias | Usage |
|----------|----------|----------|
|--key-name | -k <value> | This flag is used to specify the key name |
|--key-type | -t <value> | This flag is used to specify the keystore type (w3secretkeys/gocryptfs) |
|--insecure | -i | This flag is used to bypass password strength validations |

1. `init` - This command is used to initiate a new keystore file system and the type will be dependent on the key-type chosen. In case the type is `gocryptfs`, it will ask you to input a password which will be required whenever we need to access these file. It also performs password validations for strong passwords. To bypass this validation in the test environment, use `--insecure(-i)` flag.

```
init operation for key-type gocryptfs

$ watchtower-operator keys init -t gocryptfs -i

Creating directory:  .gocryptfs
Creating directory:  .gocryptfs/.encrypted_keys
Creating directory:  .gocryptfs/.decrypted_keys
Enter password to init: ****
```
After this command, two directories `.encrypted_keys` and `.decrypted_keys` are created inside a directory `.gocryptfs`. The names indicate their functions. Once this is done, we don't need to do it again, unless the `.encrypted_keys` or `.decrypted_keys` are tampered with. Once the command is successfully run, all other actions to create/import/export/delete `gocryptfs` type will need this password.

```
init operation for key-type gocryptfs

$ watchtower-operator keys init -t w3secretkeys -i

Creating directory:  .w3secretkeys
Init keystore done
```
Contrary to `gocryptfs` type, `w3secretkeys` does not need a password during init operation. Instead you will be asked for password while you create/import/export/delete keys. The key difference between these two types is `gocryptfs` uses the password used during the init operation. And with `w3secretkeys` type, individual keys can have their own password. You will need that password while you create/import/export keys. For consistency when using multiple keys in the config, it is mandatory for all the `w3secretkeys` keys to have the same password.

2. `create` - This command will create a new key. The value you pass with the `-k` flag will be the name of the key which will be referred in the future by the CLI. This name should be mentioned in the config file to extract the corresponding private key. When you run this command, it will ask for password to mount and then ask you to enter the private key
```
$ watchtower-operator keys create -k wt1 -t gocryptfs -i

Enter password to mount: **********
Created key: wt1
```

3. `import` - This command will import an existing key. The value you pass with the `-k` flag will be the name of the key which will be referred in the future by the CLI. This name should be mentioned in the config file to extract the corresponding private key. When you run this command, it will ask for password to mount and then ask you to enter the private key
```
$ watchtower-operator keys create -k wt1 -t gocryptfs -i

Enter password to mount: **********
Enter private key: ****************************************************************
Created key: wt1
```

4. `list` - This command will list all the keys currently present (only by file name and created date)
```
$ watchtower-operator keys list -t w3secretkeys
   -------------------------------------------------------
   Name                           Created
   -------------------------------------------------------
   wt1                            30-04-2024 11:36:16
   -------------------------------------------------------
```

5. `delete` - This command will delete a key. This command will need a flag --key-name(-k). The name given will be searched in the `.decrypted_keys` and deleted
```
$ ./watchtower-operator keys delete -t w3secretkeys

Deleted key: wt1

$ watchtower-operator keys list
   -------------------------------------------------------
   Name                           Created
   -------------------------------------------------------
   -------------------------------------------------------
```

Going forward, the CLI will ask for password to decrpyt and use these keys. This is how the config will look like when using encrypted keys and the keys are present in the default location i.e. `~/.witnesschain/cli/.gocryptfs/.encrypted_keys`

> NOTE -
> If you want to use encrypted keys, use the field `watchtower_encrypted_keys` and use `watchtower_private_keys` when you want to use raw private keys. You need to either give the full path(if you want an alternate path to be used) or the name(for the default path) of the encrypted keys as show in the example below

The below example shows how you can use the key names which will be taken from the default path
```
{
  "watchtower_private_keys": [
    "<raw-watchtower-private-key>"
  ],
  "operator_private_key": "<raw-operator-private-key>",
  "eth_rpc_url": "<Mainnet RPC URL>",
  "expiry": 1
}
```

The below example shows how you can use encrypted keys using `w3secretkeys` type. The type can also be `gocryptfs`
```
{
  "watchtower_encrypted_keys": [
    "~/alternate/path/to/your/keys/wt1"
  ],
  "operator_encrypted_key": "~/alternate/path/to/your/keys/op1",
  "encrypted_key_type": "w3secretkeys",
  "eth_rpc_url": "<Mainnet RPC URL>"
}
```

The below example shows how you can use the key names which will be taken from an alternate path using `gocryptfs` type. The type can also be `w3secretkeys`
```
{
  "watchtower_encrypted_keys": [
    "~/alternate/path/to/your/keys/.encrypted_keys/wt1"
  ],
  "operator_encrypted_key": "~/alternate/path/to/your/keys/.encrypted_keys/op1",
  "encrypted_key_type": "gocryptfs",
  "eth_rpc_url": "<Mainnet RPC URL>"
}
```
> NOTE -
> If you are using an alternate path, all the keys present in the config have to be present in the same path. You cannot save the keys in different locations in the same config
