package apis

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk/api"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service"
	"go-admin/app/admin/service/dto"
	authMiddleware "go-admin/common/auth/middleware"
	"go-admin/common/middleware"
)

// Approval 审批实例 Handler
type Approval struct {
	api.Api
}

// getCurrentUserID 从 auth-rbac AuthContext 获取当前用户ID
func getCurrentUserID(c *gin.Context) int64 {
	authCtx := authMiddleware.GetAuthContext(c)
	if authCtx == nil {
		return 0
	}
	return authCtx.UserID
}

// GetPage 审批实例分页列表
// @Router /api/v1/admin/approvals [get]
// @Security Bearer
func (e Approval) GetPage(c *gin.Context) {
	db := getDB()
	if db == nil {
		fail(c, 500, "数据库连接未就绪")
		return
	}

	tenantID := int64(middleware.GetTenantId(c))
	currentUserID := getCurrentUserID(c)
	view := c.DefaultQuery("view", "mine")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	list := make([]models.AdminApproval, 0)
	var count int64

	switch view {
	case "pending":
		userIDStr := fmt.Sprintf("%d", currentUserID)
		rawSQL := `SELECT DISTINCT a.*
FROM admin_approval a
JOIN admin_approval_node n ON a.id = n.approval_id
WHERE n.status = 'PENDING'
  AND n.tenant_id = ?
  AND JSON_CONTAINS(n.assignee_user_ids, JSON_QUOTE(?))
ORDER BY a.created_at DESC
LIMIT ? OFFSET ?`
		if err := db.Raw(rawSQL, tenantID, userIDStr, pageSize, offset).Scan(&list).Error; err != nil {
			fail(c, 500, "查询待审批列表失败: "+err.Error())
			return
		}
		countSQL := `SELECT COUNT(DISTINCT a.id)
FROM admin_approval a
JOIN admin_approval_node n ON a.id = n.approval_id
WHERE n.status = 'PENDING'
  AND n.tenant_id = ?
  AND JSON_CONTAINS(n.assignee_user_ids, JSON_QUOTE(?))`
		db.Raw(countSQL, tenantID, userIDStr).Scan(&count)

	case "mine":
		if err := db.Where("applicant_id = ? AND tenant_id = ?", currentUserID, tenantID).
			Order("created_at DESC").
			Limit(pageSize).Offset(offset).
			Find(&list).Error; err != nil {
			fail(c, 500, "查询我发起的列表失败: "+err.Error())
			return
		}
		db.Model(&models.AdminApproval{}).
			Where("applicant_id = ? AND tenant_id = ?", currentUserID, tenantID).
			Count(&count)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "msg": "view 参数无效，可选值：pending、mine"})
		return
	}

	pageOK(c, list, int(count), page, pageSize, "查询成功")
}

// Get 审批实例详情（含节点列表）
// @Router /api/v1/admin/approvals/:id [get]
// @Security Bearer
func (e Approval) Get(c *gin.Context) {
	db := getDB()
	if db == nil {
		fail(c, 500, "数据库连接未就绪")
		return
	}

	idStr := c.Param("id")
	approvalID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		fail(c, 40000, "ID 格式错误")
		return
	}

	tenantID := int64(middleware.GetTenantId(c))

	var approval models.AdminApproval
	if err := db.Where("id = ? AND tenant_id = ?", approvalID, tenantID).First(&approval).Error; err != nil {
		fail(c, 500, "审批实例不存在或无权访问")
		return
	}

	nodes := make([]models.AdminApprovalNode, 0)
	db.Where("approval_id = ?", approvalID).Order("node_order ASC").Find(&nodes)

	ok(c, gin.H{"approval": approval, "nodes": nodes}, "查询成功")
}

// Insert 发起审批
// @Router /api/v1/admin/approvals [post]
// @Security Bearer
func (e Approval) Insert(c *gin.Context) {
	db := getDB()
	if db == nil {
		fail(c, 500, "数据库连接未就绪")
		return
	}

	s := service.ApprovalService{}
	s.Orm = db

	req := dto.InitApprovalReq{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		fail(c, 40000, "参数错误: "+err.Error())
		return
	}

	req.TenantID = int64(middleware.GetTenantId(c))
	req.ApplicantID = getCurrentUserID(c)

	approvalID, err := s.InitApproval(&req)
	if err != nil {
		fail(c, 500, "发起审批失败: "+err.Error())
		return
	}

	ok(c, gin.H{"id": fmt.Sprintf("%d", approvalID)}, "发起审批成功")
}

// Approve 审批通过
// @Router /api/v1/admin/approvals/:id/approve [post]
// @Security Bearer
func (e Approval) Approve(c *gin.Context) {
	db := getDB()
	if db == nil {
		fail(c, 500, "数据库连接未就绪")
		return
	}

	s := service.ApprovalService{}
	s.Orm = db

	req := dto.ApproveReq{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		fail(c, 40000, "参数错误: "+err.Error())
		return
	}

	idStr := c.Param("id")
	approvalID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		fail(c, 40000, "ID 格式错误")
		return
	}
	currentUserID := getCurrentUserID(c)

	if err := s.Approve(approvalID, currentUserID, req.Comment); err != nil {
		fail(c, 500, "审批失败: "+err.Error())
		return
	}

	ok(c, nil, "审批通过")
}

// Reject 驳回审批
// @Router /api/v1/admin/approvals/:id/reject [post]
// @Security Bearer
func (e Approval) Reject(c *gin.Context) {
	db := getDB()
	if db == nil {
		fail(c, 500, "数据库连接未就绪")
		return
	}

	s := service.ApprovalService{}
	s.Orm = db

	req := dto.RejectReq{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		fail(c, 40000, "参数错误: "+err.Error())
		return
	}

	idStr := c.Param("id")
	approvalID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		fail(c, 40000, "ID 格式错误")
		return
	}
	currentUserID := getCurrentUserID(c)

	if err := s.Reject(approvalID, currentUserID, req.Reason); err != nil {
		fail(c, 500, "驳回失败: "+err.Error())
		return
	}

	ok(c, nil, "驳回成功")
}

// Cancel 撤销审批
// @Router /api/v1/admin/approvals/:id/cancel [post]
// @Security Bearer
func (e Approval) Cancel(c *gin.Context) {
	db := getDB()
	if db == nil {
		fail(c, 500, "数据库连接未就绪")
		return
	}

	s := service.ApprovalService{}
	s.Orm = db

	req := dto.CancelReq{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		fail(c, 40000, "参数错误: "+err.Error())
		return
	}

	idStr := c.Param("id")
	approvalID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		fail(c, 40000, "ID 格式错误")
		return
	}
	currentUserID := getCurrentUserID(c)
	isSuperAdmin := isCurrentUserSuperAdmin(c)

	if err := s.Cancel(approvalID, currentUserID, req.CancelReason, isSuperAdmin); err != nil {
		fail(c, 500, "撤销失败: "+err.Error())
		return
	}

	ok(c, nil, "撤销成功")
}
