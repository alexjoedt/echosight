// zog is a wrapper for zerolog
package logger

import (
	"github.com/rs/zerolog"
)

const (
	logFilename string = "echosight.log"
)

// Global Logger
var globalLogger *zerolog.Logger

func Init(loglevel string, stdout bool) (*zerolog.Logger, error) {
	var trace bool

	level := logeLevel(loglevel)
	if level == zerolog.TraceLevel {
		trace = true
	}

	wr := logWriter(stdout, FormatConsole)
	zlCtx := zerolog.New(wr).With()

	if trace {
		zlCtx = zlCtx.Caller()
		zlCtx = zlCtx.CallerWithSkipFrameCount(4)
	}
	zl := zlCtx.Timestamp().Logger().Level(level)
	globalLogger = &zl
	return globalLogger, nil
}

func GloabalLogger() *zerolog.Logger {
	return globalLogger
}

func Tracef(format string, args ...any) {
	writeLog(globalLogger, zerolog.TraceLevel, nil, format, args...)
}

func Trace(message string, err error) {
	writeLog(globalLogger, zerolog.TraceLevel, err, message)
}

func Debugf(format string, args ...any) {
	writeLog(globalLogger, zerolog.DebugLevel, nil, format, args...)
}

func Infof(format string, args ...any) {
	writeLog(globalLogger, zerolog.InfoLevel, nil, format, args...)
}

func Warnf(format string, args ...any) {
	writeLog(globalLogger, zerolog.WarnLevel, nil, format, args...)
}

func Errorf(format string, args ...any) {
	writeLog(globalLogger, zerolog.ErrorLevel, nil, format, args...)
}

func Fatalf(format string, args ...any) {
	writeLog(globalLogger, zerolog.FatalLevel, nil, format, args...)
}

func Panicf(format string, args ...any) {
	writeLog(globalLogger, zerolog.PanicLevel, nil, format, args...)
}

func Error(message string, err error) {
	e := globalLogger.Error()

	e.Err(err).
		Msg(message)
}

func writeLog(l *zerolog.Logger, level zerolog.Level, err error, format string, args ...any) {
	var event *zerolog.Event
	switch level {
	case zerolog.TraceLevel:
		event = l.Trace().Caller(2).Stack().Err(err)
	case zerolog.DebugLevel:
		event = l.Debug()
	case zerolog.InfoLevel:
		event = l.Info()
	case zerolog.WarnLevel:
		event = l.Warn()
	case zerolog.ErrorLevel:
		event = l.Error().Caller(2).Stack().Err(err)
	case zerolog.FatalLevel:
		event = l.Fatal().Caller(2).Err(err)
	case zerolog.PanicLevel:
		event = l.Panic().Caller(2).Err(err)
	}

	event.Msgf(format, args...)
}
