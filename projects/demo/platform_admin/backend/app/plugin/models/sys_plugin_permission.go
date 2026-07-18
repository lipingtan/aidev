package models

// SysPluginPermission 插件权限注册表
type SysPluginPermission struct {
	ID          int    `json:"id" gorm:"primaryKey;autoIncrement"`
	PluginName  string `json:"pluginName" gorm:"column:plugin_name;type:varchar(64);not null;index"`
	Code        string `json:"code" gorm:"type:varchar(200);not null"`
	DisplayName string `json:"displayName" gorm:"column:display_name;type:varchar(200)"`
	GroupName   string `json:"groupName" gorm:"column:group_name;type:varchar(100)"`
}

func (SysPluginPermission) TableName() string {
	return "sys_plugin_permission"
}
