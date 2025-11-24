package models

import (
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// PermissionRoute is used by pop to map your permission_routes database table to your go code.
type PermissionRoute struct {
	ID           uuid.UUID `gorm:"primaryKey" json:"id"`
	PermissionID uuid.UUID `json:"permission_id" gorm:"column:permission_id"`
	Path         string    `json:"path" gorm:"column:path"`
	Alias        string    `json:"alias" gorm:"column:alias"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// String is not required by pop and may be deleted
func (p PermissionRoute) String() string {
	jp, _ := json.Marshal(p)
	return string(jp)
}

// PermissionRoutes is not required by pop and may be deleted
type PermissionRoutes []PermissionRoute

//LoadAllAccessPermissions loads all access permissions
func LoadAllAccessPermissions(tx *gorm.DB) PermissionRoutes {
	permissions := make(PermissionRoutes, 0)
	tx.Find(&permissions)
	return permissions
}
