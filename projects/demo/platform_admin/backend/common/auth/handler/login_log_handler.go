package handler

import (
	"strconv"

	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
)

// LoginLogHandler 登录日志 HTTP Handler
type LoginLogHandler struct {
	svc *service.LoginLogService
}

// NewLoginLogHandler 创建 LoginLogHandler 实例
func NewLoginLogHandler(svc *service.LoginLogService) *LoginLogHandler {
	return &LoginLogHandler{svc: svc}
}

// RegisterRoutes 注册登录日志路由
func (h *LoginLogHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/login-logs", h.List)
	rg.DELETE("/login-logs/:id", h.Delete)
}

// List 分页查询登录日志
func (h *LoginLogHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	params := service.LoginLogListParams{
		Page:     page,
		PageSize: pageSize,
		Username: c.Query("username"),
		IP:       c.Query("ip"),
	}

	if statusStr := c.Query("status"); statusStr != "" {
		if v, err := strconv.Atoi(statusStr); err == nil {
			params.Status = &v
		}
	}

	result, err := h.svc.List(params)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, result)
}

// Delete 删除登录日志
func (h *LoginLogHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, &errBadRequest{message: "无效的 ID"})
		return
	}

	if err := h.svc.Delete(id); err != nil {
		Error(c, err)
		return
	}
	Success(c, gin.H{"id": id})
}
