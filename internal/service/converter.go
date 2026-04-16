package service

import (
	"encoding/json"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/hobbyGG/Dawnix/internal/domain"
)

// 定义了req->model的转换方法

func paramsToProcessDef(params *biz.ProcessDefinitionCreateParams) (*domain.ProcessDefinition, error) {
	strctureJson, err := json.Marshal(params.Structure)
	if err != nil {
		return nil, fmt.Errorf("fail to marshal structure: %w", err)
	}

	return &domain.ProcessDefinition{
		Code:      params.Code,
		Name:      params.Name,
		Structure: strctureJson,
	}, nil
}
