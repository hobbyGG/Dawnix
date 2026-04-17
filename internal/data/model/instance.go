package model

import (
	"time"

	domain "github.com/hobbyGG/Dawnix/internal/domain"
	"gorm.io/datatypes"
)

type ProcessInstance struct {
	BaseModel

	DefinitionID      int64          `gorm:"index;not null" json:"definition_id"`
	ProcessCode       string         `gorm:"type:varchar(64);index;not null" json:"process_code"`
	SnapshotStructure datatypes.JSON `gorm:"type:jsonb;not null" json:"snapshot_structure"`
	ParentID          int64          `gorm:"index;default:0" json:"parent_id"`
	ParentNodeID      string         `gorm:"type:varchar(64)" json:"parent_node_id"`
	FormData          datatypes.JSON `gorm:"type:jsonb;column:form_data;default:'{}'" json:"form_data"`
	Status            string         `gorm:"type:varchar(32);index;default:'PENDING'" json:"status"`
	SubmitterID       string         `gorm:"type:varchar(64);index" json:"submitter_id"`
	FinishedAt        *time.Time     `json:"finished_at"`
}

func (ProcessInstance) TableName() string {
	return "process_instances"
}

func (p *ProcessInstance) ToDomain() *domain.ProcessInstance {
	if p == nil {
		return nil
	}
	return &domain.ProcessInstance{
		BaseModel: domain.BaseModel{
			ID:        p.ID,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
			CreatedBy: p.CreatedBy,
			UpdatedBy: p.UpdatedBy,
		},
		DefinitionID:      p.DefinitionID,
		ProcessCode:       p.ProcessCode,
		SnapshotStructure: p.SnapshotStructure,
		ParentID:          p.ParentID,
		ParentNodeID:      p.ParentNodeID,
		FormData:          p.FormData,
		Status:            p.Status,
		SubmitterID:       p.SubmitterID,
		FinishedAt:        p.FinishedAt,
	}
}
