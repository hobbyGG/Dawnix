package workflow

import (
	"encoding/json"
	"time"
)

type ProcessDefinitionListItem struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Version   int       `json:"version"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type ProcessDefinitionListReply struct {
	Total int64                       `json:"total"`
	List  []ProcessDefinitionListItem `json:"list"`
}

type ProcessDefinitionDetailReply struct {
	ID             int64           `json:"id"`
	Code           string          `json:"code"`
	Version        int             `json:"version"`
	Name           string          `json:"name"`
	Structure      json.RawMessage `json:"structure"`
	FormDefinition json.RawMessage `json:"form_definition"`
	IsActive       bool            `json:"is_active"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	CreatedBy      string          `json:"created_by"`
	UpdatedBy      string          `json:"updated_by"`
}

type InstanceListItem struct {
	ID          int64      `json:"id"`
	ProcessCode string     `json:"process_code"`
	ProcessName string     `json:"process_name"`
	Status      string     `json:"status"`
	SubmitterID string     `json:"submitter_id"`
	CreatedAt   time.Time  `json:"created_at"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
}

type InstanceListReply struct {
	Total int64              `json:"total"`
	List  []InstanceListItem `json:"list"`
}

type InstanceDetailItem struct {
	ID                int64           `json:"id"`
	DefinitionID      int64           `json:"definition_id"`
	ProcessCode       string          `json:"process_code"`
	SnapshotStructure json.RawMessage `json:"snapshot_structure"`
	ParentID          int64           `json:"parent_id"`
	ParentNodeID      string          `json:"parent_node_id"`
	FormData          json.RawMessage `json:"form_data"`
	Status            string          `json:"status"`
	SubmitterID       string          `json:"submitter_id"`
	FinishedAt        *time.Time      `json:"finished_at,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	CreatedBy         string          `json:"created_by"`
	UpdatedBy         string          `json:"updated_by"`
}

type ExecutionReply struct {
	ID        int64     `json:"id"`
	InstID    int64     `json:"inst_id"`
	ParentID  int64     `json:"parent_id"`
	NodeID    string    `json:"node_id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by"`
	UpdatedBy string    `json:"updated_by"`
}

type InstanceDetailReply struct {
	Instance   InstanceDetailItem `json:"instance"`
	Executions []ExecutionReply   `json:"executions"`
}

type TaskViewReply struct {
	ID            int64     `json:"id"`
	TaskName      string    `json:"task_name"`
	Status        string    `json:"status"`
	ProcessTitle  string    `json:"process_title"`
	SubmitterName string    `json:"submitter_name"`
	ArrivedAt     time.Time `json:"arrived_at"`
}

type TaskDetailReply struct {
	ID           int64           `json:"id"`
	InstanceID   int64           `json:"instance_id"`
	ExecutionID  int64           `json:"execution_id"`
	NodeID       string          `json:"node_id"`
	Type         string          `json:"type"`
	Assignee     string          `json:"assignee"`
	Candidates   []string        `json:"candidates"`
	Status       string          `json:"status"`
	Action       string          `json:"action"`
	Comment      string          `json:"comment"`
	FormData     json.RawMessage `json:"form_data"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	CreatedBy    string          `json:"created_by"`
	UpdatedBy    string          `json:"updated_by"`
	ProcessTitle string          `json:"process_title"`
	SubmitterID  string          `json:"submitter_id"`
}

type TaskListReply struct {
	Total int64           `json:"total"`
	List  []TaskViewReply `json:"list"`
}
