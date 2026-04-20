package biz

import "context"

const (
	EmailTaskTopic = "email_tasks"
	EmailTaskGroup = "email_task_group"
	EmailTaskKey   = "email_info"
)

// MQ 定义消息队列端口。
// biz 层仅依赖行为，不承载具体中间件实现。
type MQ interface {
	Produce(ctx context.Context, topic string, key string, value []byte) (msgID string, err error)
	Consume(ctx context.Context, topic string, groupID, consumerID string, handler func(key string, value []byte) error) error
}

// ServiceTaskMQ 定义服务任务消息生产端口。
type ServiceTaskMQ interface {
	ProduceEmailTask(ctx context.Context, emailTaskJSON []byte) error
}
