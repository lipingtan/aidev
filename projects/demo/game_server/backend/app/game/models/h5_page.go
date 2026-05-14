package models

import "go-admin/common/models"

// H5Page H5 嵌入页面
type H5Page struct {
	models.Model
	GameId      int    `json:"gameId" gorm:"not null;index;comment:所属游戏ID"`
	PageKey     string `json:"pageKey" gorm:"size:64;not null;comment:页面标识"`
	Name        string `json:"name" gorm:"size:128;not null;comment:页面名称"`
	PageType    string `json:"pageType" gorm:"size:32;default:'custom';comment:页面类型"`
	Content     string `json:"content" gorm:"type:longtext;comment:页面内容"`
	ExternalUrl string `json:"externalUrl" gorm:"size:256;comment:外部URL"`
	UseExternal int    `json:"useExternal" gorm:"default:2;comment:是否使用外链 1是 2否"`
	Status      int    `json:"status" gorm:"default:1;comment:状态 1启用 2禁用"`
	Remark      string `json:"remark" gorm:"size:256;comment:备注"`

	models.ModelTime
	models.ControlBy
	models.TenantBy
}

func (H5Page) TableName() string { return "game_h5_page" }
