package main

import (
	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/dropbox"
	"github.com/MaxBlaushild/poltergeist/pkg/googledrive"
	"github.com/MaxBlaushild/poltergeist/travel-angels/internal/config"
	"github.com/MaxBlaushild/poltergeist/travel-angels/internal/server"
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

	googleDriveClient := googledrive.NewClient(googledrive.ClientConfig{
		ClientID:     config.Secret.GoogleDriveClientID,
		ClientSecret: config.Secret.GoogleDriveClientSecret,
		RedirectURI:  config.Public.GoogleDriveRedirectURI,
	}, dbClient)

	dropboxClient := dropbox.NewClient(dropbox.ClientConfig{
		ClientID:     config.Secret.DropboxClientID,
		ClientSecret: config.Secret.DropboxClientSecret,
		RedirectURI:  config.Public.DropboxRedirectURI,
	}, dbClient)

	server.NewServer(authClient, dbClient, googleDriveClient, dropboxClient).ListenAndServe("8083")
}
