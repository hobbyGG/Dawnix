package biz

import "github.com/hobbyGG/Dawnix/internal/biz/model"

type Navigator struct {
}

func NewNavigator() *Navigator {
	return &Navigator{}
}

// 根据runtimeGraph与当前节点找到下一条edge
func (navi *Navigator) FindPaths(rg *RuntimeGraph, currentNodeID string) []*model.EdgeModel {
	// 仅考虑userNode类型的情况
	// 返回userNode连接的所有边
	return rg.Next[currentNodeID]
}
