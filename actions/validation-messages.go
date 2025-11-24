package actions

func PopulateAndGetValidationMsg() map[string]string {
	msg := make(map[string]string)

	msg["required"] = "is required"
	msg["min"] = "should be more than"
	msg["max"] = "should be less than"
	msg["has-numeric"] = "should contain numbers (1-9)"
	msg["has-uppercase"] = "should contain uppercase letters (A-Z)"
	msg["has-lowercase"] = "contain lowercase letters (a-z)"
	msg["has-special-chars"] = `should contain special characters (!,@,#,$,"%" etc)`
	msg["password-confirmation"] = "Password and password confirmation do not match"
	msg["current-password"] = "New password cannot be the same as current password"

	return msg
}
