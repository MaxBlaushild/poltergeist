package main

import (
	"context"
	"fmt"
	"time"
	_ "time/tzdata"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/billing"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/email"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	"github.com/MaxBlaushild/poltergeist/trivai/internal/config"
	"github.com/MaxBlaushild/poltergeist/trivai/internal/server"
	"github.com/MaxBlaushild/poltergeist/trivai/internal/trivai"
)

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

	deepPriest := deep_priest.SummonDeepPriest()
	texterClient := texter.NewClient()
	emailClient := email.NewClient(email.ClientConfig{
		ApiKey:      cfg.Secret.SendgridApiKey,
		FromAddress: cfg.Public.EmailFromAddress,
		WebHost:     cfg.Public.WebHost,
	})

	trivaiClient := trivai.NewClient(deepPriest)
	billingClient := billing.NewClient()
	authClient := auth.NewClient()

	go server.NewServer(dbClient, emailClient, trivaiClient, texterClient, billingClient, *cfg, authClient)

	// poll for new questions
	loc, _ := time.LoadLocation("America/New_York")
	for {
		fmt.Println("Polling for current question updates")
		// Get the current time in EST
		now := time.Now().In(loc)

		// Calculate the next time the task should run
		next := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, loc)
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}

		// Wait until the next scheduled time
		time.Sleep(next.Sub(now))

		// Do the task
		currentQuestion, err := dbClient.HowManyQuestion().FindTodaysQuestion(ctx)
		if err != nil {
			fmt.Println("fetch current question error")
			fmt.Println(err)
		}

		if currentQuestion != nil {
			var newQuestion *trivai.HowManyQuestion = nil
			var promptIndex int = currentQuestion.PromptSeedIndex + 1
			var prompt string = ""
			var err error
			for newQuestion == nil {
				prompt = trivai.PromptSeeds[promptIndex]
				newQuestion, err = trivaiClient.GenerateNewHowManyQuestion(ctx, prompt)
				if err != nil {
					fmt.Println("error generating new question")
					fmt.Println(err.Error())
				}
				promptIndex++
			}

			_, err = dbClient.HowManyQuestion().Insert(ctx, newQuestion.Text, newQuestion.Explanation, newQuestion.HowMany, promptIndex, prompt)
			if err != nil {
				fmt.Println("error inserting new question")
				fmt.Println(err.Error())
			}

			fmt.Println("going to mark the last question as done")

			if err := dbClient.HowManyQuestion().MarkDone(ctx, currentQuestion.ID); err != nil {
				fmt.Println("mark current question done error")
				fmt.Println(err)
			}

			fmt.Println("successfully marked the question as done")
		}
	}
}
