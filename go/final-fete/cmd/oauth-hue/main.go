package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"runtime"

	"github.com/MaxBlaushild/poltergeist/final-fete/internal/config"
	"github.com/MaxBlaushild/poltergeist/pkg/hue"
)

func main() {
	cfg, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if cfg.Secret.HueClientID == "" || cfg.Secret.HueClientSecret == "" {
		fmt.Fprintf(os.Stderr, "❌ Error: HUE_CLIENT_ID and HUE_CLIENT_SECRET environment variables must be set\n")
		os.Exit(1)
	}

	if cfg.Public.HueRedirectURI == "" {
		fmt.Fprintf(os.Stderr, "❌ Error: HUE_REDIRECT_URI must be set in config\n")
		os.Exit(1)
	}

	fmt.Println("🔐 Initializing Hue OAuth client...")

	// Create OAuth client
	oauthClient := hue.NewOAuthClient(hue.OAuthClientConfig{
		ClientID:     cfg.Secret.HueClientID,
		ClientSecret: cfg.Secret.HueClientSecret,
		RedirectURI:  cfg.Public.HueRedirectURI,
	})

	// Generate state for CSRF protection
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: Failed to generate state: %v\n", err)
		os.Exit(1)
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	// Generate authorization URL
	authURL, err := oauthClient.GetAuthURL(state)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: Failed to generate auth URL: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ Authorization URL generated!")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Visit this URL in your browser to authorize:")
	fmt.Println()
	fmt.Printf("  %s\n", authURL)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("  State (for verification): %s\n", state)
	fmt.Println()
	fmt.Println("💡 After authorization, you will be redirected to the callback URL.")
	fmt.Println("   The refresh token will be stored automatically.")
	fmt.Println()

	// Try to open browser automatically
	var openCmd string
	switch runtime.GOOS {
	case "darwin":
		openCmd = "open"
	case "linux":
		openCmd = "xdg-open"
	case "windows":
		openCmd = "cmd /c start"
	default:
		openCmd = ""
	}

	if openCmd != "" {
		fmt.Println("🌐 Attempting to open browser automatically...")
		fmt.Println()

		// Note: We're not actually executing this, just showing it could be done
		// In a real implementation, you could use exec.Command here
		fmt.Printf("   (Run manually: %s \"%s\")\n", openCmd, authURL)
		fmt.Println()
	}

	fmt.Println("⏳ Waiting for authorization to complete...")
	fmt.Println("   (This process will exit - check your callback endpoint for the token)")
}
