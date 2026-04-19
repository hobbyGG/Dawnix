package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/redis/go-redis/v9"
)

var _ biz.MQ = (*redisMQ)(nil)

// ErrorHandler 是消息处理失败时的回调策略，例如写入死信队列。
// 实现者负责记录/转移消息，但不负责 ACK 原消息（由 processMessage 统一 ACK）。
type ErrorHandler func(ctx context.Context, topic string, msg redis.XMessage, handlerErr error) error

// MQOption 是 NewRedisMQ 的函数式选项。
type MQOption func(*redisMQ)

// WithErrorHandler 替换默认的死信队列策略。
func WithErrorHandler(fn ErrorHandler) MQOption {
	return func(r *redisMQ) {
		r.onError = fn
	}
}

type redisMQ struct {
	rdb     *redis.Client
	onError ErrorHandler
}

func NewRedisMQ(rdb *redis.Client, opts ...MQOption) biz.MQ {
	r := &redisMQ{rdb: rdb}
	for _, o := range opts {
		o(r)
	}
	if r.onError == nil {
		r.onError = r.defaultMoveToDLQ
	}
	return r
}

// defaultMoveToDLQ 是默认错误处理策略：通过 Produce（含重试）将失败消息写入 topic:dlq。
func (rmq *redisMQ) defaultMoveToDLQ(ctx context.Context, topic string, msg redis.XMessage, handlerErr error) error {
	key, _ := msg.Values["key"].(string)
	var value []byte
	switch v := msg.Values["value"].(type) {
	case string:
		value = []byte(v)
	case []byte:
		value = v
	}

	dlqPayload, err := marshalDLQPayload(msg.ID, key, value, handlerErr)
	if err != nil {
		return fmt.Errorf("defaultMoveToDLQ marshal failed, msgID=%s: %w", msg.ID, err)
	}

	_, err = rmq.Produce(ctx, topic+":dlq", key, dlqPayload)
	if err != nil {
		return fmt.Errorf("defaultMoveToDLQ produce failed, topic=%s, msgID=%s: %w", topic, msg.ID, err)
	}
	return nil
}

// ensureConsumerGroup 创建 Consumer Group，若已存在则忽略 BUSYGROUP 错误。
func (rmq *redisMQ) ensureConsumerGroup(ctx context.Context, topic, groupID string) error {
	err := rmq.rdb.XGroupCreateMkStream(ctx, topic, groupID, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}
	return nil
}

const dedupTTL = 72 * time.Hour

const (
	consumeBlockTimeout = 2 * time.Second
	reclaimInterval     = 30 * time.Second
	reclaimMinIdle      = 60 * time.Second
	reclaimBatchSize    = 50
)

func dedupSetKey(topic, groupID string, now time.Time) string {
	// 按小时分桶，避免单个 Set 过大。
	return fmt.Sprintf("mq:dedup:%s:%s:%s", topic, groupID, now.UTC().Format("2006010215"))
}

// Produce 使用 Stream 生产一条消息到指定topic。
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

		select {
		case <-ctx.Done():
			return "", fmt.Errorf("produce context cancelled: %w", ctx.Err())
		case <-time.After(20 * time.Millisecond):
		}
	}
	return "", fmt.Errorf("redisMQ produce failed after retries: %w", err)
}

// Consume 使用 Stream 消息队列消费消息。
// 幂等性 - 内置去重，消息可安全重投。
// 可靠性 - 消息处理完毕后统一 ACK；失败消息走 onError 策略（默认: DLQ）。
func (rmq *redisMQ) Consume(
	ctx context.Context,
	topic string,
	groupID string,
	consumerID string,
	handler func(key string, value []byte) error,
) error {
	cs := &consumerSession{
		rmq:        rmq,
		topic:      topic,
		groupID:    groupID,
		consumerID: consumerID,
		handler:    handler,
	}
	return cs.run(ctx)
}

type consumerSession struct {
	rmq        *redisMQ
	topic      string
	groupID    string
	consumerID string
	handler    func(string, []byte) error
}

func (cs *consumerSession) run(ctx context.Context) error {
	if err := cs.rmq.ensureConsumerGroup(ctx, cs.topic, cs.groupID); err != nil {
		return err
	}

	reclaimCtx, cancelReclaim := context.WithCancel(ctx)
	defer cancelReclaim()

	errCh := make(chan error, 1)
	go cs.reclaimLoop(reclaimCtx, errCh)

	return cs.readLoop(ctx, errCh)
}

func (cs *consumerSession) reclaimLoop(ctx context.Context, errCh chan<- error) {
	ticker := time.NewTicker(reclaimInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := cs.reclaimPending(ctx); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				select {
				case errCh <- err:
				default:
				}
				return
			}
		}
	}
}

func (cs *consumerSession) reclaimPending(ctx context.Context) error {
	start := "0-0"
	for {
		claimed, nextStart, err := cs.rmq.rdb.XAutoClaim(ctx, &redis.XAutoClaimArgs{
			Stream:   cs.topic,
			Group:    cs.groupID,
			Consumer: cs.consumerID,
			MinIdle:  reclaimMinIdle,
			Start:    start,
			Count:    reclaimBatchSize,
		}).Result()
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			return fmt.Errorf("redisMQ XAutoClaim failed, topic=%s: %w", cs.topic, err)
		}

		if len(claimed) == 0 {
			return nil
		}

		if err := cs.processMessages(ctx, claimed); err != nil {
			return err
		}

		if nextStart == start || len(claimed) < reclaimBatchSize {
			return nil
		}
		start = nextStart
	}
}

func (cs *consumerSession) readLoop(ctx context.Context, errCh <-chan error) error {
	readArg := &redis.XReadGroupArgs{
		Group:    cs.groupID,
		Consumer: cs.consumerID,
		Streams:  []string{cs.topic, ">"},
		Block:    consumeBlockTimeout,
	}

	for {
		select {
		case err := <-errCh:
			return fmt.Errorf("redisMQ reclaim failed, topic=%s: %w", cs.topic, err)
		default:
		}

		if err := ctx.Err(); err != nil {
			return fmt.Errorf("redisMQ consumer stopped: %w", err)
		}

		streams, err := cs.rmq.rdb.XReadGroup(ctx, readArg).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("redisMQ consume canceled: %w", err)
			}
			return fmt.Errorf("redisMQ XReadGroup failed, topic=%s: %w", cs.topic, err)
		}

		for _, stream := range streams {
			if err := cs.processMessages(ctx, stream.Messages); err != nil {
				return err
			}
		}
	}
}

func (cs *consumerSession) processMessages(ctx context.Context, messages []redis.XMessage) error {
	for _, msg := range messages {
		if err := cs.processMessage(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

func (cs *consumerSession) processMessage(ctx context.Context, msg redis.XMessage) error {
	key, _ := msg.Values["key"].(string)
	var value []byte
	switch v := msg.Values["value"].(type) {
	case string:
		value = []byte(v)
	case []byte:
		value = v
	}

	dedupKey := dedupSetKey(cs.topic, cs.groupID, time.Now())

	seen, err := cs.checkDedup(ctx, dedupKey, msg.ID)
	if err != nil {
		return err
	}
	if seen {
		if err := cs.rmq.rdb.XAck(ctx, cs.topic, cs.groupID, msg.ID).Err(); err != nil {
			return fmt.Errorf("redisMQ XAck duplicate failed, topic=%s, msgID=%s: %w", cs.topic, msg.ID, err)
		}
		return nil
	}

	if handlerErr := cs.handler(key, value); handlerErr != nil {
		if err := cs.rmq.onError(ctx, cs.topic, msg, handlerErr); err != nil {
			return fmt.Errorf("redisMQ onError failed, topic=%s, msgID=%s: %w", cs.topic, msg.ID, err)
		}
		if err := cs.rmq.rdb.XAck(ctx, cs.topic, cs.groupID, msg.ID).Err(); err != nil {
			return fmt.Errorf("redisMQ XAck after onError failed, topic=%s, msgID=%s: %w", cs.topic, msg.ID, err)
		}
		return nil
	}

	return cs.ackWithDedup(ctx, msg.ID, dedupKey)
}

func (cs *consumerSession) checkDedup(ctx context.Context, dedupKey, msgID string) (bool, error) {
	seen, err := cs.rmq.rdb.SIsMember(ctx, dedupKey, msgID).Result()
	if err != nil {
		return false, fmt.Errorf("redisMQ dedup check failed, topic=%s, msgID=%s: %w", cs.topic, msgID, err)
	}
	return seen, nil
}

func (cs *consumerSession) ackWithDedup(ctx context.Context, msgID, dedupKey string) error {
	pipe := cs.rmq.rdb.TxPipeline()
	pipe.SAdd(ctx, dedupKey, msgID)
	pipe.Expire(ctx, dedupKey, dedupTTL)
	pipe.XAck(ctx, cs.topic, cs.groupID, msgID)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redisMQ ackWithDedup failed, topic=%s, msgID=%s: %w", cs.topic, msgID, err)
	}
	return nil
}

type dlqPayload struct {
	OriginalMsgID string `json:"original_msg_id"`
	ErrorReason   string `json:"error_reason"`
	Key           string `json:"key"`
	Value         []byte `json:"value"`
}

func marshalDLQPayload(msgID, key string, value []byte, handlerErr error) ([]byte, error) {
	return json.Marshal(dlqPayload{
		OriginalMsgID: msgID,
		ErrorReason:   handlerErr.Error(),
		Key:           key,
		Value:         value,
	})
}
