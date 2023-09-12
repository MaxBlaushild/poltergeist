package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/slack"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/MaxBlaushild/poltergeist/scorekeeper/internal/config"
	"github.com/MaxBlaushild/poltergeist/scorekeeper/internal/scorekeeper"
	"github.com/gin-gonic/gin"
)

type SlackEventCallback struct {
	Token     string     `json:"token"`
	TeamID    string     `json:"team_id"`
	APIAppID  string     `json:"api_app_id"`
	Event     SlackEvent `json:"event"`
	Type      string     `json:"type"`
	EventID   string     `json:"event_id"`
	EventTime int64      `json:"event_time"`
	Challenge string     `json:"challenge"`
}

type SlackEvent struct {
	Type    string `json:"type"`
	Channel string `json:"channel"`
	User    string `json:"user"`
	Text    string `json:"text"`
	Ts      string `json:"ts"`
}

func main() {
	ctx := context.Background()

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

	if err := dbClient.Migrate(
		ctx,
		&models.Score{},
	); err != nil {
		panic(err)
	}

	router := gin.Default()

	slackClient := slack.NewSlackClient(
		cfg.Secret.SlackScorekeeperWebhookUrl,
	)

	skpr := scorekeeper.NewScorekeeper(dbClient)

	router.POST("/scorekeeper/handle-event", func(c *gin.Context) {
		var eventCallback SlackEventCallback

		// Decode the incoming request into the struct
		if err := c.BindJSON(&eventCallback); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode json"})
			return
		}

		// If this is a URL verification challenge from Slack
		if eventCallback.Type == "url_verification" {
			c.JSON(http.StatusOK, gin.H{"challenge": eventCallback.Challenge})
			return
		}

		go func() {
			// Only handle message types
			if eventCallback.Event.Type == "app_mention" {
				text := eventCallback.Event.Text
				fmt.Println(text)
				// 			// checks if asking to give a point
				matches, username := util.MatchesGivePointPattern(text)

				if matches {
					if strings.Contains(strings.ToLower(eventCallback.Event.User), strings.ToLower(username)) || strings.Contains(strings.ToLower(username), strings.ToLower(eventCallback.Event.User)) {
						fmt.Println("chiding")
						res, err := skpr.ChideCheating(ctx)
						if err != nil {
							fmt.Println(err)
						}

						if err := slackClient.Post(ctx, &slack.SlackMessage{
							Text: res,
						}); err != nil {
							fmt.Println(err)
						}

					} else {
						fmt.Println("updating score")
						res, err := skpr.UpdateScore(ctx, username)
						if err != nil {
							fmt.Println(err)
						}

						if err := slackClient.Post(ctx, &slack.SlackMessage{
							Text: res,
						}); err != nil {
							fmt.Println(err)
						}
					}

				} else if strings.Contains(text, "score") {
					res, err := skpr.GetScores(ctx)
					if err != nil {
						fmt.Println(err)
					}

					if err := slackClient.Post(ctx, &slack.SlackMessage{
						Text: res,
					}); err != nil {
						fmt.Println(err)
					}
				}
			}
		}()

		c.JSON(http.StatusOK, gin.H{"status": "Received"})
	})

	router.Run(":8086")
}
