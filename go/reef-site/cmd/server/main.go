package main

import (
	"log"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/reef-site/internal/config"
	"github.com/MaxBlaushild/poltergeist/reef-site/internal/server"
)

func main() {
	cfg, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		panic(err)
	}

	dbClient, err := db.NewClient(db.ClientConfig{
		Name:     cfg.Public.DbName,
		Host:     cfg.Public.DbHost,
		Port:     cfg.Public.DbPort,
		User:     cfg.Public.DbUser,
		Password: cfg.Secret.DbPassword,
	})
	if err != nil {
		panic(err)
	}

	log.Println("reef-site listening on :8091")
	server.NewServer(server.Deps{
		DbClient: dbClient,
		Config:   cfg,
	}).ListenAndServe("8091")
}
