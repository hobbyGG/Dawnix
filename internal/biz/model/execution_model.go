package model

type Execution struct {
	BaseModel

	InstID   int64  `gorm:"column:inst_id;type:bigint;not null;index:idx_inst_id"`
	ParentID int64  `gorm:"column:parent_id;type:bigint;not null;default:0"` // 父execution ID，支持子流程调用
	NodeID   string `gorm:"column:node_id;type:varchar(64);not null"`

	IsActive bool `gorm:"column:is_active;type:boolean;not null;default:true"` // 是否处于活动状态，子流程调用时父流程execution会被置为非活动
}
