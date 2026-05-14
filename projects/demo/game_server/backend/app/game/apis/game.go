package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"

	"go-admin/app/game/models"
	"go-admin/app/game/service"
	"go-admin/app/game/service/dto"
	"go-admin/common/actions"
	"go-admin/common/middleware"
)

type Game struct {
	api.Api
}

// GetPage 游戏列表
// @Summary 游戏列表
// @Tags 游戏管理
// @Param name query string false "游戏名称"
// @Param status query int false "状态"
// @Success 200 {object} response.Response
// @Router /api/v1/game [get]
// @Security Bearer
func (e Game) GetPage(c *gin.Context) {
	s := service.Game{}
	req := dto.GameGetPageReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	// 注入租户 ID
	req.TenantId = middleware.GetTenantId(c)
	p := actions.GetPermissionFromContext(c)
	list := make([]models.Game, 0)
	var count int64
	if err = s.GetPage(&req, p, &list, &count); err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get 获取游戏详情
// @Summary 获取游戏详情
// @Tags 游戏管理
// @Param id path int true "游戏ID"
// @Success 200 {object} response.Response
// @Router /api/v1/game/{id} [get]
// @Security Bearer
func (e Game) Get(c *gin.Context) {
	s := service.Game{}
	req := dto.GameById{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, nil).MakeService(&s.Service).Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	var object models.Game
	if err = s.Get(&req, p, &object); err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	// 隐藏 AppSecret
	object.AppSecret = "***"
	e.OK(object, "查询成功")
}

// Insert 创建游戏
// @Summary 创建游戏
// @Tags 游戏管理
// @Accept json
// @Param data body dto.GameInsertReq true "游戏数据"
// @Success 200 {object} response.Response
// @Router /api/v1/game [post]
// @Security Bearer
func (e Game) Insert(c *gin.Context) {
	s := service.Game{}
	req := dto.GameInsertReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.SetCreateBy(user.GetUserId(c))
	req.TenantId = middleware.GetTenantId(c)
	if err = s.Insert(&req); err != nil {
		e.Error(500, err, "创建失败")
		return
	}
	e.OK(nil, "创建成功")
}

// Update 更新游戏
// @Summary 更新游戏
// @Tags 游戏管理
// @Accept json
// @Param id path int true "游戏ID"
// @Param data body dto.GameUpdateReq true "游戏数据"
// @Success 200 {object} response.Response
// @Router /api/v1/game/{id} [put]
// @Security Bearer
func (e Game) Update(c *gin.Context) {
	s := service.Game{}
	req := dto.GameUpdateReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	req.SetUpdateBy(user.GetUserId(c))
	p := actions.GetPermissionFromContext(c)
	if err = s.Update(&req, p); err != nil {
		e.Error(500, err, "更新失败")
		return
	}
	e.OK(nil, "更新成功")
}

// Delete 删除游戏
// @Summary 删除游戏
// @Tags 游戏管理
// @Param id path int true "游戏ID"
// @Success 200 {object} response.Response
// @Router /api/v1/game/{id} [delete]
// @Security Bearer
func (e Game) Delete(c *gin.Context) {
	s := service.Game{}
	req := dto.GameById{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	if err = s.Remove(&req, p); err != nil {
		e.Error(500, err, "删除失败")
		return
	}
	e.OK(nil, "删除成功")
}

// RegenerateSecret 重新生成 AppSecret
// @Summary 重新生成 AppSecret
// @Tags 游戏管理
// @Param id path int true "游戏ID"
// @Success 200 {object} response.Response
// @Router /api/v1/game/{id}/secret [post]
// @Security Bearer
func (e Game) RegenerateSecret(c *gin.Context) {
	s := service.Game{}
	req := dto.GameById{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, nil).MakeService(&s.Service).Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	secret, err := s.RegenerateSecret(&req, p)
	if err != nil {
		e.Error(500, err, "生成失败")
		return
	}
	e.OK(gin.H{"appSecret": secret}, "生成成功，请妥善保存")
}
