package service

import "github.com/hobbyGG/Dawnix/internal/domain"

type InstanceDetail struct {
	Inst       *domain.ProcessInstance `json:"inst"`
	Executions []domain.Execution      `json:"executions"`
}
