package main

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"platform-admin/plugin-sdk/proto"
)

// handleRequest 路由分发：根据路径前缀委托到对应子模块
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

	// 按前缀路由到子模块
	switch {
	// 游戏列表页面（前端菜单入口）
	case path == "list" && method == "GET":
		return handleGameGetPage(req)

	// 游戏 CRUD: /game, /game/:id
	case strings.HasPrefix(path, "game"):
		sub := strings.TrimPrefix(path, "game")
		sub = strings.TrimPrefix(sub, "/")
		return routeGame(method, sub, req)

	// DLC 管理: /dlc, /dlc/:id, /dlc/:id/upload, /dlc/:id/download, /dlc/stats
	case strings.HasPrefix(path, "dlc"):
		sub := strings.TrimPrefix(path, "dlc")
		sub = strings.TrimPrefix(sub, "/")
		return routeDlc(method, sub, req)

	// 玩家管理: /player, /player/:id, /player/:id/ban
	case strings.HasPrefix(path, "player"):
		sub := strings.TrimPrefix(path, "player")
		sub = strings.TrimPrefix(sub, "/")
		return routePlayer(method, sub, req)

	// 订单管理: /order, /order/:id, /order/:id/refund
	case strings.HasPrefix(path, "order"):
		sub := strings.TrimPrefix(path, "order")
		sub = strings.TrimPrefix(sub, "/")
		return routeOrder(method, sub, req)

	// 支付配置: /payment, /payment/:id
	case strings.HasPrefix(path, "payment"):
		sub := strings.TrimPrefix(path, "payment")
		sub = strings.TrimPrefix(sub, "/")
		return routePayment(method, sub, req)

	// H5 页面: /h5, /h5/:id
	case strings.HasPrefix(path, "h5"):
		sub := strings.TrimPrefix(path, "h5")
		sub = strings.TrimPrefix(sub, "/")
		return routeH5(method, sub, req)

	default:
		return jsonResp(404, "接口不存在: "+method+" /"+path, nil)
	}
}

// ===== 路由辅助函数 =====

// routeGame 游戏子路由
func routeGame(method, sub string, req *proto.HttpRequest) (*proto.HttpResponse, error) {
	switch {
	case method == "GET" && sub == "":
		return handleGameGetPage(req)
	case method == "GET" && isNumeric(sub):
		return handleGameGet(req, toInt(sub))
	case method == "POST" && sub == "":
		return handleGameInsert(req)
	case method == "PUT" && isNumeric(sub):
		return handleGameUpdate(req, toInt(sub))
	case method == "DELETE" && isNumeric(sub):
		return handleGameDelete(req, toInt(sub))
	// 重置密钥：POST /game/:id/regen-secret
	case method == "POST" && hasNumericPrefix(sub) && strings.HasSuffix(sub, "/regen-secret"):
		id := toInt(strings.TrimSuffix(sub, "/regen-secret"))
		return handleGameRegenSecret(req, id)
	default:
		return jsonResp(404, "游戏接口不存在: "+method+" /game/"+sub, nil)
	}
}

// routeDlc DLC 子路由
func routeDlc(method, sub string, req *proto.HttpRequest) (*proto.HttpResponse, error) {
	switch {
	case method == "GET" && sub == "":
		return handleDlcGetPage(req)
	case method == "GET" && sub == "stats":
		return handleDlcStats(req)
	case method == "GET" && isNumeric(sub):
		return handleDlcGet(req, toInt(sub))
	case method == "GET" && hasNumericPrefix(sub) && strings.HasSuffix(sub, "/download"):
		id := toInt(strings.TrimSuffix(sub, "/download"))
		return handleDlcDownload(req, id)
	case method == "POST" && sub == "":
		return handleDlcInsert(req)
	case method == "POST" && hasNumericPrefix(sub) && strings.HasSuffix(sub, "/upload"):
		id := toInt(strings.TrimSuffix(sub, "/upload"))
		return handleDlcUpload(req, id)
	case method == "PUT" && isNumeric(sub):
		return handleDlcUpdate(req, toInt(sub))
	case method == "DELETE" && isNumeric(sub):
		return handleDlcDelete(req, toInt(sub))
	default:
		return jsonResp(404, "DLC接口不存在: "+method+" /dlc/"+sub, nil)
	}
}

// routePlayer 玩家子路由
func routePlayer(method, sub string, req *proto.HttpRequest) (*proto.HttpResponse, error) {
	switch {
	case method == "GET" && sub == "":
		return handlePlayerGetPage(req)
	case method == "GET" && isNumeric(sub):
		return handlePlayerGet(req, toInt(sub))
	case method == "PUT" && hasNumericPrefix(sub) && strings.HasSuffix(sub, "/ban"):
		id := toInt(strings.TrimSuffix(sub, "/ban"))
		return handlePlayerBan(req, id)
	default:
		return jsonResp(404, "玩家接口不存在: "+method+" /player/"+sub, nil)
	}
}

// routeOrder 订单子路由
func routeOrder(method, sub string, req *proto.HttpRequest) (*proto.HttpResponse, error) {
	switch {
	case method == "GET" && sub == "":
		return handleOrderGetPage(req)
	case method == "GET" && isNumeric(sub):
		return handleOrderGet(req, toInt(sub))
	case method == "POST" && hasNumericPrefix(sub) && strings.HasSuffix(sub, "/refund"):
		id := toInt(strings.TrimSuffix(sub, "/refund"))
		return handleOrderRefund(req, id)
	default:
		return jsonResp(404, "订单接口不存在: "+method+" /order/"+sub, nil)
	}
}

// routePayment 支付配置子路由
func routePayment(method, sub string, req *proto.HttpRequest) (*proto.HttpResponse, error) {
	switch {
	case method == "GET" && sub == "":
		return handlePaymentGetPage(req)
	case method == "POST" && sub == "":
		return handlePaymentSave(req)
	// 编辑已有配置：PUT /payment/:id
	case method == "PUT" && isNumeric(sub):
		return handlePaymentUpdate(req, toInt(sub))
	default:
		return jsonResp(404, "支付配置接口不存在: "+method+" /payment/"+sub, nil)
	}
}

// routeH5 H5 页面子路由
func routeH5(method, sub string, req *proto.HttpRequest) (*proto.HttpResponse, error) {
	switch {
	case method == "GET" && sub == "":
		return handleH5GetPage(req)
	case method == "GET" && isNumeric(sub):
		return handleH5Get(req, toInt(sub))
	case method == "POST" && sub == "":
		return handleH5Insert(req)
	case method == "PUT" && isNumeric(sub):
		return handleH5Update(req, toInt(sub))
	case method == "DELETE" && isNumeric(sub):
		return handleH5Delete(req, toInt(sub))
	default:
		return jsonResp(404, "H5页面接口不存在: "+method+" /h5/"+sub, nil)
	}
}

// ===== 通用工具函数 =====

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

// isNumeric 判断字符串是否为纯数字
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	_, err := strconv.Atoi(s)
	return err == nil
}

// hasNumericPrefix 判断路径是否以数字开头（如 "123/upload"）
func hasNumericPrefix(s string) bool {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) == 0 {
		return false
	}
	_, err := strconv.Atoi(parts[0])
	return err == nil
}

// toInt 字符串转 int
func toInt(s string) int {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) == 0 {
		return 0
	}
	id, _ := strconv.Atoi(parts[0])
	return id
}

// PageResponse 分页响应包装
type PageResponse struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// parsePage 从 query 解析分页参数，兼容 page/pageIndex 两种 key
func parsePage(params map[string]string) (page, pageSize int) {
	// 优先读 pageIndex（前端统一用法），兜底读 page
	if v := params["pageIndex"]; v != "" {
		page, _ = strconv.Atoi(v)
	} else {
		page, _ = strconv.Atoi(params["page"])
	}
	pageSize, _ = strconv.Atoi(params["pageSize"])
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return
}
