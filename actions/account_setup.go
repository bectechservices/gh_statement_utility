package actions

import (
	"net/http"

	"gh-statement-app/constants"
	"gh-statement-app/requests"

	"github.com/gobuffalo/buffalo"
)

// ShowAccountSetup show account setup page
func ShowAccountSetup(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("account-setup.html"))
}

// HandleAccountSetup setup user account
func HandleAccountSetup(c buffalo.Context) error {
	request := requests.AccountSetupRequest{}
	dataIsValid, validator, err := ValidateFormRequest(c, request)
	if err == nil {
		if dataIsValid {
			dbConnection := DBConnection(c)
			data := validator.SafeData()
			user := AuthUser(c)
			user.LogAccountActivity(constants.FirstTimeLoginPasswordChange, dbConnection)
			user.ResetPassword(user.ID, data["new_password"].(string), data["confirm_password"].(string), data["token"].(string), dbConnection)
			c.Redirect(http.StatusFound, user.DashboardURL())
		}
		return RedirectWithErrors(&c, validator, request.GetBoundValue(c))
	}
	return err
}
