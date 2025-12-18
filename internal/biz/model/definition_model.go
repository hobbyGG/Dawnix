package model

import (
	"gorm.io/datatypes"
)

type ProcessDefinition struct {
	BaseModel // 嵌入基础字段

	// Code + Version 必须唯一
	// uniqueIndex:idx_code_ver 定义复合唯一索引名称
	// where:deleted_at IS NULL 是 PG 特有的，保证删除了的记录不占坑
	Code    string `gorm:"type:varchar(64);not null;uniqueIndex:idx_code_ver,where:deleted_at IS NULL;comment:流程标识" json:"code"`
	Version int    `gorm:"default:1;uniqueIndex:idx_code_ver,where:deleted_at IS NULL;comment:版本号" json:"version"`

	Name string `gorm:"type:varchar(128);not null" json:"name"`

	// 流程图结构 (JSONB)
	Structure datatypes.JSON `gorm:"type:jsonb;not null" json:"structure"`

	// 全局配置 (JSONB)
	Config datatypes.JSON `gorm:"type:jsonb" json:"config"`

	IsActive bool `gorm:"default:true" json:"is_active"`
}

func (ProcessDefinition) TableName() string {
	return "process_definitions"
}
