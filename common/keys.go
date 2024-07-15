package wc_common

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"

	sdkEcdsa "github.com/Layr-Labs/eigensdk-go/crypto/ecdsa"
)

var m_useEncryptedKeys bool = false
var m_isFullPath bool = false
var m_retryMounting bool = false
var m_encryptedDir string = GoCryptFSDirName + "/" + EncryptedDirName
var m_decryptedDir string = GoCryptFSDirName + "/" + DecryptedDirName
var m_goCryptFSConfig string = GoCryptFSDirName + "/" + GoCryptFSConfigName
var m_goCryptFSDir string = GoCryptFSDirName
var m_keystoreDir string = KeyStoreDirName
var m_keystoreSuffix string = KeyStoreSuffixName

func KeysCmd() *cli.Command {
	var keysCmd = &cli.Command{
		Name:  "keys",
		Usage: "Manage the operator's keys",
		Subcommands: []*cli.Command{
			InitCmd(),
			CreateCmd(),
			ImportCmd(),
			ExportCmd(),
			DeleteCmd(),
			ListCmd(),
		},
	}
	return keysCmd
}

func InitCmd() *cli.Command {
	var initCmd = &cli.Command{
		Name:      "init",
		Usage:     "init local keystore",
		UsageText: "init",
		Flags: []cli.Flag{
			&InsecureFlag,
			&KeyStoreType,
		},
		Action: func(cCtx *cli.Context) error {
			CheckIfGocryptfsIsInstalled()
			InitKeyStore(cCtx)
			return nil
		},
	}
	return initCmd
}

func CreateCmd() *cli.Command {
	var createCmd = &cli.Command{
		Name:      "create",
		Usage:     "create encrypted key in local keystore",
		UsageText: "create <keyName>",
		Flags: []cli.Flag{
			&KeyNameFlag,
			&KeyStoreType,
			&InsecureFlag,
		},
		Action: func(cCtx *cli.Context) error {
			CreateKeyCmd(cCtx)
			return nil
		},
	}
	return createCmd
}

func ImportCmd() *cli.Command {
	var importCmd = &cli.Command{
		Name:      "import",
		Usage:     "import existing key into local keystore",
		UsageText: "import <keyName>",
		Flags: []cli.Flag{
			&KeyNameFlag,
			&KeyStoreType,
			&InsecureFlag,
		},
		Action: func(cCtx *cli.Context) error {
			ImportKeyCmd(cCtx)
			return nil
		},
	}

	return importCmd
}

func DeleteCmd() *cli.Command {
	var deleteCmd = &cli.Command{
		Name:      "delete",
		Usage:     "delete encrypted key from local keystore",
		UsageText: "delete <keyName>",
		Flags: []cli.Flag{
			&KeyNameFlag,
			&KeyStoreType,
		},
		Action: func(cCtx *cli.Context) error {
			DeleteKeyCmd(cCtx)
			return nil
		},
	}
	return deleteCmd
}

func ListCmd() *cli.Command {
	var listCmd = &cli.Command{
		Name:      "list",
		Usage:     "list all encrypted keys from local keystore",
		UsageText: "list",
		Flags: []cli.Flag{
			&KeyStoreType,
		},
		Action: func(cCtx *cli.Context) error {
			ListKeyCmd(cCtx)
			return nil
		},
	}
	return listCmd
}

func InitKeyStore(cCtx *cli.Context) {
	insecure := cCtx.Bool("insecure")
	keyType := cCtx.String("key-type")

	//Use enums instead of strings
	//Nested directory creation. Check full path
	switch keyType {
	case KeyTypeGoCryptFS:
		if !DirectoryExists(m_goCryptFSDir) {
			CreateDirectory(m_goCryptFSDir)
		}
		if !DirectoryExists(m_encryptedDir) {
			CreateDirectory(m_encryptedDir)
		}

		if !DirectoryExists(m_decryptedDir) {
			CreateDirectory(m_decryptedDir)
		}
		InitGocryptfs(insecure)
	case KeyTypeKeystore:
		if !DirectoryExists(m_keystoreDir) {
			CreateDirectory(m_keystoreDir)
		}
		fmt.Println("Init keystore done")
	default:
		var err error
		err = errors.New("invalid key type")
		CheckError(err, "error initializing key store")
	}
}

func CreateKeyCmd(cCtx *cli.Context) {
	keyName := cCtx.String("key-name")
	keyType := cCtx.String("key-type")
	insecure := cCtx.Bool("insecure")

	err := ValidateKeyName(keyName)
	CheckError(err, "Error validating key name")

	switch keyType {
	case KeyTypeGoCryptFS:
		ValidateAndMount()
		CreateGoCryptfsKey(keyName)
	case KeyTypeKeystore:
		CreateKeystoreKey(keyName, insecure)
	default:
		var err error
		err = errors.New("invalid key type")
		CheckError(err, "error creating key")
	}
}

func ImportKeyCmd(cCtx *cli.Context) {
	keyName := cCtx.String("key-name")
	keyType := cCtx.String("key-type")
	insecure := cCtx.Bool("insecure")

	err := ValidateKeyName(keyName)
	CheckError(err, "Error validating key name")

	switch keyType {
	case KeyTypeGoCryptFS:
		ValidateAndMount()
		ImportGoCryptfsKey(keyName)
	case KeyTypeKeystore:
		ImportKeystoreKey(keyName, insecure)
	default:
		var err error
		err = errors.New("invalid key type")
		CheckError(err, "error importing key")
	}
}

func ExportCmd() *cli.Command {
	var exportCmd = &cli.Command{
		Name:      "export",
		Usage:     "export existing key from local keystore",
		UsageText: "export <keyName>",
		Flags: []cli.Flag{
			&KeyNameFlag,
			&KeyStoreType,
		},
		Action: func(cCtx *cli.Context) error {
			ExportKeyCmd(cCtx)
			return nil
		},
	}

	return exportCmd
}

func DeleteKeyCmd(cCtx *cli.Context) {
	keyName := cCtx.String("key-name")
	keyType := cCtx.String("key-type")

	err := ValidateKeyName(keyName)
	CheckError(err, "Error validating key name")

	switch keyType {
	case KeyTypeGoCryptFS:
		ValidateAndMount()
		DeleteGoCryptfsKey(keyName)
	case KeyTypeKeystore:
		DeleteKeystoreKey(keyName)
	default:
		var err error
		err = errors.New("invalid key type")
		CheckError(err, "error deleting key")
	}

	fmt.Printf("Deleted key: %s\n", keyName)
}

func ListKeyCmd(cCtx *cli.Context) {
	keyType := cCtx.String("key-type")

	var err error
	var dir *os.File

	switch keyType {
	case KeyTypeGoCryptFS:
		ValidateAndMount()
		dir, err = os.Open(m_decryptedDir)
	case KeyTypeKeystore:
		dir, err = os.Open(m_keystoreDir)
	default:
		err = errors.New("invalid key type")
		CheckError(err, "error listing keys")
	}

	CheckError(err, "Error opening directory")
	defer dir.Close()

	files, err := dir.Readdir(-1)
	CheckError(err, "Error reading directory")

	fmt.Printf("   " + strings.Repeat("-", 55) + "\n")
	fmt.Printf("   %-30s %-25s\n", "Name", "Created")
	fmt.Printf("   " + strings.Repeat("-", 55) + "\n")

	for _, file := range files {
		createdTime := file.ModTime().Format("02-01-2006 15:04:05")
		fmt.Printf("   %-30s %-25s\n", file.Name(), createdTime)
	}

	fmt.Printf("   " + strings.Repeat("-", 55) + "\n")
}

func InitGocryptfs(insecure bool) {
	initCmd := exec.Command("gocryptfs", "-init", "-plaintextnames", m_encryptedDir)

	RunCommandWithPassword(initCmd, "init", insecure)
}

func ValidateKeyName(keyName string) error {
	if len(keyName) == 0 {
		return ErrEmptyKeyName
	}

	if match, _ := regexp.MatchString("\\s", keyName); match {
		return ErrKeyContainsWhitespaces
	}
	return nil
}

func ExportKeyCmd(cCtx *cli.Context) {
	keyName := cCtx.String("key-name")
	keyType := cCtx.String("key-type")

	err := ValidateKeyName(keyName)
	CheckError(err, "Error validating key name")

	switch keyType {
	case KeyTypeGoCryptFS:
		ValidateAndMount()
		ExportGoCryptfsKey(keyName)
	case KeyTypeKeystore:
		ExportKeystoreKey(keyName)
	default:
		var err error
		err = errors.New("invalid key type")
		CheckError(err, "error exporting key")
	}

	fmt.Printf("Exported key: %s\n", keyName)
}

func DeleteKey(keyName string) {
	keyFile := m_decryptedDir + "/" + keyName
	err := os.Remove(keyFile)
	CheckError(err, "Error deleting key\n")

	fmt.Printf("Deleted key: %s\n", keyName)
}

func GetPrivateKeyFromUser() string {
	fmt.Print("Enter private key: ")
	return ReadHiddenInput()
}

func CreateGoCryptfsKey(keyName string) {
	keyFile := m_decryptedDir + "/" + keyName

	if !AllowKeyOverwrite(keyFile) {
		return
	}

	privateKey := GenerateRandomKey()
	privateKeyHex := hex.EncodeToString(privateKey.D.Bytes())
	CreateKeyFileAndStoreKey(keyFile, privateKeyHex)

	fmt.Printf("Created key: %s\n", keyName)
}

func CreateKeyFileAndStoreKey(keyFile string, privateKey string) {
	file, err := os.Create(keyFile)
	CheckError(err, "Error creating file")
	defer file.Close()

	_, err = file.WriteString(privateKey)
	CheckError(err, "Error writing to file")
}

func CreateKeystoreKey(keyName string, insecure bool) {
	keyFileName := keyName + m_keystoreSuffix
	keyFile := m_keystoreDir + "/" + keyFileName

	if !AllowKeyOverwrite(keyFile) {
		return
	}

	password := GetPasswordFromPrompt(insecure, "create")
	privateKey := GenerateRandomKey()
	err := sdkEcdsa.WriteKey(keyFile, privateKey, password)
	CheckError(err, "Error Writing ecdsa key")

	fmt.Printf("Created key: %s\n", keyName)
}

func ImportGoCryptfsKey(keyName string) {
	keyFile := m_decryptedDir + "/" + keyName

	if !AllowKeyOverwrite(keyFile) {
		return
	}

	privateKey := GetPrivateKeyFromUser()
	privateKey = strings.TrimPrefix(privateKey, "0x")
	CreateKeyFileAndStoreKey(keyFile, privateKey)
	fmt.Printf("Imported key: %s\n", keyName)
}

func ImportKeystoreKey(keyName string, insecure bool) {
	keyFileName := keyName + m_keystoreSuffix
	keyFile := m_keystoreDir + "/" + keyFileName

	if !AllowKeyOverwrite(keyFile) {
		return
	}

	password := GetPasswordFromPrompt(insecure, "import")

	privateKey := GetPrivateKeyFromUser()
	privateKey = strings.TrimPrefix(privateKey, "0x")
	privateKeyPair, _ := crypto.HexToECDSA(privateKey)

	err := sdkEcdsa.WriteKey(keyFile, privateKeyPair, password)
	CheckError(err, "Error Writing ecdsa key")
	fmt.Printf("Imported key: %s\n", keyName)
}

func ExportGoCryptfsKey(keyName string) {
	privateKey := GetPrivateKeyFromFile(keyName)
	_, publicKey := GetECDSAPrivateAndPublicKey(privateKey)

	fmt.Println("Public key : ", publicKey)
	fmt.Println("Private key : ", privateKey)
}

func ExportKeystoreKey(keyName string) {
	keyFileName := keyName + m_keystoreSuffix
	keyFile := m_keystoreDir + "/" + keyFileName

	password := GetPasswordFromPrompt(true, "export")

	key, err := sdkEcdsa.ReadKey(keyFile, password)
	CheckError(err, "Error reading ecdsa key")

	privateKey := hex.EncodeToString(key.D.Bytes())

	fmt.Println("Public key : ", GetPublicAddressFromPrivateKey(key))
	fmt.Println("Private key : ", privateKey)
}

func DeleteGoCryptfsKey(keyName string) {
	keyFile := m_decryptedDir + "/" + keyName
	err := os.Remove(keyFile)
	CheckError(err, "Error deleting key\n")
}

func DeleteKeystoreKey(keyName string) {
	keyFileName := keyName + m_keystoreSuffix
	keyFile := m_keystoreDir + "/" + keyFileName
	err := os.Remove(keyFile)
	CheckError(err, "Error deleting key\n")
}

func ValidEncryptedDir() bool {
	_, err := os.Stat(m_goCryptFSConfig)

	return !os.IsNotExist(err)
}

func GetPrivateKeyFromFile(keyName string) string {
	keyFile := m_decryptedDir + "/" + keyName
	data, err := os.ReadFile(keyFile)
	CheckError(err, "Error reading key file")
	return string(data)
}

func GetKeystorePrivateKey(keyName string) string {
	keyFileName := keyName + m_keystoreSuffix
	keyFile := filepath.Join(m_keystoreDir, keyFileName)

	password := GetPasswordFromPrompt(true, "export "+keyFileName)

	key, err := sdkEcdsa.ReadKey(keyFile, password)
	CheckError(err, "Error reading ecdsa key")

	privateKey := hex.EncodeToString(key.D.Bytes())

	return privateKey
}

func UseEncryptedKeys(keyType string) {
	m_useEncryptedKeys = true
	if keyType == KeyTypeGoCryptFS {
		ValidateAndMount() //This can be moved into ProcessConfigKeyPath but user needs to make sure proper keytype is giving for a given custom path
	}
}

func ProcessConfigKeyPath(keyPath string, keyType string) {
	dir, file := filepath.Split(keyPath)

	var defaultPath string

	switch keyType {
	case KeyTypeGoCryptFS:
		defaultPath = m_encryptedDir
	case KeyTypeKeystore:
		defaultPath = m_keystoreDir
	default:
		var err error
		err = errors.New("invalid key type")
		CheckError(err, "error processing key path")
	}

	if file == keyPath {
		fmt.Printf("Using the default key path : %s\n", defaultPath)
		// this means they have given only the key name only,
		// do nothing and use default path
		return
	}

	// go to the grand parent directory of the key path to get the .encrypted_keys path
	parentPathGoCryptFS := filepath.Dir(filepath.Dir(dir))

	m_encryptedDir = filepath.Join(parentPathGoCryptFS, EncryptedDirName)
	m_decryptedDir = filepath.Join(parentPathGoCryptFS, DecryptedDirName)
	m_goCryptFSConfig = filepath.Join(parentPathGoCryptFS, GoCryptFSConfigName)

	m_keystoreDir = dir

	m_isFullPath = true

	if keyType == KeyTypeGoCryptFS {
		fmt.Printf("Using the key path : %s\n", m_encryptedDir)
	} else {
		fmt.Printf("Using the key path : %s\n", m_keystoreDir)
	}

}

func RetryMounting() {
	m_retryMounting = true
}

func GetPrivateKey(key string, keyType string) string {
	if m_useEncryptedKeys {
		keyName := key
		if m_isFullPath {
			_, keyName = filepath.Split(key)
		}

		if keyType == "gocryptfs" {
			return GetPrivateKeyFromFile(keyName)
		} else {
			return GetKeystorePrivateKey(keyName)
		}
	}
	return key
}

func GenerateRandomKey() *ecdsa.PrivateKey {
	privateKey, err := crypto.GenerateKey()
	CheckError(err, "Error generating key")

	return privateKey
}
