package hue

type Client interface {
	TurnOnLights() error
	TurnOffLights() error
}
