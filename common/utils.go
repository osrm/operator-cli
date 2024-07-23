package wc_common

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/witnesschain-com/operator-cli/common/bindings/AvsDirectory"
	"github.com/witnesschain-com/operator-cli/common/bindings/OperatorRegistry"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var m_isMounted bool = false

type Filesystem struct {
	Target string `json:"target"`
}

type MountResult struct {
	Filesystems []Filesystem `json:"filesystems"`
}

var password string

func ConnectToUrl(url string) (*ethclient.Client, *big.Int) {
	client, err := ethclient.Dial(url)
	CheckError(err, "Connection to RPC failed")

	id, err := client.ChainID(context.Background())
	CheckError(err, "Unable to retrive chainID for : "+url)

	fmt.Println("Connection successful : ", id)

	return client, id
}

func GetECDSAPrivateKey(privateKeyString string) *ecdsa.PrivateKey {
	ecdsaPrivateKey, err := crypto.HexToECDSA(privateKeyString)
	CheckError(err, "Converting private key to ECDSA format failed")
	return ecdsaPrivateKey
}

func GetPublicAddressFromPrivateKey(privateKey *ecdsa.PrivateKey) common.Address {
	publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		FatalError("Error casting public key to ECDSA")
	}

	return crypto.PubkeyToAddress(*publicKeyECDSA)
}

func GetECDSAPrivateAndPublicKey(privateKeyString string) (*ecdsa.PrivateKey, common.Address) {
	ecdsaPrivateKey := GetECDSAPrivateKey(privateKeyString)
	ecdsaPublicKey := GetPublicAddressFromPrivateKey(ecdsaPrivateKey)
	return ecdsaPrivateKey, ecdsaPublicKey
}

func GetPaddedValue(value []byte) [32]byte {
	var paddedValue [32]byte
	startIndex := len(paddedValue) - len(value)
	copy(paddedValue[startIndex:], value)

	return paddedValue
}

func CalculateExpiry(client *ethclient.Client, expectedExpiryDays uint64) *big.Int {
	// Get the latest block header
	header, err := client.HeaderByNumber(context.Background(), nil)
	CheckError(err, "Could not get HeaderByNumber")

	// Get the current timestamp from the latest block header
	currentTimestamp := big.NewInt(int64(header.Time))

	expiryInSeconds := int64(expectedExpiryDays * 24 * 60 * 60)
	timeToElapse := big.NewInt(expiryInSeconds)

	expiry := new(big.Int).Add(currentTimestamp, timeToElapse)
	return expiry
}

func GenerateSalt() [32]byte {
	var salt [32]byte

	// Generate random bytes
	_, err := rand.Read(salt[:])
	CheckError(err, "Generating salt failed")

	return salt
}

func ValidateAndMount() {
	CheckIfGocryptfsIsInstalled()

	if !ValidEncryptedDir() {
		FatalErrorWithoutUnmount(fmt.Sprintf("%v: %s\n", ErrInvalidEncryptedDirectory,
			" : check if "+m_goCryptFSConfig+" exist. Or try initiating again after deleting those directories"))
	}

	if IsAlreadyMounted() {
		if !m_retryMounting {
			FatalErrorWithoutUnmount(m_gocryptfsDecDir + " already mounted")
		}

		fmt.Println("GoCryptFS filesystem already mounted")
		for i := 0; i < MaxMountRetries; i++ {
			fmt.Printf("Retrying in %v seconds\n", RetryPeriodInSeconds)

			// RetryPeriodInSeconds
			time.Sleep(time.Duration(RetryPeriodInSeconds * uint(time.Second)))
			Mount()
			if m_isMounted {
				return
			}
		}
		FatalErrorWithoutUnmount("Giving up, " + m_gocryptfsDecDir + " already mounted")
	} else {
		Mount()
	}
}

func Mount() {
	if IsAlreadyMounted() {
		return
	}

	mountCmd := exec.Command("gocryptfs", m_gocryptfsEncDir, m_gocryptfsDecDir)
	RunCommandWithPassword(mountCmd, "mount", true)

	m_isMounted = true
}

func Unmount() {
	if !m_isMounted {
		return
	}

	umountCmd := exec.Command("fusermount", "-u", m_gocryptfsDecDir)
	err := umountCmd.Run()
	if err != nil {
		CheckErrorWithoutUnmount(err, "Error unmounting GoCryptFS filesystem")
	}

	m_isMounted = false
}

func IsAlreadyMounted() bool {
	cmd := exec.Command("findmnt", "-n", "-o", "TARGET", "--type", "fuse.gocryptfs", "-J")
	output, err := cmd.CombinedOutput()
	CheckError(err, "Error checking if filesystem is mounted. Output - "+string(output))

	var mountResult MountResult
	err = json.Unmarshal(output, &mountResult)

	CheckError(err, "Error checking if filesystem is mounted. Output - "+string(output))

	for _, fs := range mountResult.Filesystems {
		absolutePath, err := filepath.Abs(m_gocryptfsDecDir)
		CheckError(err, "Error getting absolute path")

		if absolutePath == fs.Target {
			return true
		}
	}
	return false
}

func DirectoryExists(path string) bool {
	fileInfo, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	}

	CheckError(err, "Error checking directory")

	if fileInfo.Mode().IsRegular() {
		CheckError(ErrNotADirectory, path)
	}

	return true
}

func CreateDirectory(path string) {
	fmt.Println("Creating directory: ", path)
	err := os.Mkdir(path, 0755)
	CheckError(err, "Error creating directory")
}

func RunCommandWithPassword(cmd *exec.Cmd, desc string, insecure bool) {
	fmt.Printf("Enter password to %s: ", desc)
	if len(password) == 0 {
		password = ReadHiddenInput()
	}

	if !insecure {
		ValidatePassword(password)
	}

	cmdStdin, err := cmd.StdinPipe()
	CheckError(err, "Error creating stdin pipe for "+desc)

	err = cmd.Start()
	CheckError(err, "Error starting command for "+desc)

	_, err = cmdStdin.Write([]byte(password))
	CheckError(err, "Error writing to command stdin for "+desc)

	err = cmdStdin.Close()
	CheckError(err, "Error closing command stdin for "+desc)

	err = cmd.Wait()
	CheckError(err, "Command failed for "+desc)
}

func AllowKeyOverwrite(fileLoc string) bool {
	_, err := os.Stat(fileLoc)
	if !os.IsNotExist(err) {
		fmt.Printf("Key already exists, do you want to overwrite? (y/n): ")
		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) != "y" {
			return false
		}
	}

	return true
}

func GetPasswordFromPrompt(insecure bool, desc string) string {
	fmt.Printf("Enter password to %s: ", desc)
	password := ReadHiddenInput()

	if !insecure {
		ValidatePassword(password)
	}

	return password
}

func IsWatchtowerRegistered(watchtower common.Address, operatorRegistry *OperatorRegistry.OperatorRegistry) bool {
	registered, err := operatorRegistry.IsValidWatchtower(&bind.CallOpts{}, watchtower)
	CheckError(err, "Error checking if watchtower is already registered")
	return registered
}

func IsOperatorWhitelisted(operator common.Address, operatorRegistry *OperatorRegistry.OperatorRegistry) bool {
	active, err := operatorRegistry.IsActiveOperator(&bind.CallOpts{}, operator)
	CheckError(err, "Error checking if operator is whitelisted")
	return active
}

func IsOperatorRegistered(witnessHubAddress common.Address, operator common.Address, avsDirectory *AvsDirectory.AvsDirectory) bool {
	status, err := avsDirectory.AvsOperatorStatus(&bind.CallOpts{}, witnessHubAddress, operator)
	CheckError(err, "Checking operator status failed")
	return status != 0
}

func CheckIfGocryptfsIsInstalled() {
	cmd := exec.Command("gocryptfs", "--version")
	err := cmd.Run()

	CheckErrorWithoutUnmount(err, "Check if gocryptfs is installed")
}
