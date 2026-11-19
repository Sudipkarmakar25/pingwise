package worker

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type WorkerConfig struct {
	Environment          string
	PrimaryDbDSN         string
	MaxOpenConnsToDB     int
	LogLevel             string
	EncryptionKey        string
	EncryptionKeyVersion int
	CheckWorkers         int // number of health checks that run at the same time

	// alert emails; when SmtpHost is empty no emails are sent
	SmtpHost       string
	SmtpPort       int
	SmtpUsername   string
	SmtpPassword   string
	AlertFromEmail string // e.g. Pingwise <alerts@example.com>
	AppUrl         string // the web app's address, for links in emails
}

var cfg *WorkerConfig

func LoadWorkerConfig() (*WorkerConfig, error) {
	// Load environment variables from .env file
	if os.Getenv("ENVIRONMENT") == "" {
		err := godotenv.Load(".env")
		if err != nil {
			return nil, err
		}
	}

	cfg = &WorkerConfig{
		Environment:          os.Getenv("ENVIRONMENT"),
		PrimaryDbDSN:         os.Getenv("PRIMARY_DB_DSN"),
		MaxOpenConnsToDB:     intFromEnv("MAX_OPEN_CONNS_TO_DB", 10),
		LogLevel:             os.Getenv("LOG_LEVEL"),
		EncryptionKey:        os.Getenv("ENCRYPTION_KEY"),
		EncryptionKeyVersion: intFromEnv("ENCRYPTION_KEY_VERSION", 1),
		CheckWorkers:         intFromEnv("CHECK_WORKERS", 20),
		SmtpHost:             os.Getenv("SMTP_HOST"),
		SmtpPort:             intFromEnv("SMTP_PORT", 587),
		SmtpUsername:         os.Getenv("SMTP_USERNAME"),
		SmtpPassword:         os.Getenv("SMTP_PASSWORD"),
		AlertFromEmail:       os.Getenv("ALERT_FROM_EMAIL"),
		AppUrl:               os.Getenv("APP_URL"),
	}
	return cfg, nil
}

// GetWorkerConfig returns the loaded configuration
func GetWorkerConfig() *WorkerConfig {
	if cfg == nil {
		if _, err := LoadWorkerConfig(); err != nil {
			log.Fatalf("failed to load config: %v", err.Error())
		}
	}
	return cfg
}

func intFromEnv(key string, defaultValue int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return defaultValue
	}
	return value
}
