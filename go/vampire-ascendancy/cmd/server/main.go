package main

import (
	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/vampire-ascendancy/internal/config"
	"github.com/MaxBlaushild/poltergeist/vampire-ascendancy/internal/server"
)

func main() {
	config, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		panic(err)
	}

	authClient := auth.NewClient()
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

	server.NewServer(authClient, dbClient, config.Public.RedisUrl).ListenAndServe("8090")
}
