package request

import "github.com/hobbyGG/Dawnix/internal/biz"

type CreateInstanceReq struct {
	// 流程标识 (必填)
	// 前端只传 Code，后端负责查 Definition 表找最新版
	ProcessCode string `json:"process_code" binding:"required"`

	// 发起人 ID (必填)
	SubmitterID string `json:"submitter_id" binding:"required"`

	// 业务表单数据 (可选)
	Variables map[string]interface{} `json:"variables"`

	// 父流程相关 (可选，用于子流程场景)
	ParentID     int64  `json:"parent_id"`
	ParentNodeID string `json:"parent_node_id"`
}

func (r *CreateInstanceReq) ToBizCmd() biz.StartProcessInstanceCmd {
	return biz.StartProcessInstanceCmd{
		ProcessCode:  r.ProcessCode,
		SubmitterID:  r.SubmitterID,
		Variables:    r.Variables,
		ParentID:     r.ParentID,
		ParentNodeID: r.ParentNodeID,
	}
}

type ListInstancesReq struct {
	Page int `json:"page" binding:"required,min=1"`
	Size int `json:"size" binding:"required,min=1"`
}

func (r *ListInstancesReq) ToBizParams() *biz.ListInstancesParams {
	return &biz.ListInstancesParams{
		Page: r.Page,
		Size: r.Size,
	}
}

type GetInstanceDetailReq struct {
	ID int64 `uri:"id" binding:"required"`
}
