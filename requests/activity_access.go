package requests

import (
	"fmt"

	"github.com/gobuffalo/buffalo"
	"github.com/gookit/validate"
)

// PasswordExpirySetupRequest request type
type ActivityAccessSetupRequest struct {
	Monday    bool   `form:"monday"`
	Tuesday   bool   `form:"tuesday"`
	Wednesday bool   `form:"wednesday"`
	Thursday  bool   `form:"thursday"`
	Friday    bool   `form:"friday"`
	Saturday  bool   `form:"saturday"`
	Sunday    bool   `form:"sunday"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}

// Validate validates the request
func (ActivityAccessSetupRequest) Validate(context buffalo.Context) (*validate.Validation, error) {
	request := context.Request()
	fmt.Println("######## Request level 0 ##########", request)
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = false
	})

	data, err := validate.FromRequest(request)
	fmt.Println("######## Request level 1 ##########", data)
	if err != nil {
		return nil, err
	}
	validator := data.Create()
	fmt.Println("######## Request level 2 ##########", validator)
	validator.AddRule("start_time", "required")
	validator.AddRule("end_time", "required")
	//add more rules

	validator.AddMessages(map[string]string{
		"start_time.required": "Start Time is required",
		"end_time.required":   "End Time is required",
	})
	return validator, nil
}

// GetBoundValue returns the values in the request
func (ActivityAccessSetupRequest) GetBoundValue(c buffalo.Context) interface{} {
	request := ActivityAccessSetupRequest{}
	if err := c.Bind(&request); err == nil {
		return request
	}
	return request
}
