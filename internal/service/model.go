package service

import "github.com/hobbyGG/Dawnix/internal/biz/model"

type InstanceDetail struct {
	Inst       *model.ProcessInstance `json:"inst"`
	Executions []model.Execution      `json:"executions"`
}
