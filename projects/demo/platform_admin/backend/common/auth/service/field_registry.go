package service

import (
	"log"
	"reflect"
	"strings"

	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"

	"gorm.io/gorm"
)

// FieldRegistry 字段对象自动注册服务
type FieldRegistry struct {
	db   *gorm.DB
	repo repository.FieldObjectRepository
}

// NewFieldRegistry 创建 FieldRegistry 实例
func NewFieldRegistry(db *gorm.DB, repo repository.FieldObjectRepository) *FieldRegistry {
	return &FieldRegistry{db: db, repo: repo}
}

// AutoRegister 通过反射扫描 struct 的 fieldperm tag，自动注册字段对象和字段定义
// objectCode: 业务对象标识
// objectName: 对象显示名
// m: struct 实例（传值或指针均可）
func (r *FieldRegistry) AutoRegister(objectCode, objectName string, m interface{}) {
	t := reflect.TypeOf(m)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		log.Printf("[field-registry] AutoRegister 跳过非 struct 类型: %s", t.Name())
		return
	}

	// 检查 struct 级别是否标注 fieldperm:"-"（通过内嵌匿名字段或第一个字段约定）
	// 实际场景中通过不调用 AutoRegister 来跳过不需要注册的 model

	// 注册字段对象（INSERT IGNORE）
	obj := &model.FieldObject{
		ObjectCode: objectCode,
		ObjectName: objectName,
		Source:     "AUTO",
	}
	if err := r.repo.UpsertObject(r.db, obj); err != nil {
		log.Printf("[field-registry] 注册字段对象失败 object_code=%s: %v", objectCode, err)
		return
	}

	// 扫描字段
	registered := 0
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		desc := field.Tag.Get("fieldperm")
		// 跳过无 fieldperm tag 或标记为 "-" 的字段
		if desc == "" || desc == "-" {
			continue
		}

		// 提取 json tag 作为 field_name
		jsonTag := parseJSONFieldName(field.Tag.Get("json"))
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// 写入数据库（INSERT IGNORE，不覆盖已有自定义描述）
		def := &model.FieldDefinition{
			ObjectCode:  objectCode,
			FieldName:   jsonTag,
			Description: desc,
			Source:      "AUTO",
		}
		if err := r.repo.UpsertDefinition(r.db, def); err != nil {
			log.Printf("[field-registry] 注册字段定义失败 object_code=%s field=%s: %v", objectCode, jsonTag, err)
			continue
		}
		registered++
	}

	log.Printf("[field-registry] 自动注册完成 object_code=%s object_name=%s fields=%d", objectCode, objectName, registered)
}

// parseJSONFieldName 从 json tag 中提取字段名
// 例如 "phone" → "phone", "id,string" → "id", "-" → "-", "" → ""
func parseJSONFieldName(tag string) string {
	if tag == "" {
		return ""
	}
	parts := strings.Split(tag, ",")
	return parts[0]
}
