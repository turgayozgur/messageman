package logging

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitZerolog initialize the zerolog logger.
func InitZerolog(logLevel string, humanize bool) {
	if humanize {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	level := zerolog.DebugLevel
	switch logLevel {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	case "disabled":
		level = zerolog.Disabled
	}

	zerolog.SetGlobalLevel(level)
}
