package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hobbyGG/Dawnix/client"
	"github.com/hobbyGG/Dawnix/internal/workflow/biz"
	"github.com/hobbyGG/Dawnix/internal/workflow/domain"
	"github.com/hobbyGG/Dawnix/util"
	"go.uber.org/zap"
)

type EmailSendWorker struct {
	eCli       client.EmailSender
	mq         biz.MQ
	consumerID string

	mu      sync.Mutex
	running bool
	done    chan struct{}

	// 使用 context 的 cancel 函数来代替 signal channel
	cancel context.CancelFunc
}

func NewEmailSender(emailCli client.EmailSender, mq biz.MQ) *EmailSendWorker {
	id := util.Generator.Generate().String()
	return &EmailSendWorker{
		eCli:       emailCli,
		mq:         mq,
		consumerID: id,
		done:       make(chan struct{}),
	}
}

func (s *EmailSendWorker) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("email send worker already running")
	}
	s.running = true
	s.done = make(chan struct{})
	s.mu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	s.cancel = cancel
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.running = false
		s.cancel = nil
		close(s.done)
		s.mu.Unlock()
	}()

	handler := func(key string, value []byte) error {
		emailTask := domain.EmailNodeParams{}
		if err := json.Unmarshal(value, &emailTask); err != nil {
			zap.L().Error("unmarshal failed, discarding poison message", zap.Error(err), zap.Int("payload_bytes", len(value)))
			return err
		}

		// 给邮件发送操作一个独立的超时时间，这 5 秒内也要把手里这封信发完再走。
		execCtx, execCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer execCancel()

		return s.eCli.Send(execCtx, emailTask.To, emailTask.Subject, emailTask.Body)
	}

	for {
		select {
		case <-ctx.Done():
			zap.L().Info("email send worker received stop signal, exiting loop")
			s.cleanup(ctx)
			return nil
		default:
		}

		// 这样当调用 Stop() 触发 ctx 取消时，底层的 MQ 库会自动感知并打断阻塞，返回 context.Canceled 错误
		err := s.mq.Consume(ctx, biz.EmailTaskTopic, biz.EmailTaskGroup, s.consumerID, handler)
		if err != nil {
			// 如果是主动取消导致的退出，直接结束
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				zap.L().Info("consumer context canceled")
				s.cleanup(ctx)
				return nil
			}

			// 避免底层报错时疯狂重试打爆 CPU，加一个简单的休眠补偿
			zap.L().Error("consume error", zap.Error(err))
			select {
			case <-ctx.Done():
				zap.L().Info("email send worker context canceled during backoff")
				s.cleanup(ctx)
				return nil
			case <-time.After(1 * time.Second):
			}
		}
	}
}

func (s *EmailSendWorker) Stop() {
	s.mu.Lock()
	cancel := s.cancel
	done := s.done
	running := s.running
	s.mu.Unlock()

	if cancel != nil {
		cancel() // 调用 cancel() 不会阻塞，且多次调用也是安全的
	}
	if !running {
		return
	}

	select {
	case <-done:
		zap.L().Info("email send worker stopped")
	case <-time.After(3 * time.Second):
		zap.L().Warn("email send worker stop timed out")
	}
}

func (s *EmailSendWorker) cleanup(ctx context.Context) {
	zap.L().Info("email send worker cleanup...")
	// 清理资源...
	_ = ctx
}
