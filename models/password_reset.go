package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// PasswordReset is used by pop to map your password_resets database table to your go code.
type PasswordReset struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	UserID    uuid.UUID `json:"user_id" gorm:"column:user_id"`
	Token     string    `json:"token" gorm:"column:token"`
	User      User
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt nulls.Time `json:"updated_at" gorm:"column:updated_at"`
}

// String is not required by pop and may be deleted
func (p PasswordReset) String() string {
	jp, _ := json.Marshal(p)
	return string(jp)
}

// Update the query method
func FindValidPasswordReset(token string, db *gorm.DB) (*PasswordReset, error) {
	var reset PasswordReset
	expiryTime := time.Now().Add(-24 * time.Hour) // 24 hour expiry

	err := db.Where("token = ? AND created_at >= ?", token, expiryTime).
		First(&reset).Error

	return &reset, err
}

// PasswordResets is not required by pop and may be deleted
type PasswordResets []PasswordReset

// String is not required by pop and may be deleted
func (p PasswordResets) String() string {
	jp, _ := json.Marshal(p)
	return string(jp)
}

// PasswordResetTokenIsValid checks if a given token is valid and hasnt been used
func PasswordResetTokenIsValid(token string, tx *gorm.DB) bool {
	record := PasswordReset{}
	result := tx.Where("token=? and created_at >= ? and updated_at is null", token, time.Now().Add(-5*time.Hour)).Limit(1).Find(&record)
	if result.Error != nil {
		panic(result.Error)
	}
	return result.RowsAffected == 0
}

// // GetUserFromResetToken retrieves the user who owns the token
// func GetUserFromResetToken(token string, tx *gorm.DB) User {
// 	record := PasswordReset{}
// 	fmt.Println("----------Get User Token Access from DB-----------", token, record)
// 	tx.Where("token=?", token).Preload("User").First(&record)
// 	return record.User
// }

// DestroyResetToken marks the token as used
func DestroyResetToken(token string, tx *gorm.DB) {
	tx.Exec("update password_resets set updated_at=? where token=?", time.Now(), token)
}

// In models/password_reset.go
func GetUserFromResetToken(token string, db *gorm.DB) *User {
	var passwordReset PasswordReset

	err := db.Where("token = ? AND created_at >= ?", token, time.Now().Add(-24*time.Hour)).
		First(&passwordReset).Error

	if err != nil {
		return nil // Now this returns nil pointer
	}

	var user User
	err = db.Where("id = ?", passwordReset.UserID).First(&user).Error
	if err != nil {
		return nil
	}

	return &user
}
