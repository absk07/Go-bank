package worker

import (
	"context"

	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/hibiken/asynq"
)

const (
	QueueCrirical = "critical"
	QueueDefault = "default"
)

type TaskProcessor interface {
	Start() error
	Shutdown()
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store *db.Store
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store *db.Store) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCrirical: 10,
				QueueDefault: 5,
			},
		},
	)

	return &RedisTaskProcessor{
		server: server,
		store: store,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)

	return processor.server.Start(mux)
}

func (processor *RedisTaskProcessor) Shutdown() {
	processor.server.Shutdown()
}