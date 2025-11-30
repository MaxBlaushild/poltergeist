package hue

import "context"

// Client provides methods for interacting with Philips Hue lights
type Client interface {
	// DiscoverBridge discovers a Hue bridge on the local network
	DiscoverBridge(ctx context.Context) (*Bridge, error)

	// Connect connects to a bridge using the provided hostname/IP and username
	Connect(ctx context.Context, hostname, username string) error

	// CreateUser creates a new API user on the bridge
	// The bridge button must be pressed before calling this
	CreateUser(ctx context.Context, deviceType string) (string, error)

	// GetLights returns all lights connected to the bridge
	GetLights(ctx context.Context) ([]*Light, error)

	// GetLight returns a specific light by ID
	GetLight(ctx context.Context, id int) (*Light, error)

	// TurnOn turns on a light by ID
	TurnOn(ctx context.Context, id int) error

	// TurnOff turns off a light by ID
	TurnOff(ctx context.Context, id int) error

	// SetBrightness sets the brightness of a light (0-254)
	SetBrightness(ctx context.Context, id int, brightness uint8) error

	// SetColorRGB sets the color of a light using RGB values (0-255)
	SetColorRGB(ctx context.Context, id int, r, g, b uint8) error

	// SetColorXY sets the color of a light using CIE XY coordinates
	SetColorXY(ctx context.Context, id int, x, y float32) error

	// SetColorTemperature sets the color temperature in mireds (154-500)
	SetColorTemperature(ctx context.Context, id int, mireds uint16) error

	// SetState sets multiple properties of a light at once
	SetState(ctx context.Context, id int, state *LightState) error

	// TurnOnAll turns on all lights
	TurnOnAll(ctx context.Context) error

	// TurnOffAll turns off all lights
	TurnOffAll(ctx context.Context) error
}

