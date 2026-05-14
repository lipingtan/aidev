package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/app/game/models"
	"go-admin/app/game/service/dto"
	"go-admin/common/actions"
)

type Game struct {
	service.Service
}

// GetPage 获取游戏列表
func (e *Game) GetPage(c *dto.GameGetPageReq, p *actions.DataPermission, list *[]models.Game, count *int64) error {
	var err error
	var data models.Game

	db := e.Orm.Model(&data).
		Scopes(actions.Permission(data.TableName(), p)).
		Where("name like ?", "%"+c.Name+"%")

	// 租户隔离：非超级管理员只能看自己租户的数据
	if c.TenantId > 0 {
		db = db.Where("tenant_id = ?", c.TenantId)
	}

	err = db.Find(list).Limit(-1).Offset(-1).Count(count).Error
	if err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	return nil
}

// Get 获取单个游戏
func (e *Game) Get(c *dto.GameById, p *actions.DataPermission, model *models.Game) error {
	var data models.Game
	err := e.Orm.Model(&data).
		Scopes(actions.Permission(data.TableName(), p)).
		First(model, c.GetId()).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = errors.New("查看对象不存在或无权查看")
		e.Log.Errorf("db error: %s", err)
		return err
	}
	if err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	return nil
}

// Insert 创建游戏
func (e *Game) Insert(c *dto.GameInsertReq) error {
	var err error
	var data models.Game
	c.Generate(&data)

	// 自动生成 AppSecret
	if data.AppSecret == "" {
		secret, err := generateSecret()
		if err != nil {
			return err
		}
		data.AppSecret = secret
	}

	err = e.Orm.Create(&data).Error
	if err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	return nil
}

// Update 更新游戏
func (e *Game) Update(c *dto.GameUpdateReq, p *actions.DataPermission) error {
	var err error
	var model = models.Game{}
	e.Orm.Scopes(actions.Permission(model.TableName(), p)).
		First(&model, c.GetId())
	c.Generate(&model)
	db := e.Orm.Save(&model)
	if err = db.Error; err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	return nil
}

// Remove 删除游戏
func (e *Game) Remove(c *dto.GameById, p *actions.DataPermission) error {
	var err error
	var data models.Game
	db := e.Orm.Model(&data).
		Scopes(actions.Permission(data.TableName(), p)).
		Delete(&data, c.GetId())
	if err = db.Error; err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	return nil
}

// RegenerateSecret 重新生成 AppSecret
func (e *Game) RegenerateSecret(c *dto.GameById, p *actions.DataPermission) (string, error) {
	var model models.Game
	e.Orm.Scopes(actions.Permission(model.TableName(), p)).First(&model, c.GetId())
	if model.Id == 0 {
		return "", errors.New("游戏不存在")
	}
	secret, err := generateSecret()
	if err != nil {
		return "", err
	}
	model.AppSecret = secret
	e.Orm.Save(&model)
	return secret, nil
}

// generateSecret 生成随机 32 字节 hex 密钥
func generateSecret() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
