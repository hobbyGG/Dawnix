package domain

import (
	"time"

	"gorm.io/datatypes"
)

const (
	InstanceStatusPending   = "PENDING"
	InstanceStatusApproved  = "APPROVED"
	InstanceStatusRejected  = "REJECTED"
	InstanceStatusCanceled  = "CANCELED"
	InstanceStatusSuspended = "SUSPENDED"
)

type ProcessInstance struct {
	BaseModel

	DefinitionID      int64
	ProcessCode       string
	SnapshotStructure datatypes.JSON
	ParentID          int64
	ParentNodeID      string
	FormData          datatypes.JSON
	Status            string
	SubmitterID       string
	FinishedAt        *time.Time
}
