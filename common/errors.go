package wc_common

import (
	"errors"
	"fmt"
	"log"
)

var (
	ErrEmptyKeyName              = errors.New("key name cannot be empty")
	ErrKeyContainsWhitespaces    = errors.New("key name cannot contain spaces")
	ErrInvalidPassword           = errors.New("invalid password")
	ErrNotADirectory             = errors.New("is not a directory")
	ErrInvalidEncryptedDirectory = errors.New("invalid gocryptfs encrypted directory")
	ErrInvalidKeyType            = errors.New("invalid key type")
)

func CheckErrorWithoutUnmount(err error, description string) {
	if err != nil {
		FatalErrorWithoutUnmount(fmt.Sprintf("%s: %v\n", description, err))
	}
}

func CheckError(err error, description string) {
	if err != nil {
		FatalError(fmt.Sprintf("%s: %v\n", description, err))
	}
}

func FatalError(description string) {
	Unmount()
	log.Fatalf(description)
}

func FatalErrorWithoutUnmount(description string) {
	log.Fatalf(description)
}
