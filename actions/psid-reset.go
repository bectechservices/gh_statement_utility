package actions

import (
	"fmt"
	"gh-statement-app/models"
	"net/http"
	"strings"

	"github.com/gobuffalo/buffalo"
)

// ShowPSIDPasswordResetPage shows the PSID password Reset page
func ShowPSIDPasswordResetPage(c buffalo.Context) error {
	c.Set("ab_number_error", "")
	return c.Render(http.StatusOK, r.HTML("psid-reset.html"))
}

func HandlePsidReset(c buffalo.Context) error {
	fmt.Println("=== HANDLE PSID RESET CALLED ===")
	fmt.Printf("Method: %s\n", c.Request().Method)
	fmt.Printf("Content-Type: %s\n", c.Request().Header.Get("Content-Type"))
	abNumber := c.Request().FormValue("ab_number")
	fmt.Printf("AB Number received: '%s'\n", abNumber)
	fmt.Println("############### Checking PSID Reset 1 ###############")
	var user models.User

	// Store the input value for repopulation
	c.Set("ab_number", abNumber)
	// Validate empty field
	if strings.TrimSpace(abNumber) == "" {
		fmt.Println("######## Testing Empty PSID #################")
		c.Set("ab_number_error", "PSID Number is required")
		return c.Render(200, r.HTML("psid-reset.html"))
	}
	// Try querying the DB
	err := models.GormDB.
		Raw(
			"SELECT id, ab_number, privileged FROM users WHERE ab_number = ?",
			abNumber,
		).
		First(&user).Error
	if err == nil {
		// var privilege_user models.User

		err = models.GormDB.
			Raw(
				"SELECT id, ab_number, privileged FROM users WHERE ab_number = ? and privileged = 1",
				abNumber,
			).
			First(&user).Error
		if err != nil {
			fmt.Println("######## Existing ID/PSID but not privileged #################", err)
			c.Set("ab_number_error", "your username is NOt a PSId Number.")
			return c.Render(200, r.HTML("psid-reset.html"))
		}

		// Retry query after reconnecting

	} else {
		c.Set("ab_number_error", "Database error: "+err.Error())
		return c.Render(200, r.HTML("psid-reset.html"))
	}
	// }

	// Check privilege
	if !user.Privileged {
		c.Set("ab_number_error", "You are NOT allowed to use PSID Reset.")
		return c.Render(200, r.HTML("psid-reset.html"))
	}
	token := RandomString(90)
	user.StoreForgottenPasswordToken(token, models.GormDB)
	c.Session().Set("auth_id", user.ID)
	c.Session().Set("reset_type", "PSID")
	// Continue with password reset logic...

	//c.Flash().Add("success", "Password reset completed successfully.")
	return c.Redirect(302, PasswordResetTokenURL(token))
}
