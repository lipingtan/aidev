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

type Order struct{ api.Api }

func (e Order) GetPage(c *gin.Context) {
	s := service.Order{}
	req := dto.OrderGetPageReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	p := actions.GetPermissionFromContext(c)
	list := make([]models.Order, 0)
	var count int64
	if err = s.GetPage(&req, p, &list, &count); err != nil { e.Error(500, err, "查询失败"); return }
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

func (e Order) Get(c *gin.Context) {
	s := service.Order{}
	req := dto.OrderById{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, nil).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	p := actions.GetPermissionFromContext(c)
	var object models.Order
	if err = s.Get(&req, p, &object); err != nil { e.Error(500, err, "查询失败"); return }
	e.OK(object, "查询成功")
}

func (e Order) Refund(c *gin.Context) {
	s := service.Order{}
	req := dto.OrderRefundReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON, nil).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	p := actions.GetPermissionFromContext(c)
	if err = s.Refund(&req, p); err != nil { e.Error(500, err, "退款失败"); return }
	e.OK(nil, "退款成功")
}
