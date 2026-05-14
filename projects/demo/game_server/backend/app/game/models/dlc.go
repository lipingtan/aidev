package models

import "go-admin/common/models"

// Dlc DLC 包
type Dlc struct {
	models.Model
	GameId         int    `json:"gameId" gorm:"not null;index;comment:所属游戏ID"`
	DlcKey         string `json:"dlcKey" gorm:"size:64;not null;comment:DLC标识"`
	Name           string `json:"name" gorm:"size:128;not null;comment:DLC名称"`
	Version        string `json:"version" gorm:"size:32;default:'1.0.0';comment:版本号"`
	Description    string `json:"description" gorm:"size:512;comment:描述"`
	FilePath       string `json:"filePath" gorm:"size:256;comment:PCK文件路径"`
	FileSize       int64  `json:"fileSize" gorm:"default:0;comment:文件大小字节"`
	Sha256         string `json:"sha256" gorm:"size:64;comment:SHA256哈希"`
	Price          int    `json:"price" gorm:"default:0;comment:价格分"`
	IsFree         int    `json:"isFree" gorm:"default:1;comment:是否免费 1是 2否"`
	Status         int    `json:"status" gorm:"default:1;comment:状态 1上架 2下架"`
	DownloadCount  int    `json:"downloadCount" gorm:"default:0;comment:下载次数"`
	MinGameVersion string `json:"minGameVersion" gorm:"size:32;default:'0.0.0';comment:最低游戏版本"`

	models.ModelTime
	models.ControlBy
	models.TenantBy
}

func (Dlc) TableName() string { return "game_dlc" }
