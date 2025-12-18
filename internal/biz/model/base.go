package model

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/optimisticlock"
)

// BaseModel 通用基础结构体
// 包含 ID, 创建/更新时间, 软删除
type BaseModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 软删除: 删除时并不物理删除，而是写入时间。查询时 GORM 会自动过滤掉已删除的。
	// json:"-" 表示前端接口不需要看到这个字段
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 审计
	CreatedBy string `gorm:"type:varchar(64)" json:"created_by"`
	UpdatedBy string `gorm:"type:varchar(64)" json:"updated_by"`

	// 【乐观锁】
	// 使用官方插件类型，GORM 会自动处理 version+1 和 where version=?
	Revision optimisticlock.Version `json:"revision"`
}
