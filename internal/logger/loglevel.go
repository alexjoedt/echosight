package logger

import (
	"log"
	"strings"

	"github.com/rs/zerolog"
)

func logeLevel(l string) zerolog.Level {
	l = strings.ToLower(l)
	switch l {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		log.Printf("invalid loglevel '%s', set to default: WARN", l)
		return zerolog.WarnLevel
	}
}
