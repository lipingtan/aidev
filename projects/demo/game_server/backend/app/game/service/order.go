package service

import (
	"errors"
	"time"

	"github.com/go-admin-team/go-admin-core/sdk/service"

	"go-admin/app/game/models"
	"go-admin/app/game/service/dto"
	"go-admin/common/actions"
)

type Order struct {
	service.Service
}

func (e *Order) GetPage(c *dto.OrderGetPageReq, p *actions.DataPermission, list *[]models.Order, count *int64) error {
	var data models.Order
	err := e.Orm.Model(&data).
		Scopes(actions.Permission(data.TableName(), p)).
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error
	if err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	return nil
}

func (e *Order) Get(c *dto.OrderById, p *actions.DataPermission, model *models.Order) error {
	var data models.Order
	return e.Orm.Model(&data).
		Scopes(actions.Permission(data.TableName(), p)).
		First(model, c.GetId()).Error
}

// Refund 退款
func (e *Order) Refund(c *dto.OrderRefundReq, p *actions.DataPermission) error {
	var model models.Order
	err := e.Orm.Scopes(actions.Permission(model.TableName(), p)).First(&model, c.Id).Error
	if err != nil {
		return errors.New("订单不存在")
	}
	if model.Status != "paid" {
		return errors.New("只有已支付的订单才能退款")
	}
	now := time.Now()
	model.Status = "refunded"
	model.RefundedAt = &now
	model.RefundReason = c.RefundReason
	return e.Orm.Save(&model).Error
}
