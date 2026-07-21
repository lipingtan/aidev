package handler

import (
	"net/http"
	"strconv"

	"go-admin/common/auth/middleware"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
)

// OrgUnitHandler 组织架构 HTTP Handler
type OrgUnitHandler struct {
	svc *service.OrgUnitService
}

// NewOrgUnitHandler 创建 OrgUnitHandler 实例
func NewOrgUnitHandler(svc *service.OrgUnitService) *OrgUnitHandler {
	return &OrgUnitHandler{svc: svc}
}

// GetTree 获取组织架构树
func (h *OrgUnitHandler) GetTree(c *gin.Context) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "data": nil, "message": "未认证"})
		return
	}

	tree, err := h.svc.GetTree(authCtx.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "data": nil, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": tree, "message": "success"})
}

// Create 创建组织节点
func (h *OrgUnitHandler) Create(c *gin.Context) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "data": nil, "message": "未认证"})
		return
	}

	var req service.CreateOrgUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": err.Error()})
		return
	}
	req.TenantID = authCtx.TenantID

	unit, err := h.svc.CreateOrgUnit(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": unit, "message": "success"})
}

// Update 更新组织节点
func (h *OrgUnitHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "无效的 ID"})
		return
	}

	var req service.UpdateOrgUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": err.Error()})
		return
	}

	unit, err := h.svc.UpdateOrgUnit(id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": unit, "message": "success"})
}

// Delete 删除组织节点
func (h *OrgUnitHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "无效的 ID"})
		return
	}

	if err := h.svc.DeleteOrgUnit(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": nil, "message": "success"})
}

// GetNodeUsers 获取节点下用户
func (h *OrgUnitHandler) GetNodeUsers(c *gin.Context) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "data": nil, "message": "未认证"})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "无效的 ID"})
		return
	}

	userOrgs, err := h.svc.GetNodeUsers(id, authCtx.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "data": nil, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": userOrgs, "message": "success"})
}

// SetNodeUsers 设置节点用户
func (h *OrgUnitHandler) SetNodeUsers(c *gin.Context) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "data": nil, "message": "未认证"})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "无效的 ID"})
		return
	}

	var req service.SetNodeUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": err.Error()})
		return
	}

	if err := h.svc.SetNodeUsers(id, authCtx.TenantID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": nil, "message": "success"})
}
