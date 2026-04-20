package model

import (
	"gorm.io/datatypes"

	"github.com/hobbyGG/Dawnix/internal/workflow/domain"
)

type ProcessDefinition struct {
	BaseModel

	Code           string         `gorm:"type:varchar(64);not null;uniqueIndex:idx_code_ver,where:deleted_at IS NULL;comment:流程标识" json:"code"`
	Version        int            `gorm:"default:1;uniqueIndex:idx_code_ver,where:deleted_at IS NULL;comment:版本号" json:"version"`
	Name           string         `gorm:"type:varchar(128);not null" json:"name"`
	Structure      datatypes.JSON `gorm:"type:jsonb;not null" json:"structure"`
	FormDefinition datatypes.JSON `gorm:"type:jsonb;column:form_definition" json:"form_definition"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
}

func (ProcessDefinition) TableName() string {
	return "process_definitions"
}

func (p *ProcessDefinition) ToDomain() *domain.ProcessDefinition {
	if p == nil {
		return nil
	}
	return &domain.ProcessDefinition{
		BaseModel: domain.BaseModel{
			ID:        p.ID,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
			CreatedBy: p.CreatedBy,
			UpdatedBy: p.UpdatedBy,
		},
		Code:           p.Code,
		Version:        p.Version,
		Name:           p.Name,
		Structure:      p.Structure,
		FormDefinition: p.FormDefinition,
		IsActive:       p.IsActive,
	}
}
