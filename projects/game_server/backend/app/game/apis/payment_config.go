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

type PaymentConfig struct{ api.Api }

func (e PaymentConfig) GetPage(c *gin.Context) {
	s := service.PaymentConfig{}
	req := dto.PaymentConfigGetPageReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	p := actions.GetPermissionFromContext(c)
	list := make([]models.PaymentConfig, 0)
	var count int64
	if err = s.GetPage(&req, p, &list, &count); err != nil { e.Error(500, err, "查询失败"); return }
	// 隐藏敏感密钥
	for i := range list {
		list[i].StripeSecretKey = "***"
		list[i].AlipayPrivateKey = "***"
		list[i].WechatApiKey = "***"
	}
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

func (e PaymentConfig) Save(c *gin.Context) {
	s := service.PaymentConfig{}
	req := dto.PaymentConfigSaveReq{}
	err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors
	if err != nil { e.Error(500, err, err.Error()); return }
	req.SetCreateBy(user.GetUserId(c))
	req.SetUpdateBy(user.GetUserId(c))
	p := actions.GetPermissionFromContext(c)
	if err = s.Save(&req, p); err != nil { e.Error(500, err, "保存失败"); return }
	e.OK(nil, "保存成功")
}
