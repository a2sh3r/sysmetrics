// Package crypto provides encryption and decryption functionality using RSA keys.
package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// Decryptor handles decryption operations using a private RSA key.
type Decryptor struct {
	privateKey *rsa.PrivateKey
}

// NewDecryptor creates a new Decryptor from a private key file.
func NewDecryptor(privateKeyPath string) (*Decryptor, error) {
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	var privateKey *rsa.PrivateKey
	switch block.Type {
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		var priv interface{}
		priv, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err == nil {
			privateKey = priv.(*rsa.PrivateKey)
		}
	default:
		return nil, fmt.Errorf("unsupported key type: %s", block.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &Decryptor{privateKey: privateKey}, nil
}

// Decrypt decrypts data using the private RSA key.
func (d *Decryptor) Decrypt(encryptedData []byte) ([]byte, error) {
	if len(encryptedData) == 0 {
		return nil, fmt.Errorf("encrypted data is empty")
	}

	if len(encryptedData) > 4 {
		aesKeyLength := int(encryptedData[0])<<24 | int(encryptedData[1])<<16 | int(encryptedData[2])<<8 | int(encryptedData[3])
		if len(encryptedData) > aesKeyLength+4 {
			encryptedAESKey := encryptedData[4 : aesKeyLength+4]
			encryptedPayload := encryptedData[aesKeyLength+4:]

			aesKey, err := rsa.DecryptPKCS1v15(nil, d.privateKey, encryptedAESKey)
			if err != nil {
				return rsa.DecryptPKCS1v15(nil, d.privateKey, encryptedData)
			}

			decryptedData := make([]byte, len(encryptedPayload))
			for i := range encryptedPayload {
				decryptedData[i] = encryptedPayload[i] ^ aesKey[i%len(aesKey)]
			}

			return decryptedData, nil
		}
	}

	return rsa.DecryptPKCS1v15(nil, d.privateKey, encryptedData)
}
