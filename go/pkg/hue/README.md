# Hue SDK

A Go SDK for programmatic control of Philips Hue lights. This package provides a clean interface for discovering Hue bridges, authenticating, and controlling lights.

## Prerequisites

- Go 1.18 or later
- A Philips Hue bridge on your local network
- Network access to the bridge (same network or configured routing)

## Installation

```bash
go get github.com/MaxBlaushild/poltergeist/pkg/hue
```

## Setup

### 1. Discover Your Bridge

The Hue bridge must be on the same local network as your application. The SDK can automatically discover bridges:

```go
package main

import (
    "context"
    "fmt"
    "github.com/MaxBlaushild/poltergeist/pkg/hue"
)

func main() {
    ctx := context.Background()
    client := hue.NewClient()
    
    bridge, err := client.DiscoverBridge(ctx)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found bridge at: %s\n", bridge.Hostname)
}
```

### 2. Create API Credentials

Before you can control lights, you need to create an API user on the bridge:

1. **Press the button on your Hue bridge** (this must be done within 30 seconds)
2. Call `CreateUser` to generate credentials:

```go
ctx := context.Background()
client := hue.NewClient()

// Discover bridge
bridge, err := client.DiscoverBridge(ctx)
if err != nil {
    panic(err)
}

// Connect to bridge (without username initially)
err = client.Connect(ctx, bridge.Hostname, "")
if err != nil {
    panic(err)
}

// Create user (bridge button must be pressed!)
username, err := client.CreateUser(ctx, "my-app-name")
if err != nil {
    panic(err)
}

fmt.Printf("Save this username for future use: %s\n", username)
```

**Important:** The bridge button must be pressed within 30 seconds before calling `CreateUser`. You only need to do this once - save the username for future connections.

### 3. Connect with Credentials

Once you have a username, you can connect directly:

```go
ctx := context.Background()
client := hue.NewClient()

// Connect using known credentials
err := client.Connect(ctx, "192.168.1.100", "your-username-here")
if err != nil {
    panic(err)
}
```

Or use the convenience constructor:

```go
client := hue.NewClientWithBridge("192.168.1.100", "your-username-here")
```

## Usage Examples

### Get All Lights

```go
lights, err := client.GetLights(ctx)
if err != nil {
    panic(err)
}

for _, light := range lights {
    fmt.Printf("Light %d: %s (On: %v, Brightness: %d)\n", 
        light.ID, light.Name, light.On, light.Brightness)
}
```

### Get a Specific Light

```go
light, err := client.GetLight(ctx, 1)
if err != nil {
    panic(err)
}

fmt.Printf("Light: %s\n", light.Name)
```

### Turn Lights On/Off

```go
// Turn on light 1
err := client.TurnOn(ctx, 1)
if err != nil {
    panic(err)
}

// Turn off light 1
err = client.TurnOff(ctx, 1)
if err != nil {
    panic(err)
}

// Turn on all lights
err = client.TurnOnAll(ctx)

// Turn off all lights
err = client.TurnOffAll(ctx)
```

### Set Brightness

Brightness ranges from 0 (off) to 254 (maximum brightness):

```go
// Set light 1 to 50% brightness (127)
err := client.SetBrightness(ctx, 1, 127)
if err != nil {
    panic(err)
}
```

### Set Color (RGB)

Set a light to a specific RGB color (0-255 for each component):

```go
// Set light 1 to red
err := client.SetColorRGB(ctx, 1, 255, 0, 0)

// Set light 1 to blue
err = client.SetColorRGB(ctx, 1, 0, 0, 255)

// Set light 1 to white
err = client.SetColorRGB(ctx, 1, 255, 255, 255)
```

### Set Color (XY Coordinates)

Set color using CIE XY coordinates (useful for precise color matching):

```go
// Set to a specific XY color
err := client.SetColorXY(ctx, 1, 0.5, 0.4)
```

### Set Color Temperature

Set the color temperature in mireds (154-500, where lower is warmer and higher is cooler):

```go
// Warm white (around 2000K = ~500 mireds)
err := client.SetColorTemperature(ctx, 1, 500)

// Cool white (around 6500K = ~154 mireds)
err = client.SetColorTemperature(ctx, 1, 154)
```

### Set Multiple Properties at Once

Use `SetState` to change multiple properties in a single operation:

```go
state := &hue.LightState{
    On:         boolPtr(true),
    Brightness: uint8Ptr(200),
    ColorRGB:   &[3]uint8{255, 100, 50}, // Orange
}

err := client.SetState(ctx, 1, state)
if err != nil {
    panic(err)
}
```

### Complete Example

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/MaxBlaushild/poltergeist/pkg/hue"
)

func main() {
    ctx := context.Background()
    
    // Create client and connect
    client := hue.NewClientWithBridge("192.168.1.100", "your-username")
    
    // Get all lights
    lights, err := client.GetLights(ctx)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d lights\n", len(lights))
    
    // Turn on first light with red color
    if len(lights) > 0 {
        lightID := lights[0].ID
        
        err = client.TurnOn(ctx, lightID)
        if err != nil {
            panic(err)
        }
        
        err = client.SetColorRGB(ctx, lightID, 255, 0, 0)
        if err != nil {
            panic(err)
        }
        
        err = client.SetBrightness(ctx, lightID, 200)
        if err != nil {
            panic(err)
        }
        
        fmt.Printf("Light %d is now red at 78%% brightness\n", lightID)
        
        // Wait a bit
        time.Sleep(2 * time.Second)
        
        // Turn it off
        err = client.TurnOff(ctx, lightID)
        if err != nil {
            panic(err)
        }
    }
}

// Helper functions for creating pointers
func boolPtr(b bool) *bool { return &b }
func uint8Ptr(u uint8) *uint8 { return &u }
```

## API Reference

### Client Interface

- `DiscoverBridge(ctx) (*Bridge, error)` - Discover a bridge on the network
- `Connect(ctx, hostname, username string) error` - Connect to a bridge
- `CreateUser(ctx, deviceType string) (string, error)` - Create API credentials
- `GetLights(ctx) ([]*Light, error)` - Get all lights
- `GetLight(ctx, id int) (*Light, error)` - Get a specific light
- `TurnOn(ctx, id int) error` - Turn on a light
- `TurnOff(ctx, id int) error` - Turn off a light
- `SetBrightness(ctx, id int, brightness uint8) error` - Set brightness (0-254)
- `SetColorRGB(ctx, id int, r, g, b uint8) error` - Set RGB color
- `SetColorXY(ctx, id int, x, y float32) error` - Set XY color coordinates
- `SetColorTemperature(ctx, id int, mireds uint16) error` - Set color temperature
- `SetState(ctx, id int, state *LightState) error` - Set multiple properties
- `TurnOnAll(ctx) error` - Turn on all lights
- `TurnOffAll(ctx) error` - Turn off all lights

## Troubleshooting

### Bridge Not Found

- Ensure the bridge is powered on and connected to your network
- Make sure your application is on the same network as the bridge
- Try manually specifying the bridge IP address instead of using discovery

### Authentication Failed

- Verify you're using the correct username
- If you've lost your username, you'll need to create a new one (press bridge button first)

### Light Not Responding

- Check that the light is powered on
- Verify the light ID is correct (use `GetLights` to see available IDs)
- Check the `Reachable` property on the Light object

## License

This package is part of the Poltergeist project.

