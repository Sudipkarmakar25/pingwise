package api

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type ApiConfig struct {
	Environment          string
	ServerAddress        string
	PrimaryDbDSN         string
	MaxOpenConnsToDB     int
	LogLevel             string
	EncryptionKey        string
	EncryptionKeyVersion int
}

var cfg *ApiConfig

func LoadApiConfig() (*ApiConfig, error) {
	// Load environment variables from .env file
	if os.Getenv("ENVIRONMENT") == "" {
		err := godotenv.Load(".env")
		if err != nil {
			return nil, err
		}
	}

	maxOpenConns, err := strconv.Atoi(os.Getenv("MAX_OPEN_CONNS_TO_DB"))
	if err != nil {
		maxOpenConns = 10
	}
	encryptionKeyVersion, err := strconv.Atoi(os.Getenv("ENCRYPTION_KEY_VERSION"))
	if err != nil {
		encryptionKeyVersion = 1
	}

	cfg = &ApiConfig{
		Environment:          os.Getenv("ENVIRONMENT"),
		ServerAddress:        os.Getenv("SERVER_ADDRESS"),
		PrimaryDbDSN:         os.Getenv("PRIMARY_DB_DSN"),
		MaxOpenConnsToDB:     maxOpenConns,
		LogLevel:             os.Getenv("LOG_LEVEL"),
		EncryptionKey:        os.Getenv("ENCRYPTION_KEY"),
		EncryptionKeyVersion: encryptionKeyVersion,
	}
	return cfg, nil
}

// GetApiConfig returns the loaded configuration
func GetApiConfig() *ApiConfig {
	if cfg == nil {
		if _, err := LoadApiConfig(); err != nil {
			log.Fatalf("failed to load config: %v", err.Error())
		}
	}
	return cfg
}
