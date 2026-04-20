package service

import "github.com/hobbyGG/Dawnix/internal/workflow/domain"

type InstanceDetail struct {
	Inst       *domain.ProcessInstance `json:"inst"`
	Executions []domain.Execution      `json:"executions"`
}
