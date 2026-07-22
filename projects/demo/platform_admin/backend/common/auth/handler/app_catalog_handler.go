package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go-admin/common/auth/model"
	"go-admin/common/auth/middleware"
	"go-admin/common/plugin"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AppCatalogHandler 应用目录与订阅管理
type AppCatalogHandler struct {
	db      *gorm.DB
	manager *plugin.PluginManager // 获取插件运行状态（可选，nil 则跳过）
}

// NewAppCatalogHandler 创建应用目录 Handler
func NewAppCatalogHandler(db *gorm.DB, mgr *plugin.PluginManager) *AppCatalogHandler {
	return &AppCatalogHandler{db: db, manager: mgr}
}

// RegisterRoutes 注册应用目录相关路由
func (h *AppCatalogHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/app-catalog", h.GetAppCatalog)
	rg.GET("/app-subscriptions", h.GetSubscriptions)
	rg.POST("/app-subscriptions", h.SubscribeApp)
	rg.DELETE("/app-subscriptions/:app_code", h.UnsubscribeApp)
	rg.PUT("/app-subscriptions/:app_code/modules", h.UpdateModules)
}

// ---- 响应 DTO ----

// AppCatalogItem 应用目录项
type AppCatalogItem struct {
	model.Application
	Subscribed    bool   `json:"subscribed"`               // 当前租户是否已订阅
	PluginStatus  string `json:"plugin_status,omitempty"`  // PLUGIN 类型附加运行状态
}

// SubscriptionItem 订阅详情
type SubscriptionItem struct {
	model.TenantApp
	AppName     string `json:"app_name"`
	AppType     string `json:"app_type"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// ---- 请求 DTO ----

type subscribeRequest struct {
	AppCode string `json:"app_code" binding:"required"`
}

type updateModulesRequest struct {
	EnabledModules []string `json:"enabled_modules"`
}

// ---- Handler 方法 ----

// GetAppCatalog 获取应用目录列表（免权限检查，仅需登录）
func (h *AppCatalogHandler) GetAppCatalog(c *gin.Context) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		Error(c, &errBadRequest{message: "未认证"})
		return
	}
	tenantID := authCtx.TenantID

	// 查所有启用应用
	var apps []model.Application
	if err := h.db.Where("status = ?", 1).Order("sort_order ASC, id ASC").Find(&apps).Error; err != nil {
		Error(c, err)
		return
	}

	// 查当前租户已订阅的 app_code 集合
	var tenantApps []model.TenantApp
	h.db.Where("tenant_id = ?", tenantID).Find(&tenantApps)
	subscribedSet := make(map[string]bool, len(tenantApps))
	for _, ta := range tenantApps {
		subscribedSet[ta.AppCode] = true
	}

	// 组装返回
	items := make([]AppCatalogItem, 0, len(apps))
	for _, app := range apps {
		item := AppCatalogItem{
			Application: app,
			Subscribed:  subscribedSet[app.AppCode],
		}
		// PLUGIN 类型附加运行状态
		if app.AppType == "PLUGIN" && h.manager != nil {
			if inst, ok := h.manager.GetPlugin(app.AppCode); ok {
				switch inst.Status {
				case plugin.StatusRunning:
					item.PluginStatus = "RUNNING"
				case plugin.StatusStopped:
					item.PluginStatus = "STOPPED"
				case plugin.StatusError:
					item.PluginStatus = "ERROR"
				default:
					item.PluginStatus = "UNKNOWN"
				}
			} else {
				item.PluginStatus = "NOT_INSTALLED"
			}
		}
		items = append(items, item)
	}

	Success(c, items)
}

// SubscribeApp 订阅应用
func (h *AppCatalogHandler) SubscribeApp(c *gin.Context) {
	authCtx, err := MustGetAuthContext(c)
	if err != nil {
		Error(c, err)
		return
	}
	tenantID := authCtx.TenantID

	var req subscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, &errBadRequest{message: "参数错误: " + err.Error()})
		return
	}

	// 验证应用存在且启用
	var app model.Application
	if err := h.db.Where("app_code = ? AND status = 1", req.AppCode).First(&app).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			Error(c, &errBadRequest{message: "应用不存在或未启用"})
			return
		}
		Error(c, err)
		return
	}

	// 检查是否已订阅
	var existing model.TenantApp
	if err := h.db.Where("tenant_id = ? AND app_code = ?", tenantID, req.AppCode).First(&existing).Error; err == nil {
		Error(c, &errBadRequest{message: "该应用已订阅"})
		return
	}

	// 检查配额
	maxApps := h.getQuotaMaxApps(tenantID)
	var currentCount int64
	h.db.Model(&model.TenantApp{}).Where("tenant_id = ?", tenantID).Count(&currentCount)
	if currentCount >= int64(maxApps) {
		Error(c, &errBadRequest{message: "已达订阅配额上限（最多 " + strconv.Itoa(maxApps) + " 个应用）"})
		return
	}

	// 创建订阅记录
	tenantApp := model.TenantApp{
		TenantID: tenantID,
		AppCode:  req.AppCode,
	}
	if err := h.db.Create(&tenantApp).Error; err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, Response{
		Code:    0,
		Data:    tenantApp,
		Message: "ok",
	})
}

// UnsubscribeApp 退订应用
func (h *AppCatalogHandler) UnsubscribeApp(c *gin.Context) {
	authCtx, err := MustGetAuthContext(c)
	if err != nil {
		Error(c, err)
		return
	}
	tenantID := authCtx.TenantID
	appCode := c.Param("app_code")

	// 查应用类型
	var app model.Application
	if err := h.db.Where("app_code = ?", appCode).First(&app).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			Error(c, &errBadRequest{message: "应用不存在"})
			return
		}
		Error(c, err)
		return
	}

	// BUILTIN 拒绝退订
	if app.AppType == "BUILTIN" {
		Error(c, &errBadRequest{message: "内置应用不允许退订"})
		return
	}

	// 删除订阅记录
	result := h.db.Where("tenant_id = ? AND app_code = ?", tenantID, appCode).Delete(&model.TenantApp{})
	if result.Error != nil {
		Error(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		Error(c, &errBadRequest{message: "未找到订阅记录"})
		return
	}

	// 级联清理角色绑定
	h.cascadeCleanup(tenantID, appCode)

	Success(c, gin.H{"message": "退订成功"})
}

// GetSubscriptions 获取当前租户已订阅列表（免权限检查）
func (h *AppCatalogHandler) GetSubscriptions(c *gin.Context) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		Error(c, &errBadRequest{message: "未认证"})
		return
	}
	tenantID := authCtx.TenantID

	var results []SubscriptionItem
	h.db.Table("admin_tenant_app ta").
		Select("ta.*, app.name AS app_name, app.app_type, app.description, app.icon").
		Joins("JOIN admin_application app ON app.app_code = ta.app_code AND app.deleted_at IS NULL").
		Where("ta.tenant_id = ?", tenantID).
		Order("ta.created_at ASC").
		Find(&results)

	Success(c, results)
}

// UpdateModules 更新订阅应用的启用模块
func (h *AppCatalogHandler) UpdateModules(c *gin.Context) {
	authCtx, err := MustGetAuthContext(c)
	if err != nil {
		Error(c, err)
		return
	}
	tenantID := authCtx.TenantID
	appCode := c.Param("app_code")

	var req updateModulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, &errBadRequest{message: "参数错误: " + err.Error()})
		return
	}

	// 序列化 enabled_modules
	var modulesJSON datatypes.JSON
	if req.EnabledModules != nil {
		data, _ := json.Marshal(req.EnabledModules)
		modulesJSON = datatypes.JSON(data)
	}

	result := h.db.Model(&model.TenantApp{}).
		Where("tenant_id = ? AND app_code = ?", tenantID, appCode).
		Update("enabled_modules", modulesJSON)
	if result.Error != nil {
		Error(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		Error(c, &errBadRequest{message: "未找到订阅记录"})
		return
	}

	Success(c, gin.H{"message": "更新成功"})
}

// ---- 内部方法 ----

// getQuotaMaxApps 获取租户最大应用订阅数配额
func (h *AppCatalogHandler) getQuotaMaxApps(tenantID int64) int {
	var cfg model.AdminConfig
	// 先查租户级配置
	err := h.db.Where("config_key = ? AND scope = ? AND scope_id = ?", "quota.max_apps", "TENANT", tenantID).
		First(&cfg).Error
	if err == nil {
		if v, e := strconv.Atoi(cfg.ConfigValue); e == nil {
			return v
		}
	}
	// 再查系统级配置
	err = h.db.Where("config_key = ? AND scope = ? AND scope_id = 0", "quota.max_apps", "SYSTEM").
		First(&cfg).Error
	if err == nil {
		if v, e := strconv.Atoi(cfg.ConfigValue); e == nil {
			return v
		}
	}
	// 默认值
	return 10
}

// cascadeCleanup 级联清理退订应用的角色绑定
func (h *AppCatalogHandler) cascadeCleanup(tenantID int64, appCode string) {
	// 查出该 app_code 的资源 IDs
	var resourceIDs []int64
	h.db.Model(&model.Resource{}).Where("app_code = ?", appCode).Pluck("id", &resourceIDs)

	// 查出该 app_code 的 API 权限 IDs
	var apiIDs []int64
	h.db.Model(&model.ApiPermission{}).Where("app_code = ?", appCode).Pluck("id", &apiIDs)

	// 查出该租户的角色 IDs
	var roleIDs []int64
	h.db.Model(&model.Role{}).Where("tenant_id = ?", tenantID).Pluck("id", &roleIDs)

	if len(roleIDs) == 0 {
		return
	}

	// 删除 admin_role_resource
	if len(resourceIDs) > 0 {
		h.db.Where("role_id IN ? AND resource_id IN ?", roleIDs, resourceIDs).
			Delete(&model.RoleResource{})
	}

	// 删除 admin_role_api
	if len(apiIDs) > 0 {
		h.db.Where("role_id IN ? AND api_permission_id IN ?", roleIDs, apiIDs).
			Delete(&model.RoleApi{})
	}
}
