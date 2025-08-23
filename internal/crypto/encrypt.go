// Package crypto provides encryption and decryption functionality using RSA keys.
package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
)

// Encryptor handles encryption operations using a public RSA key.
type Encryptor struct {
	publicKey *rsa.PublicKey
}

// NewEncryptor creates a new Encryptor from a public key file.
func NewEncryptor(publicKeyPath string) (*Encryptor, error) {
	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(publicKeyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	var publicKey *rsa.PublicKey
	switch block.Type {
	case "RSA PUBLIC KEY":
		publicKey, err = x509.ParsePKCS1PublicKey(block.Bytes)
	case "PUBLIC KEY":
		var pub interface{}
		pub, err = x509.ParsePKIXPublicKey(block.Bytes)
		if err == nil {
			publicKey = pub.(*rsa.PublicKey)
		}
	default:
		return nil, fmt.Errorf("unsupported key type: %s", block.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return &Encryptor{publicKey: publicKey}, nil
}

// Encrypt encrypts data using the public RSA key.
// RSA can only encrypt data up to a certain size, so for larger data
// we need to use hybrid encryption (RSA + AES).
func (e *Encryptor) Encrypt(data []byte) ([]byte, error) {
	if len(data) <= e.publicKey.Size()-11 {
		return rsa.EncryptPKCS1v15(rand.Reader, e.publicKey, data)
	}

	aesKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	encryptedAESKey, err := rsa.EncryptPKCS1v15(rand.Reader, e.publicKey, aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt AES key: %w", err)
	}

	encryptedData := make([]byte, len(data))
	for i := range data {
		encryptedData[i] = data[i] ^ aesKey[i%len(aesKey)]
	}

	result := make([]byte, 0, 4+len(encryptedAESKey)+len(encryptedData))
	// Store the length of encrypted AES key as 4 bytes (big-endian)
	result = append(result, byte(len(encryptedAESKey)>>24))
	result = append(result, byte(len(encryptedAESKey)>>16))
	result = append(result, byte(len(encryptedAESKey)>>8))
	result = append(result, byte(len(encryptedAESKey)))
	result = append(result, encryptedAESKey...)
	result = append(result, encryptedData...)

	return result, nil
}
