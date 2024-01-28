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
	PostgresHost     string `yaml:"postgreshost"`
	PostgresPort     int    `yaml:"postgresport"`
	PostgresUser     string `yaml:"postgresuser"`
	PostgresPassword string `yaml:"postgrespassword"`
	PostgresDB       string `yaml:"postgresdb"`
}

var emptyConfig = Config{}

var defaultConfig = Config{
	Port:             8080,
	LogLevel:         string(Info),
	LogFormat:        string(Plain),
	LogToFile:        false,
	PostgresHost:     "localhost",
	PostgresPort:     5432,
	PostgresUser:     "root",
	PostgresPassword: "secret", // will be generated if no supplied by the user
	PostgresDB:       "bank",   // must be user supplied

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
		cfg = defaultConfig
		return
	}

	cfg = config

	if config.Port == 0 {
		cfg.Port = defaultConfig.Port
	}

	if config.LogLevel == "" {
		cfg.LogLevel = defaultConfig.LogLevel
	}

	if config.LogFormat == "" {
		cfg.LogFormat = defaultConfig.LogFormat
	}

	if config.PostgresHost == "" {
		cfg.PostgresHost = defaultConfig.PostgresHost
	}

	if config.PostgresPort == 0 {
		cfg.PostgresPort = defaultConfig.PostgresPort
	}

	if config.PostgresUser == "" {
		cfg.PostgresUser = defaultConfig.PostgresUser
	}

	if config.PostgresPassword == "" {
		cfg.PostgresPassword = defaultConfig.PostgresPassword
	}

	if config.PostgresDB == "" {
		cfg.PostgresDB = defaultConfig.PostgresDB
	}
	return
}
