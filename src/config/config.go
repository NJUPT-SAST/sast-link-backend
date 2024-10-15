// Package config supply a global variable `Config`
// to get configuration message.
// usage example: CONFIG_FILE=dev-ask go test ./model
package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/NJUPT-SAST/sast-link-backend/version"
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
	LogLevel          string
	LogFile           string
	LogFileMaxSize    int
	LogFileMaxBackups int
	LogFileMaxAge     int
	LogCompress       bool
	// Version is the current version of server
	Version string

	// System settings, it will store to database
	SystemSettings map[string]string // key is system setting type, value is setting value which is json string
}

func (c *Config) IsDev() bool {
	return c.Mode != "prod"
}

// NewConfig create a new Config instance.
func NewConfig() *Config {
	instanceConfig := &Config{
		ConfigFile:        viper.GetString("config_file"),
		Mode:              viper.GetString("mode"),
		Addr:              viper.GetString("addr"),
		Port:              viper.GetInt("port"),
		PostgresHost:      viper.GetString("postgres.host"),
		PostgresPort:      viper.GetInt("postgres.port"),
		PostgresUser:      viper.GetString("postgres.user"),
		PostgresPWD:       viper.GetString("postgres.password"),
		PostgresDB:        viper.GetString("postgres.db"),
		RedisHost:         viper.GetString("redis.host"),
		RedisPort:         viper.GetInt("redis.port"),
		RedisPWD:          viper.GetString("redis.password"),
		RedisDB:           viper.GetInt("redis.db"),
		Secret:            viper.GetString("jwt.secret"),
		LogLevel:          viper.GetString("log.level"),
		LogFile:           viper.GetString("log.file"),
		LogFileMaxSize:    viper.GetInt("log.max_size"),
		LogFileMaxBackups: viper.GetInt("log.max_backups"),
		LogFileMaxAge:     viper.GetInt("log.max_age"),
		LogCompress:       viper.GetBool("log.compress"),
		Version:           version.GetCurrentVersion(viper.GetString("mode")),
	}
	return instanceConfig
}

func SetupConfig() error {
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
		viper.AddConfigPath("../config")
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
			return err
		}

		fmt.Printf("Config file not found: %s\n", err)
		return err
	}
	return nil
}

// SystemSettingType represents the system setting type.
type SystemSettingType string

const (
	// WebsiteSettingType represents the website setting type.
	WebsiteSettingType SystemSettingType = "website"
	// EmailSettingType represents the email setting type.
	EmailSettingType SystemSettingType = "email"
	// StorageSettingType represents the storage setting type.
	StorageSettingType SystemSettingType = "storage"
	// IdpSettingType represents the identity provider setting type.
	IdpSettingType SystemSettingType = "idp"
)

// String converts the SystemSettingType to string.
func (t SystemSettingType) String() string {
	return string(t)
}

func TypeFromString(t string) SystemSettingType {
	switch t {
	case "website":
		return WebsiteSettingType
	case "email":
		return EmailSettingType
	case "storage":
		return StorageSettingType
	case "idp":
		return IdpSettingType
	}
	return ""
}

// LoadSystemSettings loads system settings from config file.
func (c *Config) LoadSystemSettings() {
	// Load system settings from config file
	websiteSettings := viper.GetStringMapString(WebsiteSettingType.String())
	emailSettings := viper.GetStringMapString(EmailSettingType.String())
	storageSettings := viper.GetStringMapString(StorageSettingType.String())
	idpSettings := viper.GetStringMap(IdpSettingType.String())

	c.SystemSettings = make(map[string]string)
	// Transform map to JSON string
	if jsonString, err := util.MapToJSONString(websiteSettings); err == nil {
		c.SystemSettings[WebsiteSettingType.String()] = jsonString
	} else {
		fmt.Printf("Error converting website settings to JSON: %v\n", err)
	}

	if jsonString, err := util.MapToJSONString(emailSettings); err == nil {
		c.SystemSettings[EmailSettingType.String()] = jsonString
	} else {
		fmt.Printf("Error converting email settings to JSON: %v\n", err)
	}

	if jsonString, err := util.MapToJSONString(storageSettings); err == nil {
		c.SystemSettings[StorageSettingType.String()] = jsonString
	} else {
		fmt.Printf("Error converting storage settings to JSON: %v\n", err)
	}

	for _, v := range idpSettings {
		idpSetting, ok := v.(map[string]interface{})
		if !ok {
			fmt.Printf("IDP setting is not map[string]interface{}\n")
			continue
		}
		if jsonString, err := util.MapToJSONStringInterface(idpSetting); err == nil {
			// All idp settings are stored format like "idp-xxx"
			if idpSetting["name"] == nil {
				fmt.Printf("IDP setting name is nil\n")
				continue
			}
			key := fmt.Sprintf("%s-%s", IdpSettingType.String(), idpSetting["name"])
			c.SystemSettings[key] = jsonString
		} else {
			fmt.Printf("Error converting IDP settings to JSON: %v\n", err)
		}
	}
}
