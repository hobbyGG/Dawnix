package biz

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type MQ interface {
	Produce(ctx context.Context, topic string, key string, value []byte) (msgID string, err error)
	Consume(ctx context.Context, topic string, groupID, consumerID string, handler func(key string, value []byte) error) error
}

var _ MQ = (*redisMQ)(nil)

type redisMQ struct {
	rdb *redis.Client
}

func NewRedisMQ(rdb *redis.Client) MQ {
	return &redisMQ{
		rdb: rdb,
	}
}

// Produce 使用 Stream 生产一条消息到指定topic
// 生产者需要考虑：
// 可靠性 - 生产消息时可能会失败，调用方需要知道是否成功了，以便重试
func (rmq *redisMQ) Produce(ctx context.Context, topic string, key string, value []byte) (string, error) {
	xArg := &redis.XAddArgs{
		Stream: topic,
		Values: map[string]interface{}{
			"key":   key,
			"value": value,
		},
	}
	retryCount := 3

	var err error
	var msgID string
	for range retryCount {
		msgID, err = rmq.rdb.XAdd(ctx, xArg).Result()
		if err == nil {
			return msgID, nil
		}
		// 失败等待20ms
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("produce context cancelled: %w", ctx.Err())
		case <-time.After(time.Millisecond * 20):
			// 正常休眠 20ms 后继续下一次 for 循环重试
		}
	}
	// 重试失败全部失败
	zap.L().Error("redisMQ produce failed", zap.Error(err))
	return "", err
}

// Consume 使用 Stream 消息队列消费消息
// 消费者需要考虑：
// 幂等性 - 消息可能会被重复消费，处理逻辑需要保证幂等
// 可靠性 - 消息处理必须 ack
func (rmq *redisMQ) Consume(
	ctx context.Context,
	topic string,
	groupID string,
	consumerID string,
	handler func(key string, value []byte) error,
) error {
	// 确保 Stream 和消费者组存在
	err := rmq.rdb.XGroupCreateMkStream(ctx, topic, groupID, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}

	readArg := &redis.XReadGroupArgs{
		Group:    groupID,
		Consumer: consumerID,
		Streams:  []string{topic, ">"},
		Block:    0, // 0 表示永久阻塞，直到有消息或 ctx 被取消
	}

	for {
		// 每次发起网络请求前，检查 ctx 是否已经结束
		if err := ctx.Err(); err != nil {
			zap.L().Info("redisMQ consumer stopped due to context done", zap.String("topic", topic))
			return err
		}

		streams, err := rmq.rdb.XReadGroup(ctx, readArg).Result()

		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}

			// 捕获 ctx 取消错误
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				zap.L().Info("redisMQ consume canceled", zap.String("topic", topic))
				return err // 打断循环，优雅退出
			}

			// 其他真实的 Redis 报错
			zap.L().Error("redisMQ consume XReadGroup failed",
				zap.String("topic", topic),
				zap.Error(err))

			// 发生未知异常时，稍微休眠一下防抖，避免瞬间打爆 CPU
			time.Sleep(500 * time.Millisecond)
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				key, _ := msg.Values["key"].(string)

				var value []byte
				switch v := msg.Values["value"].(type) {
				case string:
					value = []byte(v)
				case []byte:
					value = v
				}

				if err := handler(key, value); err != nil {
					zap.L().Error("redisMQ consume handler failed",
						zap.String("topic", topic),
						zap.String("msgID", msg.ID),
						zap.Error(err),
					)

					// 进入死信队列
					dlpTopic := topic + ":dlq"
					fallbackCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					dlqArgs := &redis.XAddArgs{
						Stream: dlpTopic,
						MaxLen: 100000, // 给 DLQ 也要设个上限，防止积压撑爆内存
						Approx: true,   // 控制在大约这么多
						Values: map[string]interface{}{
							"original_msg_id": msg.ID,
							"error_reason":    err.Error(),
							"key":             key,
							"value":           value,
						},
					}
					if err := rmq.rdb.XAdd(fallbackCtx, dlqArgs).Err(); err != nil {
						// 如果加入死信队列失败则放在pending里，等待兜底重试
						zap.L().Error("failed to push message to DLQ", zap.Error(err))
						cancel()
					}
					if ackErr := rmq.rdb.XAck(fallbackCtx, topic, groupID, msg.ID).Err(); ackErr != nil {
						zap.L().Error("failed to ACK original message after moving to DLQ", zap.Error(ackErr))
					}

					cancel()
					continue
				}

				if err := rmq.rdb.XAck(ctx, topic, groupID, msg.ID).Err(); err != nil {
					zap.L().Error("redisMQ XAck failed",
						zap.String("topic", topic),
						zap.String("msgID", msg.ID),
						zap.Error(err),
					)
				}
			}
		}
	}
}

type ServiceTaskMQ interface {
	ProduceEmailTask(ctx context.Context, emailTaskJson []byte) error
}

type ServiceTaskMQImpl struct {
	// 需要嵌入mq
	mq MQ
}

func NewServiceTaskMQImpl(mq MQ) ServiceTaskMQ {
	return &ServiceTaskMQImpl{
		mq: mq,
	}
}

func (stmq *ServiceTaskMQImpl) ProduceEmailTask(ctx context.Context, emailTaskJson []byte) error {
	topic := "email_tasks"
	key := "email_info" // 消息的key

	_, err := stmq.mq.Produce(ctx, topic, key, emailTaskJson)
	if err != nil {
		return err
	}
	return nil
}

type EmailTask struct {
	// 发送邮件所需的字段
	InstanceID    int64  // 记录哪个流程发送的邮件
	NodeID        string // 记录哪个节点发送的邮件
	ReceiverEmail string // 收件人
	Subject       string // 邮件主题
	Content       string // 邮件内容
}
