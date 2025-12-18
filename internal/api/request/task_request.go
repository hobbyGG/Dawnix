package request

type GetTaskDetailReq struct {
	ID int64 `uri:"id" binding:"required"`
}

type CompleteTaskReq struct {
	// 任务 ID (路径参数)
	ID int64 `uri:"id" binding:"required" json:"id"`

	// 动作: "agree", "reject", "transfer"
	Action string `json:"action" binding:"required"`

	// 审批意见
	Comment string `json:"comment"`

	// 变量: 比如请假表单里的实际数据，或者审批人填写的新字段
	Variables map[string]interface{} `json:"variables"`

	// 当前操作人 (Middleware 注入)
	CurrentUserID string `json:"-"`
}
