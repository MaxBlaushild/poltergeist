package hue

import (
	"context"
	"crypto/tls"
	"fmt"
	"math"
	"net/http"

	"github.com/amimof/huego"
)

func init() {
	// Set the default HTTP client transport to skip TLS verification
	// This is needed because Hue bridges use self-signed certificates
	// that don't pass standard TLS verification
	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.InsecureSkipVerify = true
	} else {
		// If DefaultTransport is not *http.Transport, create a new one
		http.DefaultTransport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
}

type client struct {
	bridge *huego.Bridge
}

// NewClient creates a new Hue client
func NewClient() Client {
	return &client{}
}

// NewClientWithBridge creates a new Hue client with a pre-configured bridge
func NewClientWithBridge(hostname, username string) Client {
	bridge := huego.New(hostname, username)
	return &client{bridge: bridge}
}

// DiscoverBridge discovers a Hue bridge on the local network
func (c *client) DiscoverBridge(ctx context.Context) (*Bridge, error) {
	bridge, err := huego.Discover()
	if err != nil {
		return nil, fmt.Errorf("failed to discover bridge: %w", err)
	}

	return &Bridge{
		Hostname: bridge.Host,
		Username: "",
	}, nil
}

// Connect connects to a bridge using the provided hostname/IP and username
func (c *client) Connect(ctx context.Context, hostname, username string) error {
	c.bridge = huego.New(hostname, username)
	return nil
}

// CreateUser creates a new API user on the bridge
// The bridge button must be pressed before calling this
func (c *client) CreateUser(ctx context.Context, deviceType string) (string, error) {
	if c.bridge == nil {
		return "", fmt.Errorf("bridge not connected, call Connect or DiscoverBridge first")
	}

	if deviceType == "" {
		deviceType = "poltergeist-hue-client"
	}

	user, err := c.bridge.CreateUser(deviceType)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetLights returns all lights connected to the bridge
func (c *client) GetLights(ctx context.Context) ([]*Light, error) {
	if c.bridge == nil {
		return nil, fmt.Errorf("bridge not connected, call Connect or DiscoverBridge first")
	}

	hueLights, err := c.bridge.GetLights()
	if err != nil {
		return nil, fmt.Errorf("failed to get lights: %w", err)
	}

	lights := make([]*Light, 0, len(hueLights))
	for _, hueLight := range hueLights {
		light := c.convertLight(&hueLight)
		lights = append(lights, light)
	}

	return lights, nil
}

// GetLight returns a specific light by ID
func (c *client) GetLight(ctx context.Context, id int) (*Light, error) {
	if c.bridge == nil {
		return nil, fmt.Errorf("bridge not connected, call Connect or DiscoverBridge first")
	}

	hueLight, err := c.bridge.GetLight(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get light %d: %w", id, err)
	}

	return c.convertLight(hueLight), nil
}

// TurnOn turns on a light by ID
func (c *client) TurnOn(ctx context.Context, id int) error {
	if c.bridge == nil {
		return fmt.Errorf("bridge not connected, call Connect or DiscoverBridge first")
	}

	light, err := c.bridge.GetLight(id)
	if err != nil {
		return fmt.Errorf("failed to get light %d: %w", id, err)
	}

	if err := light.On(); err != nil {
		return fmt.Errorf("failed to turn on light %d: %w", id, err)
	}

	return nil
}

// TurnOff turns off a light by ID
func (c *client) TurnOff(ctx context.Context, id int) error {
	if c.bridge == nil {
		return fmt.Errorf("bridge not connected, call Connect or DiscoverBridge first")
	}

	light, err := c.bridge.GetLight(id)
	if err != nil {
		return fmt.Errorf("failed to get light %d: %w", id, err)
	}

	if err := light.Off(); err != nil {
		return fmt.Errorf("failed to turn off light %d: %w", id, err)
	}

	return nil
}

// SetBrightness sets the brightness of a light (0-254)
func (c *client) SetBrightness(ctx context.Context, id int, brightness uint8) error {
	if c.bridge == nil {
		return fmt.Errorf("bridge not connected, call Connect or DiscoverBridge first")
	}

	if brightness > 254 {
		brightness = 254
	}

	light, err := c.bridge.GetLight(id)
	if err != nil {
		return fmt.Errorf("failed to get light %d: %w", id, err)
	}

	if err := light.Bri(brightness); err != nil {
		return fmt.Errorf("failed to set brightness for light %d: %w", id, err)
	}

	return nil
}

// SetColorRGB sets the color of a light using RGB values (0-255)
// Converts RGB to CIE XY coordinates internally
func (c *client) SetColorRGB(ctx context.Context, id int, r, g, b uint8) error {
	if c.bridge == nil {
		return fmt.Errorf("bridge not connected, call Connect or DiscoverBridge first")
	}

	light, err := c.bridge.GetLight(id)
	if err != nil {
		return fmt.Errorf("failed to get light %d: %w", id, err)
	}

	// Convert RGB to XY coordinates
	x, y := rgbToXY(float64(r), float64(g), float64(b))

	if err := light.Xy([]float32{float32(x), float32(y)}); err != nil {
		return fmt.Errorf("failed to set RGB color for light %d: %w", id, err)
	}

	return nil
}

// SetColorXY sets the color of a light using CIE XY coordinates
func (c *client) SetColorXY(ctx context.Context, id int, x, y float32) error {
	if c.bridge == nil {
		return fmt.Errorf("bridge not connected, call Connect or DiscoverBridge first")
	}

	light, err := c.bridge.GetLight(id)
	if err != nil {
		return fmt.Errorf("failed to get light %d: %w", id, err)
	}

	if err := light.Xy([]float32{x, y}); err != nil {
		return fmt.Errorf("failed to set XY color for light %d: %w", id, err)
	}

	return nil
}

// SetColorTemperature sets the color temperature in mireds (154-500)
func (c *client) SetColorTemperature(ctx context.Context, id int, mireds uint16) error {
	if c.bridge == nil {
		return fmt.Errorf("bridge not connected, call Connect or DiscoverBridge first")
	}

	if mireds < 154 {
		mireds = 154
	}
	if mireds > 500 {
		mireds = 500
	}

	light, err := c.bridge.GetLight(id)
	if err != nil {
		return fmt.Errorf("failed to get light %d: %w", id, err)
	}

	if err := light.Ct(mireds); err != nil {
		return fmt.Errorf("failed to set color temperature for light %d: %w", id, err)
	}

	return nil
}

// SetState sets multiple properties of a light at once
func (c *client) SetState(ctx context.Context, id int, state *LightState) error {
	if c.bridge == nil {
		return fmt.Errorf("bridge not connected, call Connect or DiscoverBridge first")
	}

	light, err := c.bridge.GetLight(id)
	if err != nil {
		return fmt.Errorf("failed to get light %d: %w", id, err)
	}

	if state.On != nil {
		if *state.On {
			if err := light.On(); err != nil {
				return fmt.Errorf("failed to turn on light %d: %w", id, err)
			}
		} else {
			if err := light.Off(); err != nil {
				return fmt.Errorf("failed to turn off light %d: %w", id, err)
			}
		}
	}

	if state.Brightness != nil {
		brightness := *state.Brightness
		if brightness > 254 {
			brightness = 254
		}
		if err := light.Bri(brightness); err != nil {
			return fmt.Errorf("failed to set brightness for light %d: %w", id, err)
		}
	}

	if state.ColorRGB != nil {
		// Convert RGB to XY coordinates
		x, y := rgbToXY(float64(state.ColorRGB[0]), float64(state.ColorRGB[1]), float64(state.ColorRGB[2]))
		if err := light.Xy([]float32{float32(x), float32(y)}); err != nil {
			return fmt.Errorf("failed to set RGB color for light %d: %w", id, err)
		}
	}

	if state.ColorXY != nil {
		if err := light.Xy([]float32{state.ColorXY[0], state.ColorXY[1]}); err != nil {
			return fmt.Errorf("failed to set XY color for light %d: %w", id, err)
		}
	}

	if state.ColorTemperature != nil {
		ct := *state.ColorTemperature
		if ct < 154 {
			ct = 154
		}
		if ct > 500 {
			ct = 500
		}
		if err := light.Ct(ct); err != nil {
			return fmt.Errorf("failed to set color temperature for light %d: %w", id, err)
		}
	}

	if state.TransitionTime != nil {
		if err := light.TransitionTime(*state.TransitionTime); err != nil {
			return fmt.Errorf("failed to set transition time for light %d: %w", id, err)
		}
	}

	return nil
}

// TurnOnAll turns on all lights
func (c *client) TurnOnAll(ctx context.Context) error {
	if c.bridge == nil {
		return fmt.Errorf("bridge not connected, call Connect or DiscoverBridge first")
	}

	lights, err := c.bridge.GetLights()
	if err != nil {
		return fmt.Errorf("failed to get lights: %w", err)
	}

	for _, light := range lights {
		if err := light.On(); err != nil {
			return fmt.Errorf("failed to turn on light %d: %w", light.ID, err)
		}
	}

	return nil
}

// TurnOffAll turns off all lights
func (c *client) TurnOffAll(ctx context.Context) error {
	if c.bridge == nil {
		return fmt.Errorf("bridge not connected, call Connect or DiscoverBridge first")
	}

	lights, err := c.bridge.GetLights()
	if err != nil {
		return fmt.Errorf("failed to get lights: %w", err)
	}

	for _, light := range lights {
		if err := light.Off(); err != nil {
			return fmt.Errorf("failed to turn off light %d: %w", light.ID, err)
		}
	}

	return nil
}

// convertLight converts a huego.Light to our Light type
func (c *client) convertLight(hueLight *huego.Light) *Light {
	light := &Light{
		ID:         hueLight.ID,
		Name:       hueLight.Name,
		On:         hueLight.State.On,
		Brightness: uint8(hueLight.State.Bri),
		ColorMode:  hueLight.State.ColorMode,
		Reachable:  hueLight.State.Reachable,
	}

	if len(hueLight.State.Xy) >= 2 {
		light.ColorXY = [2]float32{float32(hueLight.State.Xy[0]), float32(hueLight.State.Xy[1])}
	}

	if hueLight.State.Ct > 0 {
		light.Temperature = uint16(hueLight.State.Ct)
	}

	// Convert XY to RGB if available
	if len(hueLight.State.Xy) >= 2 {
		r, g, b := xyToRGB(float64(hueLight.State.Xy[0]), float64(hueLight.State.Xy[1]), uint8(hueLight.State.Bri))
		light.ColorRGB = [3]uint8{r, g, b}
	}

	return light
}

// rgbToXY converts RGB values (0-255) to CIE XY coordinates
// Uses the standard sRGB to CIE 1931 color space conversion
func rgbToXY(r, g, b float64) (x, y float64) {
	// Normalize RGB values to 0-1
	rNorm := r / 255.0
	gNorm := g / 255.0
	bNorm := b / 255.0

	// Apply gamma correction
	var rLin, gLin, bLin float64
	if rNorm > 0.04045 {
		rLin = math.Pow((rNorm+0.055)/1.055, 2.4)
	} else {
		rLin = rNorm / 12.92
	}
	if gNorm > 0.04045 {
		gLin = math.Pow((gNorm+0.055)/1.055, 2.4)
	} else {
		gLin = gNorm / 12.92
	}
	if bNorm > 0.04045 {
		bLin = math.Pow((bNorm+0.055)/1.055, 2.4)
	} else {
		bLin = bNorm / 12.92
	}

	// Convert to XYZ color space (using sRGB matrix)
	xVal := rLin*0.4124 + gLin*0.3576 + bLin*0.1805
	yVal := rLin*0.2126 + gLin*0.7152 + bLin*0.0722
	zVal := rLin*0.0193 + gLin*0.1192 + bLin*0.9505

	// Convert XYZ to xy
	sum := xVal + yVal + zVal
	if sum == 0 {
		return 0.3127, 0.3290 // Default to D65 white point
	}

	x = xVal / sum
	y = yVal / sum

	return x, y
}

// xyToRGB converts CIE XY coordinates to RGB
// This is a simplified conversion - for production use, consider a more accurate algorithm
func xyToRGB(x, y float64, brightness uint8) (r, g, b uint8) {
	// Simplified conversion - in production, use proper color space conversion
	// This is a basic approximation
	rVal := (x * float64(brightness)) / 255.0
	gVal := (y * float64(brightness)) / 255.0
	bVal := ((1.0 - x - y) * float64(brightness)) / 255.0

	// Clamp values to valid uint8 range
	if rVal > 255 {
		rVal = 255
	}
	if gVal > 255 {
		gVal = 255
	}
	if bVal > 255 {
		bVal = 255
	}
	if rVal < 0 {
		rVal = 0
	}
	if gVal < 0 {
		gVal = 0
	}
	if bVal < 0 {
		bVal = 0
	}

	r = uint8(rVal)
	g = uint8(gVal)
	b = uint8(bVal)

	return r, g, b
}
