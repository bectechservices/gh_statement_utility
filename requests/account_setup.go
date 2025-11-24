package requests

import (
	"fmt"

	"github.com/gobuffalo/buffalo"
	"github.com/gookit/validate"
)

// AccountSetupRequest request type
type AccountSetupRequest struct {
	NewPassword     string `form:"new_password"`
	ConfirmPassword string `form:"confirm_password"`
}

// Validate validates the request
func (AccountSetupRequest) Validate(context buffalo.Context) (*validate.Validation, error) {
	request := context.Request()
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = false
	})

	data, err := validate.FromRequest(request)
	if err != nil {
		return nil, err
	}
	fmt.Println("######### Testing Password Reset (data) ################", data)
	validator := data.Create()
	validator.AddRule("new_password", "required")
	validator.AddRule("confirm_password", "required")

	validator.AddMessages(map[string]string{
		"new_password.required":     "new Password is required",
		"confirm_password.required": "confirm Password is required",
	})
	fmt.Println("######### Testing Password Reset (Validator) ################", validator)
	return validator, nil
}

// GetBoundValue returns the values in the request
func (AccountSetupRequest) GetBoundValue(c buffalo.Context) interface{} {
	request := AccountSetupRequest{}
	fmt.Println("######### Testing Password Reset (GetBoundValue) ################", request)
	if err := c.Bind(&request); err == nil {
		return request
	}
	return request
}
