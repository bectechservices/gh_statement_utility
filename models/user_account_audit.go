package models

import (
	"encoding/json"
	"gh-statement-app/pagination"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// UserAccountAudit is used by pop to map your user_account_audits database table to your go code.
type UserAccountAudit struct {
	ID          uuid.UUID `gorm:"primaryKey"`
	Activity    string    `json:"activity" gorm:"column:activity"`
	Description string    `json:"description" gorm:"column:description"`
	ActivityBy  uuid.UUID `json:"activity_by" gorm:"column:activity_by"`
	ActivityFor uuid.UUID `json:"activity_for" gorm:"column:activity_for"`
	For         User      `gorm:"foreignKey:activity_for"`
	By          User      `gorm:"foreignKey:activity_by"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// String is not required by pop and may be deleted
func (u UserAccountAudit) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// UserAccountAudits is not required by pop and may be deleted
type UserAccountAudits []UserAccountAudit

// CreateActivityAudit creates a new activity log
func CreateActivityAudit(activity, description string, activityBy, activityFor uuid.UUID, tx *gorm.DB) {
	tx.Create(&UserAccountAudit{
		ID:          NewUUID(),
		Activity:    activity,
		Description: description,
		ActivityBy:  activityBy,
		ActivityFor: activityFor,
	})
}

// LoadAccountAudits loads the n recent activities
func LoadAccountAudits(n int, tx *gorm.DB) UserAccountAudits {
	audits := make(UserAccountAudits, 0)
	tx.Limit(n).Preload("For").Preload("By").Order("created_at desc").Find(&audits)
	return audits
}

// LoadAccountAudits loads all activities
func LoadAllAccountAudits(tx *gorm.DB) UserAccountAudits {
	audits := make(UserAccountAudits, 0)
	tx.Preload("For").Preload("By").Order("created_at desc").Find(&audits)
	return audits
}

// PaginateUserAccountAudits pagniates the records
func PaginateUserAccountAudits(tx *gorm.DB, pagination pagination.Pagination) *pagination.Pagination {
	audits := make(UserAccountAudits, 0)
	tx.Scopes(Paginate(audits, &pagination, tx)).Preload("For").Preload("By").Order("created_at desc").Find(&audits)
	pagination.Rows = audits
	return &pagination
}
