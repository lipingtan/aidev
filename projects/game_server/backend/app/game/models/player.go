package models

import (
	"time"
	"go-admin/common/models"
)

// Player 游戏玩家
type Player struct {
	models.Model
	GameId      int        `json:"gameId" gorm:"not null;index;comment:所属游戏ID"`
	Uid         string     `json:"uid" gorm:"size:64;not null;index;comment:玩家UID"`
	Nickname    string     `json:"nickname" gorm:"size:64;comment:玩家昵称"`
	Email       string     `json:"email" gorm:"size:128;comment:邮箱"`
	Platform    string     `json:"platform" gorm:"size:32;comment:登录平台"`
	PlatformId  string     `json:"platformId" gorm:"size:128;comment:平台ID"`
	DeviceId    string     `json:"deviceId" gorm:"size:128;comment:设备ID"`
	Status      int        `json:"status" gorm:"default:1;comment:状态 1正常 2封禁"`
	BanReason   string     `json:"banReason" gorm:"size:256;comment:封禁原因"`
	LastLoginAt *time.Time `json:"lastLoginAt" gorm:"comment:最后登录时间"`
	LastLoginIp string     `json:"lastLoginIp" gorm:"size:64;comment:最后登录IP"`

	models.ModelTime
	models.ControlBy
	models.TenantBy
}

func (Player) TableName() string { return "game_player" }
