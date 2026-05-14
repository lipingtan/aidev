package service

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/app/game/models"
	"go-admin/app/game/service/dto"
	"go-admin/common/actions"
)

type Dlc struct {
	service.Service
}

func (e *Dlc) GetPage(c *dto.DlcGetPageReq, p *actions.DataPermission, list *[]models.Dlc, count *int64) error {
	var data models.Dlc
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

func (e *Dlc) Get(c *dto.DlcById, p *actions.DataPermission, model *models.Dlc) error {
	var data models.Dlc
	err := e.Orm.Model(&data).
		Scopes(actions.Permission(data.TableName(), p)).
		First(model, c.GetId()).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("DLC不存在或无权查看")
	}
	return err
}

func (e *Dlc) Insert(c *dto.DlcInsertReq) error {
	var data models.Dlc
	c.Generate(&data)
	return e.Orm.Create(&data).Error
}

func (e *Dlc) Update(c *dto.DlcUpdateReq, p *actions.DataPermission) error {
	var model = models.Dlc{}
	e.Orm.Scopes(actions.Permission(model.TableName(), p)).First(&model, c.GetId())
	c.Generate(&model)
	return e.Orm.Save(&model).Error
}

func (e *Dlc) Remove(c *dto.DlcById, p *actions.DataPermission) error {
	var data models.Dlc
	return e.Orm.Model(&data).
		Scopes(actions.Permission(data.TableName(), p)).
		Delete(&data, c.GetId()).Error
}

// UploadPck 上传 PCK 文件并更新哈希和大小
func (e *Dlc) UploadPck(dlcId int, srcPath string, p *actions.DataPermission) error {
	var dlc models.Dlc
	err := e.Orm.Scopes(actions.Permission(dlc.TableName(), p)).First(&dlc, dlcId).Error
	if err != nil {
		return errors.New("DLC不存在")
	}

	// 目标路径
	destDir := "dlc_files"
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	destPath := filepath.Join(destDir, fmt.Sprintf("%d_%s.pck", dlcId, dlc.DlcKey))

	// 复制文件
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	hasher := sha256.New()
	writer := io.MultiWriter(dst, hasher)
	size, err := io.Copy(writer, src)
	if err != nil {
		return err
	}

	// 更新数据库
	dlc.FilePath = destPath
	dlc.FileSize = size
	dlc.Sha256 = fmt.Sprintf("%x", hasher.Sum(nil))
	return e.Orm.Save(&dlc).Error
}
