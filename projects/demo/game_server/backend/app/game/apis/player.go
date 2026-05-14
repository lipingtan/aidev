package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk/api"

	"go-admin/app/game/models"
	"go-admin/app/game/service"
	"go-admin/app/game/service/dto"
	"go-admin/common/actions"
)

type Player struct{ api.Api }

func (e Player) GetPage(c *gin.Context) {
	s := service.Player{}
	req := dto.PlayerGetPageReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	p := actions.GetPermissionFromContext(c)
	list := make([]models.Player, 0)
	var count int64
	if err = s.GetPage(&req, p, &list, &count); err != nil { e.Error(500, err, "查询失败"); return }
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

func (e Player) Get(c *gin.Context) {
	s := service.Player{}
	req := dto.PlayerById{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, nil).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	p := actions.GetPermissionFromContext(c)
	var object models.Player
	if err = s.Get(&req, p, &object); err != nil { e.Error(500, err, "查询失败"); return }
	e.OK(object, "查询成功")
}

func (e Player) Ban(c *gin.Context) {
	s := service.Player{}
	req := dto.PlayerBanReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON, nil).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	p := actions.GetPermissionFromContext(c)
	if err = s.Ban(&req, p); err != nil { e.Error(500, err, "操作失败"); return }
	e.OK(nil, "操作成功")
}
