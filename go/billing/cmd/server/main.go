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

const (
	sessionCompletedEventType    = "checkout.session.completed"
	subscriptionDeletedEventType = "customer.subscription.deleted"
)

func forwardCreateSubscription(ctx *gin.Context, session *stripe.CheckoutSession, url string) error {
	onSubscribe := billing.OnSubscribe{
		Metadata:       session.Metadata,
		SubscriptionID: session.Subscription.ID,
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

func forwardCancelSubscription(ctx *gin.Context, subscription *stripe.Subscription, url string) error {
	onUnsubscribe := billing.OnSubscriptionDelete{
		Metadata:       subscription.Metadata,
		SubscriptionID: subscription.ID,
	}
	jsonBody, err := json.Marshal(onUnsubscribe)
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
		params := &stripe.SubscriptionListParams{}
	})

	router.POST("/billing/checkout-session", func(ctx *gin.Context) {
		var params billing.CheckoutSessionParams

		if err := ctx.Bind(&params); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		params.Metadata["create_callback_url"] = params.SubscriptionCreateCallbackUrl
		params.Metadata["cancel_callback_url"] = params.SubscriptionCancelCallbackUrl

		session, err := session.New(&stripe.CheckoutSessionParams{
			SuccessURL: &params.SessionSuccessRedirectUrl,
			CancelURL:  &params.SessionCancelRedirectUrl,
			Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					Price:    stripe.String(params.PlanID),
					Quantity: stripe.Int64(1),
				},
			},
			Metadata: params.Metadata,
			SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
				Metadata: params.Metadata,
			},
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

		if event.Type == sessionCompletedEventType {
			session := stripe.CheckoutSession{}
			if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
				fmt.Println("garbage checkout session from stripe")
				fmt.Println(string(event.Data.Raw))
			}

			createCallbackUrl, ok := session.Metadata["create_callback_url"]
			if ok {
				if err := forwardCreateSubscription(ctx, &session, createCallbackUrl); err != nil {
					fmt.Println(err.Error())
					ctx.JSON(500, gin.H{
						"message": err.Error(),
					})
				}
				return
			}
		}

		if event.Type == subscriptionDeletedEventType {
			subscription := stripe.Subscription{}
			if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
				fmt.Println("garbage subscription from stripe")
				fmt.Println(string(event.Data.Raw))
			}

			cancelCallbackUrl, ok := subscription.Metadata["cancel_callback_url"]
			if ok {
				if err := forwardCancelSubscription(ctx, &subscription, cancelCallbackUrl); err != nil {
					fmt.Println(err.Error())
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
