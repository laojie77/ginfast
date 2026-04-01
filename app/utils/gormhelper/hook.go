package gormhelper

import (
	"gin-fast/app/global/app"
	"gin-fast/app/global/myerrors"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

// 这里的函数都是 gorm 的 hook 函数，用于补充一些通用字段处理。

// MaskNotDataError 解决 gorm v2 查询无数据时抛错的问题。
func MaskNotDataError(gormDB *gorm.DB) {
	gormDB.Statement.RaiseErrorOnNotFound = false
}

// InterceptCreatePramsNotPtrError 拦截 create 参数不是指针的情况。
func CreateBeforeHook(gormDB *gorm.DB) {
	if reflect.TypeOf(gormDB.Statement.Dest).Kind() != reflect.Ptr {
		app.ZapLog.Warn(myerrors.ErrorsGormDBCreateParamsNotPtr)
		return
	}

	destValueOf := reflect.ValueOf(gormDB.Statement.Dest).Elem()
	switch destValueOf.Type().Kind() {
	case reflect.Slice, reflect.Array:
		inLen := destValueOf.Len()
		for i := 0; i < inLen; i++ {
			row := destValueOf.Index(i)
			switch row.Type().Kind() {
			case reflect.Struct:
				setStructUintField(row, "TenantID", GetTenantIDFromContext(gormDB.Statement.Context))
				setStructUintField(row, "CreatedBy", GetCurrentUserIDFromContext(gormDB.Statement.Context))
			case reflect.Map:
				if b, column := structHasSpecialField("tenant_id", row); b {
					if tenantID := GetTenantIDFromContext(gormDB.Statement.Context); tenantID > 0 {
						row.SetMapIndex(reflect.ValueOf(column), reflect.ValueOf(tenantID))
					}
				}
				if b, column := structHasSpecialField("created_by", row); b {
					if userID := GetCurrentUserIDFromContext(gormDB.Statement.Context); userID > 0 {
						row.SetMapIndex(reflect.ValueOf(column), reflect.ValueOf(userID))
					}
				}
			}
		}
	case reflect.Struct:
		setStatementUintColumn(gormDB, "TenantID", GetTenantIDFromContext(gormDB.Statement.Context))
		setStatementUintColumn(gormDB, "CreatedBy", GetCurrentUserIDFromContext(gormDB.Statement.Context))
	case reflect.Map:
		if b, column := structHasSpecialField("tenant_id", gormDB.Statement.Dest); b {
			if tenantID := GetTenantIDFromContext(gormDB.Statement.Context); tenantID > 0 {
				destValueOf.SetMapIndex(reflect.ValueOf(column), reflect.ValueOf(tenantID))
			}
		}
		if b, column := structHasSpecialField("created_by", gormDB.Statement.Dest); b {
			if userID := GetCurrentUserIDFromContext(gormDB.Statement.Context); userID > 0 {
				destValueOf.SetMapIndex(reflect.ValueOf(column), reflect.ValueOf(userID))
			}
		}
	}
}

func setStructUintField(row reflect.Value, fieldName string, value uint) {
	if value == 0 {
		return
	}

	field := row.FieldByName(fieldName)
	if !field.IsValid() || !field.CanSet() {
		return
	}

	switch field.Kind() {
	case reflect.Uint:
		field.SetUint(uint64(value))
	case reflect.Ptr:
		if field.Type().Elem().Kind() != reflect.Uint {
			return
		}
		ptr := reflect.New(field.Type().Elem())
		ptr.Elem().SetUint(uint64(value))
		field.Set(ptr)
	}
}

func setStatementUintColumn(gormDB *gorm.DB, fieldName string, value uint) {
	if value == 0 {
		return
	}

	if b, column := structHasSpecialField(fieldName, gormDB.Statement.Dest); b {
		gormDB.Statement.SetColumn(column, value)
	}
}

// UpdateBeforeHook
// InterceptUpdatePramsNotPtrError 拦截 save、update 参数不是指针的情况。
func UpdateBeforeHook(gormDB *gorm.DB) {
	if reflect.TypeOf(gormDB.Statement.Dest).Kind() == reflect.Struct {
		app.ZapLog.Warn(myerrors.ErrorsGormDBUpdateParamsNotPtr)
	} else if reflect.TypeOf(gormDB.Statement.Dest).Kind() == reflect.Map {
		// gorm.Update / Updates 在 map 场景下无需额外处理。
	}
}

// DeleteBeforeHook 删除前 hook。
func DeleteBeforeHook(gormDB *gorm.DB) {
	_ = gormDB
}

// structHasSpecialField 检查结构体或 map 是否包含指定字段。
func structHasSpecialField(fieldName string, anyStructPtr interface{}) (bool, string) {
	var tmp reflect.Type
	if reflect.TypeOf(anyStructPtr).Kind() == reflect.Ptr && reflect.ValueOf(anyStructPtr).Elem().Kind() == reflect.Map {
		destValueOf := reflect.ValueOf(anyStructPtr).Elem()
		for _, item := range destValueOf.MapKeys() {
			if item.String() == fieldName {
				return true, fieldName
			}
		}
	} else if reflect.TypeOf(anyStructPtr).Kind() == reflect.Ptr && reflect.ValueOf(anyStructPtr).Elem().Kind() == reflect.Struct {
		destValueOf := reflect.ValueOf(anyStructPtr).Elem()
		tf := destValueOf.Type()
		for i := 0; i < tf.NumField(); i++ {
			if !tf.Field(i).Anonymous && tf.Field(i).Type.Kind() != reflect.Struct {
				if tf.Field(i).Name == fieldName {
					return true, getColumnNameFromGormTag(fieldName, tf.Field(i).Tag.Get("gorm"))
				}
			} else if tf.Field(i).Type.Kind() == reflect.Struct {
				tmp = tf.Field(i).Type
				for j := 0; j < tmp.NumField(); j++ {
					if tmp.Field(j).Name == fieldName {
						return true, getColumnNameFromGormTag(fieldName, tmp.Field(j).Tag.Get("gorm"))
					}
				}
			}
		}
	} else if reflect.Indirect(anyStructPtr.(reflect.Value)).Type().Kind() == reflect.Struct {
		destValueOf := anyStructPtr.(reflect.Value)
		tf := destValueOf.Type()
		for i := 0; i < tf.NumField(); i++ {
			if !tf.Field(i).Anonymous && tf.Field(i).Type.Kind() != reflect.Struct {
				if tf.Field(i).Name == fieldName {
					return true, getColumnNameFromGormTag(fieldName, tf.Field(i).Tag.Get("gorm"))
				}
			} else if tf.Field(i).Type.Kind() == reflect.Struct {
				tmp = tf.Field(i).Type
				for j := 0; j < tmp.NumField(); j++ {
					if tmp.Field(j).Name == fieldName {
						return true, getColumnNameFromGormTag(fieldName, tmp.Field(j).Tag.Get("gorm"))
					}
				}
			}
		}
	} else if reflect.Indirect(anyStructPtr.(reflect.Value)).Type().Kind() == reflect.Map {
		destValueOf := anyStructPtr.(reflect.Value)
		for _, item := range destValueOf.MapKeys() {
			if item.String() == fieldName {
				return true, fieldName
			}
		}
	}
	return false, ""
}

// getColumnNameFromGormTag 从 gorm tag 中获取列名。
func getColumnNameFromGormTag(defaultColumn, tagValue string) (str string) {
	pos1 := strings.Index(tagValue, "column:")
	if pos1 == -1 {
		str = defaultColumn
		return
	}

	tagValue = tagValue[pos1+7:]
	pos2 := strings.Index(tagValue, ";")
	if pos2 == -1 {
		str = tagValue
	} else {
		str = tagValue[:pos2]
	}
	return strings.ReplaceAll(str, " ", "")
}
