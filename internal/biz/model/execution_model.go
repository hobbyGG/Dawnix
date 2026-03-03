package model

type Execution struct {
	BaseModel

	InstID int64 `gorm:"column:inst_id;type:bigint;not null;index:idx_inst_id"`
	
}