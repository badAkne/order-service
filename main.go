package main

import (
	"context"
	"fmt"
	"log"

	"github.com/badAkne/order-service/internal/app/config"
	rhealth "github.com/badAkne/order-service/internal/app/handler/health"
	rprocessor "github.com/badAkne/order-service/internal/app/processor/http"
	"github.com/badAkne/order-service/internal/app/repository/postgres"
)

func main() {
	config.Load()
	fmt.Println(config.Root)
	fmt.Println(config.Root.Repository.Postgres.DSN())

	_, err := postgres.NewConn(context.Background(), config.Root.Repository.Postgres)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Printf("Connected to db: %s", config.Root.Repository.Postgres.Name)

	health := rhealth.NewHandler()

	proc := rprocessor.NewHTTP(health, config.Root.Processor.WebServer)

	if err = proc.Run(); err != nil {
		log.Fatal(err.Error())
	}

}
