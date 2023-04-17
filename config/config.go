// Package config supply a global variable `Config`
// to get configuration message.
package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var Config *viper.Viper = viper.New()

func init() {
	fileName := os.Getenv("CONFIG_FILE")
	Config.AddConfigPath(".")
	Config.SetConfigName(fileName)
	Config.SetConfigType("toml")

	if err := Config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Sprintf("File [config/%s].toml Not Found\n", fileName))
		} else {
			panic(err.Error())
		}
	}
}
