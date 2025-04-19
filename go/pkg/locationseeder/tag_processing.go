package locationseeder

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

const unknownTagGroupName = "unknown"

func (c *client) ProccessPlaceTypes(ctx context.Context, placeType []string) ([]*models.Tag, error) {
	unknownTagGroup, err := c.dbClient.TagGroup().FindByName(ctx, unknownTagGroupName)
	if err != nil {
		return nil, err
	}

	tags := []*models.Tag{}
	for _, tag := range placeType {
		existingTag, err := c.dbClient.Tag().FindByValue(ctx, tag)
		if err != nil {
			return nil, err
		}

		if existingTag != nil {
			tags = append(tags, existingTag)
			continue
		}

		newTag := &models.Tag{
			Value:      tag,
			TagGroupID: unknownTagGroup.ID,
			ID:         uuid.New(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		err = c.dbClient.Tag().Create(ctx, newTag)
		if err != nil {
			return nil, err
		}

		tags = append(tags, newTag)
	}

	return tags, nil
}
