package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	v1 "pingwise/cmd/api/handlers/v1"
	customMiddlewares "pingwise/cmd/api/middlewares"
	"pingwise/common/database"
	"pingwise/common/encryption"
	"pingwise/common/logger"
	"pingwise/config/api"
	"pingwise/internal/service"
)

func main() {
	// Load configuration
	cfg, err := api.LoadApiConfig()
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

	apiServer := InitAPIServer()

	// Start the api server in a go routine
	go func() {
		logger.Infof("Starting server on %s", cfg.ServerAddress)
		if err := apiServer.Start(cfg.ServerAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("Could not listen on %s: %v", cfg.ServerAddress, err)
			os.Exit(1)
		}
	}()

	// Set up signal handling for graceful shutdown
	sigintOrTerm := make(chan os.Signal, 1)
	signal.Notify(sigintOrTerm, syscall.SIGINT, syscall.SIGTERM)
	<-sigintOrTerm

	logger.Infof("Received SIGINT/SIGTERM; giving 20 seconds for api server to close gracefully")
	softStopCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(softStopCtx); err != nil {
		logger.Errorf("Server Shutdown Failed: %v", err)
	} else {
		logger.Infof("Server exited properly through soft stop")
	}
}

func initializeDependencies(cfg *api.ApiConfig) error {
	err := database.InitDb(cfg.PrimaryDbDSN, cfg.Environment, cfg.MaxOpenConnsToDB)
	if err != nil {
		return errors.Wrap(err, "failed to initialize database")
	}
	err = encryption.Init(cfg.EncryptionKey, cfg.EncryptionKeyVersion)
	if err != nil {
		return errors.Wrap(err, "failed to initialize encryption")
	}
	return nil
}

func InitAPIServer() *echo.Echo {
	// Initialize Echo
	e := echo.New()
	e.HideBanner = true
	e.Pre(middleware.AddTrailingSlash())

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(customMiddlewares.CustomCtxMiddleware)

	e.Validator = &service.CustomValidator{Validator: validator.New()}

	// Set up routes
	registerHandlers(e)

	return e
}

func registerHandlers(e *echo.Echo) {
	e.GET("api/health/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	v1Routes := e.Group("api/v1")
	v1.RegisterHandlers(v1Routes)
}
