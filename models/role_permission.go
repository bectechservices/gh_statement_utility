package models

import (
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
)

// RolePermission is used by pop to map your role_permissions database table to your go code.
type RolePermission struct {
	ID           uuid.UUID `gorm:"primaryKey" json:"id"`
	RoleID       uuid.UUID `json:"role_id" gorm:"column:role_id"`
	PermissionID uuid.UUID `json:"permission_id" gorm:"column:permission_id"`
	Role         Role
	Permission   Permission
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// String is not required by pop and may be deleted
func (r RolePermission) String() string {
	jr, _ := json.Marshal(r)
	return string(jr)
}

// RolePermissions is not required by pop and may be deleted
type RolePermissions []RolePermission
