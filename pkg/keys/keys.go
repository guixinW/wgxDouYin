package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

// CreateKeyPair 创建一个私钥，并通过这个私钥返回公钥
func CreateKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	publicKey := &privateKey.PublicKey
	return privateKey, publicKey, nil
}

func GetKey(keyName string) (*ecdsa.PrivateKey, error) {
	path, err := getKeyPath()
	if err != nil {
		return nil, err
	}
	keyPath := filepath.Join(path, keyName)
	if _, err = os.Stat(keyPath); os.IsNotExist(err) {
		privateKey, _, err := CreateKeyPair()
		if err != nil {
			return nil, err
		}
		err = savePrivateKey(keyPath, privateKey)
		if err != nil {
			return nil, err
		}
		return privateKey, nil
	}
	return loadPrivateKey(keyPath)
}

func getKeyPath() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	keysDir := filepath.Join(currentDir, "keys")
	privateKeyFilePath := filepath.Join(keysDir)
	return privateKeyFilePath, nil
}

func savePrivateKey(path string, privateKey *ecdsa.PrivateKey) error {
	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal ECDSA private key: %w", err)
	}
	block := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer file.Close()
	err = pem.Encode(file, block)
	if err != nil {
		return fmt.Errorf("failed to write ECDSA key to file: %w", err)
	}

	return nil
}

func loadPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("invalid ECDSA key file format")
	}
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ECDSA private key: %w", err)
	}

	return privateKey, nil
}
