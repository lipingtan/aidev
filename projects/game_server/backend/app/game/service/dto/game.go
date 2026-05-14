package dto

import (
	"go-admin/app/game/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
)

// GameGetPageReq 游戏列表查询
type GameGetPageReq struct {
	dto.Pagination `search:"-"`
	Name           string `form:"name" search:"type:contains;column:name;table:game" comment:"游戏名称"`
	Status         int    `form:"status" search:"type:exact;column:status;table:game" comment:"状态"`
	TenantId       int    `form:"-" search:"-"` // 从 context 注入，不从请求参数读取
}

func (m *GameGetPageReq) GetNeedSearch() interface{} {
	return *m
}

// GameInsertReq 创建游戏
type GameInsertReq struct {
	Name        string `json:"name" vd:"len($)>0;msg:'游戏名称不能为空'"`
	AppKey      string `json:"appKey" vd:"len($)>0;msg:'游戏标识不能为空'"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Version     string `json:"version"`
	Status      int    `json:"status"`
	TenantId    int    `json:"-"` // 从 context 注入
	common.ControlBy
}

func (s *GameInsertReq) Generate(m *models.Game) {
	m.Name = s.Name
	m.AppKey = s.AppKey
	m.Description = s.Description
	m.Icon = s.Icon
	m.Version = s.Version
	m.Status = s.Status
	m.CreateBy = s.CreateBy
	m.UpdateBy = s.UpdateBy
	m.TenantId = s.TenantId
}

func (s *GameInsertReq) GetId() interface{} {
	return nil
}

// GameUpdateReq 更新游戏
type GameUpdateReq struct {
	Id          int    `uri:"id" json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Version     string `json:"version"`
	Status      int    `json:"status"`
	common.ControlBy
}

func (s *GameUpdateReq) Generate(m *models.Game) {
	m.Name = s.Name
	m.Description = s.Description
	m.Icon = s.Icon
	m.Version = s.Version
	m.Status = s.Status
	m.UpdateBy = s.UpdateBy
}

func (s *GameUpdateReq) GetId() interface{} {
	return s.Id
}

// GameById 按 ID 查询
type GameById struct {
	dto.ObjectById
}
