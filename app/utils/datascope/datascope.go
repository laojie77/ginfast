package datascope

import (
	"gin-fast/app/global/app"
	"gin-fast/app/models"
	"gin-fast/app/utils/common"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// getUserRoles 获取用户角色列表
func getUserRoles(userID uint) ([]*models.SysRole, error) {
	var user models.User
	err := app.DB().Preload("Roles").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return user.Roles, nil
}

// getDepartmentAndChildrenIDs 获取部门及其所有子部门 ID
func getDepartmentAndChildrenIDs(departmentTree models.SysDepartmentList, deptID uint) ([]uint, error) {
	if departmentTree.IsEmpty() {
		return []uint{deptID}, nil
	}

	var findDepartment func(depts models.SysDepartmentList, targetID uint) *models.SysDepartment
	findDepartment = func(depts models.SysDepartmentList, targetID uint) *models.SysDepartment {
		for _, dept := range depts {
			if dept.ID == targetID {
				return dept
			}
			if len(dept.Children) > 0 {
				if found := findDepartment(dept.Children, targetID); found != nil {
					return found
				}
			}
		}
		return nil
	}

	targetDept := findDepartment(departmentTree, deptID)
	if targetDept == nil {
		return []uint{deptID}, nil
	}

	var getAllChildrenIDs func(dept *models.SysDepartment, ids *[]uint)
	getAllChildrenIDs = func(dept *models.SysDepartment, ids *[]uint) {
		*ids = append(*ids, dept.ID)
		if len(dept.Children) > 0 {
			for _, child := range dept.Children {
				getAllChildrenIDs(child, ids)
			}
		}
	}

	var deptIDs []uint
	getAllChildrenIDs(targetDept, &deptIDs)
	return deptIDs, nil
}

// stringToUintSlice 字符串转 uint 切片
func stringToUintSlice(s string) ([]uint, error) {
	if s == "" {
		return []uint{}, nil
	}

	parts := strings.Split(s, ",")
	var result []uint
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			return nil, err
		}
		result = append(result, uint(id))
	}
	return result, nil
}

// getUserDepartmentID 获取用户所属部门 ID
func getUserDepartmentID(userID uint) (uint, error) {
	var user models.User
	err := app.DB().Select("dept_id").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return 0, err
	}
	return user.DeptID, nil
}

// getUserIDsByDepartmentIDs 根据部门 ID 获取用户 ID 列表
func getUserIDsByDepartmentIDs(deptIDs []uint) ([]uint, error) {
	if len(deptIDs) == 0 {
		return []uint{}, nil
	}

	var users []models.User
	err := app.DB().Select("id").Where("dept_id IN ?", deptIDs).Find(&users).Error
	if err != nil {
		return nil, err
	}

	var userIDs []uint
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}
	return userIDs, nil
}

func getDataScopeByColumn(c *gin.Context, column string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		claims := common.GetClaims(c)
		if claims == nil {
			return db.Where("1 = 0")
		}

		userID := claims.UserID
		if userID == 0 {
			return db.Where("1 = 0")
		}

		notCheckUserIDs := app.ConfigYml.GetUintSlice("server.notcheckuser")
		for _, notCheckUserID := range notCheckUserIDs {
			if notCheckUserID == userID {
				return db
			}
		}

		roles, err := getUserRoles(userID)
		if err != nil || len(roles) == 0 {
			return db.Where(column+" = ?", userID)
		}

		hasFullPermission := false
		for _, role := range roles {
			if role.DataScope == 1 {
				hasFullPermission = true
				break
			}
		}
		if hasFullPermission {
			return db
		}

		allowedDeptIDs := make(map[uint]bool)
		userDeptID, _ := getUserDepartmentID(userID)

		var allDepartments models.SysDepartmentList
		err = app.DB().Find(&allDepartments).Error
		if err != nil {
			return db.Where(column+" = ?", userID)
		}
		departmentTree := allDepartments.BuildTree()

		var allDeptIDs []uint
		for _, role := range roles {
			switch role.DataScope {
			case 2:
				if role.CheckedDepts != "" {
					deptIDs, err := stringToUintSlice(role.CheckedDepts)
					if err == nil && len(deptIDs) > 0 {
						allDeptIDs = append(allDeptIDs, deptIDs...)
					}
				}
			case 3:
				if userDeptID != 0 {
					allDeptIDs = append(allDeptIDs, userDeptID)
				}
			case 4:
				if userDeptID != 0 {
					deptIDs, err := getDepartmentAndChildrenIDs(departmentTree, userDeptID)
					if err == nil && len(deptIDs) > 0 {
						allDeptIDs = append(allDeptIDs, deptIDs...)
					}
				}
			}
		}

		for _, deptID := range allDeptIDs {
			allowedDeptIDs[deptID] = true
		}

		var deptIDSlice []uint
		for deptID := range allowedDeptIDs {
			deptIDSlice = append(deptIDSlice, deptID)
		}

		var allowedUserIDs []uint
		if len(deptIDSlice) > 0 {
			userIDs, err := getUserIDsByDepartmentIDs(deptIDSlice)
			if err == nil {
				allowedUserIDs = append(allowedUserIDs, userIDs...)
			}
		}
		allowedUserIDs = append(allowedUserIDs, userID)

		userIDMap := make(map[uint]bool)
		var userIDSlice []uint
		for _, uid := range allowedUserIDs {
			if !userIDMap[uid] {
				userIDMap[uid] = true
				userIDSlice = append(userIDSlice, uid)
			}
		}

		return db.Where(column+" IN ?", userIDSlice)
	}
}

// GetDataScope 数据权限，默认按 created_by 过滤
func GetDataScope(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return getDataScopeByColumn(c, "created_by")
}

// GetDataScopeUser 数据权限，默认按 user_id 过滤
func GetDataScopeUser(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return getDataScopeByColumn(c, "user_id")
}
