package app

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/alexjoedt/echosight/osutils"
	flag "github.com/spf13/pflag"
)

func LoadConfigFromFile(filename string) (*Config, error) {
	return loadConfig(filename)
}

func LoadConfig() (*Config, error) {
	configFile := flag.String("config", "", "path to the config file")
	flag.Parse()

	return loadConfig(*configFile)
}

func loadConfig(filename string) (*Config, error) {

	logger.Infof("Load config file '%s'", filename)

	if !osutils.IsFileExists(filename) {
		log.Fatal("no config file found")
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config Config
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		return nil, err
	}

	// load config from environment
	// TODO: extend config with environment variables, flags and defaults
	config.LogLevel = getStringConfig(os.Getenv("ES_LOGLEVEL"), config.LogLevel, "info")
	config.Environment = getStringConfig(os.Getenv("ES_ENVIRONMENT"), config.Environment, "prod")
	config.Secret = mustStringConfig(os.Getenv("ES_SECRET"), config.Secret)
	config.HTTP.Port = getStringConfig(config.HTTP.Port, "8080")

	return &config, nil
}

func mustStringConfig(key string, defaults ...string) string {
	for _, d := range defaults {
		if d != "" {
			return d
		}
	}

	log.Fatalf("[FATAL] invalid configuration: '%s' must be provided", key)
	return ""
}

func getStringConfig(v string, defaults ...string) string {
	if v != "" {
		return v
	}

	for _, d := range defaults {
		if d != "" {
			return d
		}
	}

	return ""
}

func (c *Config) isDev() bool {
	return strings.Contains(c.Environment, "dev")
}
