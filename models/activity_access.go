package models

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

const (
	Sunday time.Weekday = 0 + iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

type ActivityAccess struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Monday    bool      `db:"monday" json:"monday"`
	Tuesday   bool      `db:"tuesday" json:"tuesday"`
	Wednesday bool      `db:"wednesday" json:"wednesday"`
	Thursday  bool      `db:"thursday" json:"thursday"`
	Friday    bool      `db:"friday" json:"friday"`
	Saturday  bool      `db:"saturday" json:"saturday"`
	Sunday    bool      `db:"sunday" json:"sunday"`
	StartTime string    `db:"start_time" json:"start_time"`
	EndTime   string    `db:"end_time" json:"end_time"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func (ActivityAccess) TableName() string {
	return "activity_access"
}

// Roles is not required by pop and may be deleted
type ActivityAccesses []ActivityAccess

// GetRoleRestrictionByID gets a role by ID
func GetRoleRestrictionByID(id uuid.UUID, tx *gorm.DB) ActivityAccess {
	activity := ActivityAccess{}
	if err := tx.Where("id=?", id).First(&activity); err != nil {
		panic(err)
	}
	return activity
}

// LoadAllRoleRestriction loads all branches
func LoadAllRoleRestriction(tx *gorm.DB) ActivityAccesses {
	activities := make(ActivityAccesses, 0)

	if err := tx.Order("created_at ASC").Find(&activities).Error; err != nil {
		return nil
	}

	return activities
}

func GrantRolesAccessActivity(ab_number string, tx *gorm.DB) bool {
	RolesactivityAccess := ActivityAccess{}

	result := tx.Raw(`SELECT activity_access.* FROM users 
                    INNER JOIN user_roles ON users.id = user_roles.user_id
                    INNER JOIN roles ON roles.id = user_roles.role_id 
                    INNER JOIN activity_access ON activity_access.id = roles.activity_access_id
                    WHERE users.ab_number = ?`, ab_number).Scan(&RolesactivityAccess)

	if result.Error != nil {
		return false
	}

	// Check if any record was found
	if result.RowsAffected == 0 {
		return false
	}

	// Debug logging
	fmt.Printf("DEBUG - User: %s, StartTime: %s, EndTime: %s\n",
		ab_number, RolesactivityAccess.StartTime, RolesactivityAccess.EndTime)
	fmt.Printf("DEBUG - Current time: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	dayAccess := RolesactivityAccess.RolesCanWorkDuringActiveDays()
	hourAccess := RolesactivityAccess.RoleCanWorkDuringActiveHours()

	fmt.Printf("DEBUG - Day access: %v, Hour access: %v\n", dayAccess, hourAccess)

	return dayAccess && hourAccess
}

// Check whether user can work during selected active days
func (activity *ActivityAccess) RolesCanWorkDuringActiveDays() bool {
	currentDateTime := time.Now()
	currentDay := currentDateTime.Weekday()
	isActive := false

	switch currentDay {
	case time.Monday:
		isActive = activity.Monday
	case time.Tuesday:
		isActive = activity.Tuesday
	case time.Wednesday:
		isActive = activity.Wednesday
	case time.Thursday:
		isActive = activity.Thursday
	case time.Friday:
		isActive = activity.Friday
	case time.Saturday:
		isActive = activity.Saturday
	case time.Sunday:
		isActive = activity.Sunday
	}
	return isActive
}

func (activity ActivityAccess) RoleCanWorkDuringActiveHours() bool {
	current := time.Now()

	// Extract time part from ISO string (assuming format is like "0001-01-01T01:00:00Z")
	startTimeStr := activity.StartTime[11:19] // Extract "01:00:00"
	endTimeStr := activity.EndTime[11:19]     // Extract "23:59:00"

	layout := "15:04:05"
	startTime, err := time.Parse(layout, startTimeStr)
	if err != nil {
		fmt.Printf("ERROR parsing start time '%s': %v\n", startTimeStr, err)
		return false
	}

	endTime, err := time.Parse(layout, endTimeStr)
	if err != nil {
		fmt.Printf("ERROR parsing end time '%s': %v\n", endTimeStr, err)
		return false
	}

	fmt.Printf("DEBUG - Current: %s, Start: %s, End: %s\n",
		current.Format("15:04:05"),
		startTime.Format("15:04:05"),
		endTime.Format("15:04:05"))

	// Create comparable time objects
	currentTime := time.Date(0, 1, 1, current.Hour(), current.Minute(), current.Second(), 0, time.UTC)
	startComparable := time.Date(0, 1, 1, startTime.Hour(), startTime.Minute(), startTime.Second(), 0, time.UTC)
	endComparable := time.Date(0, 1, 1, endTime.Hour(), endTime.Minute(), endTime.Second(), 0, time.UTC)

	// Handle overnight schedule
	if endComparable.Before(startComparable) {
		return currentTime.After(startComparable) || currentTime.Before(endComparable)
	}

	return currentTime.After(startComparable) && currentTime.Before(endComparable)
}

// GetRolesAccessActivityID gets a branch by ID
func GetRolesAccessActivityID(id uuid.UUID, tx *gorm.DB) ActivityAccess {
	accessactivity := ActivityAccess{}
	if err := tx.Where("id=?", id).First(&accessactivity); err != nil {
		panic(err)
	}
	return accessactivity
}

// CreateBranch creates a new branch
func CreateRoleRestriction(monday, tuesday, wednesday, thursday, friday, saturday, sunday bool, start_time, end_time string, tx *gorm.DB) ActivityAccess {
	activity := ActivityAccess{}
	fmt.Println("########################### start_time, end_time ###########", start_time, end_time)
	activity.ID, _ = uuid.NewV4()
	activity.Monday = monday
	activity.Tuesday = tuesday
	activity.Wednesday = wednesday
	activity.Thursday = thursday
	activity.Friday = friday
	activity.Saturday = saturday
	activity.Sunday = sunday
	activity.StartTime = start_time
	activity.EndTime = end_time
	fmt.Println("########################### ID###########", activity.ID)
	if err := tx.Create(&activity).Error; err != nil {
		panic(err)
	}
	return activity
}

func (acw ActivityAccess) UpdateRoleAccessRestriction(monday, tuesday, wednesday, thursday, friday, saturday, sunday bool, start_time, end_time string, tx *gorm.DB) {
	acw.Monday = monday
	acw.Tuesday = tuesday
	acw.Wednesday = wednesday
	acw.Thursday = thursday
	acw.Friday = friday
	acw.Saturday = saturday
	acw.Sunday = sunday
	acw.StartTime = start_time
	acw.EndTime = end_time
	if err := tx.Model(&acw).Updates(&acw).Error; err != nil {
		panic(err)
	}
}

func GetLastUpdatedAccessActivity(tx *gorm.DB) (*ActivityAccess, error) {
	activity := &ActivityAccess{}

	if err := tx.Order("created_at DESC").First(activity).Error; err != nil {
		fmt.Println("############GetLastUpdatedAccessActivity", err)
		return nil, err
	}

	return activity, nil
}
