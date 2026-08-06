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
	"github.com/stripe/stripe-go/v75/client"
	"github.com/stripe/stripe-go/v75/subscription"
	"github.com/stripe/stripe-go/v75/webhook"
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

func forwardPaymentComplete(ctx *gin.Context, session *stripe.CheckoutSession, url string) error {
	// Get amount from session amount_total (in cents)
	amountInCents := session.AmountTotal

	onPaymentComplete := billing.OnPaymentComplete{
		Metadata:      session.Metadata,
		SessionID:     session.ID,
		AmountInCents: amountInCents,
	}
	if session.CustomerDetails != nil {
		onPaymentComplete.CustomerEmail = session.CustomerDetails.Email
	}
	if session.ShippingDetails != nil && session.ShippingDetails.Address != nil {
		addr := session.ShippingDetails.Address
		onPaymentComplete.ShippingAddress = &billing.ShippingAddress{
			Name:       session.ShippingDetails.Name,
			Line1:      addr.Line1,
			Line2:      addr.Line2,
			City:       addr.City,
			State:      addr.State,
			PostalCode: addr.PostalCode,
			Country:    addr.Country,
		}
	}
	jsonBody, err := json.Marshal(onPaymentComplete)
	if err != nil {
		fmt.Printf("[forwardPaymentComplete] ERROR marshaling JSON: %v\n", err)
		return err
	}

	fmt.Printf("[forwardPaymentComplete] POSTing to %s with body: %s\n", url, string(jsonBody))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("[forwardPaymentComplete] ERROR making POST request: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("[forwardPaymentComplete] Response status: %d\n", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		bodyBytes := make([]byte, 1024)
		n, _ := resp.Body.Read(bodyBytes)
		fmt.Printf("[forwardPaymentComplete] Response body: %s\n", string(bodyBytes[:n]))
		return fmt.Errorf("received non-OK response from server: %d", resp.StatusCode)
	}
	fmt.Printf("[forwardPaymentComplete] Successfully forwarded payment complete\n")
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
		var params billing.CancelSubscriptionParams

		if err := ctx.Bind(&params); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		if _, err := subscription.Cancel(params.StripeID, &stripe.SubscriptionCancelParams{}); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(200, billing.CancelSubscriptionResponse{})
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

	router.POST("/billing/payment-checkout-session", func(ctx *gin.Context) {
		var params billing.PaymentCheckoutSessionParams

		if err := ctx.Bind(&params); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		if params.Metadata == nil {
			params.Metadata = make(map[string]string)
		}
		params.Metadata["payment_complete_callback_url"] = params.PaymentCompleteCallbackUrl

		var lineItems []*stripe.CheckoutSessionLineItemParams
		if len(params.LineItems) > 0 {
			// Itemized session (R-2.8/R-6.2) — the customer sees each line,
			// not one lump sum.
			for _, item := range params.LineItems {
				lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
					PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
						Currency: stripe.String("usd"),
						ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
							Name: stripe.String(item.Name),
						},
						UnitAmount: stripe.Int64(item.AmountInCents),
					},
					Quantity: stripe.Int64(item.Quantity),
				})
			}
		} else {
			// Original single-line-item shape, unchanged for existing callers.
			lineItems = []*stripe.CheckoutSessionLineItemParams{
				{
					PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
						Currency: stripe.String("usd"),
						ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
							Name: stripe.String("Travel Angels Credits"),
						},
						UnitAmount: stripe.Int64(params.AmountInCents),
					},
					Quantity: stripe.Int64(1),
				},
			}
		}

		sessionParams := &stripe.CheckoutSessionParams{
			SuccessURL: &params.SessionSuccessRedirectUrl,
			CancelURL:  &params.SessionCancelRedirectUrl,
			Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
			LineItems:  lineItems,
			Metadata:   params.Metadata,
		}
		if params.AutomaticTax {
			sessionParams.AutomaticTax = &stripe.CheckoutSessionAutomaticTaxParams{
				Enabled: stripe.Bool(true),
			}
		}
		if params.CollectShippingAddress {
			// v1 is US-only (R-1.2: no international shipping).
			sessionParams.ShippingAddressCollection = &stripe.CheckoutSessionShippingAddressCollectionParams{
				AllowedCountries: stripe.StringSlice([]string{"US"}),
			}
		}

		// Platform "reef" routes through reef-site's own Stripe account
		// (separate business, separate credentials) rather than the
		// shared key every other caller of this endpoint (travel-angels)
		// still uses.
		var sess *stripe.CheckoutSession
		if params.Platform == "reef" {
			// New Stripe accounts default to "Managed Payments," which
			// rejects the classic shipping_address_collection param this
			// service builds above ("Shipping parameters cannot be used
			// with Managed Payments"). stripe-go v75 predates a typed field
			// for this, so it's set via the SDK's raw-param escape hatch
			// rather than bumping the (shared, multi-domain) stripe-go
			// dependency just for one new account's default.
			sessionParams.AddExtra("managed_payments[enabled]", "false")
			sc := client.New(cfg.Secret.ReefStripeSecretKey, nil)
			sess, err = sc.CheckoutSessions.New(sessionParams)
		} else {
			sess, err = session.New(sessionParams)
		}
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(200, billing.CheckoutSessionResponse{
			URL: sess.URL,
		})
	})

	router.POST("/billing/stripe-webhook", func(ctx *gin.Context) {
		// Previously this bound the POST body directly with no check that
		// it actually came from Stripe — anyone who could guess/observe an
		// order's metadata could forge a "checkout.session.completed"
		// event and get an order marked paid (fulfillment triggered,
		// confirmation email sent) without paying. ConstructEvent verifies
		// the Stripe-Signature header against the raw body using the
		// webhook's real signing secret (Stripe dashboard -> Developers ->
		// Webhooks -> this endpoint -> "Signing secret"), so a forged
		// request now fails here instead of being processed.
		payload, err := ctx.GetRawData()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, err.Error())
			return
		}

		// This one endpoint receives events from two separate Stripe
		// accounts (the original shared one and reef-site's own), each
		// signing with its own secret — try both, succeed on whichever
		// matches. IgnoreAPIVersionMismatch is safe here: the API-version
		// check only runs after the HMAC signature has already been
		// validated (see stripe-go's webhook.constructEvent), so this
		// doesn't weaken auth — it just stops rejecting genuine events
		// from reef's newer Stripe account, which defaults to a newer
		// API version than this pinned stripe-go release expects.
		sigHeader := ctx.GetHeader("Stripe-Signature")
		opts := webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true}
		event, err := webhook.ConstructEventWithOptions(payload, sigHeader, cfg.Secret.StripeWebhookSecret, opts)
		if err != nil && cfg.Secret.ReefStripeWebhookSecret != "" {
			event, err = webhook.ConstructEventWithOptions(payload, sigHeader, cfg.Secret.ReefStripeWebhookSecret, opts)
		}
		if err != nil {
			fmt.Printf("[StripeWebhook] signature verification failed: %v\n", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
			return
		}

		if event.Type == sessionCompletedEventType {
			session := stripe.CheckoutSession{}
			if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
				fmt.Println("garbage checkout session from stripe")
				fmt.Println(string(event.Data.Raw))
			}

			// Check if this is a subscription checkout
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

			// Check if this is a payment checkout
			paymentCallbackUrl, ok := session.Metadata["payment_complete_callback_url"]
			if ok {
				fmt.Printf("[StripeWebhook] Payment checkout detected, forwarding to: %s\n", paymentCallbackUrl)
				fmt.Printf("[StripeWebhook] Session ID: %s, Amount: %d cents\n", session.ID, session.AmountTotal)
				fmt.Printf("[StripeWebhook] Metadata: %+v\n", session.Metadata)
				if err := forwardPaymentComplete(ctx, &session, paymentCallbackUrl); err != nil {
					fmt.Printf("[StripeWebhook] ERROR forwarding payment complete: %v\n", err)
					ctx.JSON(500, gin.H{
						"message": err.Error(),
					})
					return
				}
				fmt.Printf("[StripeWebhook] Successfully forwarded payment complete\n")
				return
			} else {
				fmt.Printf("[StripeWebhook] Payment checkout session but no payment_complete_callback_url in metadata\n")
				fmt.Printf("[StripeWebhook] Available metadata keys: %+v\n", session.Metadata)
			}
		}

		if event.Type == subscriptionDeletedEventType {
			subscription := stripe.Subscription{}
			if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
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
