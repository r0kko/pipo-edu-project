package logging

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init(env string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	level := zerolog.InfoLevel
	if strings.ToLower(env) == "dev" {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)

	if strings.ToLower(env) == "dev" {
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		logger := zerolog.New(output).With().Timestamp().Logger()
		log.Logger = logger
		return logger
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	log.Logger = logger
	return logger
}
