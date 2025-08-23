package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
)

// generateTestKeys creates temporary RSA key pair for testing
func generateTestKeys(t *testing.T) (string, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	publicKey := &privateKey.PublicKey

	privateKeyFile, err := os.CreateTemp("", "private_key_*.pem")
	if err != nil {
		t.Fatalf("Failed to create temp private key file: %v", err)
	}
	defer func() {
		if closeErr := privateKeyFile.Close(); closeErr != nil {
			t.Logf("Warning: failed to close private key file: %v", closeErr)
		}
	}()

	publicKeyFile, err := os.CreateTemp("", "public_key_*.pem")
	if err != nil {
		t.Fatalf("Failed to create temp public key file: %v", err)
		defer func() {
			if closeErr := publicKeyFile.Close(); closeErr != nil {
				t.Logf("Warning: failed to close public key file: %v", closeErr)
			}
		}()
		if removeErr := os.Remove(privateKeyFile.Name()); removeErr != nil {
			t.Logf("Warning: failed to remove private key file: %v", removeErr)
		}
		t.Fatalf("Failed to create temp public key file: %v", err)
	}
	defer func() {
		if closeErr := publicKeyFile.Close(); closeErr != nil {
			t.Logf("Warning: failed to close public key file: %v", closeErr)
		}
	}()

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		t.Fatalf("Failed to encode private key: %v", err)
	}

	publicKeyPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(publicKey),
	}
	if err := pem.Encode(publicKeyFile, publicKeyPEM); err != nil {
		t.Fatalf("Failed to encode public key: %v", err)
	}

	return privateKeyFile.Name(), publicKeyFile.Name()
}

func TestEncryptDecrypt(t *testing.T) {
	privateKeyPath, publicKeyPath := generateTestKeys(t)
	defer func() {
		if removeErr := os.Remove(privateKeyPath); removeErr != nil {
			t.Logf("Warning: failed to remove private key file: %v", removeErr)
		}
		if removeErr := os.Remove(publicKeyPath); removeErr != nil {
			t.Logf("Warning: failed to remove public key file: %v", removeErr)
		}
	}()

	encryptor, err := NewEncryptor(publicKeyPath)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	decryptor, err := NewDecryptor(privateKeyPath)
	if err != nil {
		t.Fatalf("Failed to create decryptor: %v", err)
	}

	testData := []byte("Hello, World! This is a test message for encryption.")

	encrypted, err := encryptor.Encrypt(testData)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	decrypted, err := decryptor.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if !bytes.Equal(testData, decrypted) {
		t.Errorf("Decrypted data doesn't match original. Got: %s, Want: %s", decrypted, testData)
	}
}

func TestEncryptDecryptLargeData(t *testing.T) {
	privateKeyPath, publicKeyPath := generateTestKeys(t)
	defer func() {
		if removeErr := os.Remove(privateKeyPath); removeErr != nil {
			t.Logf("Warning: failed to remove private key file: %v", removeErr)
		}
		if removeErr := os.Remove(publicKeyPath); removeErr != nil {
			t.Logf("Warning: failed to remove public key file: %v", removeErr)
		}
	}()

	encryptor, err := NewEncryptor(publicKeyPath)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	decryptor, err := NewDecryptor(privateKeyPath)
	if err != nil {
		t.Fatalf("Failed to create decryptor: %v", err)
	}

	largeData := make([]byte, 1000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	encrypted, err := encryptor.Encrypt(largeData)
	if err != nil {
		t.Fatalf("Failed to encrypt large data: %v", err)
	}

	decrypted, err := decryptor.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt large data: %v", err)
	}

	if !bytes.Equal(largeData, decrypted) {
		t.Errorf("Decrypted large data doesn't match original")
	}
}

func TestNewEncryptorInvalidPath(t *testing.T) {
	_, err := NewEncryptor("/nonexistent/path")
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}

func TestNewDecryptorInvalidPath(t *testing.T) {
	_, err := NewDecryptor("/nonexistent/path")
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}
