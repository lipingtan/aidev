package handler

import (
	"net/http"
	"strconv"

	"go-admin/common/auth/middleware"
	"go-admin/common/auth/model"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
)

// AdminConfigHandler 三级配置 HTTP Handler
type AdminConfigHandler struct {
	svc *service.AdminConfigService
}

// NewAdminConfigHandler 创建 AdminConfigHandler 实例
func NewAdminConfigHandler(svc *service.AdminConfigService) *AdminConfigHandler {
	return &AdminConfigHandler{svc: svc}
}

// List 配置列表
func (h *AdminConfigHandler) List(c *gin.Context) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "data": nil, "message": "未认证"})
		return
	}

	scope := c.Query("scope")
	key := c.Query("key")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var isFeatureFlag *int
	if ff := c.Query("is_feature_flag"); ff != "" {
		v, _ := strconv.Atoi(ff)
		isFeatureFlag = &v
	}

	list, total, err := h.svc.List(scope, key, isFeatureFlag, authCtx.TenantID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "data": nil, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"list": list, "total": total, "page": page, "page_size": pageSize},
		"message": "success",
	})
}

// Create 创建配置
func (h *AdminConfigHandler) Create(c *gin.Context) {
	var cfg model.AdminConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": err.Error()})
		return
	}

	if err := h.svc.Create(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": cfg, "message": "success"})
}

// Update 更新配置
func (h *AdminConfigHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "无效的 ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": err.Error()})
		return
	}

	if err := h.svc.Update(id, updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": nil, "message": "success"})
}

// Delete 删除配置
func (h *AdminConfigHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "无效的 ID"})
		return
	}

	if err := h.svc.Delete(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": nil, "message": "success"})
}

// Resolve 按三级逻辑解析有效值
func (h *AdminConfigHandler) Resolve(c *gin.Context) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "data": nil, "message": "未认证"})
		return
	}

	key := c.Param("key")
	val, ok := h.svc.Resolve(key, authCtx.TenantID, authCtx.UserID)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"key": key, "value": nil, "found": false}, "message": "success"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"key": key, "value": val, "found": true}, "message": "success"})
}

// GetFeatureFlags 获取当前租户功能开关列表
func (h *AdminConfigHandler) GetFeatureFlags(c *gin.Context) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "data": nil, "message": "未认证"})
		return
	}

	flags, err := h.svc.GetFeatureFlags(authCtx.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "data": nil, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": flags, "message": "success"})
}
