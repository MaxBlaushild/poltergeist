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
	primary := strings.TrimSpace(os.Getenv("UNCLAIMED_STREETS_APPLICATION_CREDENTIALS"))
	fallback := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	log.Printf(
		"[push][init] env vars: UNCLAIMED_STREETS_APPLICATION_CREDENTIALS set=%t, GOOGLE_APPLICATION_CREDENTIALS set=%t",
		primary != "",
		fallback != "",
	)

	if primary == "" && fallback == "" {
		log.Printf("[push][init] no credential env vars set; push client disabled")
		return nil
	}

	ctx := context.Background()
	tryPath := func(credPath string) Client {
		log.Printf("[push][init] attempting firebase credentials from %s", credPath)
		if info, err := os.Stat(credPath); err != nil {
			log.Printf("[push][init] credentials file check failed for %s: %v", credPath, err)
			return nil
		} else {
			log.Printf("[push][init] credentials file found at %s (%d bytes)", credPath, info.Size())
		}

		app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(credPath))
		if err != nil {
			log.Printf("[push] failed to initialize firebase app with %s: %v", credPath, err)
			return nil
		}
		client, err := app.Messaging(ctx)
		if err != nil {
			log.Printf("[push] failed to create firebase messaging client with %s: %v", credPath, err)
			return nil
		}
		log.Printf("[push][init] firebase messaging client configured successfully with %s", credPath)
		return &fcmClient{client: client}
	}

	// If an Unclaimed Streets-specific credential path is set, fail closed on that path.
	// Falling back to another Firebase project causes SenderId mismatch for iOS tokens.
	if primary != "" {
		client := tryPath(primary)
		if client != nil {
			return client
		}
		log.Printf("[push][init] UNCLAIMED_STREETS_APPLICATION_CREDENTIALS is set but invalid; refusing fallback")
		return nil
	}

	if fallback != "" {
		client := tryPath(fallback)
		if client != nil {
			return client
		}
	}
	log.Printf("[push][init] all credential paths failed; push client disabled")
	return nil
}
