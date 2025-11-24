package models

import (
	"encoding/json"
	"gh-statement-app/pagination"
	"strings"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// Permission is used by pop to map your permissions database table to your go code.
type Permission struct {
	ID               uuid.UUID        `gorm:"primaryKey" json:"id"`
	Name             string           `json:"name" gorm:"column:name"`
	Description      nulls.String     `json:"description" gorm:"column:description"`
	PermissionRoutes PermissionRoutes `json:"permission_routes"`
	CreatedAt        time.Time        `json:"created_at" gorm:"column:created_at"`
	UpdatedAt        time.Time        `json:"updated_at" gorm:"column:updated_at"`
}

// String is not required by pop and may be deleted
func (p Permission) String() string {
	jp, _ := json.Marshal(p)
	return string(jp)
}

// Permissions is not required by pop and may be deleted
type Permissions []Permission

// Create creates a new permission
func (p Permission) Create(tx *gorm.DB) Permission {
	tx.Create(&p)
	return p
}

// CreateAccessPolicies creates an access policy
func (p Permission) CreateAccessPolicies(routes []string, tx *gorm.DB) {
	for _, route := range routes {
		parsedRoute := strings.Split(route, "|")
		tx.Create(&PermissionRoute{
			ID:           NewUUID(),
			PermissionID: p.ID,
			Path:         strings.Trim(parsedRoute[0], "/"),
			Alias:        parsedRoute[1],
		})
	}
}

// LoadAllPermissions loads all permissions
func LoadAllPermissions(tx *gorm.DB) Permissions {
	permissions := make(Permissions, 0)
	tx.Order("created_at desc").Find(&permissions)
	return permissions
}

// PaginatePermissions paginates records
func PaginatePermissions(tx *gorm.DB, pagination pagination.Pagination) *pagination.Pagination {
	permissions := make(Permissions, 0)
	tx.Scopes(Paginate(permissions, &pagination, tx)).Preload("PermissionRoutes").Order("created_at desc").Find(&permissions)
	pagination.Rows = permissions
	return &pagination
}

// GetPermissionByID gets a permission by ID
func GetPermissionByID(id uuid.UUID, tx *gorm.DB) Permission {
	permission := Permission{}
	tx.Where("id=?", id).First(&permission)
	return permission
}

// DeleteAllAccessPolicies removes all routes
func (p Permission) DeleteAllAccessPolicies(tx *gorm.DB) {
	tx.Exec("delete from permission_routes where permission_id = ?", p.ID)
}
