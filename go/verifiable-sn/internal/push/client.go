package push

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// Client sends FCM push notifications. Implementations may be nil (no-op) when Firebase is not configured.
type Client interface {
	Send(ctx context.Context, token, title, body string, data map[string]string) error
}

type fcmClient struct {
	client *messaging.Client
}

func (c *fcmClient) Send(ctx context.Context, token, title, body string, data map[string]string) error {
	msg := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data:  data,
		Token: token,
	}
	_, err := c.client.Send(ctx, msg)
	return err
}

// NewClient returns a push client. When GOOGLE_APPLICATION_CREDENTIALS is set to a path
// to a service account JSON file, returns an FCM client. Otherwise returns nil (no-op).
func NewClient() Client {
	credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credPath == "" {
		return nil
	}
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(credPath))
	if err != nil {
		log.Printf("[push] failed to init Firebase: %v", err)
		return nil
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		log.Printf("[push] failed to get Messaging client: %v", err)
		return nil
	}
	return &fcmClient{client: client}
}
