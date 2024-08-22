// Package config supply a global variable `Config`
// to get configuration message.
// usage example: CONFIG_FILE=dev-ask go test ./model
package config

import (
	"fmt"
	"os"

	"github.com/NJUPT-SAST/sast-link-backend/version"
	"github.com/spf13/viper"
)

// Config is the configuration to start main server.
type Config struct {
	ConfigFile string
	// Mode can be "prod" or "dev" or "demo"
	Mode string
	// Addr is the binding address for server
	Addr string
	// Port is the binding port for server
	Port int
	// FIXME: Maybe can add DSN for database connection, but it need to parse.

	// For Postgres
	PostgresHost string
	PostgresPort int
	PostgresUser string
	PostgresPWD  string
	PostgresDB   string
	// For Redis
	RedisHost string
	RedisPort int
	RedisDB   int
	RedisPWD  string
	// For JWT
	Secret string
	// For log
	LogLevel string
	LogFile  string
	// Version is the current version of server
	Version string
}

func (p *Config) IsDev() bool {
	return p.Mode != "prod"
}

// NewConfig create a new Config instance
func NewConfig() *Config {
	instanceConfig := &Config{
		ConfigFile:   viper.GetString("config_file"),
		Mode:         viper.GetString("mode"),
		Addr:         viper.GetString("addr"),
		Port:         viper.GetInt("port"),
		PostgresHost: viper.GetString("postgres.host"),
		PostgresPort: viper.GetInt("postgres.port"),
		PostgresUser: viper.GetString("postgres.user"),
		PostgresPWD:  viper.GetString("postgres.password"),
		PostgresDB:   viper.GetString("postgres.db"),
		RedisHost:    viper.GetString("redis.host"),
		RedisPort:    viper.GetInt("redis.port"),
		RedisPWD:     viper.GetString("redis.password"),
		RedisDB:      viper.GetInt("redis.db"),
		Secret:       viper.GetString("secret"),
		LogLevel:     viper.GetString("log.level"),
		LogFile:      viper.GetString("log.file"),
		Version:      version.GetCurrentVersion(viper.GetString("mode")),
	}
	return instanceConfig
}

func SetupConfig() {
	// Load the specified config file if provided
	if configFile := viper.GetString("config_file"); configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		fileName := os.Getenv("CONFIG_FILE")
		if fileName == "" {
			fileName = viper.GetString("config_file")
			if fileName == "" {
				fileName = "dev-prod"
			}
		}
		// Otherwise use default paths
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.sast-link/")
		viper.AddConfigPath("/etc/sast-link/")
		viper.AddConfigPath("./config")
		viper.SetConfigName(fileName) // dev-prod.toml by default
		viper.SetConfigType("toml")
	}

	// Get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Current working directory: %s\n", currentDir)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Failed to read config file: %s\n", err)
			os.Exit(1)
		} else {
			fmt.Printf("Config file not found: %s\n", err)
			os.Exit(1)
		}
	}
}
