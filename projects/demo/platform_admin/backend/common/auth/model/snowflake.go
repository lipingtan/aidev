package model

import (
	"github.com/bwmarrin/snowflake"
)

// node 包级雪花节点实例，线程安全
var node *snowflake.Node

func init() {
	var err error
	// 使用 node ID 1 初始化，后续可通过配置文件或环境变量调整
	node, err = snowflake.NewNode(1)
	if err != nil {
		panic("初始化雪花 ID 生成器失败: " + err.Error())
	}
}

// NextID 生成下一个雪花 ID（线程安全）
func NextID() int64 {
	return node.Generate().Int64()
}
