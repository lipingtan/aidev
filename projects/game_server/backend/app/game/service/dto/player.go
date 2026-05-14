package dto

import (
	"go-admin/common/dto"
)

type PlayerGetPageReq struct {
	dto.Pagination `search:"-"`
	GameId         int    `form:"gameId" search:"type:exact;column:game_id;table:game_player" comment:"游戏ID"`
	Uid            string `form:"uid" search:"type:contains;column:uid;table:game_player" comment:"玩家UID"`
	Nickname       string `form:"nickname" search:"type:contains;column:nickname;table:game_player" comment:"昵称"`
	Status         int    `form:"status" search:"type:exact;column:status;table:game_player" comment:"状态"`
}

func (m *PlayerGetPageReq) GetNeedSearch() interface{} {
	return *m
}

// PlayerBanReq 封禁/解封玩家
type PlayerBanReq struct {
	Id        int    `uri:"id" json:"-"`
	Status    int    `json:"status"`
	BanReason string `json:"banReason"`
}

func (s *PlayerBanReq) GetId() interface{} {
	return s.Id
}

type PlayerById struct {
	dto.ObjectById
}
