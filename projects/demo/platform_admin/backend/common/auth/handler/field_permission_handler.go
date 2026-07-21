package handler

import (
	"strconv"

	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
)

// FieldPermissionHandler 字段权限 HTTP Handler
type FieldPermissionHandler struct {
	svc *service.FieldPermissionService
}

// NewFieldPermissionHandler 创建 FieldPermissionHandler 实例
func NewFieldPermissionHandler(svc *service.FieldPermissionService) *FieldPermissionHandler {
	return &FieldPermissionHandler{svc: svc}
}

// RegisterRoutes 注册字段权限路由
func (h *FieldPermissionHandler) RegisterRoutes(rg *gin.RouterGroup) {
	// 字段对象管理
	rg.GET("/field-objects", h.ListObjects)
	rg.POST("/field-objects", h.ManualRegisterObject)
	rg.GET("/field-objects/:objectCode/fields", h.ListFields)
	rg.PUT("/field-objects/:objectCode/fields/:fieldName", h.UpdateFieldDesc)
	rg.POST("/field-objects/:objectCode/fields", h.ManualRegisterField)

	// 字段权限管理
	rg.GET("/field-permissions", h.GetPermissions)
	rg.PUT("/field-permissions", h.SetPermissions)
	rg.DELETE("/field-permissions/:id", h.DeletePermission)
}

// ListObjects 获取已注册的字段对象列表
func (h *FieldPermissionHandler) ListObjects(c *gin.Context) {
	list, err := h.svc.ListObjects()
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, list)
}

// ListFields 获取对象的字段定义列表
func (h *FieldPermissionHandler) ListFields(c *gin.Context) {
	objectCode := c.Param("objectCode")
	if objectCode == "" {
		Error(c, &errBadRequest{message: "object_code 不能为空"})
		return
	}

	list, err := h.svc.ListFields(objectCode)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, list)
}

// updateFieldDescRequest 修改字段描述请求体
type updateFieldDescRequest struct {
	Description string `json:"description" binding:"required"`
}

// UpdateFieldDesc 修改字段描述
func (h *FieldPermissionHandler) UpdateFieldDesc(c *gin.Context) {
	objectCode := c.Param("objectCode")
	fieldName := c.Param("fieldName")
	if objectCode == "" || fieldName == "" {
		Error(c, &errBadRequest{message: "参数不完整"})
		return
	}

	var req updateFieldDescRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, &errBadRequest{message: "参数错误: " + err.Error()})
		return
	}

	if err := h.svc.UpdateFieldDesc(objectCode, fieldName, req.Description); err != nil {
		Error(c, err)
		return
	}
	Success(c, gin.H{"object_code": objectCode, "field_name": fieldName})
}

// manualRegisterObjectRequest 手动注册对象请求体
type manualRegisterObjectRequest struct {
	ObjectCode string `json:"object_code" binding:"required"`
	ObjectName string `json:"object_name" binding:"required"`
}

// ManualRegisterObject 手动注册字段对象
func (h *FieldPermissionHandler) ManualRegisterObject(c *gin.Context) {
	var req manualRegisterObjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, &errBadRequest{message: "参数错误: " + err.Error()})
		return
	}

	if err := h.svc.ManualRegisterObject(req.ObjectCode, req.ObjectName); err != nil {
		Error(c, err)
		return
	}
	Success(c, gin.H{"object_code": req.ObjectCode})
}

// manualRegisterFieldRequest 手动注册字段请求体
type manualRegisterFieldRequest struct {
	FieldName   string `json:"field_name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

// ManualRegisterField 手动注册字段
func (h *FieldPermissionHandler) ManualRegisterField(c *gin.Context) {
	objectCode := c.Param("objectCode")
	if objectCode == "" {
		Error(c, &errBadRequest{message: "object_code 不能为空"})
		return
	}

	var req manualRegisterFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, &errBadRequest{message: "参数错误: " + err.Error()})
		return
	}

	if err := h.svc.ManualRegisterField(objectCode, req.FieldName, req.Description); err != nil {
		Error(c, err)
		return
	}
	Success(c, gin.H{"object_code": objectCode, "field_name": req.FieldName})
}

// setPermissionsRequest 批量设置字段权限请求体
type setPermissionsRequest struct {
	RoleID     int64                      `json:"role_id,string" binding:"required"`
	ObjectCode string                     `json:"object_code" binding:"required"`
	Items      []service.SetPermissionItem `json:"items" binding:"required"`
}

// GetPermissions 查询角色字段权限
func (h *FieldPermissionHandler) GetPermissions(c *gin.Context) {
	roleIDStr := c.Query("role_id")
	objectCode := c.Query("object_code")
	if roleIDStr == "" || objectCode == "" {
		Error(c, &errBadRequest{message: "role_id 和 object_code 不能为空"})
		return
	}

	roleID, err := strconv.ParseInt(roleIDStr, 10, 64)
	if err != nil {
		Error(c, &errBadRequest{message: "无效的 role_id"})
		return
	}

	list, err := h.svc.GetPermissions(roleID, objectCode)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, list)
}

// SetPermissions 批量设置角色字段权限（全量替换）
func (h *FieldPermissionHandler) SetPermissions(c *gin.Context) {
	var req setPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, &errBadRequest{message: "参数错误: " + err.Error()})
		return
	}

	if err := h.svc.SetPermissions(req.RoleID, req.ObjectCode, req.Items); err != nil {
		Error(c, err)
		return
	}
	Success(c, gin.H{"role_id": req.RoleID, "object_code": req.ObjectCode})
}

// DeletePermission 删除字段权限配置
func (h *FieldPermissionHandler) DeletePermission(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, &errBadRequest{message: "无效的 ID"})
		return
	}

	if err := h.svc.DeletePermission(id); err != nil {
		Error(c, err)
		return
	}
	Success(c, gin.H{"id": id})
}
