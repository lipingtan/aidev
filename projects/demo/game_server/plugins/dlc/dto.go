package main

// DlcListReq 分页列表请求
type DlcListReq struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	GameId   int    `json:"gameId"`
	Name     string `json:"name"`
	Status   int    `json:"status"`
}

// DlcCreateReq 创建请求
type DlcCreateReq struct {
	GameId         int    `json:"gameId"`
	DlcKey         string `json:"dlcKey"`
	Name           string `json:"name"`
	Version        string `json:"version"`
	Description    string `json:"description"`
	Price          int    `json:"price"`
	IsFree         int    `json:"isFree"`
	Status         int    `json:"status"`
	MinGameVersion string `json:"minGameVersion"`
}

// DlcUpdateReq 更新请求
type DlcUpdateReq struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	Description    string `json:"description"`
	Price          int    `json:"price"`
	IsFree         int    `json:"isFree"`
	Status         int    `json:"status"`
	MinGameVersion string `json:"minGameVersion"`
}

// PageResponse 分页响应包装
type PageResponse struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}
