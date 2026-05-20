package apis

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"go-admin/app/plugin/models"
	"go-admin/app/plugin/service"
)

// Plugin 插件管理 API handler
type Plugin struct{}

// InstallRequest 通过 URL 安装插件的请求体
type InstallRequest struct {
	Name string `json:"name" binding:"required"`
	URL  string `json:"url" binding:"required"`
}

// List 获取插件列表
// GET /api/v1/plugins
func (p Plugin) List(c *gin.Context) {
	records, err := service.Installer.ListPluginRecords()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  "获取插件列表失败: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": records,
	})
}

// Install 安装插件
// POST /api/v1/plugins/install
// 支持两种方式：
//   - multipart/form-data 上传文件（field: "file", "name"）
//   - JSON body（{"name": "xxx", "url": "https://..."}）
func (p Plugin) Install(c *gin.Context) {
	contentType := c.GetHeader("Content-Type")

	// multipart/form-data 方式：上传文件安装
	if strings.Contains(contentType, "multipart/form-data") {
		name := c.PostForm("name")
		if name == "" {
			c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "插件名称不能为空"})
			return
		}
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "获取上传文件失败: " + err.Error()})
			return
		}
		defer file.Close()

		if err = service.Installer.InstallFromFile(c.Request.Context(), name, file, header.Filename); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "安装插件失败: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "插件安装成功", "data": gin.H{"name": name}})
		return
	}

	// JSON body 方式：从 URL 安装
	var req InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	if err := service.Installer.InstallFromURL(c.Request.Context(), req.Name, req.URL); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "安装插件失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "插件安装成功", "data": gin.H{"name": req.Name}})
}

// Start 启动插件
// POST /api/v1/plugins/:name/start
func (p Plugin) Start(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "插件名称不能为空"})
		return
	}

	if err := service.Manager.Start(name); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "启动插件失败: " + err.Error()})
		return
	}

	// 更新数据库状态为运行中
	if err := service.Installer.UpdateStatus(name, models.PluginStatusRunning); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "更新插件状态失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "插件启动成功", "data": gin.H{"name": name}})
}

// Stop 停止插件
// POST /api/v1/plugins/:name/stop
func (p Plugin) Stop(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "插件名称不能为空"})
		return
	}

	if err := service.Manager.Stop(name); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "停止插件失败: " + err.Error()})
		return
	}

	// 更新数据库状态为已停止
	if err := service.Installer.UpdateStatus(name, models.PluginStatusStopped); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "更新插件状态失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "插件停止成功", "data": gin.H{"name": name}})
}

// Uninstall 卸载插件
// DELETE /api/v1/plugins/:name
// Query param: clean_data=true 清除插件数据表
func (p Plugin) Uninstall(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "插件名称不能为空"})
		return
	}

	// 如果插件正在运行，先停止
	if inst, exists := service.Manager.GetPlugin(name); exists && inst.Status == 1 {
		_ = service.Manager.Stop(name)
	}

	cleanData := c.Query("clean_data") == "true"
	if err := service.Installer.Uninstall(c.Request.Context(), name, cleanData); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "卸载插件失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "插件卸载成功", "data": gin.H{"name": name}})
}

// Health 插件健康检查
// GET /api/v1/plugins/:name/health
func (p Plugin) Health(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "插件名称不能为空"})
		return
	}

	resp, err := service.Manager.Healthcheck(name)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "健康检查失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{
			"healthy": resp.Healthy,
			"message": resp.Message,
		},
	})
}
