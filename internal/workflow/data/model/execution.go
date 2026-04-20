package model

import domain "github.com/hobbyGG/Dawnix/internal/workflow/domain"

type Execution struct {
	BaseModel

	InstID   int64  `gorm:"column:inst_id;type:bigint;not null;index:idx_inst_id"`
	ParentID int64  `gorm:"column:parent_id;type:bigint;not null;default:0"`
	NodeID   string `gorm:"column:node_id;type:varchar(64);not null"`
	IsActive bool   `gorm:"column:is_active;type:boolean;not null;default:true"`
}

func (p *Execution) ToDomain() *domain.Execution {
	if p == nil {
		return nil
	}
	return &domain.Execution{
		BaseModel: domain.BaseModel{
			ID:        p.ID,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
			CreatedBy: p.CreatedBy,
			UpdatedBy: p.UpdatedBy,
		},
		InstID:   p.InstID,
		ParentID: p.ParentID,
		NodeID:   p.NodeID,
		IsActive: p.IsActive,
	}
}
