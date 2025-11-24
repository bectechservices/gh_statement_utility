package requests

import (
	"fmt"
	"gh-statement-app/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gofrs/uuid"
	"github.com/gookit/validate"
)

// LoginRequest login request
type LoginRequest struct {
	ABNumber string `form:"ab_number"`
	Password string `form:"password"`
}

// Validate validates the login request
func (LoginRequest) Validate(context buffalo.Context) (*validate.Validation, error) {
	request := context.Request()
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = false
	})

	validate.AddValidator("user_exists", func(ab_number interface{}) bool {
		id := ab_number.(string)
		user := models.User{}

		models.GormDB.Where("ab_number=?", id).First(&user)
		return user.ID != uuid.Nil
	})
	fmt.Println("##########Request info##########", request)
	data, err := validate.FromRequest(request)
	if err != nil {
		return nil, err
	}
	fmt.Println("##########Data##########", data)

	validator := data.Create()
	validator.AddRule("ab_number", "required")
	validator.AddRule("ab_number", "user_exists")
	validator.AddRule("password", "required")

	validator.AddMessages(map[string]string{
		"ab_number.required":    "A StaffID is required to login",
		"ab_number.user_exists": "No user exists with the StaffID provided",
		"password.required":     "Password required",
	})
	return validator, nil
}

// GetBoundValue gets a value from the request
func (LoginRequest) GetBoundValue(c buffalo.Context) interface{} {
	request := LoginRequest{}

	if err := c.Bind(&request); err == nil {
		return request
	}
	return request
}

// LogoutRequest logout request type
type LogoutRequest struct {
	ActionType string `form:"action_type"`
}
