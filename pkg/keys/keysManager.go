package keys

import (
	"crypto/ecdsa"
	"github.com/pkg/errors"
	"time"
)

var (
	timeout = time.Duration(1)
)

type KeyManager struct {
	serverToPublicKey map[string]*ecdsa.PublicKey
	serverPrivateKey  *ecdsa.PrivateKey
}

// NewKeyManager 创建一个keys manager,用于管理密钥；
// 若私钥为空，则说明这个keys manager被api router使用，用于存储各类微服务的密钥；
// 若私钥不为空，则说明这个keys manager被各类微服务使用，用于存储微服务自身的私钥
// 以及与其相关服务的公钥
func NewKeyManager(privateKey *ecdsa.PrivateKey, serverName string) (*KeyManager, error) {
	if privateKey == nil {
		var keyManager KeyManager
		keyManager.serverToPublicKey = make(map[string]*ecdsa.PublicKey)
		keyManager.serverPrivateKey = nil
		return &keyManager, nil
	}
	return &KeyManager{serverPrivateKey: privateKey,
		serverToPublicKey: map[string]*ecdsa.PublicKey{serverName: &privateKey.PublicKey}}, nil
}

func (j *KeyManager) Update(key, value []byte) error {
	publicKey, err := PEMToPublicKey(string(value))
	if err != nil {
		return errors.Wrap(err, "KeyManager.Update failed")
	}
	j.serverToPublicKey[string(key)] = publicKey
	return nil
}

// GetServerPublicKey retrieves the public key associated with specified service.
func (j *KeyManager) GetServerPublicKey(serverName string) (*ecdsa.PublicKey, error) {
	serverPublicKey, ok := j.serverToPublicKey[serverName]
	if !ok {
		return nil, errors.New("can't find service's public key")
	}
	return serverPublicKey, nil
}

// GetPrivateKey get the KeyManager private key.
func (j *KeyManager) GetPrivateKey() *ecdsa.PrivateKey {
	return j.serverPrivateKey
}

// addServerPublicKey saves the public key associated with specified service.
func (j *KeyManager) addServerPublicKey(serverName string, serverPublicKey *ecdsa.PublicKey) error {
	if j == nil {
		return errors.New("KeyManager is nil object")
	}
	j.serverToPublicKey[serverName] = serverPublicKey
	return nil
}

func (j *KeyManager) updatePublicKey(serverName string, publicKey *ecdsa.PublicKey) {
	j.serverToPublicKey[serverName] = publicKey
}
