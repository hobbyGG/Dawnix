package data

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/biz"
	"go.uber.org/zap"
)

func (repo *HelloRepo) Hello(ctx context.Context, data interface{}) error {
	// 假装存储到数据库
	return nil
}

type HelloRepo struct {
	// gorm连接db
}

func NewHelloRepo(logger *zap.Logger) biz.HelloRepo {
	return &HelloRepo{}
}
