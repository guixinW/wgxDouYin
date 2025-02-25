package keys

import (
	"crypto/ecdsa"
	"github.com/pkg/errors"
	"sync"
	"time"
)

var (
	timeout = time.Duration(1)
)

type ServerToKeyMap struct {
	data sync.Map
}

func (p *ServerToKeyMap) Store(key string, value *ecdsa.PublicKey) {
	p.data.Store(key, value)
}

func (p *ServerToKeyMap) Load(key string) (*ecdsa.PublicKey, bool) {
	value, ok := p.data.Load(key)
	if !ok {
		return nil, false
	}
	return value.(*ecdsa.PublicKey), true
}

type KeyManager struct {
	serverToPublicKey ServerToKeyMap
	serverPrivateKey  *ecdsa.PrivateKey
}

// NewKeyManager 创建一个keys manager,用于管理密钥；
// 若私钥为空，则说明这个keys manager被api router使用，用于存储各类微服务的密钥；
// 若私钥不为空，则说明这个keys manager被各类微服务使用，用于存储微服务自身的私钥
// 以及与其相关服务的公钥
func NewKeyManager(privateKey *ecdsa.PrivateKey, serverName string) (*KeyManager, error) {
	if privateKey == nil {
		var keyManager KeyManager
		keyManager.serverPrivateKey = nil
		return &keyManager, nil
	}
	newKeyManager := KeyManager{}
	newKeyManager.serverToPublicKey = ServerToKeyMap{}
	newKeyManager.serverPrivateKey = privateKey
	err := newKeyManager.addServerPublicKey(serverName, &privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	return &newKeyManager, nil
}

func (j *KeyManager) Update(key, value []byte) error {
	publicKey, err := PEMToPublicKey(string(value))
	serverName := string(key)
	if err != nil {
		return errors.Wrap(err, "KeyManager.Update failed")
	}
	err = j.addServerPublicKey(serverName, publicKey)
	if err != nil {
		return err
	}
	return nil
}

// GetServerPublicKey 获取服务的公钥
func (j *KeyManager) GetServerPublicKey(serverName string) (*ecdsa.PublicKey, error) {
	serverPublicKey, ok := j.serverToPublicKey.Load(serverName)
	if !ok {
		return nil, errors.New("can't find service's public key")
	}
	return serverPublicKey, nil
}

// GetPrivateKey 获取使用该KeyManager的服务私钥
func (j *KeyManager) GetPrivateKey() *ecdsa.PrivateKey {
	if j != nil {
		return j.serverPrivateKey
	}
	return nil
}

// addServerPublicKey 通过服务名保存公钥
func (j *KeyManager) addServerPublicKey(serverName string, serverPublicKey *ecdsa.PublicKey) error {
	if j == nil {
		return errors.New("KeyManager is nil object")
	}
	if key, ok := j.serverToPublicKey.Load(serverName); ok {
		if key == serverPublicKey {
			return nil
		}
	}
	j.serverToPublicKey.Store(serverName, serverPublicKey)
	return nil
}
