package model

import (
	"fmt"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db     *gorm.DB
	conf   = config.Config
	logger = log.Log
)

func init() {
	connect()
}

func connect() {
	var err error
	database := conf.Sub("database")
	username := database.GetString("username")
	password := database.GetString("password")
	databasename := database.GetString("dbname")
	host := database.GetString("host")
	port := database.GetInt("port")

	dsn := fmt.Sprintf(`host=%s user=%s
		password=%s dbname=%s
		port=%d sslmode=disable 
		TimeZone=Asia/Shanghai`,
		host,
		username,
		password,
		databasename,
		port,
	)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Panicln(err)
	}
}
