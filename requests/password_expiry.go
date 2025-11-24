package requests

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gookit/validate"
)

// PasswordExpirySetupRequest request type
type PasswordExpirySetupRequest struct {
	Days     int `form:"days"`
	RemindIn int `form:"remindin"`
	Length   int `form:"length"`
	Dormancy int `form:"dormancy"`
	Tries    int `form:"tries"`
}

// Validate validates the request
func (PasswordExpirySetupRequest) Validate(context buffalo.Context) (*validate.Validation, error) {
	request := context.Request()
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = false
	})

	data, err := validate.FromRequest(request)
	if err != nil {
		return nil, err
	}
	validator := data.Create()
	validator.AddRule("days", "required")
	validator.AddRule("remindin", "required")
	//add more rules

	validator.AddMessages(map[string]string{
		"days.required":     "Password expiry days is required",
		"remindin.required": "Password expiry reminder is required",
	})
	return validator, nil
}

// GetBoundValue returns the values in the request
func (PasswordExpirySetupRequest) GetBoundValue(c buffalo.Context) interface{} {
	request := PasswordExpirySetupRequest{}
	if err := c.Bind(&request); err == nil {
		return request
	}
	return request
}
