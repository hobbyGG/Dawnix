package domain

type Execution struct {
	BaseModel

	InstID   int64
	ParentID int64
	NodeID   string
	IsActive bool
}
