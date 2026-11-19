package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"pingwise/async/river"
	"pingwise/common/database"
	"pingwise/common/email"
	"pingwise/common/encryption"
	"pingwise/common/logger"
	"pingwise/config/worker"
)

func main() {
	// Load configuration
	cfg, err := worker.LoadWorkerConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// Initialize logger
	logger.Init(cfg.LogLevel)

	// Initialise dependencies
	err = initializeDependencies(cfg)
	if err != nil {
		logger.Errorf("failed to initialise dependencies : %v", err)
		os.Exit(1)
	}

	mainCtx, cancelMainCtx := context.WithCancel(context.Background())
	defer cancelMainCtx()

	riverClient, err := river.InitRiverClient(mainCtx, cfg.PrimaryDbDSN, cfg.CheckWorkers)
	if err != nil {
		logger.Errorf("failed to initialise river client : %v", err)
		os.Exit(1)
	}
	if err := riverClient.Start(mainCtx); err != nil {
		logger.Errorf("failed to start river client : %v", err)
		os.Exit(1)
	}
	logger.Infof("Worker started with %d check workers", cfg.CheckWorkers)

	// Set up signal handling for graceful shutdown
	sigintOrTerm := make(chan os.Signal, 1)
	signal.Notify(sigintOrTerm, syscall.SIGINT, syscall.SIGTERM)
	<-sigintOrTerm

	logger.Infof("Received SIGINT/SIGTERM; giving 30 seconds for running checks to finish")
	softStopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := riverClient.Stop(softStopCtx); err != nil {
		logger.Errorf("River client did not stop gracefully: %v", err)
	} else {
		logger.Infof("Worker exited properly through soft stop")
	}
}

func initializeDependencies(cfg *worker.WorkerConfig) error {
	err := database.InitDb(cfg.PrimaryDbDSN, cfg.Environment, cfg.MaxOpenConnsToDB)
	if err != nil {
		return errors.Wrap(err, "failed to initialize database")
	}
	err = encryption.Init(cfg.EncryptionKey, cfg.EncryptionKeyVersion)
	if err != nil {
		return errors.Wrap(err, "failed to initialize encryption")
	}
	err = email.Init(email.Config{
		Host:     cfg.SmtpHost,
		Port:     cfg.SmtpPort,
		Username: cfg.SmtpUsername,
		Password: cfg.SmtpPassword,
		From:     cfg.AlertFromEmail,
		AppUrl:   cfg.AppUrl,
	})
	if err != nil {
		return errors.Wrap(err, "failed to initialize email")
	}
	if !email.Enabled() {
		logger.Warnf("SMTP_HOST is not set, alert emails are turned off")
	}
	return nil
}
