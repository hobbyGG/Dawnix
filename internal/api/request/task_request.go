package request

import "github.com/hobbyGG/Dawnix/internal/biz"

type GetTaskDetailReq struct {
	ID int64 `uri:"id" binding:"required"`
}

type ListTasksReq struct {
	// 分页参数
	Page  int    `form:"page" binding:"omitempty,min=1"`
	Size  int    `form:"size" binding:"omitempty,min=1,max=100"`
	Scope string `form:"scope" binding:"omitempty"` // 列表页范围：my_pending, my_completed, all_pending, all_completed...
}

func (req *ListTasksReq) ToBizParams() *biz.ListTasksParams {
	return &biz.ListTasksParams{
		Page:  req.Page,
		Size:  req.Size,
		Scope: req.Scope,
	}
}

type CompleteTaskReq struct {
	// 任务 ID (路径参数)
	ID int64 `uri:"id"`

	// 动作: "agree", "reject", "transfer"
	Action string `json:"action" binding:"required"`

	// 审批意见
	Comment string `json:"comment"`

	// 变量: 比如请假表单里的实际数据，或者审批人填写的新字段
	Variables map[string]interface{} `json:"variables"`

	// 当前操作人 (Middleware 注入)
	CurrentUserID int64 `json:"-"`
}

func (req *CompleteTaskReq) ToBizParams() *biz.CompleteTaskParams {
	return &biz.CompleteTaskParams{
		TaskID:  req.ID,
		UserID:  req.CurrentUserID,
		Action:  req.Action,
		Comment: req.Comment,
	}
}
