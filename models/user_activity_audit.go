package models

import (
	"gh-statement-app/constants"
	"gh-statement-app/pagination"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// UserActivityAudit is used by pop to map your .model.Name.Proper.Pluralize.Underscore database table to your go code.
type UserActivityAudit struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	UserID    uuid.UUID `json:"user_id" gorm:"column:user_id"`
	BranchID  uuid.UUID `json:"branch_id" gorm:"column:branch_id"`
	Activity  string    `json:"activity" gorm:"column:activity"`
	User      User
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}

// UserActivityAudits is not required by pop and may be deleted
type UserActivityAudits []UserActivityAudit

// LoadRecentActivities loads the n recent activities
func LoadRecentActivities(n int, tx *gorm.DB) UserActivityAudits {
	audits := make(UserActivityAudits, 0)
	tx.Limit(n).Preload("User").Order("created_at desc").Find(&audits)
	return audits
}

// LoadUserActivityAudit loads the activities
func LoadUserActivityAudit(tx *gorm.DB) UserActivityAudits {
	audits := make(UserActivityAudits, 0)
	tx.Preload("User").Order("created_at desc").Find(&audits)
	return audits
}

// LoadUserLastLogin loads the last time a user logged in
func LoadUserLastLogin(id uuid.UUID, tx *gorm.DB) UserActivityAudit {
	audit := UserActivityAudit{}
	tx.Order("created_at desc").Where("user_id=? and activity=?", id, constants.UserLogin).Limit(1).Find(&audit)
	return audit
}

// PaginateUserActivityAudits pagniates the records
func PaginateUserActivityAudits(tx *gorm.DB, pagination pagination.Pagination) *pagination.Pagination {
	audits := make(UserActivityAudits, 0)
	tx.Scopes(Paginate(audits, &pagination, tx)).Where("activity <> ?", "").Preload("User").Order("created_at desc").Find(&audits)
	pagination.Rows = audits
	return &pagination
}
