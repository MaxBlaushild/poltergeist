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
			fmt.Println("going to mark the last question as done")

			if err := dbClient.HowManyQuestion().MarkDone(ctx, currentQuestion.ID); err != nil {
				fmt.Println("mark current question done error")
				fmt.Println(err)
			}

			fmt.Println("successfully marked the question as done")

			// Get new question
			newQuestion, err := dbClient.HowManyQuestion().FindTodaysQuestion(ctx)
			if err != nil {
				fmt.Println("fetch current question error")
				fmt.Println(err)
			}

			fmt.Println("fetched the new question")

			subscriptions, err := dbClient.HowManySubscription().FindAll(ctx)
			if err != nil {
				fmt.Println("fetch subscriptions error")
				fmt.Println(err)
			}

			fmt.Println("fetched all them subscriptions")

			if newQuestion != nil {
				for _, subscription := range subscriptions {
					var shouldSend bool = false
					if subscription.Subscribed {
						shouldSend = true
					}

					if subscription.NumFreeQuestions < 7 {
						shouldSend = true
					}

					if shouldSend {
						if err := texterClient.Text(ctx, &texter.Text{
							Body:     newQuestion.Text,
							To:       subscription.User.PhoneNumber,
							From:     cfg.Secret.GuessHowManyPhoneNumber,
							TextType: "guess-how-many-question",
						}); err != nil {
							fmt.Println("error sending text")
							fmt.Println(subscription.User.PhoneNumber)
						}
					}

					fmt.Println("sent message to: ")
					fmt.Println(subscription.User.PhoneNumber)

					if shouldSend && !subscription.Subscribed {
						if err := dbClient.HowManySubscription().IncrementNumFreeQuestions(ctx, subscription.UserID); err != nil {
							fmt.Println("error incrementing user id")
							fmt.Println(subscription.UserID)
						}
					}

					if !subscription.Subscribed && subscription.NumFreeQuestions == 7 {
						if err := texterClient.Text(ctx, &texter.Text{
							Body:     fmt.Sprintf("Sorry, you're all out of free questions! Keep the daily questions coming by subscribing:\n\nhttps://api.guesswith.us/trivai/users/%s/subscribe", subscription.UserID.String()),
							To:       subscription.User.PhoneNumber,
							From:     cfg.Secret.GuessHowManyPhoneNumber,
							TextType: "guess-how-many-out-of-free-questions",
						}); err != nil {
							fmt.Println("error sending text")
							fmt.Println(subscription.User.PhoneNumber)
						}

						if err := dbClient.HowManySubscription().IncrementNumFreeQuestions(ctx, subscription.UserID); err != nil {
							fmt.Println("error incrementing user id after sending last free text")
							fmt.Println(subscription.UserID)
						}
					}
				}
			}

		}

		if countLeft, err := dbClient.HowManyQuestion().ValidQuestionsRemaining(ctx); err != nil {
			fmt.Println("error getting num valid subscriptions left")
			fmt.Println(err.Error())
		} else {
			if countLeft < 3 {
				if err := texterClient.Text(ctx, &texter.Text{
					Body:     fmt.Sprintf("Hey dumbass! You only have %d questions left. Make some new ones.", countLeft),
					To:       "+14407858475",
					From:     cfg.Secret.GuessHowManyPhoneNumber,
					TextType: "idiot-reminded",
				}); err != nil {
					fmt.Println("error sending text")
					fmt.Println(err.Error())
				}
			}
		}
	}
}
