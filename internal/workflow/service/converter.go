package service

import (
	"encoding/json"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/workflow/biz"
	"github.com/hobbyGG/Dawnix/internal/workflow/domain"
)

// 定义了req->model的转换方法

func paramsToProcessDef(params *biz.ProcessDefinitionCreateParams) (*domain.ProcessDefinition, error) {
	strctureJson, err := json.Marshal(params.Structure)
	if err != nil {
		return nil, fmt.Errorf("fail to marshal structure: %w", err)
	}
	if params.FormDefinition == nil {
		params.FormDefinition = []biz.FormDataItem{}
	}
	formDefinitionJSON, err := json.Marshal(params.FormDefinition)
	if err != nil {
		return nil, fmt.Errorf("fail to marshal form_definition: %w", err)
	}

	return &domain.ProcessDefinition{
		Code:           params.Code,
		Name:           params.Name,
		Structure:      strctureJson,
		FormDefinition: formDefinitionJSON,
	}, nil
}
