package models

import (
	"encoding/json"
	"ng-statement-app/pagination"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// Role is used by pop to map your roles database table to your go code.
type Role struct {
	ID               uuid.UUID       `gorm:"primaryKey" json:"id"`
	Name             string          `json:"name" gorm:"column:name"`
	Description      nulls.String    `json:"description" gorm:"column:description"`
	RolePermissions  RolePermissions `json:"role_permissions"`
	ActivityAccessID nulls.UUID      `json:"activity_access_id" db:"activity_access_id"`
	ActivityAccess   ActivityAccess  `belongs_to:"activity_access"`
	CreatedAt        time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt        time.Time       `json:"updated_at" gorm:"column:updated_at"`
}

// String is not required by pop and may be deleted
func (r Role) String() string {
	jr, _ := json.Marshal(r)
	return string(jr)
}

// Roles is not required by pop and may be deleted
type Roles []Role

// Create creates a new role
func (r Role) Create(tx *gorm.DB) Role {
	tx.Create(&r)
	return r
}

// AddPermissions adds a permission to a role
func (r Role) AddPermissions(permissions []string, tx *gorm.DB) {
	for _, permission := range permissions {
		uid, _ := uuid.FromString(permission)
		tx.Create(&RolePermission{
			ID:           NewUUID(),
			PermissionID: uid,
			RoleID:       r.ID,
		})
	}
}

// PaginateRoles pagniates the records
func PaginateRoles(tx *gorm.DB, pagination pagination.Pagination) *pagination.Pagination {
	roles := make(Roles, 0)
	tx.Scopes(Paginate(roles, &pagination, tx)).Preload("RolePermissions.Role").Preload("RolePermissions.Permission").Preload("ActivityAccess").Order("created_at desc").Find(&roles)
	pagination.Rows = roles
	return &pagination
}

// LoadAllRoles loads all roles
func LoadAllRoles(tx *gorm.DB) Roles {
	roles := make(Roles, 0)
	tx.Order("created_at desc").Find(&roles)
	return roles
}

// GetRoleByID gets a role by ID
func GetRoleByID(id uuid.UUID, tx *gorm.DB) Role {
	role := Role{}
	tx.Where("id=?", id).First(&role)
	return role
}

// DeleteAllPermissions removes all permissions
func (r Role) DeleteAllPermissions(tx *gorm.DB) {
	tx.Exec("delete from role_permissions where role_id = ?", r.ID)
}
