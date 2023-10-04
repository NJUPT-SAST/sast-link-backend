package main

import (
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/router"
)

func main() {
	router := router.InitRouter()
	// _ = router.Run()
	log.Logger.Errorln(router.Run())
}
