// Package config supply a global variable `Config`
// to get configuration message.
// usage example: CONFIG_FILE=dev-ask go test ./model
package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config is global variable
// where you can get configuration
// by something like `Config.GetInt`
var Config *viper.Viper = viper.New()

func init() {
	fileName := os.Getenv("CONFIG_FILE")
	if fileName == "" {
		fileName = "dev-xun"
	}
	Config.AddConfigPath(".")
	Config.AddConfigPath("../../config")
	Config.AddConfigPath("../../../config")
	Config.AddConfigPath("../config")
	Config.AddConfigPath("./config")
	Config.AddConfigPath("$HOME/Workspace/go/sast/sast-link-backend/config/")
	Config.SetConfigName(fileName)
	Config.SetConfigType("toml")

	if err := Config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Sprintf("File [config/%s.toml] Not Found\n", fileName))
		} else {
			panic(err.Error())
		}
	}
	fmt.Printf("Config file [config/%s.toml] loaded\n", fileName)
}
