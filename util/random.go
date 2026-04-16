package util

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/domain"
	"gorm.io/datatypes"
)

func RandomString(prefix string) string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%s_fallback", prefix)
	}
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}

func RandomGraphStartUserEnd() *domain.GraphModel {
	startID := RandomString("start")
	userID := RandomString("user")
	endID := RandomString("end")

	return &domain.GraphModel{
		Nodes: []domain.NodeModel{
			{ID: startID, Type: domain.NodeTypeStart, Name: "Start"},
			{ID: userID, Type: domain.NodeTypeUserTask, Name: "Approve", Candidates: domain.Candidates{Users: []string{RandomString("user")}}},
			{ID: endID, Type: domain.NodeTypeEnd, Name: "End"},
		},
		Edges: []domain.EdgeModel{
			{ID: RandomString("edge"), SourceNode: startID, TargetNode: userID},
			{ID: RandomString("edge"), SourceNode: userID, TargetNode: endID},
		},
	}
}

func RandomProcessDefinition(graph *domain.GraphModel) *domain.ProcessDefinition {
	if graph == nil {
		graph = RandomGraphStartUserEnd()
	}

	return &domain.ProcessDefinition{
		Code:      RandomString("def_code"),
		Version:   1,
		Name:      RandomString("def_name"),
		Structure: MustJSON(graph),
		Config:    datatypes.JSON([]byte("{}")),
		IsActive:  true,
	}
}

func RandomProcessInstance(definitionID int64, processCode string, graph *domain.GraphModel) *domain.ProcessInstance {
	if graph == nil {
		graph = RandomGraphStartUserEnd()
	}
	if processCode == "" {
		processCode = RandomString("inst_code")
	}

	return &domain.ProcessInstance{
		DefinitionID:      definitionID,
		ProcessCode:       processCode,
		SnapshotStructure: MustJSON(graph),
		ParentID:          0,
		ParentNodeID:      "",
		Variables:         datatypes.JSON([]byte(`{"amount":100}`)),
		Status:            domain.InstanceStatusPending,
		SubmitterID:       RandomString("submitter"),
	}
}

func RandomProcessTask(instanceID, executionID int64, nodeID string) *domain.ProcessTask {
	if nodeID == "" {
		nodeID = RandomString("node")
	}

	return &domain.ProcessTask{
		InstanceID:  instanceID,
		ExecutionID: executionID,
		Type:        domain.TaskTypeUser,
		Assignee:    RandomString("assignee"),
		Candidates:  []string{RandomString("user")},
		Status:      domain.TaskStatusPending,
	}
}

func MustJSON(v interface{}) datatypes.JSON {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return datatypes.JSON(b)
}
