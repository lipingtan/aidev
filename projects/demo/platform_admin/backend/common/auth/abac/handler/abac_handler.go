// Package handler 提供 ABAC 策略管理的 HTTP Handler。
package handler

import (
	"net/http"
	"strconv"

	"go-admin/common/auth/abac/engine"
	"go-admin/common/auth/abac/service"
	autherrors "go-admin/common/auth/errors"
	"go-admin/common/auth/middleware"

	"github.com/gin-gonic/gin"
)

// AbacHandler ABAC 策略 HTTP Handler
type AbacHandler struct {
	svc *service.AbacService
}

// NewAbacHandler 创建 AbacHandler
func NewAbacHandler(svc *service.AbacService) *AbacHandler {
	return &AbacHandler{svc: svc}
}

// RegisterRoutes 注册路由
func (h *AbacHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/abac")
	{
		g.GET("/policies", h.List)
		g.POST("/policies", h.Create)
		g.GET("/policies/:id", h.GetByID)
		g.PUT("/policies/:id", h.Update)
		g.DELETE("/policies/:id", h.Delete)
		g.GET("/resources", h.ListResources)
		g.POST("/evaluate", h.Evaluate)
	}
}

// List GET /api/v1/admin/abac/policies
func (h *AbacHandler) List(c *gin.Context) {
	tenantID := getTenantID(c)
	req := &service.ListAbacPoliciesRequest{
		TenantID:     tenantID,
		ResourceType: c.Query("resource_type"),
		SubjectType:  c.Query("subject_type"),
	}
	req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	req.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.svc.List(req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, gin.H{"list": list, "total": total})
}

// Create POST /api/v1/admin/abac/policies
func (h *AbacHandler) Create(c *gin.Context) {
	var req service.CreateAbacPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, autherrors.NewAuthError(autherrors.ErrInvalidParam, "请求参数无效: "+err.Error()))
		return
	}
	tenantID := getTenantID(c)
	policy, err := h.svc.Create(tenantID, &req)
	if err != nil {
		Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 200, "data": policy})
}

// GetByID GET /api/v1/admin/abac/policies/:id
func (h *AbacHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, autherrors.NewAuthError(autherrors.ErrEntityNotFound, "无效的策略 ID"))
		return
	}
	detail, err := h.svc.GetByID(id)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, detail)
}

// Update PUT /api/v1/admin/abac/policies/:id
func (h *AbacHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, autherrors.NewAuthError(autherrors.ErrEntityNotFound, "无效的策略 ID"))
		return
	}
	var req service.UpdateAbacPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, autherrors.NewAuthError(autherrors.ErrInvalidParam, "请求参数无效: "+err.Error()))
		return
	}
	tenantID := getTenantID(c)
	if err := h.svc.Update(tenantID, id, &req); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// Delete DELETE /api/v1/admin/abac/policies/:id
func (h *AbacHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, autherrors.NewAuthError(autherrors.ErrEntityNotFound, "无效的策略 ID"))
		return
	}
	tenantID := getTenantID(c)
	if err := h.svc.Delete(tenantID, id); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// ListResources GET /api/v1/admin/abac/resources
func (h *AbacHandler) ListResources(c *gin.Context) {
	resources := engine.ListResources()
	// 只返回前端需要的字段（Type/DisplayName/Attributes.AttrName/Display/DataType）
	type attrVO struct {
		AttrName string `json:"attr_name"`
		Display  string `json:"display"`
		DataType string `json:"data_type"`
	}
	type resVO struct {
		Type        string   `json:"type"`
		DisplayName string   `json:"display_name"`
		Attributes  []attrVO `json:"attributes"`
	}
	result := make([]resVO, 0, len(resources))
	for _, r := range resources {
		attrs := make([]attrVO, 0, len(r.Attributes))
		for _, a := range r.Attributes {
			attrs = append(attrs, attrVO{AttrName: a.AttrName, Display: a.Display, DataType: a.DataType})
		}
		result = append(result, resVO{Type: r.Type, DisplayName: r.DisplayName, Attributes: attrs})
	}
	// 主体属性列表
	subjectAttrs := engine.ListSubjectAttrs()
	type subjectAttrVO struct {
		AttrName string `json:"attr_name"`
		Display  string `json:"display"`
		DataType string `json:"data_type"`
	}
	subjectResult := make([]subjectAttrVO, 0, len(subjectAttrs))
	for _, s := range subjectAttrs {
		subjectResult = append(subjectResult, subjectAttrVO{AttrName: s.AttrName, Display: s.Display, DataType: s.DataType})
	}
	Success(c, gin.H{"resources": result, "subject_attrs": subjectResult})
}

// EvaluateRequest 策略评估请求
type EvaluateRequest struct {
	ResourceType string            `json:"resource_type" binding:"required"`
	Action       string            `json:"action"        binding:"required"`
	Subject      map[string]interface{} `json:"subject" binding:"required"`
}

// Evaluate POST /api/v1/admin/abac/evaluate
func (h *AbacHandler) Evaluate(c *gin.Context) {
	var req EvaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, autherrors.NewAuthError(autherrors.ErrInvalidParam, "请求参数无效: "+err.Error()))
		return
	}

	// 从 subject 中解析 tenant_id
	tenantID := getTenantID(c)
	userID := ""
	if v, ok := req.Subject["user_id"].(string); ok {
		userID = v
	}

	rowEntries, colEntries := h.svc.LoadPoliciesForCallback(tenantID, userID, req.ResourceType)

	resDef, ok := engine.GetResource(req.ResourceType)
	if !ok {
		resDef = &engine.ResourceDef{Type: req.ResourceType}
	}

	rowResult := engine.MergeRowPolicies(rowEntries, req.Action, req.Subject, resDef)
	colEffects := engine.MergeColPolicies(colEntries)

	result := engine.BuildEvaluateResult(rowResult, colEffects)
	Success(c, result)
}

// ---- 辅助函数 ----

func getTenantID(c *gin.Context) int64 {
	authInfo := middleware.GetAuthInfo(c.Request.Context())
	if authInfo != nil {
		return authInfo.TenantID
	}
	return 0
}

// Success 统一成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// Error 统一错误响应
func Error(c *gin.Context, err error) {
	if authErr, ok := err.(*autherrors.AuthError); ok {
		statusCode := http.StatusBadRequest
		switch {
		case authErr.Code == autherrors.ErrEntityNotFound:
			statusCode = http.StatusNotFound
		case authErr.Code == autherrors.ErrDuplicateEntity:
			statusCode = http.StatusConflict
		case authErr.Code == autherrors.ErrPermissionDenied:
			statusCode = http.StatusForbidden
		case authErr.Code == autherrors.ErrOptimisticLock:
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{"code": authErr.Code, "msg": authErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
}
