package models

import "go-admin/common/models"

// Save 玩家存档
type Save struct {
	models.Model
	GameId      int    `json:"gameId" gorm:"not null;index;comment:所属游戏ID"`
	PlayerId    int    `json:"playerId" gorm:"not null;index;comment:玩家ID"`
	Slot        int    `json:"slot" gorm:"default:0;comment:存档槽位"`
	SaveName    string `json:"saveName" gorm:"size:64;comment:存档名称"`
	Data        string `json:"data" gorm:"type:longtext;comment:存档数据JSON"`
	DataVersion string `json:"dataVersion" gorm:"size:32;comment:存档数据版本"`
	PlayTime    int64  `json:"playTime" gorm:"default:0;comment:游戏内时间秒"`
	Screenshot  string `json:"screenshot" gorm:"size:256;comment:存档截图"`

	models.ModelTime
	models.ControlBy
	models.TenantBy
}

func (Save) TableName() string { return "game_save" }
