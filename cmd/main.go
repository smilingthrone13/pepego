package main

import (
	"apubot/internal/app"
	"apubot/internal/config"
	"log"
)

func main() {
	cfg, err := config.NewConfig("./config")
	if err != nil {
		log.Fatal(err)
	}

	app.New(cfg).Run()
}
