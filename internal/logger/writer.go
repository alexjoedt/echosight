package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogFormat string

const (
	FormatJSON    LogFormat = "json"
	FormatConsole LogFormat = "console"
)

// logWriter returns the log writer.
// If debug is true the logs will write to stdout and file
func logWriter(debug bool, format LogFormat) io.Writer {
	format = LogFormat(strings.ToLower(string(format)))
	if format != FormatConsole && format != FormatJSON {
		format = FormatJSON
	}

	var out io.Writer
	out = &lumberjack.Logger{
		Filename:   GetLogFile(),
		MaxSize:    30, // megabytes
		MaxBackups: 3,
		MaxAge:     1, // days
		Compress:   false,
	}

	color := false
	if debug {
		color = true
		out = zerolog.MultiLevelWriter(os.Stdout, out)
	}

	if format == FormatConsole {
		output := zerolog.ConsoleWriter{
			Out:        out,
			TimeFormat: time.RFC3339,
			NoColor:    !color,
		}

		output.FormatLevel = func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("%-6s", i))
		}
		out = output
	}

	return out
}

func GetLogFile() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	logFilepath := filepath.Join(wd, "logs", logFilename)
	os.MkdirAll(filepath.Dir(logFilepath), 0o755)
	f, err := os.OpenFile(logFilepath, os.O_CREATE, 0o644)
	if err != nil {
		log.Println("failed to create log file:", err)
	}
	defer f.Close()
	return logFilepath
}
