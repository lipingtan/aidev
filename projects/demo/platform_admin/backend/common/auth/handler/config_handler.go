package handler

import (
	"strconv"

	"go-admin/common/auth/model"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
)

// ConfigHandler 系统配置 HTTP Handler
type ConfigHandler struct {
	svc *service.ConfigService
}

// NewConfigHandler 创建 ConfigHandler 实例
func NewConfigHandler(svc *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{svc: svc}
}

// RegisterRoutes 注册系统配置路由
func (h *ConfigHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/configs", h.List)
	rg.GET("/configs/:id", h.GetByID)
	rg.GET("/configs/key/:key", h.GetByKey)
	rg.POST("/configs", h.Create)
	rg.PUT("/configs/:id", h.Update)
	rg.DELETE("/configs/:id", h.Delete)
}

// List 分页查询配置列表
func (h *ConfigHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	params := service.ConfigListParams{
		Page:       page,
		PageSize:   pageSize,
		ConfigName: c.Query("config_name"),
		ConfigKey:  c.Query("config_key"),
	}

	result, err := h.svc.List(params)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, result)
}

// GetByID 获取配置详情
func (h *ConfigHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, &errBadRequest{message: "无效的 ID"})
		return
	}

	cfg, err := h.svc.GetByID(id)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, cfg)
}

// GetByKey 按 key 查询配置
func (h *ConfigHandler) GetByKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		Error(c, &errBadRequest{message: "config_key 不能为空"})
		return
	}

	cfg, err := h.svc.GetByKey(key)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, cfg)
}

// configCreateRequest 创建配置请求体
type configCreateRequest struct {
	ConfigName  string `json:"config_name" binding:"required"`
	ConfigKey   string `json:"config_key" binding:"required"`
	ConfigValue string `json:"config_value"`
	ConfigType  int    `json:"config_type"`
	Remark      string `json:"remark"`
}

// Create 创建配置
func (h *ConfigHandler) Create(c *gin.Context) {
	var req configCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, &errBadRequest{message: "参数错误: " + err.Error()})
		return
	}

	cfg := &model.SysConfig{
		ConfigName:  req.ConfigName,
		ConfigKey:   req.ConfigKey,
		ConfigValue: req.ConfigValue,
		ConfigType:  req.ConfigType,
		Remark:      req.Remark,
	}

	if err := h.svc.Create(cfg); err != nil {
		Error(c, err)
		return
	}
	Success(c, cfg)
}

// configUpdateRequest 更新配置请求体
type configUpdateRequest struct {
	ConfigName  *string `json:"config_name"`
	ConfigKey   *string `json:"config_key"`
	ConfigValue *string `json:"config_value"`
	ConfigType  *int    `json:"config_type"`
	Remark      *string `json:"remark"`
}

// Update 更新配置
func (h *ConfigHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, &errBadRequest{message: "无效的 ID"})
		return
	}

	var req configUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, &errBadRequest{message: "参数错误: " + err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.ConfigName != nil {
		updates["config_name"] = *req.ConfigName
	}
	if req.ConfigKey != nil {
		updates["config_key"] = *req.ConfigKey
	}
	if req.ConfigValue != nil {
		updates["config_value"] = *req.ConfigValue
	}
	if req.ConfigType != nil {
		updates["config_type"] = *req.ConfigType
	}
	if req.Remark != nil {
		updates["remark"] = *req.Remark
	}

	if len(updates) == 0 {
		Error(c, &errBadRequest{message: "没有需要更新的字段"})
		return
	}

	if err := h.svc.Update(id, updates); err != nil {
		Error(c, err)
		return
	}
	Success(c, gin.H{"id": id})
}

// Delete 删除配置
func (h *ConfigHandler) Delete(c *gin.Context) {
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
