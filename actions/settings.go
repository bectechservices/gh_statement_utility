package actions

import (
	"fmt"
	"net/http"
	"ng-statement-app/models"
	"ng-statement-app/requests"

	"github.com/gobuffalo/buffalo"
)

// ShowSettingsPage loads the settings page
func ShowSettingsPage(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("settings.html"))
}

func ShowPasswordPolicyPage(c buffalo.Context) error {

	c.Set("expiry", models.GetPasswordExpiry(DBConnection(c)))
	return c.Render(http.StatusOK, r.HTML("password-policy.html"))
}

// func HandleUpdatePasswordPolicy(c buffalo.Context) error {
// 	request := requests.PasswordExpirySetupRequest{}
// 	dataIsValid, validator, err := ValidateFormRequest(c, request)
// 	if err == nil {
// 		if dataIsValid {
// 			dbConnection := DBConnection(c)
// 			request = request.GetBoundValue(c).(requests.PasswordExpirySetupRequest)
// 			password := models.GetPasswordExpiry(dbConnection)
// 			password.UpdatePasswordExpiry(request.Days, request.RemindIn, request.Length, request.Dormancy, request.Tries, dbConnection)
// 			return c.Redirect(http.StatusFound, PasswordPolicyURL)
// 		}
// 		return RedirectWithErrors(&c, validator, request.GetBoundValue(c))
// 	}
// 	return err
// }

func HandleUpdatePasswordPolicy(c buffalo.Context) error {
	request := requests.PasswordExpirySetupRequest{}
	dataIsValid, validator, err := ValidateFormRequest(c, request)

	if err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return c.Render(500, r.String("Internal server error during validation"))
	}

	if !dataIsValid {
		return RedirectWithErrors(&c, validator, request.GetBoundValue(c))
	}

	dbConnection := DBConnection(c)

	// Safely type assert the bound value
	boundValue := request.GetBoundValue(c)
	passwordRequest, ok := boundValue.(requests.PasswordExpirySetupRequest)
	if !ok {
		return c.Render(500, r.String("Invalid request data format"))
	}

	// Get the current password expiry settings (returns value and error)
	passwordExpiry := models.GetPasswordExpiry(dbConnection)
	if err != nil { // Check error instead of nil
		fmt.Printf("Error getting password expiry: %v\n", err)
		return c.Render(500, r.String("Password policy configuration not found"))
	}

	// Update with proper error handling
	passwordExpiry.UpdatePasswordExpiry(
		passwordRequest.Days,
		passwordRequest.RemindIn,
		passwordRequest.Length,
		passwordRequest.Dormancy,
		passwordRequest.Tries,
		dbConnection,
	)

	if err != nil {
		fmt.Printf("Password policy update error: %v\n", err)
		return c.Render(500, r.String("Failed to update password policy. Please try again."))
	}

	c.Flash().Add("success", "Password policy updated successfully!")
	return c.Redirect(http.StatusFound, PasswordPolicyURL)
}
