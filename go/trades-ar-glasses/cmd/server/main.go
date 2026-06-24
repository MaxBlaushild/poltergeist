package main

import (
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/trades-ar-glasses/internal/config"
	"github.com/MaxBlaushild/poltergeist/trades-ar-glasses/internal/server"
)

func main() {
	config, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		panic(err)
	}

	dbClient, err := db.NewClient(db.ClientConfig{
		Name:     config.Public.DbName,
		Host:     config.Public.DbHost,
		Port:     config.Public.DbPort,
		User:     config.Public.DbUser,
		Password: config.Secret.DbPassword,
	})
	if err != nil {
		panic(err)
	}

	server.NewServer(dbClient).ListenAndServe("8091")
}
