package requests

import (
	"gh-statement-app/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gofrs/uuid"
	"github.com/gookit/validate"
)

// CreateUserRequest request type
type CreateUserRequest struct {
	FirstName   string   `form:"first_name"`
	LastName    string   `form:"last_name"`
	Email       string   `form:"email"`
	ABNumber    string   `form:"ab_number"`
	BranchID    string   `form:"branch_id"`
	Locked      bool     `form:"locked"`
	Privileged  string   `form:"privileged"`
	Roles       []string `form:"roles"`
	Permissions []string `form:"permissions"`
}

// EditUserRequest request type
type EditUserRequest struct {
	ID          uuid.UUID `form:"id"`
	FirstName   string    `form:"first_name"`
	LastName    string    `form:"last_name"`
	Email       string    `form:"email"`
	ABNumber    string    `form:"ab_number"`
	Password    string    `form:"password"`
	BranchID    string    `form:"branch_id"`
	Locked      bool      `form:"locked"`
	Privileged  bool      `form:"privileged"`
	IsLoggedIn  bool      `form:"is_logged_in"`
	Roles       []string  `form:"roles"`
	Permissions []string  `form:"permissions"`
}

// Validate validates the form request
func (CreateUserRequest) Validate(context buffalo.Context) (*validate.Validation, error) {
	request := context.Request()
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = false
	})

	validate.AddValidator("unique_email", func(name interface{}) bool {
		email := name.(string)
		user := models.User{}
		result := models.GormDB.Where("email=?", email).Limit(1).Find(&user)
		if result.Error != nil {
			panic(result.Error)
		}
		return result.RowsAffected == 0
	})

	validate.AddValidator("unique_ab_number", func(ab interface{}) bool {
		abNumber := ab.(string)
		user := models.User{}
		result := models.GormDB.Where("ab_number=?", abNumber).Limit(1).Find(&user)
		if result.Error != nil {
			panic(result.Error)
		}
		return result.RowsAffected == 0
	})

	validate.AddValidator("branch_exists", func(name interface{}) bool {
		branchID := name.(string)
		if nul := uuid.FromStringOrNil(branchID); nul == uuid.Nil {
			return false
		}
		branch := models.Branch{}
		result := models.GormDB.Where("id=?", branchID).Limit(1).Find(&branch)
		if result.Error != nil {
			panic(result.Error)
		}
		return result.RowsAffected != 0
	})

	data, err := validate.FromRequest(request)
	if err != nil {
		return nil, err
	}
	validator := data.Create()
	validator.FilterRule("locked", "bool")
	validator.AddRule("first_name", "required")
	validator.AddRule("last_name", "required")
	validator.AddRule("email", "required")
	validator.AddRule("email", "unique_email")
	validator.AddRule("ab_number", "required")
	validator.AddRule("ab_number", "unique_ab_number")
	validator.AddRule("branch_id", "required")
	validator.AddRule("branch_id", "branch_exists")

	validator.AddMessages(map[string]string{
		"first_name.required":        "First name is required",
		"last_name.required":         "Last name is required",
		"email.required":             "User email is required",
		"email.unique_email":         "A user with the given email already exists",
		"ab_number.required":         "Staff ID is required",
		"ab_number.unique_ab_number": "A user with the given Staff ID already exists",
		"branch_id.required":         "Branch required",
		"branch_id.branch_exists":    "Unknown Branch",
	})
	return validator, nil
}

// GetBoundValue gets the values in the request
func (CreateUserRequest) GetBoundValue(c buffalo.Context) interface{} {
	request := CreateUserRequest{}
	if err := c.Bind(&request); err == nil {
		return request
	}
	return request
}

// Validate validates the form request
func (EditUserRequest) Validate(context buffalo.Context) (*validate.Validation, error) {
	request := context.Request()
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = false
	})

	validate.AddValidator("branch_exists", func(name interface{}) bool {
		branchID := name.(string)
		if nul := uuid.FromStringOrNil(branchID); nul == uuid.Nil {
			return false
		}
		branch := models.Branch{}
		result := models.GormDB.Where("id=?", branchID).Limit(1).Find(&branch)
		if result.Error != nil {
			panic(result.Error)
		}
		return result.RowsAffected != 0
	})

	data, err := validate.FromRequest(request)
	if err != nil {
		return nil, err
	}
	validator := data.Create()
	validator.FilterRule("locked", "bool")
	validator.FilterRule("is_logged_in", "bool")
	validator.AddRule("first_name", "required")
	validator.AddRule("last_name", "required")
	validator.AddRule("email", "required")
	validator.AddRule("ab_number", "required")
	validator.AddRule("branch_id", "required")
	//validator.AddRule("branch_id", "branch_exists")

	validator.AddMessages(map[string]string{
		"first_name.required": "First name is required",
		"last_name.required":  "Last name is required",
		"email.required":      "User email is required",
		"ab_number.required":  "Staff ID is required",
		"branch_id.required":  "Branch required",
		//	"branch_id.branch_exists": "Unknown Branch",
	})
	return validator, nil
}

// GetBoundValue gets the values in the request
func (EditUserRequest) GetBoundValue(c buffalo.Context) interface{} {
	request := EditUserRequest{}
	if err := c.Bind(&request); err == nil {
		return request
	}
	return request
}
