package config

import (
	"fmt"
)

type Config struct {
	AppPassword         string
	EncryptionSecretKey string // секретный ключ шифрования
	ApiPort             string
}

// NewConfig конструктор объекта конфигурации приложения
func NewConfig(appPass string, encKey string, apiPort string) (*Config, error) {
	if appPass == "" || encKey == "" {
		return nil, fmt.Errorf("Invalid config.")
	}

	if apiPort == "" {
		apiPort = "7540"
	}
	return &Config{AppPassword: appPass, EncryptionSecretKey: encKey, ApiPort: apiPort}, nil
}
