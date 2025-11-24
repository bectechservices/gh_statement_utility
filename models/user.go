package models

import (
	"encoding/json"
	"fmt"
	"log"
	"ng-statement-app/constants"
	"ng-statement-app/mailers"
	"ng-statement-app/pagination"
	"strings"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User is used by pop to map your users database table to your go code.
type User struct {
	ID                  uuid.UUID    `gorm:"primaryKey" json:"id"`
	FirstName           string       `json:"first_name" gorm:"column:first_name"`
	LastName            string       `json:"last_name" gorm:"column:last_name"`
	ABNumber            string       `json:"ab_number" gorm:"column:ab_number"`
	Email               string       `json:"email" gorm:"column:email"`
	Password            string       `json:"password" gorm:"column:password"`
	BranchID            uuid.UUID    `json:"branch_id" gorm:"column:branch_id"`
	PasswordLastChanged time.Time    `json:"password_last_changed" gorm:"column:password_last_changed"`
	MustResetPassword   bool         `json:"must_reset_password" gorm:"column:must_reset_password"`
	Locked              bool         `json:"locked" gorm:"column:locked"`
	Deleted             nulls.Time   `json:"deleted" gorm:"column:deleted"`
	LastLogin           nulls.Time   `json:"last_login" gorm:"column:last_login"`
	IsLoggedIn          bool         `json:"is_logged_in" gorm:"column:is_logged_in"`
	Privileged          bool         `json:"privileged" gorm:"column:privileged"`
	Status              nulls.String `json:"status" gorm:"column:status"`
	Branch              Branch
	Roles               UserRoles
	Permissions         UserPermissions
	FailedLoginAttempts FailedLoginAttempts `has_many:"failed_login_attempts"`
	CreatedAt           time.Time           `json:"created_at" gorm:"column:created_at"`
	UpdatedAt           nulls.Time          `json:"updated_at" gorm:"column:updated_at"`
}

// String is not required by pop and may be deleted
func (u User) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Users is not required by pop and may be deleted
type Users []User

// String is not required by pop and may be deleted
func (u Users) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// GetUserByID gets a user by ID
func GetUserByID(id uuid.UUID, tx *gorm.DB) User {
	user := User{}
	tx.First(&user, id)
	return user
}

// GetUserByABNumber finds a user with the given ABNumber
func GetUserByABNumber(ABNumber string, tx *gorm.DB) User {
	user := User{}
	tx.Model(User{}).Preload("Branch").Where("ab_number=?", ABNumber).First(&user)
	return user
}

// IsEmpty checks if the user instance is empty
func (u User) IsEmpty() bool {
	return u.ID == uuid.Nil
}

func (u User) Name() string {
	return u.FirstName + " " + u.LastName
}

// PasswordIsValid verifies the password
func (u User) PasswordIsValid(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}

// BeforeCreate gets called before a user is created
func (u *User) BeforeCreate(tx *gorm.DB) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.WithStack(err)
	}

	u.Password = string(hash)

	return nil
}

// AllUsersCount returns the total number of users
func AllUsersCount(tx *gorm.DB) int64 {
	var count int64
	tx.Model(User{}).Count(&count)
	return count
}

// AllUsersCountPerBranch returns the total number of users
func AllUsersCountPerBranch(branch_id uuid.UUID, tx *gorm.DB) int64 {
	var count int64
	tx.Where("branch_id = ?", branch_id).Model(User{}).Count(&count)
	return count
}

// BlockedUsersCount number of locked users
func BlockedUsersCount(tx *gorm.DB) int64 {
	var count int64
	tx.Model(User{}).Where("locked = ?", true).Count(&count)
	return count
}

// LoadUserDetails loads user details
func LoadUserDetails(id uuid.UUID, tx *gorm.DB) User {
	user := User{}
	tx.Where("id=?", id).Preload("Branch").Preload("Roles.Role").Preload("Permissions.Permission").First(&user)
	return user
}

// LockAccount locks the user's account
func (u User) LockAccount(tx *gorm.DB) {
	u.LogAccountActivity(constants.AccountLocked, tx)
	u.Locked = true
	tx.Save(&u)
}

// LogAccountActivity logs an account activity
func (u User) LogAccountActivity(activity string, tx *gorm.DB) {
	tx.Create(&UserActivityAudit{
		ID:       NewUUID(),
		UserID:   u.ID,
		BranchID: u.BranchID,
		Activity: activity,
	})
}

// // ResetPassword resets the users password and sets the given password as new password
// func (u User) ResetPassword(newPassword string, tx *gorm.DB) {
// 	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
// 	if err != nil {
// 		// panic(err)
// 	}

// 	u.Password = string(hash)
// 	u.MustResetPassword = false
// 	u.PasswordLastChanged = time.Now()
// 	tx.Save(&u)
// }

// HashPasswordBcrypt hashing using BCRYPT
func HashPasswordBcrypt(new_password string) string {
	// bcrypt.DefaultCost is 10; you can increase this for more security at the cost of performance
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(new_password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hashedBytes)
}

type PasswordHistory struct {
	ID           uuid.UUID `db:"id" json:"id"`
	UserID       uuid.UUID `db:"user_id" json:"user_id"`
	User         User      `belongs_to:"users"`
	PasswordHash string    `db:"password_hash" json:"password_hash"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// ResetPassword resets the users password and sets the given password as new password
func (u *User) ResetPassword(userID uuid.UUID, newPassword, confirmPassword, token string, tx *gorm.DB) error {
	var passwordHistory []PasswordHistory

	fmt.Printf("########### ResetPassword called ###########\n")
	fmt.Printf("########### UserID: %s\n", userID)
	// fmt.Printf("########### Token: %s\n", token)
	// fmt.Printf("########### New Password: %s\n", newPassword)
	// fmt.Printf("########### Confirm Password: %s\n", confirmPassword)

	// Validate inputs
	if newPassword == "" {
		return fmt.Errorf("new password is required")
	}

	if confirmPassword == "" {
		return fmt.Errorf("confirm password is required")
	}

	if newPassword != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	if token == "" {
		return fmt.Errorf("reset token is required")
	}

	// Hash the new password
	newPasswordHash := HashPasswordBcrypt(newPassword)

	// Check password history
	if err := tx.Model(&PasswordHistory{}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(12).
		Find(&passwordHistory).Error; err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	// Compare given password with each stored hash
	for _, history := range passwordHistory {
		if bcrypt.CompareHashAndPassword([]byte(history.PasswordHash), []byte(newPassword)) == nil {
			return fmt.Errorf("password has been used recently")
		}
	}

	// Generate new UUID for password history
	uid, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to generate UUID: %w", err)
	}

	fmt.Printf("########### Password History ID: %s\n", uid)

	// Insert new password into history
	newHistory := PasswordHistory{
		ID:           uid,
		UserID:       userID,
		PasswordHash: newPasswordHash,
		CreatedAt:    time.Now(),
	}

	if err := tx.Create(&newHistory).Error; err != nil {
		return fmt.Errorf("failed to log password history: %w", err)
	}

	// Clean up old password history (keep only last 12)
	var recentHistoryIDs []uuid.UUID
	if err := tx.Model(&PasswordHistory{}).
		Select("id").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(12).
		Pluck("id", &recentHistoryIDs).Error; err != nil {
		return fmt.Errorf("failed to get recent history IDs: %w", err)
	}

	// Delete old records not in the recent 12
	if len(recentHistoryIDs) > 0 {
		if err := tx.Where("user_id = ? AND id NOT IN (?)", userID, recentHistoryIDs).
			Delete(&PasswordHistory{}).Error; err != nil {
			return fmt.Errorf("failed to clean up password history: %w", err)
		}
	}

	// Update user record
	if err := tx.Model(&User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"must_reset_password":   false,
			"password_last_changed": time.Now(),
			"password":              newPasswordHash,
		}).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Mark token as used - FIXED: Use the token parameter passed to the function
	if err := tx.Model(&PasswordReset{}).
		Where("token = ?", token).
		Update("updated_at", time.Now()).Error; err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	fmt.Printf("########### Password reset successful for user: %s\n", userID)
	return nil
}

// ResetPassword resets the users password and sets the given password as new password
func (u *User) ResetPrivilegeIDPasswordm(user uuid.UUID, password string, tx *gorm.DB) string {
	var passwordHistory []PasswordHistory

	// Hash the new password
	newPasswordHash := HashPasswordBcrypt(password)
	// Check password history
	if err := tx.Model(&PasswordHistory{}).
		Where("user_id = ?", user).
		Order("created_at DESC").
		Limit(12).
		Find(&passwordHistory).Error; err != nil {
		//return fmt.Errorf("query failed: %w", err)
	}
	// Compare given password with each stored hash
	for _, history := range passwordHistory {
		if bcrypt.CompareHashAndPassword([]byte(history.PasswordHash), []byte(password)) == nil {
			//return fmt.Errorf("password has been used recently")
		}
	}
	// Generate new UUID for password history
	uid, err := uuid.NewV4()
	if err != nil {
		//return fmt.Errorf("failed to generate UUID: %w", err)
	}
	fmt.Printf("########### Password History ID: %s\n", uid)
	// Insert new password into history
	newHistory := PasswordHistory{
		ID:           uid,
		UserID:       user,
		PasswordHash: newPasswordHash,
		CreatedAt:    time.Now(),
	}
	if err := tx.Create(&newHistory).Error; err != nil {
		//return fmt.Errorf("failed to log password history: %w", err)
	}
	// Clean up old password history (keep only last 12)
	var recentHistoryIDs []uuid.UUID
	if err := tx.Model(&PasswordHistory{}).
		Select("id").
		Where("user_id = ?", user).
		Order("created_at DESC").
		Limit(12).
		Pluck("id", &recentHistoryIDs).Error; err != nil {
		//	return fmt.Errorf("failed to get recent history IDs: %w", err)
	}

	// Delete old records not in the recent 12
	if len(recentHistoryIDs) > 0 {
		if err := tx.Where("user_id = ? AND id NOT IN (?)", user, recentHistoryIDs).
			Delete(&PasswordHistory{}).Error; err != nil {
			//	return fmt.Errorf("failed to clean up password history: %w", err)
		}
	}
	// Update user record
	if err := tx.Model(&User{}).
		Where("id = ?", user).
		Updates(map[string]interface{}{
			"must_reset_password":   false,
			"password_last_changed": time.Now(),
			"password":              newPasswordHash,
		}).Error; err != nil {
		//return fmt.Errorf("failed to update user: %w", err)
	}
	fmt.Printf("########### Password reset successful for user: %s\n", user)
	return newPasswordHash
}

// SendPasswordResetEmail sends a password reset email
func (u User) SendPasswordResetEmail(url string) {
	err := mailers.SendPasswordResets(u.Email, u.Name(), url)
	fmt.Println(err)
	fmt.Println("###################### Mailers sending @###################", u.Email, u.Name(), url)
}

// DashboardURL returns the user's dashboard url
func (u User) DashboardURL() string {
	permissions := make(PermissionRoutes, 0)
	permissionIds := make([]interface{}, 0)
	for _, id := range u.LoadAllPermissionIDs(GormDB) {
		permissionIds = append(permissionIds, id)
	}
	if len(permissionIds) > 0 {
		GormDB.Where("permission_id in (?)", permissionIds...).Find(&permissions)
	}
	fmt.Printf("%+v\n", permissionIds)

	for _, route := range permissions {
		path := strings.Trim(route.Path, "/")
		if path == "dashboard" {
			return "/dashboard/"
		}
	}
	for _, route := range permissions {
		if strings.Contains(strings.ToLower(route.Alias), "view") {
			return "/" + route.Path
		}
	}
	return "/no-permission"
}

// LoadAllPermissionIDs loads all user's perm ids
func (u User) LoadAllPermissionIDs(tx *gorm.DB) []string {
	permissions := make([]string, 0)
	user := User{}
	tx.Where("id=?", u.ID).Preload("Roles").Preload("Permissions").First(&user)
	if len(user.Roles) > 0 {
		roleIds := make([]interface{}, 0)
		for _, role := range user.Roles {
			roleIds = append(roleIds, role.RoleID)
		}
		rolePermissions := make(RolePermissions, 0)
		tx.Where("role_id in (?)", roleIds...).Find(&rolePermissions)
		for _, permission := range rolePermissions {
			permissions = append(permissions, permission.PermissionID.String())
		}
	}

	for _, permission := range user.Permissions {
		permissions = append(permissions, permission.PermissionID.String())
	}

	return permissions
}

// StoreForgottenPasswordToken stores the unique token generated
func (u User) StoreForgottenPasswordToken(token string, tx *gorm.DB) {
	tx.Create(&PasswordReset{
		ID:     NewUUID(),
		Token:  token,
		UserID: u.ID,
	})
}

// IsFirstTimeLogin checks if its the user's first time login in
func (u User) IsFirstTimeLogin() bool {
	return u.UpdatedAt.Time.IsZero()
}

// PasswordHasExpired check if the user's password has expired
func (u User) PasswordHasExpired(tx *gorm.DB) bool {
	expiry := PasswordExpiry{}
	tx.First(&expiry)
	if expiry.Days == 0 {
		expiry = PasswordExpiry{
			Days:     90,
			RemindIn: 45,
		}
	}
	dateDiff := time.Since(u.PasswordLastChanged)
	return int(dateDiff.Hours()/24) > expiry.Days || u.MustResetPassword
}

// DormancyCheckLastLogin check if the user has been away more than 90 days
func (u User) DormancyCheckLastLogin(tx *gorm.DB) bool {
	dormant := PasswordExpiry{}
	tx.First(&dormant)
	dateDiff := time.Since(u.LastLogin.Time)
	fmt.Println("####### dateDiff: -", dateDiff)
	fmt.Println("####### boolean for DormancyStatusCheck: -", int(dateDiff.Hours()/24) >= dormant.Dormancy || u.Locked)
	return int(dateDiff.Hours()/24) >= dormant.Dormancy || u.Locked
}

// DormancyOnLastLogin check if the user has been away more than 90 days
func DormancyOnLastLogin(tx *gorm.DB) error {
	dormant := PasswordExpiry{}
	u := User{}
	tx.First(&dormant)
	cutoff := time.Now().AddDate(0, 0, -90) // 90 days ago
	dateDiff := time.Since(u.LastLogin.Time)
	dorm := int(dateDiff.Hours() / 24)

	tx.Exec("update users set locked = 1, status = 'Dormant' where last_login <= ? ", cutoff)
	fmt.Println("####### Cuff-off and actual: -", cutoff, dorm)

	return nil
}

// DeletedUsersUpdate check if the user has been away more than 90 days
func DeletedUsersUpdate(tx *gorm.DB) error {
	tx.Exec("update users set locked = 1, status = 'Deleted' where deleted_at Not Null ? ")
	return nil
}

// CheckLastloginAndNow check how long user has been logged in before reseting is_logged_in status
func (u User) CheckLastloginAndNow(tx *gorm.DB) bool {
	dateDiff := time.Since(u.LastLogin.Time)
	fmt.Println("####### >>>>>>>>>> dateDiff: -", dateDiff)
	fmt.Println("####### >>>>>>>>>>> boolean for CheckLastloginAndNow:-", int(dateDiff.Hours()) >= 1 || u.IsLoggedIn)
	return int(dateDiff.Hours()) >= 1 || u.IsLoggedIn
}

// DormancyCheckLastLogin check if the user has been away more than 90 days
func (u User) ForceDormancy(tx *gorm.DB) {
	//dormant := PasswordExpiry{}
	u.Locked = true
	if u.PasswordLastChanged.IsZero() {
		u.PasswordLastChanged = time.Now()
	}
	if err := tx.Save(&u); err != nil {
		//panic(err)
		fmt.Println(err)
	}
}

// DormancyCheckLastLogin check if the user has been away more than 90 days
func (u User) LockDormancyExit(tx *gorm.DB) bool {
	dormant := PasswordExpiry{}
	tx.First(&dormant)
	dateDiff := time.Since(u.LastLogin.Time)
	if int(dateDiff.Hours()/24) >= dormant.Dormancy {
		//Exit user to login prompt
	}
	return true
}

// GetAllUsers returns all users
func GetAllUsers(tx *gorm.DB, search string) Users {
	users := make(Users, 0)
	search = fmt.Sprintf("%%%s%%", search)
	tx.Order("created_at desc").Preload("Branch").Preload("Roles.Role").Where("first_name like ? or last_name like ? or email like ? or ab_number like ?", search, search, search, search).Find(&users)
	return users
}

func UserIDsInBranch(tx *gorm.DB, branchID uuid.UUID) []uuid.UUID {
	users := make(Users, 0)
	tx.Where("branch_id=?", branchID).Find(&users)

	ids := make([]uuid.UUID, 0)
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	return ids
}

// UserBranchID this get the Branch of each user
func UserBranchID(tx *gorm.DB, userID uuid.UUID) uuid.UUID {
	branches := make([]Branch, 0)
	users := make([]User, 0)
	var id uuid.UUID
	tx.Where("id=?", userID).Preload("Branch").Find(&users)
	tx.Find(&branches)
	for _, branc := range branches {
		for _, userdata := range users {
			if branc.ID == userdata.BranchID {
				id = userdata.BranchID
			}
		}

	}
	return id
}

// RolesToString convert all roles to string
func (u User) RolesToString() string {
	roles := make([]string, 0)
	for _, role := range u.Roles {
		roles = append(roles, role.Role.Name)
	}
	return strings.Join(roles, " | ")
}

// CreateUser creates a new user
func CreateUser(branch, firstName, secondName, abNumber, email, password, privileged string, locked bool, tx *gorm.DB) User {
	user := User{}
	var privilege bool
	if privileged == "on" {
		privilege = true
		email = "admin.sutility@scb.com"
	} else {
		privilege = false
	}

	user.ID = NewUUID()
	user.BranchID = uuid.FromStringOrNil(branch)
	user.FirstName = firstName
	user.LastName = secondName
	user.Email = email
	user.ABNumber = abNumber
	user.Locked = locked
	user.Password = password
	user.MustResetPassword = true
	user.CreatedAt = time.Now()
	user.Deleted = nulls.Time{}
	user.PasswordLastChanged = time.Now()
	user.LastLogin = nulls.Time{}
	user.UpdatedAt = nulls.Time{}
	user.Privileged = privilege
	user.Status = nulls.NewString("Active")
	tx.Create(&user)

	userstamp := UserStampDetails{
		ID:        NewUUID(),
		UserID:    user.ID,
		BranchID:  user.BranchID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	// working fine
	fmt.Println("############## User.ID ##################", user.ID) // cannot insert into the stampify table as foreign key to user
	fmt.Println("############## Userstamp.ID ##################", userstamp.ID)

	// Adding the user created to the user Stampify group ##################
	tx.Create(&userstamp)
	return user
}

// SendWelcomeEmail sends a welcome email to the user
func (u User) SendWelcomeEmail() {
	mailers.SendWelcomeEmails(u.Email, u.Name())
}

// CreateAccessPolicies creates the users roles and perms
func (u User) CreateAccessPolicies(roles, permissions []string, tx *gorm.DB) {
	for _, permissionID := range permissions {
		permission := UserPermission{}
		permission.ID = NewUUID()
		permission.UserID = u.ID
		permission.PermissionID = uuid.FromStringOrNil(permissionID)
		tx.Create(&permission)
	}

	for _, roleID := range roles {
		role := UserRole{}
		role.ID = NewUUID()
		role.UserID = u.ID
		role.RoleID = uuid.FromStringOrNil(roleID)
		tx.Create(&role)
	}
}

// PaginateAllUsers returns all users
func PaginateAllUsers(search string, tx *gorm.DB, pagination pagination.Pagination) *pagination.Pagination {
	users := make(Users, 0)
	query := "%" + search + "%"
	tx.Scopes(Paginate(users, &pagination, tx)).Where("EXISTS(SELECT 1 FROM branches WHERE (name like ? or code like ?) AND id=branch_id) or first_name like ? or last_name like ? or ab_number like ? or email like ?", query, query, query, query, query, query).Preload("Branch").Preload("Roles.Role").Order("created_at desc").Find(&users)
	pagination.Rows = users
	return &pagination
}

// Edit edits a user
func (u *User) Edit(branch_id, firstName, lastName, abNumber, email, password string, locked, is_logged_in bool, tx *gorm.DB) User {
	//handling of privilege ID password
	fmt.Println("##########Password##############", password)
	if password != "" && len(password) > 8 {

		newPasswordHash := u.ResetPrivilegeIDPasswordm(u.ID, password, tx)
		// newPasswordHash := ResetPrivilegeIDPassword(password)
		u.MustResetPassword = false
		u.Password = newPasswordHash
		u.BranchID = uuid.FromStringOrNil(branch_id)
		u.FirstName = firstName
		u.LastName = lastName
		u.Email = email
		u.ABNumber = abNumber
		u.Locked = locked
		u.IsLoggedIn = is_logged_in
		u.UpdatedAt = nulls.Time{}
		fmt.Println("########## HashPassword ##############", newPasswordHash)
		// // Usage
		isValid := CheckPassword(password, newPasswordHash)
		fmt.Println("########## Comparing the input against DB password ##############", isValid)

		// tx.Exec("update users set branch_id=?, first_name=?, last_name=?,email=?, ab_number=?,locked=?,is_logged_in=?, password=? where id= ? ", uuid.FromStringOrNil(branch_id), firstName, lastName, email, abNumber, locked, is_logged_in, newPasswordHash, u.ID)
		tx.Exec("update users set branch_id=?, first_name=?, last_name=?,email=?, ab_number=?,locked=?,is_logged_in=?, password=? where id= ? ", uuid.FromStringOrNil(branch_id), firstName, lastName, email, abNumber, locked, is_logged_in, newPasswordHash, u.ID)
		return *u
	} else {
		//handling of Other ID User details Update
		u.BranchID = uuid.FromStringOrNil(branch_id)
		u.FirstName = firstName
		u.LastName = lastName
		u.Email = email
		u.ABNumber = abNumber
		u.Locked = locked
		u.IsLoggedIn = is_logged_in
		u.UpdatedAt = nulls.Time{}
		tx.Exec("update users set branch_id=?, first_name=?, last_name=?,email=?, ab_number=?,locked=?,is_logged_in=? where id= ? ", uuid.FromStringOrNil(branch_id), firstName, lastName, email, abNumber, locked, is_logged_in, u.ID)
		return *u

	}

}

// Reset Privilege ID Password  a new password hash
func ResetPrivilegeIDPassword(reset_password string) string {
	//newPassword := "new_temp_password"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reset_password), bcrypt.DefaultCost)
	if err != nil {
		// handle error
	}
	return string(hashedPassword)
}

func CheckPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// SyncAccessPolicies creates the users roles and perms
func (u User) SyncAccessPolicies(roles, permissions []string, tx *gorm.DB) {
	tx.Exec("delete from user_roles where user_id=?", u.ID)
	tx.Exec("delete from user_permissions where user_id=?", u.ID)

	for _, permissionID := range permissions {
		permission := UserPermission{}
		permission.ID = NewUUID()
		permission.UserID = u.ID
		permission.PermissionID = uuid.FromStringOrNil(permissionID)
		tx.Create(&permission)
	}

	for _, roleID := range roles {
		role := UserRole{}
		role.ID = NewUUID()
		role.UserID = u.ID
		role.RoleID = uuid.FromStringOrNil(roleID)
		tx.Create(&role)
	}
}

func DeleteUser(id uuid.UUID, tx *gorm.DB) {
	tx.Exec("update users set deleted=? where id=?", time.Now(), id)
}

func RestoreUser(id uuid.UUID, tx *gorm.DB) {
	tx.Exec("update users set deleted=? where id=?", nil, id)
}

func (u User) GetRoleIDsFromUserRoles() []string {
	roles := make([]string, 0)
	for _, role := range u.Roles {
		roles = append(roles, role.RoleID.String())
	}
	return roles
}

// HasEnteredInvalidCredentials verifies the login credentials provided by the user
func (u User) HasEnteredInvalidCredentials(tx *gorm.DB) {
	attempts := u.FailedLoginAttemptsSinceLastLogin(tx)
	w := PasswordExpiry{}
	tx.First(&w)
	fmt.Println("*****************ATTEMPTS: ", len(attempts), attempts)
	tries := w.Length
	if len(attempts) >= tries {
		u.LockAccount(tx)
	} else {
		u.LogAccountActivity(constants.IncorrectPassword, tx)
		if err := tx.Create(&FailedLoginAttempt{
			ID:     NewUUID(),
			UserID: u.ID,
		}); err != nil {
			//panic(err)
		}
	}
}

// FailedLoginAttemptsSinceLastLogin returns the failed login attempts since the last login
func (u User) FailedLoginAttemptsSinceLastLogin(tx *gorm.DB) FailedLoginAttempts {
	interval := time.Minute * 5
	attempts := make(FailedLoginAttempts, 0)
	fmt.Println("######Attempting Login 1 #########", u)

	if err := tx.Debug().Where("user_id=?", u.ID).Where("created_at > ?", u.LastLogin).Where("created_at > ?", time.Now().Add(-interval)).Find(&attempts); err != nil {

		//panic(err)
	}
	return attempts
}

func (u User) GetPermissionIDsFromUserPermissions() []string {
	permissions := make([]string, 0)
	for _, permission := range u.Permissions {
		permissions = append(permissions, permission.PermissionID.String())
	}
	return permissions
}

// ForcePasswordReset forces the user to perform a password reset
func (u User) ForcePasswordReset(tx *gorm.DB) {
	u.MustResetPassword = true
	if u.PasswordLastChanged.IsZero() {
		u.PasswordLastChanged = time.Now()
	}
	if err := tx.Save(&u); err != nil {
		//panic(err)
	}
}

// SaveLastLoginTime logs the user's last login time
func (u *User) SaveLastLoginTime(tx *gorm.DB) {
	u.LastLogin = nulls.NewTime(time.Now())
	u.IsLoggedIn = true
	// log.Println("user: ", u)
	if err := tx.Save(&u); err.Error != nil {
		// panic(err.Error)
	}
}

func SaveUserLastLogin(user *User, db *gorm.DB) {
	user.LastLogin = nulls.NewTime(time.Now())
	user.IsLoggedIn = true
	// log.Println("user: ", u)
	if err := db.Save(&user); err.Error != nil {
		// panic(err.Error)
	}

	//fmt.Println("@@@@@ Affected Rows: ", db.RowsAffected)
}

// AccountIsActive checks if the account is active
func (u User) AccountIsActive() bool {
	return !u.Locked && u.Deleted.Time.IsZero()
}

func SetUserLoginStatus(userId uuid.UUID, isLoggedIn bool, db *gorm.DB) {
	// fmt.Println("isLoggin: ", isLoggedIn)
	// fmt.Println("userId: ", userId)
	if err := db.Exec("update users set is_logged_in = ? where id = ?", isLoggedIn, userId).Error; err != nil {
		log.Println(err)
	}
	//fmt.Println("@@@@@ Affected Rows: ", db.RowsAffected)
}

// PaginateAllUsersDeleted returns all users marked To be Deleted Permanently
func PaginateAllUsersDeleted(search string, tx *gorm.DB, pagination pagination.Pagination) *pagination.Pagination {
	users := make(Users, 0)
	query := "%" + search + "%"
	tx.Scopes(Paginate(users, &pagination, tx)).Where("EXISTS(SELECT 1 FROM branches WHERE (name like ? or code like ?) AND id=branch_id) AND deleted is NOT NULL AND (first_name like ? or last_name like ? or ab_number like ? or email like ?)", query, query, query, query, query, query).Preload("Branch").Preload("Roles.Role").Order("created_at desc").Find(&users)
	pagination.Rows = users
	return &pagination
}

// RemoveUser delete a specific user in a table..............
func RemoveUser(uid uuid.UUID, tx *gorm.DB) {
	user := User{}
	fmt.Println("########## uid for deleted User ##############", uid, user.ABNumber)
	//tx.Exec("delete from users where deleted is NOT Null and id=?", uid)
	tx.Exec("update users set ab_number = ?, locked = ? where deleted is NOT Null and id=?", user.ABNumber+"DEL", true, uid).First(&user)
}

// LoadUserDetails loads user details
func LoadAllToBeDeletedUsersDetails(tx *gorm.DB) User {
	user := User{}
	tx.Where("deleted is not NULL").Preload("Branch").Preload("Roles.Role").Preload("Permissions.Permission").Find(&user)
	return user
}
