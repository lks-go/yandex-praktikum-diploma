package main

import (
	"log"

	"github.com/lks-go/yandex-praktikum-diploma/internal/app"
)

func main() {
	a := app.New()

	log.Println("getting application config")
	cfg, err := a.BuildConfig()
	if err != nil {
		log.Fatalf("application starting error: %s", err)
	}

	log.Println("starting application")
	if err := a.Run(cfg); err != nil {
		log.Fatalf("application error: %s", err)
	}

	log.Print("application successfully stopped")
}
