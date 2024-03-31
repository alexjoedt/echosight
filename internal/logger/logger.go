package logger

import (
	"context"
	"fmt"
	"io"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger *zerolog.Logger
}

func New(c string, fields ...Field) *Logger {
	if globalLogger == nil {
		panic("logger is not initialized")
	}

	var logger zerolog.Logger
	if c == "" {
		logger = globalLogger.With().Logger()
	} else {
		logger = globalLogger.With().Str("context", c).Logger()
	}

	for _, f := range fields {
		logger = logger.With().Any(f.Key, f.Value).Logger()
	}

	return &Logger{logger: &logger}
}

// Output duplicates the global logger and sets w as its output.
func (l *Logger) Output(w io.Writer) zerolog.Logger {
	return l.logger.Output(w)
}

// With creates a child logger with the field added to its context.
func (l *Logger) With() zerolog.Context {
	return l.logger.With()
}

func (l *Logger) WithContext(ctx context.Context) context.Context {
	return l.logger.WithContext(ctx)
}

func (l *Logger) UpdateContext(update func(c zerolog.Context) zerolog.Context) {
	l.logger.UpdateContext(update)
}

// Level creates a child logger with the minimum accepted level set to level.
func (l *Logger) Level(level zerolog.Level) zerolog.Logger {
	return l.logger.Level(level)
}

// Sample returns a logger with the s sampler.
func (l *Logger) Sample(s zerolog.Sampler) zerolog.Logger {
	return l.logger.Sample(s)
}

// Hook returns a logger with the h Hook.
func (l *Logger) Hook(h zerolog.Hook) zerolog.Logger {
	return l.logger.Hook(h)
}

// Trace starts a new message with trace level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Trace() *zerolog.Event {
	return l.logger.Trace()
}

// Debug starts a new message with debug level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Debug() *zerolog.Event {
	return l.logger.Debug()
}

// Info starts a new message with info level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Inf() *zerolog.Event {
	return l.logger.Info()
}

// Warn starts a new message with warn level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Warn() *zerolog.Event {
	return l.logger.Warn()
}

// Error starts a new message with error level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Err() *zerolog.Event {
	return l.logger.Error()
}

// Fatal starts a new message with fatal level. The os.Exit(1) function
// is called by the Msg method.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Fatal() *zerolog.Event {
	return l.logger.Fatal()
}

// Panic starts a new message with panic level. The message is also sent
// to the panic function.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Panic() *zerolog.Event {
	return l.logger.Panic()
}

// WithLevel starts a new message with level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) WithLevel(level zerolog.Level) *zerolog.Event {
	return l.logger.WithLevel(level)
}

// Log starts a new message with no level. Setting zerolog.GlobalLevel to
// zerolog.Disabled will still disable events produced by this method.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Log() *zerolog.Event {
	return l.logger.Log()
}

// Print sends a log event using debug level and no extra field.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Print(v ...interface{}) {
	l.logger.Print(v...)
}

// Printf sends a log event using debug level and no extra field.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}

// Ctx returns the Logger associated with the ctx. If no logger
// is associated, a disabled logger is returned.
func (l *Logger) Ctx(ctx context.Context) *Logger {
	return &Logger{logger: zerolog.Ctx(ctx)}
}

// Custom f-Loggers
func (l *Logger) Tracef(format string, args ...any) {
	l.logger.Trace().Msgf(format, args...)
}

func (l *Logger) Debugf(format string, args ...any) {
	l.logger.Debug().Msgf(format, args...)
}

func (l *Logger) Infow(msg string, fields ...Field) {
	applyFields(l.logger.Info(), fields...).Msg(msg)
}

func (l *Logger) Debugw(msg string, fields ...Field) {
	applyFields(l.logger.Debug(), fields...).Msg(msg)
}

func (l *Logger) Errorw(msg string, fields ...Field) {
	applyFields(l.logger.Error(), fields...).Msg(msg)
}

func (l *Logger) Infof(format string, args ...any) {
	l.logger.Info().Msgf(format, args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.logger.Warn().Msgf(format, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.logger.Error().Msgf(format, args...)
}

func (l *Logger) Fatalf(format string, args ...any) {
	l.logger.Fatal().Msgf(format, args...)
}

func (l *Logger) Panicf(format string, args ...any) {
	l.logger.Panic().Msgf(format, args...)
}

func (l *Logger) Errorc(message string, err error, fields ...Field) {
	e := l.logger.Error()
	applyFields(e, fields...).
		Err(err).
		Caller(1).
		Msg(message)
}

func (l *Logger) Write(p []byte) (int, error) {
	return l.logger.Write(p)
}

func applyFields(logEvent *zerolog.Event, fields ...Field) *zerolog.Event {
	for _, f := range fields {
		logEvent.Any(f.Key, f.Value)
	}
	return logEvent
}

// Info implements the Logger interface for robfig/cron
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.Infof(msg, keysAndValues...)
}

// Error implements the Logger interface for robfig/cron
func (l *Logger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.Errorf(fmt.Sprintf("%s: %v", msg, err), keysAndValues...)
}
