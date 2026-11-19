package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init(logLevel string) {
	if logLevel == "" {
		logLevel = "info"
	}
	level, err := zerolog.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid log level")
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func Infof(format string, args ...interface{}) {
	log.Info().Msgf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	log.Warn().Msgf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	log.Error().Msgf(format, args...)
}
