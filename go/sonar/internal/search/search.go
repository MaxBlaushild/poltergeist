package search

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

const (
	findRelevantTagsPrompt = `
	You are a helpful assistant that finds tags that may be relevant for what a user wants to do for an activity.
	The tags apply to certain types of places that the user would want to go to depending on the activity.

	This is what the user wants to do:
	%s

	Here are the tags that apply to certain types of places that the user would want to go to depending on the activity:
	%v

	You will need to find the most relevant tags for the query.
	You will return a list of tags that are most relevant to the query.

	Here is the json format you will return the tags in:
	{
		"tags": ["tag1", "tag2", "tag3"]
	}
	`
)

type tagResponse struct {
	Tags []string `json:"tags"`
}

type SearchClient interface {
	FindRelevantTags(ctx context.Context, query string) ([]*models.Tag, error)
}

type searchClient struct {
	dbClient   db.DbClient
	deepPriest deep_priest.DeepPriest
}

func NewSearchClient(dbClient db.DbClient, deepPriest deep_priest.DeepPriest) SearchClient {
	return &searchClient{
		dbClient:   dbClient,
		deepPriest: deepPriest,
	}
}

func (c *searchClient) FindRelevantTags(ctx context.Context, query string) ([]*models.Tag, error) {
	tags, err := c.dbClient.Tag().FindAll(ctx)
	if err != nil {
		return nil, err
	}

	tagValues := make([]string, len(tags))
	for i, tag := range tags {
		tagValues[i] = tag.Value
	}
	prompt := fmt.Sprintf(findRelevantTagsPrompt, query, tagValues)
	response, err := c.deepPriest.PetitionTheFount(&deep_priest.Question{
		Question: prompt,
	})
	if err != nil {
		return nil, err
	}

	var tagResponse tagResponse
	if err := json.Unmarshal([]byte(response.Answer), &tagResponse); err != nil {
		return nil, err
	}

	relevantTags := []*models.Tag{}
	for _, tag := range tagResponse.Tags {
		for _, t := range tags {
			if t.Value == tag {
				relevantTags = append(relevantTags, t)
				break
			}
		}
	}

	return relevantTags, nil
}
