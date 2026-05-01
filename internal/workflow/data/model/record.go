package model

import (
	"time"

	domain "github.com/hobbyGG/Dawnix/internal/workflow/domain"
)

// RecordPO 数据库表结构
type Record struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	InstanceID   int64     `gorm:"index: not null"`
	TaskID       int64     `gorm:"index: not null"`
	NodeID       string    `gorm:"not null"`
	NodeName     string    `gorm:"not null"`
	ApproverUID  string    `gorm:"not null"`
	ApproverName string    `gorm:"not null"`
	Action       string    `gorm:"not null"`
	Comment      string    `gorm:"size:512"`
	CreatedAt    time.Time `gorm:"default:now()"`
}

func (Record) TableName() string {
	return "approval_record"
}

func (record *Record) ToDomain() *domain.Record {
	if record == nil {
		return nil
	}
	return &domain.Record{
		ID:           record.ID,
		InstanceID:   record.InstanceID,
		TaskID:       record.TaskID,
		NodeID:       record.NodeID,
		NodeName:     record.NodeName,
		ApproverUID:  record.ApproverUID,
		ApproverName: record.ApproverName,
		Action:       record.Action,
		Comment:      record.Comment,
		CreatedAt:    record.CreatedAt,
	}
}
