package models

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// PasswordExpiry is used by pop to map your .model.Name.Proper.Pluralize.Underscore database table to your go code.
type PasswordExpiry struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Days      int       `json:"days" db:"days"`
	Length    int       `json:"length" db:"length"`
	Dormancy  int       `json:"dormancy" db:"dormancy"`
	Tries     int       `json:"tries" db:"tries"`
	RemindIn  int       `json:"remind_in" db:"remind_in"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PasswordExpiries is not required by pop and may be deleted
type PasswordExpiries []PasswordExpiry

func GetPasswordExpiry(tx *gorm.DB) PasswordExpiry {
	expiry := PasswordExpiry{}
	tx.Last(&expiry)
	return expiry
}

func (pe PasswordExpiry) UpdatePasswordExpiry(days, remind, length, dormancy, tries int, tx *gorm.DB) {
	pe.Days = days
	pe.RemindIn = remind
	pe.Length = length
	pe.Dormancy = dormancy
	pe.Tries = tries
	if err := tx.Save(&pe); err != nil {
		panic(err)
	}
}
