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
)

type H5Page struct{ api.Api }

func (e H5Page) GetPage(c *gin.Context) {
	s := service.H5Page{}
	req := dto.H5PageGetPageReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	p := actions.GetPermissionFromContext(c)
	list := make([]models.H5Page, 0)
	var count int64
	if err = s.GetPage(&req, p, &list, &count); err != nil { e.Error(500, err, "查询失败"); return }
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

func (e H5Page) Get(c *gin.Context) {
	s := service.H5Page{}
	req := dto.H5PageById{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, nil).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	p := actions.GetPermissionFromContext(c)
	var object models.H5Page
	if err = s.Get(&req, p, &object); err != nil { e.Error(500, err, "查询失败"); return }
	e.OK(object, "查询成功")
}

func (e H5Page) Insert(c *gin.Context) {
	s := service.H5Page{}
	req := dto.H5PageInsertReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	req.SetCreateBy(user.GetUserId(c))
	if err = s.Insert(&req); err != nil { e.Error(500, err, "创建失败"); return }
	e.OK(nil, "创建成功")
}

func (e H5Page) Update(c *gin.Context) {
	s := service.H5Page{}
	req := dto.H5PageUpdateReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	req.SetUpdateBy(user.GetUserId(c))
	p := actions.GetPermissionFromContext(c)
	if err = s.Update(&req, p); err != nil { e.Error(500, err, "更新失败"); return }
	e.OK(nil, "更新成功")
}

func (e H5Page) Delete(c *gin.Context) {
	s := service.H5Page{}
	req := dto.H5PageById{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	p := actions.GetPermissionFromContext(c)
	if err = s.Remove(&req, p); err != nil { e.Error(500, err, "删除失败"); return }
	e.OK(nil, "删除成功")
}
