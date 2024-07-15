package wc_common

const (
	DefaultAdminConfig   string  = "config/admin-config.json"
	DefaultOpL1Config    string  = "config/l1-operator-config.json"
	DefaultOpL2Config    string  = "config/l2-operator-config.json"
	GoCryptFSDirName     string  = ".gocryptfs"
	EncryptedDirName     string  = ".encrypted_keys"
	DecryptedDirName     string  = ".decrypted_keys"
	GoCryptFSConfigName  string  = EncryptedDirName + "/gocryptfs.conf"
	MinEntropyBits       float64 = 50
	MaxMountRetries      int     = 5
	RetryPeriodInSeconds uint    = 1
	KeyStoreDirName      string  = ".keystore"
	KeyStoreSuffixName   string  = ".ecdsa.key.json"
	KeyTypeGoCryptFS     string  = "gocryptfs"
	KeyTypeKeystore      string  = "keystore"
)
