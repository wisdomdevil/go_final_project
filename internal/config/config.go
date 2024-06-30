package config

import (
	"fmt"
)

const (
	defaultPassword = "123456"
	defaultPort     = "7540"
)

type Config struct {
	AppPassword         string
	EncryptionSecretKey string // секретный ключ шифрования
	ApiPort             string
}

// NewConfig конструктор объекта конфигурации приложения
func NewConfig(appPass string, encKey string, apiPort string) (*Config, error) {
	if appPass == "" {
		appPass = defaultPassword
	}
	if appPass == "" || encKey == "" {
		return nil, fmt.Errorf("invalid config")
	}

	if apiPort == "" {
		apiPort = defaultPort
	}
	return &Config{AppPassword: appPass, EncryptionSecretKey: encKey, ApiPort: apiPort}, nil
}
