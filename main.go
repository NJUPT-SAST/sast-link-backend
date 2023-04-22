package main

import (
	"github.com/NJUPT-SAST/sast-link-backend/router"
	"log"
)

func main() {
	router := router.InitRouter()
	// _ = router.Run()
	log.Fatalln(router.Run())
}
