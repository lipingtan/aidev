package dto

import (
	"go-admin/app/game/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
)

type DlcGetPageReq struct {
	dto.Pagination `search:"-"`
	GameId         int    `form:"gameId" search:"type:exact;column:game_id;table:game_dlc" comment:"游戏ID"`
	Name           string `form:"name" search:"type:contains;column:name;table:game_dlc" comment:"DLC名称"`
	Status         int    `form:"status" search:"type:exact;column:status;table:game_dlc" comment:"状态"`
}

func (m *DlcGetPageReq) GetNeedSearch() interface{} {
	return *m
}

type DlcInsertReq struct {
	GameId         int    `json:"gameId" vd:"$>0;msg:'游戏ID不能为空'"`
	DlcKey         string `json:"dlcKey" vd:"len($)>0;msg:'DLC标识不能为空'"`
	Name           string `json:"name" vd:"len($)>0;msg:'DLC名称不能为空'"`
	Version        string `json:"version"`
	Description    string `json:"description"`
	Price          int    `json:"price"`
	IsFree         int    `json:"isFree"`
	Status         int    `json:"status"`
	MinGameVersion string `json:"minGameVersion"`
	common.ControlBy
}

func (s *DlcInsertReq) Generate(m *models.Dlc) {
	m.GameId = s.GameId
	m.DlcKey = s.DlcKey
	m.Name = s.Name
	m.Version = s.Version
	m.Description = s.Description
	m.Price = s.Price
	m.IsFree = s.IsFree
	m.Status = s.Status
	m.MinGameVersion = s.MinGameVersion
	m.CreateBy = s.CreateBy
	m.UpdateBy = s.UpdateBy
}

func (s *DlcInsertReq) GetId() interface{} {
	return nil
}

type DlcUpdateReq struct {
	Id             int    `uri:"id" json:"-"`
	Name           string `json:"name"`
	Version        string `json:"version"`
	Description    string `json:"description"`
	Price          int    `json:"price"`
	IsFree         int    `json:"isFree"`
	Status         int    `json:"status"`
	MinGameVersion string `json:"minGameVersion"`
	common.ControlBy
}

func (s *DlcUpdateReq) Generate(m *models.Dlc) {
	m.Name = s.Name
	m.Version = s.Version
	m.Description = s.Description
	m.Price = s.Price
	m.IsFree = s.IsFree
	m.Status = s.Status
	m.MinGameVersion = s.MinGameVersion
	m.UpdateBy = s.UpdateBy
}

func (s *DlcUpdateReq) GetId() interface{} {
	return s.Id
}

type DlcById struct {
	dto.ObjectById
}
