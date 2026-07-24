package handler

import (
	"net/http"
	"strconv"

	"go-admin/app/user_auth/dto"
	"go-admin/app/user_auth/service"
	"go-admin/common/auth/middleware"

	"github.com/gin-gonic/gin"
)

// BizUserHandler 管理端 biz_user CRUD Handler
type BizUserHandler struct {
	bizUserSvc *service.BizUserService
}

// NewBizUserHandler 创建 BizUserHandler 实例
func NewBizUserHandler(bizUserSvc *service.BizUserService) *BizUserHandler {
	return &BizUserHandler{bizUserSvc: bizUserSvc}
}

// RegisterRoutes 注册管理端 biz_user 路由
func (h *BizUserHandler) RegisterRoutes(rg *gin.RouterGroup) {
	bizUsers := rg.Group("/biz-users")
	{
		bizUsers.GET("", h.List)
		bizUsers.GET("/:id", h.GetByID)
		bizUsers.POST("", h.Create)
		bizUsers.PUT("/:id", h.Update)
		bizUsers.DELETE("/:id", h.Delete)
		bizUsers.POST("/:id/reset-password", h.ResetPassword)
		bizUsers.POST("/:id/force-logout", h.ForceLogout)
		bizUsers.POST("/:id/toggle-status", h.ToggleStatus)
	}
}

// List GET /api/v1/admin/biz-users — 分页列表
// query params: page, page_size, phone (模糊搜索)
// tenant_id 从 AuthContext 获取
func (h *BizUserHandler) List(c *gin.Context) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "data": nil, "message": "未认证"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	phone := c.Query("phone")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	list, total, err := h.bizUserSvc.ListBizUsers(authCtx.TenantID, page, pageSize, phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "data": nil, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":      list,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
		"message": "ok",
	})
}

// GetByID GET /api/v1/admin/biz-users/:id — 获取详情
func (h *BizUserHandler) GetByID(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "无效的用户 ID"})
		return
	}

	user, err := h.bizUserSvc.GetBizUser(id)
	if err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 40002, "data": nil, "message": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "data": nil, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": user, "message": "ok"})
}

// Create POST /api/v1/admin/biz-users — 创建
func (h *BizUserHandler) Create(c *gin.Context) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "data": nil, "message": "未认证"})
		return
	}

	var req dto.CreateBizUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 强制使用当前租户
	req.TenantID = authCtx.TenantID

	user, err := h.bizUserSvc.CreateBizUser(&req)
	if err != nil {
		if err == service.ErrDuplicatePhone {
			c.JSON(http.StatusConflict, gin.H{"code": 40901, "data": nil, "message": "手机号已存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "data": nil, "message": "创建失败"})
		return
	}

	// 记录 create_by
	user.CreateBy = authCtx.UserID
	user.UpdateBy = authCtx.UserID

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": user, "message": "ok"})
}

// Update PUT /api/v1/admin/biz-users/:id — 更新（乐观锁）
func (h *BizUserHandler) Update(c *gin.Context) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "data": nil, "message": "未认证"})
		return
	}

	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "无效的用户 ID"})
		return
	}

	var req dto.UpdateBizUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "请求参数错误: " + err.Error()})
		return
	}

	user, err := h.bizUserSvc.UpdateBizUser(id, &req)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"code": 40002, "data": nil, "message": "用户不存在"})
		case service.ErrVersionConflict:
			c.JSON(http.StatusConflict, gin.H{"code": 40902, "data": nil, "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "data": nil, "message": "更新失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": user, "message": "ok"})
}

// Delete DELETE /api/v1/admin/biz-users/:id — 软删除
func (h *BizUserHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "无效的用户 ID"})
		return
	}

	if err := h.bizUserSvc.DeleteBizUser(id); err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 40002, "data": nil, "message": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "data": nil, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": nil, "message": "ok"})
}

// ResetPassword POST /api/v1/admin/biz-users/:id/reset-password — 重置密码
func (h *BizUserHandler) ResetPassword(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "无效的用户 ID"})
		return
	}

	plainPassword, err := h.bizUserSvc.ResetPassword(id)
	if err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 40002, "data": nil, "message": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "data": nil, "message": "重置密码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    dto.ResetPasswordResponse{Password: plainPassword},
		"message": "ok",
	})
}

// ForceLogout POST /api/v1/admin/biz-users/:id/force-logout — 强制登出
func (h *BizUserHandler) ForceLogout(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "无效的用户 ID"})
		return
	}

	if err := h.bizUserSvc.ForceLogout(id); err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 40002, "data": nil, "message": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "data": nil, "message": "强制登出失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": nil, "message": "ok"})
}

// ToggleStatus POST /api/v1/admin/biz-users/:id/toggle-status — 启用/禁用
// 请求体: {"status": 0} 或 {"status": 1}
func (h *BizUserHandler) ToggleStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "无效的用户 ID"})
		return
	}

	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "请求参数错误: " + err.Error()})
		return
	}

	if err := h.bizUserSvc.ToggleStatus(id, req.Status); err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"code": 40002, "data": nil, "message": "用户不存在"})
		case service.ErrInvalidStatus:
			c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "data": nil, "message": "状态切换失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": nil, "message": "ok"})
}

// parseID 从 URL 参数解析 int64 ID（雪花 ID 精度保护，使用 string 解析）
func parseID(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}
