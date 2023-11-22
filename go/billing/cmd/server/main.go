package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/billing/internal/config"
	"github.com/MaxBlaushild/poltergeist/pkg/billing"
	"github.com/stripe/stripe-go/v75/checkout/session"

	"github.com/gin-gonic/gin"

	"github.com/stripe/stripe-go/v75"
)

func forwardSubscription(ctx *gin.Context, url string, metadata map[string]string) error {
	onSubscribe := billing.OnSubscribe{
		Metadata: metadata,
	}
	jsonBody, err := json.Marshal(onSubscribe)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("received non-OK response from server")
	}
	return nil
}

func main() {
	router := gin.Default()

	cfg, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		panic(err)
	}

	stripe.Key = cfg.Secret.StripeSecretKey

	router.POST("/billing/subscriptions/cancel", func(ctx *gin.Context) {

	})

	router.POST("/billing/checkout-session", func(ctx *gin.Context) {
		var params billing.CheckoutSessionParams

		if err := ctx.Bind(&params); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		params.Metadata["callback_url"] = params.CallbackUrl

		session, err := session.New(&stripe.CheckoutSessionParams{
			SuccessURL: &params.SuccessUrl,
			CancelURL:  &params.CancelUrl,
			Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					Price:    stripe.String(params.PlanID),
					Quantity: stripe.Int64(1),
				},
			},
			Metadata: params.Metadata,
		})
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(200, billing.CheckoutSessionResponse{
			URL: session.URL,
		})
	})

	router.POST("/billing/stripe-webhook", func(ctx *gin.Context) {
		var event stripe.Event

		if err := ctx.ShouldBindJSON(&event); err != nil {
			ctx.JSON(http.StatusBadRequest, err.Error())
			return
		}

		if event.Type == "checkout.session.completed" {
			session := &stripe.CheckoutSession{}
			if err := json.Unmarshal(event.Data.Raw, session); err != nil {
				fmt.Println("garbage from stripe")
				fmt.Println(string(event.Data.Raw))
			}

			callbackUrl, ok := session.Metadata["callback_url"]
			if ok {
				fmt.Println(callbackUrl)
				if err := forwardSubscription(ctx, callbackUrl, session.Metadata); err != nil {
					// (TODO): Cancel subscription
					fmt.Println(err)
					ctx.JSON(500, gin.H{
						"message": err.Error(),
					})
				}
				return
			}
		}

		ctx.JSON(200, gin.H{
			"message": "event handled successfully",
		})
	})

	router.Run(":8022")
}
