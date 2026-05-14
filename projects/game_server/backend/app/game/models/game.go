package models

import "go-admin/common/models"

// Game 游戏
type Game struct {
	models.Model
	Name        string `json:"name" gorm:"size:128;not null;comment:游戏名称"`
	AppKey      string `json:"appKey" gorm:"size:64;uniqueIndex;not null;comment:游戏标识"`
	AppSecret   string `json:"appSecret" gorm:"size:128;not null;comment:API密钥"`
	Description string `json:"description" gorm:"size:512;comment:游戏描述"`
	Icon        string `json:"icon" gorm:"size:256;comment:游戏图标"`
	Status      int    `json:"status" gorm:"default:1;comment:状态 1启用 2禁用"`
	Version     string `json:"version" gorm:"size:32;default:'1.0.0';comment:游戏版本"`

	models.ModelTime
	models.ControlBy
	models.TenantBy
}

func (Game) TableName() string { return "game" }
