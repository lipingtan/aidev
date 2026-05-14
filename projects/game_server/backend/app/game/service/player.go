package service

import (
	"errors"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/app/game/models"
	"go-admin/app/game/service/dto"
	"go-admin/common/actions"
)

type Player struct {
	service.Service
}

func (e *Player) GetPage(c *dto.PlayerGetPageReq, p *actions.DataPermission, list *[]models.Player, count *int64) error {
	var data models.Player
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

func (e *Player) Get(c *dto.PlayerById, p *actions.DataPermission, model *models.Player) error {
	var data models.Player
	err := e.Orm.Model(&data).
		Scopes(actions.Permission(data.TableName(), p)).
		First(model, c.GetId()).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("玩家不存在")
	}
	return err
}

// Ban 封禁/解封玩家
func (e *Player) Ban(c *dto.PlayerBanReq, p *actions.DataPermission) error {
	var model models.Player
	err := e.Orm.Scopes(actions.Permission(model.TableName(), p)).First(&model, c.Id).Error
	if err != nil {
		return errors.New("玩家不存在")
	}
	model.Status = c.Status
	model.BanReason = c.BanReason
	return e.Orm.Save(&model).Error
}
