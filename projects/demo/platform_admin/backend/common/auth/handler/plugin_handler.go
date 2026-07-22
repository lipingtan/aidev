package handler

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"go-admin/app/plugin/models"
	"go-admin/common/plugin"

	"github.com/gin-gonic/gin"
)

// PluginHandler 插件管理 Handler
type PluginHandler struct {
	manager   *plugin.PluginManager
	installer *plugin.Installer
}

// NewPluginHandler 创建插件管理 Handler
func NewPluginHandler(mgr *plugin.PluginManager, ins *plugin.Installer) *PluginHandler {
	return &PluginHandler{
		manager:   mgr,
		installer: ins,
	}
}

// RegisterRoutes 注册插件管理路由
func (h *PluginHandler) RegisterRoutes(rg *gin.RouterGroup) {
	plugins := rg.Group("/plugins")
	{
		plugins.GET("", h.List)
		plugins.POST("/upload", h.Upload)
		plugins.POST("/:name/start", h.Start)
		plugins.POST("/:name/stop", h.Stop)
		plugins.DELETE("/:name", h.Uninstall)
		plugins.PUT("/:name/upgrade", h.Upgrade)
		plugins.GET("/:name/health", h.Health)
	}
}

// pluginListItem 插件列表响应项
type pluginListItem struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	Description  string `json:"description"`
	Status       int    `json:"status"`
	BinaryPath   string `json:"binaryPath"`
	FrontendPath string `json:"frontendPath"`
	RunStatus    string `json:"runStatus"` // running / stopped / unknown
}

// List 获取插件列表
// GET /api/v1/admin/plugins
func (h *PluginHandler) List(c *gin.Context) {
	records, err := h.installer.ListPluginRecords()
	if err != nil {
		Error(c, err)
		return
	}

	items := make([]pluginListItem, 0, len(records))
	for _, r := range records {
		item := pluginListItem{
			ID:           r.ID,
			Name:         r.Name,
			Version:      r.Version,
			Description:  r.Description,
			Status:       r.Status,
			BinaryPath:   r.BinaryPath,
			FrontendPath: r.FrontendPath,
			RunStatus:    "unknown",
		}
		// 附加运行时状态
		if inst, ok := h.manager.GetPlugin(r.Name); ok {
			switch inst.Status {
			case plugin.StatusRunning:
				item.RunStatus = "running"
			case plugin.StatusStopped:
				item.RunStatus = "stopped"
			default:
				item.RunStatus = "error"
			}
		} else {
			item.RunStatus = "stopped"
		}
		items = append(items, item)
	}

	Success(c, gin.H{"list": items, "total": len(items)})
}

// Upload 上传安装插件
// POST /api/v1/admin/plugins/upload (multipart/form-data)
func (h *PluginHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		Error(c, &errBadRequest{message: "文件上传失败: " + err.Error()})
		return
	}
	defer file.Close()

	if err := h.installer.InstallFromFile(context.Background(), "", file, header.Filename); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"message": "安装成功"})
}

// Start 启动插件
// POST /api/v1/admin/plugins/:name/start
func (h *PluginHandler) Start(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		Error(c, &errBadRequest{message: "插件名称不能为空"})
		return
	}

	// 获取插件记录以拿到 binaryPath
	record, err := h.installer.GetPluginRecord(name)
	if err != nil {
		Error(c, &errBadRequest{message: "插件不存在: " + name})
		return
	}

	// 启动子进程
	if err := h.manager.StartProcess(name, record.BinaryPath); err != nil {
		Error(c, err)
		return
	}

	// 更新数据库状态
	_ = h.installer.UpdateStatus(name, models.PluginStatusRunning)

	Success(c, gin.H{"message": "启动成功", "name": name})
}

// Stop 停止插件
// POST /api/v1/admin/plugins/:name/stop
func (h *PluginHandler) Stop(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		Error(c, &errBadRequest{message: "插件名称不能为空"})
		return
	}

	if err := h.manager.Stop(name); err != nil {
		Error(c, err)
		return
	}

	_ = h.installer.UpdateStatus(name, models.PluginStatusStopped)

	Success(c, gin.H{"message": "停止成功", "name": name})
}

// Uninstall 卸载插件
// DELETE /api/v1/admin/plugins/:name
func (h *PluginHandler) Uninstall(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		Error(c, &errBadRequest{message: "插件名称不能为空"})
		return
	}

	// 如果插件正在运行，先停止
	if inst, ok := h.manager.GetPlugin(name); ok && inst.Status == plugin.StatusRunning {
		_ = h.manager.Stop(name)
	}

	if err := h.installer.Uninstall(context.Background(), name, false); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"message": "卸载成功", "name": name})
}

// Upgrade 升级插件
// PUT /api/v1/admin/plugins/:name/upgrade (multipart/form-data)
// 支持 confirm 参数进行二次确认
func (h *PluginHandler) Upgrade(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		Error(c, &errBadRequest{message: "插件名称不能为空"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		Error(c, &errBadRequest{message: "文件上传失败: " + err.Error()})
		return
	}
	defer file.Close()

	// 解析 confirm 参数
	confirm := c.PostForm("confirm") == "true"

	// 如果插件正在运行，先停止
	if inst, ok := h.manager.GetPlugin(name); ok && inst.Status == plugin.StatusRunning {
		if stopErr := h.manager.Stop(name); stopErr != nil {
			Error(c, stopErr)
			return
		}
		_ = h.installer.UpdateStatus(name, models.PluginStatusStopped)
	}

	// 执行升级
	result, err := h.installer.Upgrade(context.Background(), name, file, header.Filename, confirm)
	if err != nil {
		Error(c, err)
		return
	}

	// 需要二次确认，直接返回
	if result.NeedConfirm {
		c.JSON(http.StatusOK, Response{
			Code: 0,
			Data: gin.H{
				"needConfirm":    true,
				"migrationNotes": result.MigrationNotes,
				"oldVersion":     result.OldVersion,
				"newVersion":     result.NewVersion,
			},
			Message: "升级需要确认",
		})
		return
	}

	// 升级成功，重新启动插件
	record, recErr := h.installer.GetPluginRecord(name)
	if recErr != nil {
		Error(c, recErr)
		return
	}

	startErr := h.manager.StartProcess(name, record.BinaryPath)
	if startErr != nil {
		// 启动失败 → 回滚
		_ = h.installer.Rollback(name)
		Error(c, startErr)
		return
	}

	// 启动成功 → 清理 .bak 目录
	_ = h.installer.UpdateStatus(name, models.PluginStatusRunning)
	// installer.Upgrade 将旧目录命名为 pluginsDir/{name}.bak
	// binaryPath 格式为 pluginsDir/{name}/{binary}，取父目录即为 pluginsDir/{name}
	pluginDir := filepath.Dir(record.BinaryPath)
	bakDir := pluginDir + ".bak"
	_ = os.RemoveAll(bakDir)

	Success(c, gin.H{
		"message":    "升级成功",
		"oldVersion": result.OldVersion,
		"newVersion": result.NewVersion,
	})
}

// Health 插件健康检查
// GET /api/v1/admin/plugins/:name/health
func (h *PluginHandler) Health(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		Error(c, &errBadRequest{message: "插件名称不能为空"})
		return
	}

	resp, err := h.manager.Healthcheck(name)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"name":    name,
		"healthy": resp.Healthy,
		"message": resp.Message,
	})
}
