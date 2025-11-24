package requests

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gookit/validate"
)

// PasswordExpirySetupRequest request type
type StampifyUserSetupRequest struct {
	UserID            string `form:"user_id"`
	BranchID          string `form:"branch_id"`
	Position          string `form:"position"`
	BranchStamp       bool   `form:"branch_stamp"`
	SignatureLocation string `form:"file_base64" `
}

// Validate validates the request
func (stampifyUserSetupRequest StampifyUserSetupRequest) Validate(context buffalo.Context) (*validate.Validation, error) {
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

	validator.AddRule("user_id", "required")
	validator.AddRule("position", "required")
	// validator.AddRule("bank_name", "required")
	// validator.AddRule("branch_or_unit_name", "required")
	//add more rules

	validator.AddMessages(map[string]string{
		"user_id.required":  "user ID is required",
		"position.required": "Position is required",
		// "bank_name.required":           "Bank Name is required",
		// "branch_or_unit_name.required": "Branch or Unit Name is required",
	})
	return validator, nil
}

// GetBoundValue returns the values in the request
func (StampifyUserSetupRequest) GetBoundValue(c buffalo.Context) interface{} {
	request := StampifyUserSetupRequest{}
	if err := c.Bind(&request); err == nil {
		return request
	}
	return request
}
