package model

import domain "github.com/hobbyGG/Dawnix/internal/domain"

type ProcessTask struct {
	BaseModel

	InstanceID  int64    `gorm:"index;not null"`
	ExecutionID int64    `gorm:"index"`
	NodeID      string   `gorm:"type:varchar(64);index"`
	Type        string   `gorm:"type:varchar(32);default:'user_task'"`
	Assignee    string   `gorm:"type:varchar(64);index"`
	Candidates  []string `gorm:"type:json"`
	Status      string   `gorm:"type:varchar(32);default:'PENDING';index"`
	Action      string   `gorm:"type:varchar(32)"`
	Comment     string   `gorm:"type:text"`
}

func (ProcessTask) TableName() string {
	return "process_tasks"
}

func (p *ProcessTask) ToDomain() *domain.ProcessTask {
	if p == nil {
		return nil
	}
	return &domain.ProcessTask{
		BaseModel: domain.BaseModel{
			ID:        p.ID,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
			CreatedBy: p.CreatedBy,
			UpdatedBy: p.UpdatedBy,
		},
		InstanceID:  p.InstanceID,
		ExecutionID: p.ExecutionID,
		NodeID:      p.NodeID,
		Type:        p.Type,
		Assignee:    p.Assignee,
		Candidates:  p.Candidates,
		Status:      p.Status,
		Action:      p.Action,
		Comment:     p.Comment,
	}
}
