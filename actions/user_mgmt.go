package actions

import (
	"fmt"
	"net/http"
	"ng-statement-app/constants"
	"ng-statement-app/models"
	"ng-statement-app/requests"

	"github.com/gobuffalo/buffalo"
	"github.com/gofrs/uuid"
)

// ShowCreateUserPage shows the create user page
func ShowCreateUserPage(c buffalo.Context) error {
	dbConnection := DBConnection(c)
	c.Set("branches", models.LoadAllBranches(dbConnection))
	c.Set("permissions", models.LoadAllPermissions(dbConnection))
	c.Set("roles", models.LoadAllRoles(dbConnection))
	return c.Render(http.StatusOK, r.HTML("create-user.html"))
}

// HandleCreateUser creates the user
func HandleCreateUser(c buffalo.Context) error {
	request := requests.CreateUserRequest{}
	dataIsValid, validator, err := ValidateFormRequest(c, request)
	if err == nil {
		if dataIsValid {
			request = request.GetBoundValue(c).(requests.CreateUserRequest)
			dbConnection := DBConnection(c)
			user := models.CreateUser(request.BranchID, request.FirstName, request.LastName, request.ABNumber, request.Email, RandomString(15), request.Privileged, request.Locked, dbConnection)
			user.CreateAccessPolicies(request.Roles, request.Permissions, dbConnection)
			models.CreateActivityAudit(constants.UserAccountCreate, "Account created", AuthID(c), user.ID, dbConnection)
			user.SendWelcomeEmail()
			return c.Redirect(http.StatusFound, AllUsersURL)
		}
		return RedirectWithErrors(&c, validator, request.GetBoundValue(c))
	}
	return err
}

// HandleLoadAllUsers loads all users
func HandleLoadAllUsers(c buffalo.Context) error {
	search := c.Param("search")
	pagination := models.PaginateAllUsers(search, DBConnection(c), DBPaginator(c, 10))
	c.Set("pagination", pagination)
	c.Set("search", search)
	return c.Render(http.StatusOK, r.HTML("users.html"))
}

// HandleLoadSpecificUser loads a user
func HandleLoadSpecificUser(c buffalo.Context) error {
	uid := uuid.FromStringOrNil(c.Param("user"))
	dbConnection := DBConnection(c)
	c.Set("user", models.LoadUserDetails(uid, dbConnection))
	c.Set("last_login", models.LoadUserLastLogin(uid, dbConnection))
	return c.Render(http.StatusOK, r.HTML("user.html"))
}

// HandleShowUserEdit shows the edit form
func HandleShowUserEdit(c buffalo.Context) error {
	uid := uuid.FromStringOrNil(c.Param("user"))
	dbConnection := DBConnection(c)
	c.Set("user", models.LoadUserDetails(uid, dbConnection))
	c.Set("branches", models.LoadAllBranches(dbConnection))
	c.Set("permissions", models.LoadAllPermissions(dbConnection))
	c.Set("roles", models.LoadAllRoles(dbConnection))
	return c.Render(http.StatusOK, r.HTML("user-edit.html"))
}

// HandleUserEdit performs the user edit logic
func HandleUserEdit(c buffalo.Context) error {
	request := requests.EditUserRequest{}
	dataIsValid, validator, err := ValidateFormRequest(c, request)
	fmt.Print("########### DataValidity ##############", dataIsValid)
	if err == nil {
		if dataIsValid {
			request = request.GetBoundValue(c).(requests.EditUserRequest)
			dbConnection := DBConnection(c)
			//data := validator.SafeData()
			user := models.LoadUserDetails(request.ID, dbConnection)
			description := "Account edit"
			if request.FirstName != user.FirstName {
				description += "--:Change First Name"
			}
			if request.LastName != user.LastName {
				description += "--:Change Last Name"
			}
			if request.ABNumber != user.ABNumber {
				description += "--:Change Staff ID"
			}
			if request.Email != user.Email {
				description += "--:Change Email"
			}
			if request.Locked != user.Locked {
				description += "--:Change Lock Status"
			}
			if request.IsLoggedIn != user.IsLoggedIn {
				description += "--:Change Is_Logged_In Status"
			}
			if request.BranchID != user.BranchID.String() {
				description += "--:Change Branch"
			}
			if request.Password != user.Password {
				description += "--:Privileged User Password Change"
			}
			//user.ResetPrivilegeIDPasswordm(user.ID, data["password"].(string), dbConnection)
			user.Edit(request.BranchID, request.FirstName, request.LastName, request.ABNumber, request.Email, request.Password, request.Locked, request.IsLoggedIn, dbConnection)
			if AuthID(c) != user.ID {
				if !SliceContainsSameElementsWithoutOrder(request.Roles, user.GetRoleIDsFromUserRoles()) {
					description += "--:Change Roles"
				}
				if !SliceContainsSameElementsWithoutOrder(request.Permissions, user.GetPermissionIDsFromUserPermissions()) {
					description += "--:Change Permissions"
				}
				user.SyncAccessPolicies(request.Roles, request.Permissions, dbConnection)
			}
			models.CreateActivityAudit(constants.UserAccountEdit, description, AuthID(c), request.ID, dbConnection)
			return c.Redirect(http.StatusFound, AllUsersURL)
		}
		return RedirectWithErrors(&c, validator, request.GetBoundValue(c))
	}
	return err
}

func HandleRestoreUser(c buffalo.Context) error {
	req := &requests.OnlyID{}
	if err := c.Bind(req); err != nil {
		return err
	}
	dbConnection := DBConnection(c)
	uid := uuid.FromStringOrNil(req.ID)
	models.RestoreUser(uid, dbConnection)
	models.CreateActivityAudit(constants.UserAccountRestore, "Account restore", AuthID(c), uid, dbConnection)
	return RedirectBack(&c)
}

func HandleDeleteUser(c buffalo.Context) error {
	req := &requests.OnlyID{}
	if err := c.Bind(req); err != nil {
		return err
	}
	dbConnection := DBConnection(c)
	uid := uuid.FromStringOrNil(req.ID)
	models.DeleteUser(uid, dbConnection)
	models.CreateActivityAudit(constants.UserAccountDelete, "Account delete", AuthID(c), uid, dbConnection)
	return RedirectBack(&c)
}

// HandleLoadAllUsers loads all users
func HandleLoadAllUsersDeleted(c buffalo.Context) error {
	search := c.Param("search")
	pagination := models.PaginateAllUsersDeleted(search, DBConnection(c), DBPaginator(c, 10))
	c.Set("pagination", pagination)
	c.Set("search", search)
	return c.Render(http.StatusOK, r.HTML("users-deleted.html"))
}

// HandleLoadSpecificRemovedUser
func HandleLoadSpecificRemovedUser(c buffalo.Context) error {
	uid := uuid.FromStringOrNil(c.Param("user"))
	dbConnection := DBConnection(c)
	c.Set("user", models.LoadUserDetails(uid, dbConnection))
	//	c.Set("user", models.RemoveUser(uid, dbConnection))
	return c.Render(http.StatusOK, r.HTML("userremoved.html"))
}

func HandleRemoveUser(c buffalo.Context) error {
	uid := uuid.FromStringOrNil(c.Param("user"))
	// req := &requests.OnlyID{}
	// if err := c.Bind(req); err != nil {
	// 	return err
	// }
	fmt.Println("##########Action ---1-- testing Mgmt_user ##############", uid)
	dbConnection := DBConnection(c)
	//uid := uuid.FromStringOrNil(req.ID)
	models.CreateActivityAudit(constants.UserAccountRemoved, "Account Remove", AuthID(c), uid, dbConnection)
	fmt.Println("########## [AuthID(c)] Action uid for deleted User ##############", AuthID(c))
	models.RemoveUser(uid, dbConnection)
	return RedirectBack(&c)
}

// HandleLoadSpecificDeleteUser loads a user
func HandleLoadSpecificDeleteUser(c buffalo.Context) error {
	uid := uuid.FromStringOrNil(c.Param("user"))
	dbConnection := DBConnection(c)
	c.Set("user", models.LoadUserDetails(uid, dbConnection))
	c.Set("last_login", models.LoadUserLastLogin(uid, dbConnection))
	return c.Render(http.StatusOK, r.HTML("userdelete.html"))
}
