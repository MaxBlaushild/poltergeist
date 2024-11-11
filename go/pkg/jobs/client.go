package jobs

import (
	"context"

	"github.com/hibiken/asynq"
)

type Client interface {
	QueueJob(ctx context.Context, job Job) error
}

type client struct {
	async *asynq.Client
}

type Job struct {
	Type    string
	Payload []byte
}

func NewClient(redisUrl string) Client {
	async := asynq.NewClient(asynq.RedisClientOpt{Addr: redisUrl})
	defer async.Close()
	return &client{async: async}
}

func (c *client) QueueJob(ctx context.Context, job Job) error {
	if _, err := c.async.Enqueue(asynq.NewTask(job.Type, job.Payload)); err != nil {
		return err
	}
	return nil
}
