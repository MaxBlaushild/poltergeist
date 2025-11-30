package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/hue"
)

func main() {
	ctx := context.Background()

	fmt.Println("🔍 Discovering Hue bridge on local network...")

	client := hue.NewClient()

	// Discover bridge
	bridge, err := client.DiscoverBridge(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: Failed to discover Hue bridge: %v\n", err)
		fmt.Fprintf(os.Stderr, "   Make sure your bridge is powered on and connected to the same network.\n")
		os.Exit(1)
	}

	fmt.Printf("✅ Found bridge at: %s\n", bridge.Hostname)
	fmt.Println()
	fmt.Println("📋 IMPORTANT: Before continuing, press the button on your Hue bridge!")
	fmt.Println("   You have 30 seconds after pressing the button to complete registration.")
	fmt.Println()
	fmt.Print("Press Enter after you've pressed the bridge button, or wait 5 seconds to continue automatically...")

	// Wait for user input or timeout
	done := make(chan bool)
	go func() {
		var input string
		fmt.Scanln(&input)
		done <- true
	}()

	select {
	case <-done:
		// User pressed Enter
	case <-time.After(5 * time.Second):
		// Timeout, continue automatically
		fmt.Println("\n⏱️  Continuing automatically...")
	}

	fmt.Println()
	fmt.Println("🔐 Creating API user...")
	fmt.Println("   Device type: final-fete")
	fmt.Println("Bridge hostname: ", bridge.Hostname)

	// Connect to bridge (without username for initial connection)
	err = client.Connect(ctx, fmt.Sprintf("https://%s", bridge.Hostname), "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: Failed to connect to bridge: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🔐 Connected to bridge")

	// Create user
	username, err := client.CreateUser(ctx, "fete")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: Failed to create user: %v\n", err)
		fmt.Fprintf(os.Stderr, "   Make sure you pressed the bridge button within the last 30 seconds.\n")
		fmt.Fprintf(os.Stderr, "   You may need to run this command again.\n")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ Success! Your Hue username has been created.")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Username: %s\n", username)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("💡 Save this username - you'll need it to connect to your bridge.")
	fmt.Printf("   Bridge hostname: %s\n", bridge.Hostname)
	fmt.Println()
}
