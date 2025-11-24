package models

import (
	"time"

	"github.com/gofrs/uuid"
)

// UserRole is used by pop to map your .model.Name.Proper.Pluralize.Underscore database table to your go code.
type UserRole struct {
	ID        uuid.UUID `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `json:"user_id" gorm:"column:user_id"`
	RoleID    uuid.UUID `json:"role_id" gorm:"column:role_id"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
	User      User      `belongs_to:"user"`
	Role      Role      `belongs_to:"role"`
}

// UserRoles is not required by pop and may be deleted
type UserRoles []UserRole
