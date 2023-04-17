package model

import (
	"fmt"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	conf = config.Config
)

func init() {
	connect()
}

func connect() {
	database := conf.Sub("database")
	username := database.GetString("username")
	password := database.GetString("password")
	databasename := database.GetString("dbname")
	host := database.GetString("host")
	port := database.GetString("port")

	dsn := fmt.Sprintf(`host=%s user=%s
		password=%s databasename=%s
		port=%d sslmode=disable 
		TimeZone=Asia/Shanghai`,
		host,
		username,
		password,
		databasename,
		port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// TODO: use log to show ERROR with this message
		// then just panic
		panic(fmt.Sprintf(`Database Connect Failed\n
			host: %s\n
			port
			databasename: %s\n
			username: %s\n
			password: %s\n`,
			host,
			port,
			databasename,
			username,
			password,
		))
	}
}
