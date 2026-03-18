package models

// CustomerValidListParams 客户有效性标签列表查询参数
type CustomerValidListParams struct {
	BasePaging
	Type   *uint  `json:"type" form:"type"`     // 类型
	Name   string `json:"name" form:"name"`     // 名称
	Status *uint  `json:"status" form:"status"` // 状态
}

// CustomerValidCreateRequest 创建客户有效性标签请求
type CustomerValidCreateRequest struct {
	Type   uint   `json:"type" binding:"required" label:"类型"`
	Name   string `json:"name" binding:"required,max=50" label:"名称"`
	Status uint   `json:"status" binding:"required" label:"状态"`
}

// CustomerValidUpdateRequest 更新客户有效性标签请求
type CustomerValidUpdateRequest struct {
	ID     uint   `json:"id" binding:"required" label:"ID"`
	Type   uint   `json:"type" binding:"required" label:"类型"`
	Name   string `json:"name" binding:"required,max=50" label:"名称"`
	Status uint   `json:"status" binding:"required" label:"状态"`
}