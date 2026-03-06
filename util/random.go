package util

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/biz/model"
	"gorm.io/datatypes"
)

func RandomString(prefix string) string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%s_fallback", prefix)
	}
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}

func RandomGraphStartUserEnd() *model.GraphModel {
	startID := RandomString("start")
	userID := RandomString("user")
	endID := RandomString("end")

	return &model.GraphModel{
		Nodes: []model.NodeModel{
			{ID: startID, Type: model.NodeTypeStart, Name: "Start"},
			{ID: userID, Type: model.NodeTypeUserTask, Name: "Approve", Candidates: model.Candidates{Users: []string{RandomString("user")}}},
			{ID: endID, Type: model.NodeTypeEnd, Name: "End"},
		},
		Edges: []model.EdgeModel{
			{ID: RandomString("edge"), SourceNode: startID, TargetNode: userID},
			{ID: RandomString("edge"), SourceNode: userID, TargetNode: endID},
		},
	}
}

func RandomProcessDefinition(graph *model.GraphModel) *model.ProcessDefinition {
	if graph == nil {
		graph = RandomGraphStartUserEnd()
	}

	return &model.ProcessDefinition{
		Code:      RandomString("def_code"),
		Version:   1,
		Name:      RandomString("def_name"),
		Structure: MustJSON(graph),
		Config:    datatypes.JSON([]byte("{}")),
		IsActive:  true,
	}
}

func RandomProcessInstance(definitionID int64, processCode string, graph *model.GraphModel) *model.ProcessInstance {
	if graph == nil {
		graph = RandomGraphStartUserEnd()
	}
	if processCode == "" {
		processCode = RandomString("inst_code")
	}

	return &model.ProcessInstance{
		DefinitionID:      definitionID,
		ProcessCode:       processCode,
		SnapshotStructure: MustJSON(graph),
		ParentID:          0,
		ParentNodeID:      "",
		Variables:         datatypes.JSON([]byte(`{"amount":100}`)),
		Status:            model.InstanceStatusPending,
		SubmitterID:       RandomString("submitter"),
	}
}

func RandomProcessTask(instanceID, executionID int64, nodeID string) *model.ProcessTask {
	if nodeID == "" {
		nodeID = RandomString("node")
	}

	return &model.ProcessTask{
		InstanceID:  instanceID,
		ExecutionID: executionID,
		Type:        model.TaskTypeUser,
		Assignee:    RandomString("assignee"),
		Candidates:  datatypes.JSON([]byte(`{"users":[]}`)),
		Status:      model.TaskStatusPending,
		Variables:   datatypes.JSON([]byte("{}")),
	}
}

func MustJSON(v interface{}) datatypes.JSON {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return datatypes.JSON(b)
}
