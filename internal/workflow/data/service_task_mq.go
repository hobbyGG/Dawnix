package data

import (
	"context"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/workflow/biz"
)

type serviceTaskMQImpl struct {
	mq biz.MQ
}

func NewServiceTaskMQ(mq biz.MQ) biz.ServiceTaskMQ {
	return &serviceTaskMQImpl{mq: mq}
}

func (stmq *serviceTaskMQImpl) ProduceEmailTask(ctx context.Context, emailTaskJSON []byte) error {
	if _, err := stmq.mq.Produce(ctx, biz.EmailTaskTopic, biz.EmailTaskKey, emailTaskJSON); err != nil {
		return fmt.Errorf("produce email task failed: %w", err)
	}
	return nil
}
