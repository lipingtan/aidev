package handler

import (
	"strconv"
	"time"

	"go-admin/common/auth/model"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
)

// RecordShareHandler 记录共享 HTTP Handler
type RecordShareHandler struct {
	svc *service.RecordShareService
}

// NewRecordShareHandler 创建 RecordShareHandler 实例
func NewRecordShareHandler(svc *service.RecordShareService) *RecordShareHandler {
	return &RecordShareHandler{svc: svc}
}

// RegisterRoutes 注册记录共享路由
func (h *RecordShareHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/record-shares", h.Create)
	rg.GET("/record-shares", h.ListByRecord)
	rg.DELETE("/record-shares/:id", h.Delete)
}

// createRecordShareRequest 创建共享规则请求体
type createRecordShareRequest struct {
	ObjectCode  string  `json:"object_code" binding:"required"`
	RecordID    int64   `json:"record_id,string" binding:"required"`
	ShareToType string  `json:"share_to_type" binding:"required"` // USER/ROLE/DEPT
	ShareToID   int64   `json:"share_to_id,string" binding:"required"`
	AccessLevel string  `json:"access_level"` // READ/EDIT，默认 READ
	ExpireAt    *string `json:"expire_at"`    // ISO8601 时间字符串，可选
}

// Create 创建共享规则
func (h *RecordShareHandler) Create(c *gin.Context) {
	authCtx, err := MustGetAuthContext(c)
	if err != nil {
		Error(c, err)
		return
	}

	var req createRecordShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, &errBadRequest{message: "参数错误: " + err.Error()})
		return
	}

	// 校验 share_to_type
	if req.ShareToType != "USER" && req.ShareToType != "ROLE" && req.ShareToType != "DEPT" {
		Error(c, &errBadRequest{message: "share_to_type 必须为 USER/ROLE/DEPT"})
		return
	}

	// 校验 access_level
	accessLevel := req.AccessLevel
	if accessLevel == "" {
		accessLevel = "READ"
	}
	if accessLevel != "READ" && accessLevel != "EDIT" {
		Error(c, &errBadRequest{message: "access_level 必须为 READ/EDIT"})
		return
	}

	// 解析过期时间
	var expireAt *time.Time
	if req.ExpireAt != nil && *req.ExpireAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpireAt)
		if err != nil {
			Error(c, &errBadRequest{message: "expire_at 格式错误，需 ISO8601"})
			return
		}
		expireAt = &t
	}

	share := &model.RecordShare{
		TenantID:    authCtx.TenantID,
		ObjectCode:  req.ObjectCode,
		RecordID:    req.RecordID,
		ShareToType: req.ShareToType,
		ShareToID:   req.ShareToID,
		AccessLevel: accessLevel,
		ExpireAt:    expireAt,
		CreatedBy:   authCtx.UserID,
	}

	result, err := h.svc.Create(share)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, result)
}

// ListByRecord 查询记录的共享规则列表
func (h *RecordShareHandler) ListByRecord(c *gin.Context) {
	authCtx, err := MustGetAuthContext(c)
	if err != nil {
		Error(c, err)
		return
	}

	objectCode := c.Query("object_code")
	recordIDStr := c.Query("record_id")
	if objectCode == "" || recordIDStr == "" {
		Error(c, &errBadRequest{message: "object_code 和 record_id 不能为空"})
		return
	}

	recordID, err := strconv.ParseInt(recordIDStr, 10, 64)
	if err != nil {
		Error(c, &errBadRequest{message: "无效的 record_id"})
		return
	}

	list, err := h.svc.ListByRecord(authCtx.TenantID, objectCode, recordID)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, list)
}

// Delete 删除共享规则
func (h *RecordShareHandler) Delete(c *gin.Context) {
	_, err := MustGetAuthContext(c)
	if err != nil {
		Error(c, err)
		return
	}

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
