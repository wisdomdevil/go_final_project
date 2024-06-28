package config

import (
	"fmt"
)

const (
	DefaultPassword = "123456"
)

type Config struct {
	AppPassword         string
	EncryptionSecretKey string // секретный ключ шифрования
	ApiPort             string
}

// NewConfig конструктор объекта конфигурации приложения
func NewConfig(appPass string, encKey string, apiPort string) (*Config, error) {
	if appPass == "" {
		appPass = DefaultPassword
	}
	if appPass == "" || encKey == "" {
		return nil, fmt.Errorf("invalid config")
	}

	if apiPort == "" {
		apiPort = "7540"
	}
	return &Config{AppPassword: appPass, EncryptionSecretKey: encKey, ApiPort: apiPort}, nil
}
