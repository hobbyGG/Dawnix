package domain

import "time"

const (
	TaskStatusPending     = "PENDING"
	TaskStatusApproved    = "APPROVED"
	TaskStatusRejected    = "REJECTED"
	TaskStatusTransferred = "TRANSFERRED"
	TaskStatusRolledBack  = "ROLLED_BACK"
	TaskStatusCanceled    = "CANCELED"
	TaskStatusAborted     = "ABORTED"
)

const (
	TaskTypeUser    = "user_task"
	TaskTypeService = "service_task"
	TaskTypeReceive = "receive_task"
	TaskTypeCc      = "cc_task"
)

type ProcessTask struct {
	BaseModel

	InstanceID  int64
	ExecutionID int64
	NodeID      string
	Type        string
	Assignee    string
	Candidates  []string
	Status      string
	Action      string
	Comment     string
}

type TaskView struct {
	ID            int64
	TaskName      string
	Status        string
	ProcessTitle  string
	SubmitterName string
	ArrivedAt     time.Time
}
