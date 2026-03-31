package config

import (
	"log"

	"github.com/badAkne/order-service/internal/app/config/section"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	App        section.App
	Repository section.Repository
	//Broker     section.Broker
	Processor section.Processor
	Monitor   section.Monitor
}

var Root Config

func Load() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("unable to load condig: %s", err.Error())
	}

	err = envconfig.Process("APP", &Root)
	if err != nil {
		log.Fatalf("Unable to process config: %s", err.Error())
	}

	log.Println("Config processed")
}
