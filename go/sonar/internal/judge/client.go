package judge

import (
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
)

type Client interface {
	JudgeSubmission(question string, imageSubmission []byte, textSubmission string) (string, error)
}

type client struct {
	aws aws.AWSClient
	db  db.DbClient
}

func NewClient(aws aws.AWSClient, db db.DbClient) Client {
	return &client{
		aws: aws,
		db:  db,
	}
}

func (c *client) JudgeImageSubmission(question string, image []byte) (string, error) {
	return "", nil
}
