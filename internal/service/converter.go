package service

import (
	"encoding/json"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

// 定义了req->model的转换方法

func paramsToProcessDef(params *biz.ProcessDefinitionCreateParams) (*model.ProcessDefinition, error) {
	strctureJson, err := json.Marshal(params.Structure)
	if err != nil {
		return nil, fmt.Errorf("fail to marshal structure: %w", err)
	}

	return &model.ProcessDefinition{
		Code:      params.Code,
		Name:      params.Name,
		Structure: strctureJson,
	}, nil
}
