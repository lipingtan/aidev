package model

import (
	"fmt"
	"net"

	"github.com/bwmarrin/snowflake"
)

// node 包级雪花节点实例，线程安全
var node *snowflake.Node

func init() {
	nodeID := nodeIDFromPodIP()
	var err error
	node, err = snowflake.NewNode(nodeID)
	if err != nil {
		panic(fmt.Sprintf("初始化雪花 ID 生成器失败 nodeID=%d: %v", nodeID, err))
	}
}

// nodeIDFromPodIP 通过 Pod IP 最后两个字节取模生成 NodeID（0-1023）。
// 在 K8s 同一 CIDR 内不同 Pod 的 IP 后两字节通常不同，碰撞概率极低。
// 单机部署时回退到 1。
func nodeIDFromPodIP() int64 {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return 1
	}
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}
		ip4 := ipnet.IP.To4()
		if ip4 == nil {
			continue
		}
		// 取 IP 最后两个字节拼成 16-bit 整数，对 1024 取模
		nodeID := int64((int(ip4[2])<<8 + int(ip4[3])) % 1024)
		return nodeID
	}
	return 1
}

// NextID 生成下一个雪花 ID（线程安全）
func NextID() int64 {
	return node.Generate().Int64()
}
