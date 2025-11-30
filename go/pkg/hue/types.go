package hue

// Bridge represents a Philips Hue bridge
type Bridge struct {
	Hostname string
	Username string
}

// Light represents a Philips Hue light
type Light struct {
	ID          int
	Name        string
	On          bool
	Brightness  uint8
	ColorMode   string
	ColorXY     [2]float32
	ColorRGB    [3]uint8
	Temperature uint16
	Reachable   bool
}

// LightState represents a complete light state configuration
type LightState struct {
	On               *bool
	Brightness       *uint8
	ColorRGB         *[3]uint8
	ColorXY          *[2]float32
	ColorTemperature *uint16
	TransitionTime   *uint16 // in deciseconds (1/10 seconds)
}

// Color represents a color in different formats
type Color struct {
	RGB *[3]uint8
	XY  *[2]float32
	CT  *uint16 // Color temperature in mireds
}
