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

type Dlc struct{ api.Api }

func (e Dlc) GetPage(c *gin.Context) {
	s := service.Dlc{}
	req := dto.DlcGetPageReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	p := actions.GetPermissionFromContext(c)
	list := make([]models.Dlc, 0)
	var count int64
	if err = s.GetPage(&req, p, &list, &count); err != nil { e.Error(500, err, "查询失败"); return }
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

func (e Dlc) Get(c *gin.Context) {
	s := service.Dlc{}
	req := dto.DlcById{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, nil).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	p := actions.GetPermissionFromContext(c)
	var object models.Dlc
	if err = s.Get(&req, p, &object); err != nil { e.Error(500, err, "查询失败"); return }
	e.OK(object, "查询成功")
}

func (e Dlc) Insert(c *gin.Context) {
	s := service.Dlc{}
	req := dto.DlcInsertReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	req.SetCreateBy(user.GetUserId(c))
	if err = s.Insert(&req); err != nil { e.Error(500, err, "创建失败"); return }
	e.OK(nil, "创建成功")
}

func (e Dlc) Update(c *gin.Context) {
	s := service.Dlc{}
	req := dto.DlcUpdateReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	req.SetUpdateBy(user.GetUserId(c))
	p := actions.GetPermissionFromContext(c)
	if err = s.Update(&req, p); err != nil { e.Error(500, err, "更新失败"); return }
	e.OK(nil, "更新成功")
}

func (e Dlc) Delete(c *gin.Context) {
	s := service.Dlc{}
	req := dto.DlcById{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	p := actions.GetPermissionFromContext(c)
	if err = s.Remove(&req, p); err != nil { e.Error(500, err, "删除失败"); return }
	e.OK(nil, "删除成功")
}

// UploadPck 上传 PCK 文件
func (e Dlc) UploadPck(c *gin.Context) {
	s := service.Dlc{}
	req := dto.DlcById{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, nil).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }

	file, err := c.FormFile("file")
	if err != nil { e.Error(400, err, "请上传PCK文件"); return }

	// 保存临时文件
	tmpPath := "/tmp/" + file.Filename
	if err = c.SaveUploadedFile(file, tmpPath); err != nil {
		e.Error(500, err, "文件保存失败")
		return
	}

	p := actions.GetPermissionFromContext(c)
	if err = s.UploadPck(req.Id, tmpPath, p); err != nil {
		e.Error(500, err, "PCK上传失败")
		return
	}
	e.OK(nil, "PCK上传成功")
}
