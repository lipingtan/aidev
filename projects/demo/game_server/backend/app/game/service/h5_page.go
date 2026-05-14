package service

import (
	"errors"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/app/game/models"
	"go-admin/app/game/service/dto"
	"go-admin/common/actions"
)

type H5Page struct {
	service.Service
}

func (e *H5Page) GetPage(c *dto.H5PageGetPageReq, p *actions.DataPermission, list *[]models.H5Page, count *int64) error {
	var data models.H5Page
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

func (e *H5Page) Get(c *dto.H5PageById, p *actions.DataPermission, model *models.H5Page) error {
	var data models.H5Page
	err := e.Orm.Model(&data).
		Scopes(actions.Permission(data.TableName(), p)).
		First(model, c.GetId()).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("H5页面不存在")
	}
	return err
}

func (e *H5Page) Insert(c *dto.H5PageInsertReq) error {
	var data models.H5Page
	c.Generate(&data)
	return e.Orm.Create(&data).Error
}

func (e *H5Page) Update(c *dto.H5PageUpdateReq, p *actions.DataPermission) error {
	var model = models.H5Page{}
	e.Orm.Scopes(actions.Permission(model.TableName(), p)).First(&model, c.GetId())
	c.Generate(&model)
	return e.Orm.Save(&model).Error
}

func (e *H5Page) Remove(c *dto.H5PageById, p *actions.DataPermission) error {
	var data models.H5Page
	return e.Orm.Model(&data).
		Scopes(actions.Permission(data.TableName(), p)).
		Delete(&data, c.GetId()).Error
}

// GetByKey 通过 pageKey 获取 H5 页面（游戏客户端调用）
func (e *H5Page) GetByKey(gameId int, pageKey string) (*models.H5Page, error) {
	var page models.H5Page
	err := e.Orm.Where("game_id = ? AND page_key = ? AND status = 1", gameId, pageKey).First(&page).Error
	return &page, err
}
