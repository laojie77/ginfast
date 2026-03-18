package models

// CustomerValid 客户有效性标签模型
type CustomerValid struct {
	BaseModel
	Type     uint   `json:"type" gorm:"type:tinyint(3) unsigned;not null;default:1;comment:类型"`
	Name     string `json:"name" gorm:"type:varchar(50);not null;comment:名称"`
	Status   uint   `json:"status" gorm:"type:tinyint(3) unsigned;not null;default:1;comment:状态"`
	TenantID uint   `json:"tenantID" gorm:"type:int(11);column:tenant_id;comment:租户ID"`
}

// TableName 设置表名
func (CustomerValid) TableName() string {
	return "sys_customer_valids"
}
