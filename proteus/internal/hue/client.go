package hue

import (
	"github.com/amimof/huego"
)

type client struct {
	bridge *huego.Bridge
}

const (
	userID = "cbVioJMmCtp-n9-LFEOaKv93L0qobJPGl9dYazQ8"
)

func NewClient() (Client, error) {
	bridge, _ := huego.Discover()
	bridge = bridge.Login(userID)

	return &client{bridge: bridge}, nil
}

func (c *client) TurnOnLights() error {
	lights, err := c.bridge.GetLights()
	if err != nil {
		return err
	}

	for _, light := range lights {
		if err := light.On(); err != nil {
			return err
		}
	}

	return nil
}

func (c *client) TurnOffLights() error {
	lights, err := c.bridge.GetLights()
	if err != nil {
		return err
	}

	for _, light := range lights {
		if err := light.Off(); err != nil {
			return err
		}
	}

	return nil
}
