package apis

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service"
	"go-admin/app/admin/service/dto"
	authMiddleware "go-admin/common/auth/middleware"
	"go-admin/common/middleware"
)

// ApprovalFlow 审批流定义 Handler
type ApprovalFlow struct {
	api.Api
}

// getDB 从 sdk.Runtime 获取第一个可用 DB（不依赖 host key，避免 MakeOrm nil panic）
func getDB() *gorm.DB {
	for _, d := range sdk.Runtime.GetDb() {
		if d != nil {
			return d
		}
	}
	return nil
}

// ok 成功响应，与 auth-rbac 体系格式对齐（code=0）
func ok(c *gin.Context, data interface{}, msg string) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": data, "msg": msg})
}

// pageOK 分页成功响应（code=0）
func pageOK(c *gin.Context, list interface{}, total, page, pageSize int, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":     list,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
		"msg": msg,
	})
}

// fail 错误响应
func fail(c *gin.Context, bizCode int, msg string) {
	c.JSON(http.StatusOK, gin.H{"code": bizCode, "data": nil, "msg": msg})
}

// isCurrentUserSuperAdmin 判断当前用户是否为超级管理员
func isCurrentUserSuperAdmin(c *gin.Context) bool {
	if v, exists := c.Get("is_super_admin"); exists {
		if b, ok := v.(bool); ok && b {
			return true
		}
	}
	authCtx := authMiddleware.GetAuthContext(c)
	if authCtx != nil && authCtx.TenantID == 0 {
		return true
	}
	if middleware.GetTenantId(c) == 0 {
		return true
	}
	return false
}

// getCreateBy 获取当前用户 ID 用于 create_by 字段
func getCreateBy(c *gin.Context) int {
	authCtx := authMiddleware.GetAuthContext(c)
	if authCtx == nil {
		return 0
	}
	return int(authCtx.UserID)
}

// GetPage 审批流定义分页列表
// @Router /api/v1/admin/approval-flows [get]
// @Security Bearer
func (e ApprovalFlow) GetPage(c *gin.Context) {
	db := getDB()
	if db == nil {
		fail(c, 500, "数据库连接未就绪")
		return
	}

	req := dto.ApprovalFlowGetPageReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		fail(c, 40000, "参数错误: "+err.Error())
		return
	}

	s := service.ApprovalFlowService{}
	s.Orm = db

	tenantID := int64(middleware.GetTenantId(c))
	superAdmin := isCurrentUserSuperAdmin(c)
	list := make([]models.AdminApprovalFlow, 0)
	var count int64

	if err := s.GetPage(&req, tenantID, superAdmin, &list, &count); err != nil {
		fail(c, 500, "查询失败: "+err.Error())
		return
	}

	pageOK(c, list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get 审批流定义详情
// @Router /api/v1/admin/approval-flows/:id [get]
// @Security Bearer
func (e ApprovalFlow) Get(c *gin.Context) {
	db := getDB()
	if db == nil {
		fail(c, 500, "数据库连接未就绪")
		return
	}

	s := service.ApprovalFlowService{}
	s.Orm = db

	req := dto.ApprovalFlowGetReq{Id: c.Param("id")}
	tenantID := int64(middleware.GetTenantId(c))
	superAdmin := isCurrentUserSuperAdmin(c)

	var object models.AdminApprovalFlow
	if err := s.Get(&req, tenantID, superAdmin, &object); err != nil {
		fail(c, 500, "查询失败: "+err.Error())
		return
	}

	ok(c, object, "查询成功")
}

// Insert 创建审批流定义
// @Router /api/v1/admin/approval-flows [post]
// @Security Bearer
func (e ApprovalFlow) Insert(c *gin.Context) {
	db := getDB()
	if db == nil {
		fail(c, 500, "数据库连接未就绪")
		return
	}

	req := dto.ApprovalFlowInsertReq{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		fail(c, 40000, "参数错误: "+err.Error())
		return
	}

	s := service.ApprovalFlowService{}
	s.Orm = db

	req.CreateBy = getCreateBy(c)
	tenantID := int64(middleware.GetTenantId(c))
	superAdmin := isCurrentUserSuperAdmin(c)

	if err := s.Insert(&req, tenantID, superAdmin); err != nil {
		fail(c, 500, "创建失败: "+err.Error())
		return
	}

	ok(c, nil, "创建成功")
}

// Update 更新审批流定义
// @Router /api/v1/admin/approval-flows/:id [put]
// @Security Bearer
func (e ApprovalFlow) Update(c *gin.Context) {
	db := getDB()
	if db == nil {
		fail(c, 500, "数据库连接未就绪")
		return
	}

	req := dto.ApprovalFlowUpdateReq{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		fail(c, 40000, "参数错误: "+err.Error())
		return
	}

	s := service.ApprovalFlowService{}
	s.Orm = db

	req.Id = c.Param("id")
	req.UpdateBy = getCreateBy(c)
	tenantID := int64(middleware.GetTenantId(c))
	superAdmin := isCurrentUserSuperAdmin(c)

	if err := s.Update(&req, tenantID, superAdmin); err != nil {
		fail(c, 500, "更新失败: "+err.Error())
		return
	}

	ok(c, nil, "更新成功")
}

// Delete 删除审批流定义
// @Router /api/v1/admin/approval-flows/:id [delete]
// @Security Bearer
func (e ApprovalFlow) Delete(c *gin.Context) {
	db := getDB()
	if db == nil {
		fail(c, 500, "数据库连接未就绪")
		return
	}

	s := service.ApprovalFlowService{}
	s.Orm = db

	req := dto.ApprovalFlowDeleteReq{Id: c.Param("id")}
	tenantID := int64(middleware.GetTenantId(c))
	superAdmin := isCurrentUserSuperAdmin(c)

	if err := s.Delete(&req, tenantID, superAdmin); err != nil {
		fail(c, 500, "删除失败: "+err.Error())
		return
	}

	ok(c, nil, "删除成功")
}
