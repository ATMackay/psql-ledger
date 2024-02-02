package service

import (
	"bytes"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	Port             int    `yaml:"port"`
	LogLevel         string `yaml:"loglevel"`
	LogFormat        string `yaml:"logformat"`
	LogToFile        bool   `yaml:"logtofile"`
	PostgresHost     string `yaml:"postgres_host"`
	PostgresPort     int    `yaml:"postgres_port"`
	PostgresUser     string `yaml:"postgres_user"`
	PostgresPassword string `yaml:"postgres_password"`
	PostgresDB       string `yaml:"postgres_db"`
	MigrationsPath   string `yaml:"migrations_path"`
	MaxThreads       int    `yaml:"max_threads"`
}

var emptyConfig = Config{}

var DefaultConfig = Config{
	Port:             8080,
	LogLevel:         string(Info),
	LogFormat:        string(Plain),
	LogToFile:        false,
	PostgresHost:     "localhost",          // Default Postgres database configuration
	PostgresPort:     5432,                 //
	PostgresUser:     "root",               //
	PostgresPassword: "secret",             //
	PostgresDB:       "bank",               //
	MigrationsPath:   "../sqlc/migrations", // local project migrations directory
	MaxThreads:       1,                    // Not multi-threaded by default
}

func isEmpty(c Config) bool {
	b, _ := yaml.Marshal(c)
	e, _ := yaml.Marshal(emptyConfig)
	return bytes.Equal(b, e)
}

// sanitizeConfig Partially empty configs will be sanitized with default values.
func sanitizeConfig(config Config) (cfg Config, defaultUsed bool) {
	if isEmpty(config) {
		defaultUsed = true
		cfg = DefaultConfig
		return
	}

	cfg = config

	if config.Port == 0 {
		cfg.Port = DefaultConfig.Port
	}

	if config.LogLevel == "" {
		cfg.LogLevel = DefaultConfig.LogLevel
	}

	if config.LogFormat == "" {
		cfg.LogFormat = DefaultConfig.LogFormat
	}

	if config.PostgresHost == "" {
		cfg.PostgresHost = DefaultConfig.PostgresHost
	}

	if config.PostgresPort == 0 {
		cfg.PostgresPort = DefaultConfig.PostgresPort
	}

	if config.PostgresUser == "" {
		cfg.PostgresUser = DefaultConfig.PostgresUser
	}

	if config.PostgresPassword == "" {
		cfg.PostgresPassword = DefaultConfig.PostgresPassword
	}

	if config.PostgresDB == "" {
		cfg.PostgresDB = DefaultConfig.PostgresDB
	}

	if config.MigrationsPath == "" {
		cfg.MigrationsPath = DefaultConfig.MigrationsPath
	}

	if config.MaxThreads == 0 {
		cfg.MaxThreads = DefaultConfig.MaxThreads
	}
	return
}
