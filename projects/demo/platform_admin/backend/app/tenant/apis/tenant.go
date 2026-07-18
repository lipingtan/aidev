package apis

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go-admin/app/tenant/models"
	"gorm.io/gorm"
)

type TenantApi struct{}

// GetPage 租户列表（仅超级管理员）
func (TenantApi) GetPage(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var list []models.Tenant
	var count int64
	pageIndex := 1
	pageSize := 10

	db.Model(&models.Tenant{}).Count(&count)
	db.Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&list)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{"list": list, "count": count, "pageIndex": pageIndex, "pageSize": pageSize},
		"msg":  "查询成功",
	})
}

// Get 获取租户详情
func (TenantApi) Get(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")
	var obj models.Tenant
	if err := db.First(&obj, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "msg": "租户不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": obj, "msg": "查询成功"})
}

// Insert 创建租户
func (TenantApi) Insert(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var obj models.Tenant
	if err := c.ShouldBindJSON(&obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	if err := db.Create(&obj).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "创建失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": obj.Id, "msg": "创建成功"})
}

// Update 更新租户
func (TenantApi) Update(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")
	var obj models.Tenant
	if err := db.First(&obj, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "msg": "租户不存在"})
		return
	}
	var req models.Tenant
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	if err := db.Model(&obj).Updates(&req).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "更新失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

// Delete 删除租户
func (TenantApi) Delete(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")
	if err := db.Delete(&models.Tenant{}, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}
