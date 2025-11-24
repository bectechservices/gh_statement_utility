package actions

import (
	"fmt"
	"gh-statement-app/constants"
	"gh-statement-app/models"
	"gh-statement-app/requests"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gofrs/uuid"
)

// ErrorHandler shows the error page
func ErrorHandler(status int, err error, c buffalo.Context) error {
	if c.Request().Header.Get("Accept") == "application/json" || c.Request().Header.Get("Content-Type") == "application/json" {
		return c.Error(status, nil)
	}
	c.Set("error_message", err.Error())
	return c.Render(http.StatusOK, r.HTML(fmt.Sprintf("%d.html", status)))
}

// HandleLogin performs the login logic
func HandleLogin(c buffalo.Context) error {
	request := requests.LoginRequest{}
	dataIsValid, validator, err := ValidateFormRequest(c, request)
	if err == nil {
		if dataIsValid {
			dbConnection := DBConnection(c)
			data := validator.SafeData()
			user := models.GetUserByABNumber(data["ab_number"].(string), dbConnection)
			if models.GrantRolesAccessActivity(user.ABNumber, dbConnection) {
				if user.AccountIsActive() {
					if user.PasswordIsValid(data["password"].(string)) {
						c.Session().Set("auth_id", user.ID)
						user.LogAccountActivity(constants.UserLogin, dbConnection)
						if user.IsFirstTimeLogin() {
							return c.Redirect(http.StatusFound, AccountSetupURL)
						}
						if user.PasswordHasExpired(dbConnection) {
							user.ForcePasswordReset(dbConnection)
							return c.Redirect(http.StatusFound, ExpiredPasswordResetURL)
						}
						user.SaveLastLoginTime(dbConnection)
						return c.Redirect(http.StatusFound, user.DashboardURL())
					}
					user.HasEnteredInvalidCredentials(dbConnection)
					return RedirectWithCustomError(&c, CustomError{
						Field:   "password",
						Error:   "Password incorrect",
						Value:   data["password"].(string),
						IsError: true,
					}, request.GetBoundValue(c))

				}
				fmt.Println("############## Wrong Password for user ################################")
				return RedirectWithCustomError(&c, CustomError{
					Field:   "ab_number",
					Error:   "Account Locked",
					Value:   data["ab_number"].(string),
					IsError: true,
				}, request.GetBoundValue(c))

			}
			fmt.Println("############## Access Denied Feature ################################")
			user.LogAccountActivity(constants.UnauthorizedTimeAndDay, models.GormDB)
			return RedirectWithCustomError(&c, CustomError{
				Field:   "ab_number",
				Error:   "Time and Date Restriction",
				Value:   data["ab_number"].(string),
				IsError: true,
			}, request.GetBoundValue(c))

		}
		return RedirectWithErrors(&c, validator, request.GetBoundValue(c))
	}
	return err
}

// HandleLogout logout
func HandleLogout(c buffalo.Context) error {
	request := requests.LogoutRequest{}
	if err := c.Bind(&request); err != nil {
		return err
	}

	var user models.User

	if uid := c.Session().Get("auth_id"); uid != nil {
		switch id := uid.(type) {
		case uuid.UUID:
			user = models.GetUserByID(id, models.GormDB)
		}
	}

	// set user login status to false
	models.SetUserLoginStatus(user.ID, false, models.GormDB)
	c.Session().Clear()
	return c.Redirect(http.StatusFound, IndexURL)
}

func HandleNoPermissionAssigned(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("no-permission.html"))
}
