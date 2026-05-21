package main

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"game-server/plugin-sdk/proto"
)

// handleRequest 路由分发
func handleRequest(ctx context.Context, req *proto.HttpRequest) (*proto.HttpResponse, error) {
	// 初始化数据库
	if req.DbDsn != "" {
		if err := initDB(req.DbDsn); err != nil {
			return jsonResp(500, "数据库初始化失败: "+err.Error(), nil)
		}
	}
	if db == nil {
		return jsonResp(503, "数据库未初始化", nil)
	}

	path := strings.TrimPrefix(req.Path, "/")
	method := strings.ToUpper(req.Method)

	// 路由匹配
	switch {
	case method == "GET" && (path == "list" || path == ""):
		return handleGetPage(req)
	case method == "GET" && path == "stats":
		return handleStats(req)
	case method == "GET" && hasIdPrefix(path) && strings.HasSuffix(path, "/download"):
		id := extractId(strings.TrimSuffix(path, "/download"))
		return handleDownload(req, id)
	case method == "GET" && isIdPath(path):
		return handleGet(req, extractId(path))
	case method == "POST" && hasIdPrefix(path) && strings.HasSuffix(path, "/upload"):
		id := extractId(strings.TrimSuffix(path, "/upload"))
		return handleUpload(req, id)
	case method == "POST" && (path == "" || path == "/"):
		return handleInsert(req)
	case method == "PUT" && isIdPath(path):
		return handleUpdate(req, extractId(path))
	case method == "DELETE" && isIdPath(path):
		return handleDelete(req, extractId(path))
	default:
		return jsonResp(404, "接口不存在: "+method+" /"+path, nil)
	}
}

// isIdPath 判断路径是否为纯数字 ID
func isIdPath(path string) bool {
	if path == "" {
		return false
	}
	for _, c := range path {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// hasIdPrefix 判断路径是否以数字 ID 开头（如 "123/upload"）
func hasIdPrefix(path string) bool {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 {
		return false
	}
	_, err := strconv.Atoi(parts[0])
	return err == nil
}

// extractId 从路径中提取 ID
func extractId(path string) int {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 {
		return 0
	}
	id, _ := strconv.Atoi(parts[0])
	return id
}

// jsonResp 构造 JSON 响应
func jsonResp(code int, msg string, data interface{}) (*proto.HttpResponse, error) {
	body := map[string]interface{}{
		"code": code,
		"msg":  msg,
	}
	if data != nil {
		body["data"] = data
	}
	b, _ := json.Marshal(body)
	return &proto.HttpResponse{
		StatusCode: int32(code),
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       b,
	}, nil
}

// parseQuery 解析 query string 为 map
func parseQuery(query string) map[string]string {
	result := make(map[string]string)
	if query == "" {
		return result
	}
	values, err := url.ParseQuery(query)
	if err != nil {
		return result
	}
	for k, v := range values {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}
