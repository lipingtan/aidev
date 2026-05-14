package models

import (
	"go-admin/common/models"
	"time"
)

// Tenant 租户
type Tenant struct {
	models.Model
	// 租户名称
	Name string `json:"name" gorm:"size:128;not null;uniqueIndex;comment:租户名称"`
	// 租户编码（唯一标识）
	Code string `json:"code" gorm:"size:64;not null;uniqueIndex;comment:租户编码"`
	// 联系人
	Contact string `json:"contact" gorm:"size:64;comment:联系人"`
	// 联系电话
	Phone string `json:"phone" gorm:"size:20;comment:联系电话"`
	// 联系邮箱
	Email string `json:"email" gorm:"size:128;comment:联系邮箱"`
	// 状态：1=启用 2=禁用
	Status int `json:"status" gorm:"default:1;comment:状态 1启用 2禁用"`
	// 备注
	Remark string `json:"remark" gorm:"size:512;comment:备注"`
	// 过期时间（nil=永不过期）
	ExpireAt *time.Time `json:"expireAt" gorm:"comment:过期时间"`

	models.ModelTime
	models.ControlBy
}

func (Tenant) TableName() string {
	return "tenant"
}
