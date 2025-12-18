package data

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/biz"
	"gorm.io/gorm"
)

type Cache interface {
	// 这里定义缓存相关的方法
}

type Data struct {
	db    *gorm.DB
	cache Cache
}

func NewData(db *gorm.DB) (*Data, func(), error) {
	d := &Data{
		db: db,
	}

	cleanup := func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		// if d.rdb != nil { d.rdb.Close() }
	}

	return d, cleanup, nil
}

func (d *Data) DB(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	if ok {
		return tx
	}
	return d.db
}

type transactionManager struct {
	db *gorm.DB
}

// NewTransactionManager 构造函数
func NewTransactionManager(db *gorm.DB) biz.TransactionManager {
	return &transactionManager{db: db}
}

// 定义一个私有的 key，防止外部包乱用
type contextTxKey struct{}

// 实现 InTx
func (tm *transactionManager) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	// 调用 GORM 的 Transaction 方法
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// GORM 开启事务后，会传进来一个 tx 对象 (它也是 *gorm.DB)
		// 我们把这个 tx 塞进 context 里，gorm
		ctxWithTx := context.WithValue(ctx, contextTxKey{}, tx)

		// 执行 Biz 层的闭包，把带 tx 的 context 传进去
		return fn(ctxWithTx)
	})
}
