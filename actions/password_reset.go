package actions

import (
	"fmt"
	"gh-statement-app/constants"
	"gh-statement-app/models"
	"gh-statement-app/requests"
	"log"
	"net/http"
	"strings"

	"github.com/gobuffalo/buffalo"
)

// ShowExpiredPasswordReset shows the password reset page
func ShowExpiredPasswordReset(c buffalo.Context) error {
	policy := models.GetPasswordExpiry(DBConnection(c))
	c.Set("password_length", policy.Length)
	return c.Render(http.StatusOK, r.HTML("expired-password-reset.html"))
}

// ShowExpiredPasswordReset shows the password reset page
func ShowDormantLoginReset(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("expired-password-reset.html"))
}

// HandleExpiredPasswordReset handles the reset of the expired password
func HandleExpiredPasswordReset(c buffalo.Context) error {
	request := requests.AccountSetupRequest{}
	dataIsValid, validator, err := ValidateFormRequest(c, request)
	if err == nil {
		if dataIsValid {
			fmt.Println("############DataValid and Request 2")
			dbConnection := DBConnection(c)
			data := validator.SafeData()
			user := AuthUser(c)
			fmt.Println("$$$$AB User#################### ", user.ABNumber)
			user.LogAccountActivity(constants.PasswordResetAfterExpiration, dbConnection)
			//user.ResetPassword(user.ID, data["password"].(string), dbConnection)
			passwordChangeResult := user.ResetPassword(user.ID, data["new_password"].(string), data["confirm_password"].(string), data["token"].(string), dbConnection)
			fmt.Println("#########Password Results #################### ", passwordChangeResult)
			if passwordChangeResult != nil {
				return RedirectWithCustomError(&c, CustomError{
					Field:   "new_password",
					Error:   "Password Has been used Recently, Kindly Change to a Different one",
					Value:   data["new_password"].(string),
					IsError: true,
				}, request.GetBoundValue(c))
			}

			resetType := c.Session().Get("reset_type")
			log.Printf("===================================================== resertType: %v\n", resetType)
			if resetType != nil && strings.EqualFold(resetType.(string), "PSID") {
				c.Session().Delete("reset_type")
				c.Session().Delete("auth_id")
				c.Session().Clear()
			}

			c.Redirect(http.StatusFound, user.DashboardURL())
		}
		return RedirectWithErrors(&c, validator, request.GetBoundValue(c))
	}
	return err
}

// ShowForgotPasswordPage shows the forgot password page
func ShowForgotPasswordPage(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("forgot-password.html"))
}

// HandleForgotPassword handles the password reset
func HandleForgotPassword(c buffalo.Context) error {
	request := requests.ForgotPasswordRequest{}
	dataIsValid, validator, err := ValidateFormRequest(c, request)
	fmt.Println("################ Password Reset##################", dataIsValid)
	fmt.Println("################ Request ##################", request)
	if err == nil {
		if dataIsValid {
			dbConnection := DBConnection(c)
			data := validator.SafeData()
			fmt.Println("################ actions Data  ##################", data)
			user := models.GetUserByABNumber(data["ab_number"].(string), dbConnection)
			if !user.IsEmpty() {
				fmt.Println("################ user  ##################", user)

				user.LogAccountActivity(constants.ForgotPassword, dbConnection)
				token := RandomString(90)
				user.StoreForgottenPasswordToken(token, dbConnection)
				user.SendPasswordResetEmail(PasswordResetTokenURL(token))
				fmt.Println("################ token ##################", token)

			}
			return c.Redirect(http.StatusFound, IndexURL)
		}
		return RedirectWithErrors(&c, validator, request.GetBoundValue(c))
	}
	return err
}

// ShowPasswordSetupPage shows the password setup page
func ShowPasswordSetupPage(c buffalo.Context) error {
	policy := models.GetPasswordExpiry(DBConnection(c))
	c.Set("password_length", policy.Length)
	return c.Render(http.StatusOK, r.HTML("setup-password.html"))
}

// // HandlePasswordSetup handles the password setup
// func HandlePasswordSetup(c buffalo.Context) error {
// 	request := requests.AccountSetupRequest{}
// 	dataIsValid, validator, err := ValidateFormRequest(c, request)
// 	if err == nil {
// 		if dataIsValid {
// 			dbConnection := DBConnection(c)
// 			data := validator.SafeData()
// 			resetToken := GetResetTokenFromURL(c)
// 			user := models.GetUserFromResetToken(resetToken, dbConnection)
// 			user.LogAccountActivity(constants.SetupNewPassword, dbConnection)
// 			//user.ResetPassword(user.ID, request.NewPassword, request.ConfirmPassword, dbConnection)
// 			passwordChangeResult := user.ResetPassword(user.ID, data["new_password"].(string), data["confirm_password"].(string), data["token"].(string), dbConnection)
// 			if passwordChangeResult != nil {
// 				return RedirectWithCustomError(&c, CustomError{
// 					Field:   "new_password",
// 					Error:   "Password Has been Used Recently, Kindly Change to a Different one",
// 					Value:   data["new_password"].(string),
// 					IsError: true,
// 				}, request.GetBoundValue(c))
// 			}
// 			models.DestroyResetToken(resetToken, dbConnection)
// 			return c.Redirect(http.StatusFound, IndexURL)
// 		}
// 		return RedirectWithErrors(&c, validator, request.GetBoundValue(c))
// 	}
// 	return err
// }

func HandlePasswordSetup(c buffalo.Context) error {
	request := requests.AccountSetupRequest{}
	dataIsValid, validator, err := ValidateFormRequest(c, request)

	if err != nil {
		return err
	}

	if !dataIsValid {
		return RedirectWithErrors(&c, validator, request.GetBoundValue(c))
	}

	dbConnection := DBConnection(c)
	data := validator.SafeData()
	resetToken := GetResetTokenFromURL(c)

	fmt.Printf("################### dataIsValid: %v\n", dataIsValid)
	fmt.Printf("################### data: %v\n", data)
	fmt.Printf("################### resetToken: %s\n", resetToken)

	// Check if token exists and is valid
	user := models.GetUserFromResetToken(resetToken, dbConnection)
	if err != nil {
		fmt.Printf("################### Token validation error: %v\n", err)
		return RedirectWithCustomError(&c, CustomError{
			Field:   "token",
			Error:   "Invalid or expired password reset token",
			Value:   resetToken,
			IsError: true,
		}, request.GetBoundValue(c))
	}

	fmt.Printf("################### user: %+v\n", user)

	user.LogAccountActivity(constants.SetupNewPassword, dbConnection)

	// Pass the resetToken directly, not from data
	passwordChangeResult := user.ResetPassword(user.ID, data["new_password"].(string), data["confirm_password"].(string), resetToken, dbConnection)
	if passwordChangeResult != nil {
		fmt.Printf("################### Password reset error: %v\n", passwordChangeResult)
		return RedirectWithCustomError(&c, CustomError{
			Field:   "new_password",
			Error:   "Password Has been Used Recently, Kindly Change to a Different one",
			Value:   data["new_password"].(string),
			IsError: true,
		}, request.GetBoundValue(c))
	}

	models.DestroyResetToken(resetToken, dbConnection)

	resetType := c.Session().Get("reset_type")
	log.Printf("===================================================== resertType: %v\n", resetType)
	if resetType != nil && strings.EqualFold(resetType.(string), "PSID") {
		c.Session().Delete("reset_type")
		c.Session().Delete("auth_id")
		c.Session().Clear()
	}
	return c.Redirect(http.StatusFound, IndexURL)
}
