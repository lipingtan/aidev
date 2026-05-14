package dto

import (
	"go-admin/app/game/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
)

type H5PageGetPageReq struct {
	dto.Pagination `search:"-"`
	GameId         int    `form:"gameId" search:"type:exact;column:game_id;table:game_h5_page" comment:"游戏ID"`
	Name           string `form:"name" search:"type:contains;column:name;table:game_h5_page" comment:"页面名称"`
	PageType       string `form:"pageType" search:"type:exact;column:page_type;table:game_h5_page" comment:"页面类型"`
	Status         int    `form:"status" search:"type:exact;column:status;table:game_h5_page" comment:"状态"`
}

func (m *H5PageGetPageReq) GetNeedSearch() interface{} {
	return *m
}

type H5PageInsertReq struct {
	GameId      int    `json:"gameId" vd:"$>0;msg:'游戏ID不能为空'"`
	PageKey     string `json:"pageKey" vd:"len($)>0;msg:'页面标识不能为空'"`
	Name        string `json:"name" vd:"len($)>0;msg:'页面名称不能为空'"`
	PageType    string `json:"pageType"`
	Content     string `json:"content"`
	ExternalUrl string `json:"externalUrl"`
	UseExternal int    `json:"useExternal"`
	Status      int    `json:"status"`
	Remark      string `json:"remark"`
	common.ControlBy
}

func (s *H5PageInsertReq) Generate(m *models.H5Page) {
	m.GameId = s.GameId
	m.PageKey = s.PageKey
	m.Name = s.Name
	m.PageType = s.PageType
	m.Content = s.Content
	m.ExternalUrl = s.ExternalUrl
	m.UseExternal = s.UseExternal
	m.Status = s.Status
	m.Remark = s.Remark
	m.CreateBy = s.CreateBy
	m.UpdateBy = s.UpdateBy
}

func (s *H5PageInsertReq) GetId() interface{} {
	return nil
}

type H5PageUpdateReq struct {
	Id          int    `uri:"id" json:"-"`
	Name        string `json:"name"`
	PageType    string `json:"pageType"`
	Content     string `json:"content"`
	ExternalUrl string `json:"externalUrl"`
	UseExternal int    `json:"useExternal"`
	Status      int    `json:"status"`
	Remark      string `json:"remark"`
	common.ControlBy
}

func (s *H5PageUpdateReq) Generate(m *models.H5Page) {
	m.Name = s.Name
	m.PageType = s.PageType
	m.Content = s.Content
	m.ExternalUrl = s.ExternalUrl
	m.UseExternal = s.UseExternal
	m.Status = s.Status
	m.Remark = s.Remark
	m.UpdateBy = s.UpdateBy
}

func (s *H5PageUpdateReq) GetId() interface{} {
	return s.Id
}

type H5PageById struct {
	dto.ObjectById
}
