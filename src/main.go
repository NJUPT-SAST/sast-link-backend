package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/api/v1/server"
	"github.com/NJUPT-SAST/sast-link-backend/store"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	greetingBanner = `
███████  █████  ███████ ████████     ██      ██ ███    ██ ██   ██ 
██      ██   ██ ██         ██        ██      ██ ████   ██ ██  ██  
███████ ███████ ███████    ██        ██      ██ ██ ██  ██ █████   
     ██ ██   ██      ██    ██        ██      ██ ██  ██ ██ ██  ██  
███████ ██   ██ ███████    ██        ███████ ██ ██   ████ ██   ██ 
`
)

var (
	rootCmd = cobra.Command{
		Use:   "sast-link-backend",
		Short: "SAST Link Backend",
		Run:   run,
	}
)

func run(_ *cobra.Command, _ []string) {
	instanceConfig := config.NewConfig()
	ctx, cancel := context.WithCancel(context.Background())
	storeInstance, err := store.NewStore(ctx, instanceConfig)
	if err != nil {
		cancel()
		fmt.Printf("Failed to create store: %s", err)
		return
	}

	s, err := server.NewServer(ctx, instanceConfig, storeInstance)
	if err != nil {
		cancel()
		fmt.Printf("Failed to create server: %s", err)
		return
	}

	if err := s.Start(); err != nil {
		if err != http.ErrServerClosed {
			fmt.Printf("Failed to start server: %s", err)
			cancel()
		}
	}
	// After server started, we can load system settings from config file.
	instanceConfig.LoadSystemSettings()
	// Store system settings to database
	if err := storeInstance.InitSystemSetting(ctx, instanceConfig); err != nil {
		fmt.Printf("Failed to init system settings: %s", err)
		cancel()
	}

	printGreeting(instanceConfig)

	c := make(chan os.Signal, 1)
	// Shutdown server when receive signal
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		s.Shutdown(ctx)
		cancel()
	}()

	// Waiting for shutdown signal
	<-ctx.Done()
}

func init() {
	setupDefaults()
	setupCommandLine()
	config.SetupConfig()
}

func setupDefaults() {
	// Not set default value for config file
	// viper.SetDefault("config_file", "config.yaml")
	// Default values
	viper.SetDefault("mode", "demo")
	viper.SetDefault("addr", "127.0.0.1")
	viper.SetDefault("port", 8080)
	viper.SetDefault("secret", util.GenerateUUID)
	viper.SetDefault("postgres.host", "localhost")
	viper.SetDefault("postgres.port", 5432)
	viper.SetDefault("postgres.user", "user")
	viper.SetDefault("postgres.password", "password")
	viper.SetDefault("postgres.db", "dbname")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.file", "sast-link-backend.log")
}

func setupCommandLine() {
	rootCmd.PersistentFlags().String("config_file", "", "path to config file")
	rootCmd.PersistentFlags().String("mode", "", "mode of server, can be 'prod', 'dev', or 'demo'")
	rootCmd.PersistentFlags().String("addr", "", "binding address for server")
	rootCmd.PersistentFlags().Int("port", 0, "binding port for server")
	rootCmd.PersistentFlags().String("postgres.host", "", "Postgres server host")
	rootCmd.PersistentFlags().Int("postgres.port", 0, "Postgres server port")
	rootCmd.PersistentFlags().String("postgres.user", "", "Postgres username")
	rootCmd.PersistentFlags().String("postgres.password", "", "Postgres password")
	rootCmd.PersistentFlags().String("postgres.db", "", "Postgres database name")
	rootCmd.PersistentFlags().String("redis.host", "", "Redis server host")
	rootCmd.PersistentFlags().Int("redis.port", 0, "Redis server port")
	rootCmd.PersistentFlags().Int("redis.db", 0, "Redis database index")
	rootCmd.PersistentFlags().String("redis.password", "", "Redis password")
	rootCmd.PersistentFlags().String("secret", "", "JWT secret key")
	rootCmd.PersistentFlags().String("log.level", "", "Log level")
	rootCmd.PersistentFlags().String("log.file", "", "Log file path")

	// Ensure command-line parameters are bound to Viper
	bindFlagsToViper()
}

func bindFlagsToViper() {
	viper.BindPFlag("config_file", rootCmd.PersistentFlags().Lookup("config_file"))
	viper.BindPFlag("mode", rootCmd.PersistentFlags().Lookup("mode"))
	viper.BindPFlag("addr", rootCmd.PersistentFlags().Lookup("addr"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("postgres.host", rootCmd.PersistentFlags().Lookup("postgres.host"))
	viper.BindPFlag("postgres.port", rootCmd.PersistentFlags().Lookup("postgres.port"))
	viper.BindPFlag("postgres.user", rootCmd.PersistentFlags().Lookup("postgres.user"))
	viper.BindPFlag("postgres.password", rootCmd.PersistentFlags().Lookup("postgres.password"))
	viper.BindPFlag("postgres.db", rootCmd.PersistentFlags().Lookup("postgres.db"))
	viper.BindPFlag("redis.host", rootCmd.PersistentFlags().Lookup("redis.host"))
	viper.BindPFlag("redis.port", rootCmd.PersistentFlags().Lookup("redis.port"))
	viper.BindPFlag("redis.db", rootCmd.PersistentFlags().Lookup("redis.db"))
	viper.BindPFlag("redis.password", rootCmd.PersistentFlags().Lookup("redis.password"))
	viper.BindPFlag("secret", rootCmd.PersistentFlags().Lookup("secret"))
	viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log.level"))
	viper.BindPFlag("log.file", rootCmd.PersistentFlags().Lookup("log.file"))
}

func printGreeting(config *config.Config) {
	fmt.Printf(`---
Server profile
version: %s
addr:    %s
port:    %d
mode:    %s
dsn:     %s:%d
redis:   %s:%d
---
`, config.Version, config.Addr, config.Port, config.Mode, config.PostgresHost, config.PostgresPort, config.RedisHost, config.RedisPort)

	print(greetingBanner)
	if len(config.Addr) == 0 {
		fmt.Printf("Version %s has been started on port %d\n", config.Version, config.Port)
	} else {
		fmt.Printf("Version %s has been started on address '%s' and port %d\n", config.Version, config.Addr, config.Port)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
