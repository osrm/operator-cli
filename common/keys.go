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

var m_gocryptfsDirName string = filepath.Join(GetUserHomeDir(), WitnesschainCLIPath, GoCryptFSDirName)
var m_gocryptfsEncDir string = filepath.Join(m_gocryptfsDirName, GocryptfsEncDirName)
var m_gocryptfsDecDir string = filepath.Join(m_gocryptfsDirName, GocryptfsDecDirName)
var m_goCryptFSConfig string = filepath.Join(m_gocryptfsEncDir, GoCryptFSConfigName)

var m_w3SecretKeyDir string = filepath.Join(GetUserHomeDir(), WitnesschainCLIPath, W3SecretKeyDirName)
var m_w3SecretKeysPassword string = ""

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
			if cCtx.String("key-name") == "" {
				CheckError(ErrEmptyKeyName, "Required flag \"key-name\" not set")
			}
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
		CheckIfGocryptfsIsInstalled()
		if !DirectoryExists(GoCryptFSDirName) {
			CreateDirectory(GoCryptFSDirName)
		}
		if !DirectoryExists(m_gocryptfsEncDir) {
			CreateDirectory(m_gocryptfsEncDir)
		}

		if !DirectoryExists(m_gocryptfsDecDir) {
			CreateDirectory(m_gocryptfsDecDir)
		}
		InitGocryptfs(insecure)
	case KeyTypeW3SecretKey:
		if !DirectoryExists(m_w3SecretKeyDir) {
			CreateDirectory(m_w3SecretKeyDir)
		}
		fmt.Println("Init keystore done")
	default:
		CheckError(ErrInvalidKeyType, "Error initializing key store : "+keyType)
	}
}

func CreateKeyCmd(cCtx *cli.Context) {
	keyName := cCtx.String("key-name")
	keyType := cCtx.String("key-type")
	insecure := cCtx.Bool("insecure")

	switch keyType {
	case KeyTypeGoCryptFS:
		ValidateAndMount()
		CreateGoCryptfsKey(keyName)
	case KeyTypeW3SecretKey:
		CreateW3SecretKey(keyName, insecure)
	default:
		CheckError(ErrInvalidKeyType, "Error creating key : "+keyType)
	}
}

func ImportKeyCmd(cCtx *cli.Context) {
	keyName := cCtx.String("key-name")
	keyType := cCtx.String("key-type")
	insecure := cCtx.Bool("insecure")

	switch keyType {
	case KeyTypeGoCryptFS:
		ValidateAndMount()
		ImportGoCryptfsKey(keyName)
	case KeyTypeW3SecretKey:
		ImportW3SecretKey(keyName, insecure)
	default:
		CheckError(ErrInvalidKeyType, "Error importing key : "+keyType)
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
			if cCtx.String("key-name") == "" {
				CheckError(ErrEmptyKeyName, "Required flag \"key-name\" not set")
			}
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
	case KeyTypeW3SecretKey:
		DeleteW3SecretKey(keyName)
	default:
		CheckError(ErrInvalidKeyType, "Error deleting key : "+keyType)
	}

	fmt.Printf("Deleted key: %s\n", keyName)
}

func ListKeyCmd(cCtx *cli.Context) {
	keyType := cCtx.String("key-type")

	var err error = nil
	var dir *os.File
	var path string
	switch keyType {
	case KeyTypeGoCryptFS:
		dir, err = os.Open(m_gocryptfsEncDir)
		path, _ = filepath.Abs(m_gocryptfsEncDir)
	case KeyTypeW3SecretKey:
		dir, err = os.Open(m_w3SecretKeyDir)
		path, _ = filepath.Abs(m_w3SecretKeyDir)
	default:
		CheckErrorWithoutUnmount(ErrInvalidKeyType, "error listing keys")
	}

	CheckError(err, "Error opening directory")
	defer dir.Close()

	files, err := dir.Readdir(-1)
	CheckError(err, "Error reading directory")

	separatorLen := len(path) + 100
	nameLen := len(path) + 75
	fmt.Printf("   " + strings.Repeat("-", separatorLen) + "\n")
	fmt.Printf("   %-*s %-25s\n", nameLen, "Name", "Created")
	fmt.Printf("   " + strings.Repeat("-", separatorLen) + "\n")

	for _, file := range files {
		if file.Name() == GoCryptFSConfigName {
			continue
		}

		createdTime := file.ModTime().Format("02-01-2006 15:04:05")

		fmt.Printf("   %-*s %-25s\n", nameLen, filepath.Join(path, file.Name()), createdTime)
	}

	fmt.Printf("   " + strings.Repeat("-", separatorLen) + "\n")
}

func InitGocryptfs(insecure bool) {
	initCmd := exec.Command("gocryptfs", "-init", "-plaintextnames", m_gocryptfsEncDir)

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
	case KeyTypeW3SecretKey:
		ExportW3SecretKey(keyName)
	default:
		var err error
		err = errors.New("invalid key type")
		CheckError(err, "error exporting key")
	}

	fmt.Printf("Exported key: %s\n", keyName)
}

func GetPrivateKeyFromUser() string {
	fmt.Print("Enter private key: ")
	return ReadHiddenInput()
}

func CreateGoCryptfsKey(keyName string) {
	privateKey := GenerateRandomKey()

	address := GetPublicAddressFromPrivateKey(privateKey)

	if keyName == "" {
		keyName = address.String()
	}

	err := ValidateKeyName(keyName)
	CheckError(err, "Error validating key name")

	keyFile := filepath.Join(m_gocryptfsDecDir, keyName)

	if !AllowKeyOverwrite(keyFile) {
		return
	}

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

func CreateW3SecretKey(keyName string, insecure bool) {
	password := GetPasswordFromPrompt(insecure, "create")
	privateKey := GenerateRandomKey()

	address := GetPublicAddressFromPrivateKey(privateKey)

	if keyName == "" {
		keyName = address.String()
	}

	err := ValidateKeyName(keyName)
	CheckError(err, "Error validating key name")

	keyFileName := keyName + W3SecretKeySuffixName
	keyFile := filepath.Join(m_w3SecretKeyDir, keyFileName)

	if !AllowKeyOverwrite(keyFile) {
		return
	}

	err = sdkEcdsa.WriteKey(keyFile, privateKey, password)
	CheckError(err, "Error Writing ecdsa key")

	fmt.Printf("Created key: %s\n", keyName)
}

func ImportGoCryptfsKey(keyName string) {
	privateKey := GetPrivateKeyFromUser()
	privateKey = strings.TrimPrefix(privateKey, "0x")

	privKeyBytes, err := hex.DecodeString(privateKey)
	CheckError(err, "Error decoding private key hex string")

	privKey, err := crypto.ToECDSA(privKeyBytes)
	CheckError(err, "Error converting bytes to ECDSA private key")

	address := GetPublicAddressFromPrivateKey(privKey)

	if keyName == "" {
		keyName = address.String()
	}

	err = ValidateKeyName(keyName)
	CheckError(err, "Error validating key name")

	keyFile := filepath.Join(m_gocryptfsDecDir, keyName)

	if !AllowKeyOverwrite(keyFile) {
		return
	}

	CreateKeyFileAndStoreKey(keyFile, privateKey)
	fmt.Printf("Imported key: %s\n", keyName)
}

func ImportW3SecretKey(keyName string, insecure bool) {
	password := GetPasswordFromPrompt(insecure, "import")

	privateKey := GetPrivateKeyFromUser()
	privateKey = strings.TrimPrefix(privateKey, "0x")
	privateKeyPair, _ := crypto.HexToECDSA(privateKey)
	address := GetPublicAddressFromPrivateKey(privateKeyPair)

	if keyName == "" {
		keyName = address.String()
	}

	err := ValidateKeyName(keyName)
	CheckError(err, "Error validating key name")

	keyFileName := keyName + W3SecretKeySuffixName
	keyFile := filepath.Join(m_w3SecretKeyDir, keyFileName)

	if !AllowKeyOverwrite(keyFile) {
		return
	}

	err = sdkEcdsa.WriteKey(keyFile, privateKeyPair, password)
	CheckError(err, "Error Writing ecdsa key")
	fmt.Printf("Imported key: %s\n", keyName)
}

func ExportGoCryptfsKey(keyName string) {
	privateKey := GetGocryptfsPrivateKey(keyName)
	_, publicKey := GetECDSAPrivateAndPublicKey(privateKey)

	fmt.Println("Public key : ", publicKey)
	fmt.Println("Private key : ", privateKey)
}

func ExportW3SecretKey(keyName string) {
	password := GetPasswordFromPrompt(true, "export")

	key, err := sdkEcdsa.ReadKey(GetSanitizedW3SecretKeyName(keyName), password)
	CheckError(err, "Error reading ecdsa key")

	privateKey := hex.EncodeToString(key.D.Bytes())

	fmt.Println("Public key : ", GetPublicAddressFromPrivateKey(key))
	fmt.Println("Private key : ", privateKey)
}

func DeleteGoCryptfsKey(keyName string) {
	err := os.Remove(GetSanitizedGocryptfsKeyName(keyName))
	CheckError(err, "Error deleting key\n")
}

func DeleteW3SecretKey(keyName string) {
	err := os.Remove(GetSanitizedW3SecretKeyName(keyName))
	CheckError(err, "Error deleting key\n")
}

func ValidEncryptedDir() bool {
	_, err := os.Stat(m_goCryptFSConfig)

	return !os.IsNotExist(err)
}

func GetGocryptfsPrivateKey(keyName string) string {
	keyFile := GetSanitizedGocryptfsKeyName(keyName)
	data, err := os.ReadFile(keyFile)
	CheckError(err, "Error reading key file "+keyFile)
	return string(data)
}

func GetW3SecretStoragePrivateKey(keyName string) string {
	if m_w3SecretKeysPassword == "" {
		m_w3SecretKeysPassword = GetPasswordFromPrompt(true, "export web3 secret storage keys")
	}

	key, err := sdkEcdsa.ReadKey(GetSanitizedW3SecretKeyName(keyName), m_w3SecretKeysPassword)
	CheckError(err, "Error reading ecdsa key")

	privateKey := hex.EncodeToString(key.D.Bytes())

	return privateKey
}

func UseEncryptedKeys(keyType string) {
	if m_useEncryptedKeys {
		return
	}

	m_useEncryptedKeys = true
	if keyType == KeyTypeGoCryptFS {
		ValidateAndMount()
	}
}

func ProcessConfigKeyPath(keyPath string, keyType string) {
	dir, file := filepath.Split(keyPath)

	switch keyType {
	case KeyTypeGoCryptFS:
		if file != keyPath {
			// go to the grand parent directory of the key path to get the .encrypted_keys path
			parentPathGoCryptFS := filepath.Dir(filepath.Dir(dir))
			m_gocryptfsEncDir = filepath.Join(parentPathGoCryptFS, GocryptfsEncDirName)
			m_gocryptfsDecDir = filepath.Join(parentPathGoCryptFS, GocryptfsDecDirName)
			m_goCryptFSConfig = filepath.Join(m_gocryptfsEncDir, GoCryptFSConfigName)
			m_isFullPath = true
		}
		fmt.Printf("Using the key path : %s\n", m_gocryptfsEncDir)

	case KeyTypeW3SecretKey:
		if file != keyPath {
			m_w3SecretKeyDir = dir
			m_isFullPath = true
		}
		fmt.Printf("Using the key path : %s\n", m_w3SecretKeyDir)

	default:
		CheckError(ErrInvalidKeyType, "Error processing key path : "+keyType)
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

		switch keyType {
		case KeyTypeGoCryptFS:
			return GetGocryptfsPrivateKey(keyName)
		case KeyTypeW3SecretKey:
			return GetW3SecretStoragePrivateKey(keyName)
		default:
			CheckError(ErrInvalidKeyType, "Error processing key path : "+keyType)
		}
	}
	return key
}

func GenerateRandomKey() *ecdsa.PrivateKey {
	privateKey, err := crypto.GenerateKey()
	CheckError(err, "Error generating key")

	return privateKey
}

func LoadPrivateKey(path string, keyType string) (*ecdsa.PrivateKey, error) {
	priv, err := crypto.HexToECDSA(GetPrivateKey(path, keyType))
	if err != nil {
		return nil, err
	}
	return priv, nil
}

func GetSanitizedW3SecretKeyName(keyName string) string {
	keyFileName := keyName
	if len(filepath.Ext(keyName)) == 0 {
		keyFileName = keyName + W3SecretKeySuffixName
	}

	keyFile := keyFileName
	if !filepath.IsAbs(keyFile) {
		keyFile = filepath.Join(m_w3SecretKeyDir, keyFileName)
	}

	return keyFile
}

func GetSanitizedGocryptfsKeyName(keyName string) string {
	keyFile := keyName
	if !filepath.IsAbs(keyFile) {
		keyFile = filepath.Join(m_gocryptfsDecDir, keyName)
	}

	return keyFile
}
