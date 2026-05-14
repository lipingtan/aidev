package service

import (
	"github.com/go-admin-team/go-admin-core/sdk/service"

	"go-admin/app/game/models"
	"go-admin/app/game/service/dto"
	"go-admin/common/actions"
)

type PaymentConfig struct {
	service.Service
}

func (e *PaymentConfig) GetPage(c *dto.PaymentConfigGetPageReq, p *actions.DataPermission, list *[]models.PaymentConfig, count *int64) error {
	var data models.PaymentConfig
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

// Save 创建或更新支付配置（upsert）
func (e *PaymentConfig) Save(c *dto.PaymentConfigSaveReq, p *actions.DataPermission) error {
	var model models.PaymentConfig

	// 查找是否已存在同游戏同渠道的配置
	e.Orm.Where("game_id = ? AND channel = ?", c.GameId, c.Channel).First(&model)

	c.Generate(&model)

	if model.Id == 0 {
		return e.Orm.Create(&model).Error
	}
	return e.Orm.Save(&model).Error
}

// GetByGameAndChannel 获取指定游戏和渠道的配置
func (e *PaymentConfig) GetByGameAndChannel(gameId int, channel string) (*models.PaymentConfig, error) {
	var config models.PaymentConfig
	err := e.Orm.Where("game_id = ? AND channel = ? AND enabled = 1", gameId, channel).First(&config).Error
	if err != nil {
		// 尝试全局配置
		err = e.Orm.Where("game_id = 0 AND channel = ? AND enabled = 1", channel).First(&config).Error
	}
	return &config, err
}
