package request

import (
	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

type ProcessDefinitionCreateReq struct {
	Code      string                 `json:"code"`                         // 流程模板业务号，用于创建流程
	Name      string                 `json:"name" binding:"required"`      // 流程模板名称
	Structure model.ProcessStructure `json:"structure" binding:"required"` // 流程模板图结构
	Config    model.ProcessConfig    `json:"config"`                       // 流程全局配置，例如该流程结束后处理配置等
}

func (r *ProcessDefinitionCreateReq) ToBizParams() *biz.ProcessDefinitionCreateParams {
	return &biz.ProcessDefinitionCreateParams{
		Name:      r.Name,
		Code:      r.Code,
		Structure: r.Structure,
		Config:    r.Config,
	}
}

type ProcessDefinitionListReq struct {
	Page int `json:"page" binding:"required,min=1"` // 页码，从1开始
	Size int `json:"size" binding:"required,min=1"` // 每页大小
}

func (r *ProcessDefinitionListReq) ToBizParams() *biz.ProcessDefinitionListParams {
	return &biz.ProcessDefinitionListParams{
		Page: r.Page,
		Size: r.Size,
	}
}

type ProcessDefinitionDetailReq struct {
	ID int64 `uri:"id" binding:"required,min=1"` // 流程模板ID
}

type ProcessDefinitionDeleteReq struct {
	ID int64 `uri:"id" binding:"required,min=1"` // 流程模板ID
}
