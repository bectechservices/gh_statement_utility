package requests

import (
	"fmt"

	"github.com/gobuffalo/buffalo"
	"github.com/gookit/validate"
)

// ForgotPasswordRequest type
type ForgotPasswordRequest struct {
	ABNumber string `form:"ab_number"`
}

// Validate validates the request
func (ForgotPasswordRequest) Validate(context buffalo.Context) (*validate.Validation, error) {
	request := context.Request()
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = false
	})

	data, err := validate.FromRequest(request)
	fmt.Println("*************************Data*******************", data)
	if err != nil {
		return nil, err
	}
	validator := data.Create()
	validator.AddRule("ab_number", "required")

	validator.AddMessages(map[string]string{
		"ab_number.required": "Staff ID is required",
	})
	return validator, nil
}

// GetBoundValue returns the values in the request
func (ForgotPasswordRequest) GetBoundValue(c buffalo.Context) interface{} {
	request := ForgotPasswordRequest{}
	if err := c.Bind(&request); err == nil {
		return request
	}
	return request
}
