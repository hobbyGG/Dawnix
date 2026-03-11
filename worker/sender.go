package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hobbyGG/Dawnix/client"
	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/hobbyGG/Dawnix/internal/biz/model"
	"github.com/hobbyGG/Dawnix/util"
	"go.uber.org/zap"
)

const (
	emailTaskTopic = "email_tasks"
	emailTaskGroup = "email_task_group"
)

type EmailSendWorker struct {
	eCli       client.EmailSender
	mq         biz.MQ
	consumerID string

	// 使用 context 的 cancel 函数来代替 signal channel
	cancel context.CancelFunc
}

func NewEmailSender(emailCli client.EmailSender, mq biz.MQ) *EmailSendWorker {
	id := util.Generator.Generate().String()
	return &EmailSendWorker{
		eCli:       emailCli,
		mq:         mq,
		consumerID: id,
	}
}

func (s *EmailSendWorker) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	handler := func(key string, value []byte) error {
		emailTask := model.EmailNodeParmas{}
		if err := json.Unmarshal(value, &emailTask); err != nil {
			zap.L().Error("unmarshal failed, discarding poison message", zap.Error(err), zap.ByteString("val", value))
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
			s.cleanup()
			return nil
		default:
		}

		// 这样当调用 Stop() 触发 ctx 取消时，底层的 MQ 库会自动感知并打断阻塞，返回 context.Canceled 错误
		err := s.mq.Consume(ctx, emailTaskTopic, emailTaskGroup, s.consumerID, handler)
		if err != nil {
			// 如果是主动取消导致的退出，直接结束
			if err == context.Canceled || err == context.DeadlineExceeded {
				zap.L().Info("consumer context canceled")
				s.cleanup()
				return nil
			}

			// 避免底层报错时疯狂重试打爆 CPU，加一个简单的休眠补偿
			zap.L().Error("consume error", zap.Error(err))
			time.Sleep(1 * time.Second)
		}
	}
}

func (s *EmailSendWorker) Stop() {
	if s.cancel != nil {
		s.cancel() // 调用 cancel() 不会阻塞，且多次调用也是安全的
		zap.L().Info("email send worker stop signal triggered")
	}
}

func (s *EmailSendWorker) cleanup() {
	zap.L().Info("email send worker cleanup...")
	// 清理资源...
}
