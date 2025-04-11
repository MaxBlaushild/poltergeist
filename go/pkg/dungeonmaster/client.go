package dungeonmaster

import (
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

type client struct {
	googlemapsClient googlemaps.Client
	dbClient         db.Client
}

type Client interface {
	GenerateQuests() ([]*models.PointOfInterest, error)
}

func NewClient(googlemapsClient googlemaps.Client, dbClient db.Client) Client {
	return &client{
		googlemapsClient: googlemapsClient,
		dbClient:         dbClient,
	}
}

func (c *client) GenerateQuests() ([]*models.PointOfInterest, error) {

	return nil, nil
}
