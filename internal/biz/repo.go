package biz

import "context"

// TransactionManager 事务管理器接口
// 它并不依赖 GORM，只依赖 context
type TransactionManager interface {
	// InTx 开启一个事务，并在事务中执行 fn
	// fn 接收的 ctx 内部已经包含了事务句柄
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}
