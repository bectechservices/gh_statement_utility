package requests

import (
	"ng-statement-app/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gofrs/uuid"
	"github.com/gookit/validate"
)

// CreateBranchRequest request type
type CreateBranchRequest struct {
	Name       string `form:"name"`
	Code       string `form:"code"`
	BankName   string `form:"bank_name"`
	StreetName string `form:"street_name"`
}

// Validate validates the form request
func (CreateBranchRequest) Validate(context buffalo.Context) (*validate.Validation, error) {
	request := context.Request()
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = false
	})

	validate.AddValidator("unique_name", func(name interface{}) bool {
		branchName := name.(string)
		branch := models.Branch{}
		result := models.GormDB.Where("name=?", branchName).Limit(1).Find(&branch)
		if result.Error != nil {
			panic(result.Error)
		}
		return result.RowsAffected == 0
	})

	validate.AddValidator("unique_code", func(code interface{}) bool {
		branchCode := code.(string)
		branch := models.Branch{}
		result := models.GormDB.Where("code=?", branchCode).Limit(1).Find(&branch)
		if result.Error != nil {
			panic(result.Error)
		}
		return result.RowsAffected == 0
	})

	data, err := validate.FromRequest(request)
	if err != nil {
		return nil, err
	}
	validator := data.Create()
	validator.AddRule("name", "required")
	validator.AddRule("code", "required")
	validator.AddRule("name", "unique_name")
	validator.AddRule("code", "unique_code")

	validator.AddMessages(map[string]string{
		"name.required":    "Branch name is required",
		"code.required":    "Branch code is required",
		"name.unique_name": "A branch with the given name already exists",
		"code.unique_code": "A branch with the given code already exists",
	})
	return validator, nil
}

// GetBoundValue gets the values in the request
func (CreateBranchRequest) GetBoundValue(c buffalo.Context) interface{} {
	request := CreateBranchRequest{}
	if err := c.Bind(&request); err == nil {
		return request
	}
	return request
}

// EditBranchRequest request type
type EditBranchRequest struct {
	ID         string `form:"id"`
	Name       string `form:"edit_name"`
	Code       string `form:"edit_code"`
	BankName   string `form:"edit_bank_name"`
	StreetName string `form:"edit_street_name"`
}

// Validate validates the form request
func (editBranchRequest EditBranchRequest) Validate(context buffalo.Context) (*validate.Validation, error) {
	request := context.Request()
	requestData := editBranchRequest.GetBoundValue(context).(EditBranchRequest)
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = false
	})

	validate.AddValidator("unique_name", func(edit_name interface{}) bool {
		branchName := edit_name.(string)

		branch := models.Branch{}
		result := models.GormDB.Where("name=? and id <> ?", branchName, uuid.FromStringOrNil(requestData.ID)).Limit(1).Find(&branch)
		if result.Error != nil {
			panic(result.Error)
		}
		return result.RowsAffected == 0
	})

	validate.AddValidator("unique_code", func(edit_code interface{}) bool {
		branchCode := edit_code.(string)

		branch := models.Branch{}
		result := models.GormDB.Where("code=? and id <> ?", branchCode, uuid.FromStringOrNil(requestData.ID)).Limit(1).Find(&branch)
		if result.Error != nil {
			panic(result.Error)
		}
		return result.RowsAffected == 0
	})

	data, err := validate.FromRequest(request)
	if err != nil {
		return nil, err
	}
	validator := data.Create()
	validator.AddRule("edit_name", "required")
	validator.AddRule("edit_code", "required")
	validator.AddRule("edit_name", "unique_name")
	validator.AddRule("edit_code", "unique_code")

	validator.AddMessages(map[string]string{
		"edit_name.required":    "Branch name is required",
		"edit_code.required":    "Branch code is required",
		"edit_name.unique_name": "A branch with the given name already exists",
		"edit_code.unique_code": "A branch with the given code already exists",
	})
	return validator, nil
}

// GetBoundValue gets the values in the request
func (EditBranchRequest) GetBoundValue(c buffalo.Context) interface{} {
	request := EditBranchRequest{}
	if err := c.Bind(&request); err == nil {
		return request
	}
	return request
}
