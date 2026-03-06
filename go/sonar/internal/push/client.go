package push

import (
	"context"
	"log"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type Client interface {
	Send(ctx context.Context, token, title, body string, data map[string]string) error
}

type fcmClient struct {
	client *messaging.Client
}

func (c *fcmClient) Send(ctx context.Context, token, title, body string, data map[string]string) error {
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data:  data,
		Token: token,
	}
	_, err := c.client.Send(ctx, message)
	return err
}

func NewClient() Client {
	credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if strings.TrimSpace(credPath) == "" {
		return nil
	}
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(credPath))
	if err != nil {
		log.Printf("[push] failed to initialize firebase app: %v", err)
		return nil
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		log.Printf("[push] failed to create firebase messaging client: %v", err)
		return nil
	}
	return &fcmClient{client: client}
}
