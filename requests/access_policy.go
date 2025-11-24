package requests

import (
	"ng-statement-app/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gofrs/uuid"
	"github.com/gookit/validate"
)

// CreateRoleRequest request type
type CreateRoleRequest struct {
	Name        string   `form:"name"`
	Description string   `form:"description"`
	Permissions []string `form:"permissions"`
}

// Validate validates the form request
func (CreateRoleRequest) Validate(context buffalo.Context) (*validate.Validation, error) {
	request := context.Request()
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = false
	})
	validate.AddValidator("does_not_exist", func(name interface{}) bool {
		roleName := name.(string)
		role := models.Role{}
		result := models.GormDB.Where("name=?", roleName).Limit(1).Find(&role)
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
	validator.AddRule("name", "does_not_exist")

	validator.AddMessages(map[string]string{
		"name.required":       "Role name is required",
		"name.does_not_exist": "A role with the given name already exists",
	})
	return validator, nil
}

// GetBoundValue gets the values in the request
func (CreateRoleRequest) GetBoundValue(c buffalo.Context) interface{} {
	request := CreateRoleRequest{}
	if err := c.Bind(&request); err == nil {
		return request
	}
	return request
}

// CreatePermissionRequest request type
type CreatePermissionRequest struct {
	Name        string   `form:"name"`
	Description string   `form:"description"`
	Routes      []string `form:"routes"`
}

// Validate validates the form request
func (CreatePermissionRequest) Validate(context buffalo.Context) (*validate.Validation, error) {
	request := context.Request()
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = false
	})
	validate.AddValidator("does_not_exist", func(name interface{}) bool {
		permissionName := name.(string)
		permission := models.Permission{}
		result := models.GormDB.Where("name=?", permissionName).Limit(1).Find(&permission)
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
	validator.AddRule("name", "does_not_exist")

	validator.AddMessages(map[string]string{
		"name.required":       "Permission name is required",
		"name.does_not_exist": "A permission with the given name already exists",
	})
	return validator, nil
}

// GetBoundValue returns the values submitted
func (CreatePermissionRequest) GetBoundValue(c buffalo.Context) interface{} {
	request := CreatePermissionRequest{}
	if err := c.Bind(&request); err == nil {
		return request
	}
	return request
}

// EditRoleRequest request type
type EditRoleRequest struct {
	ID          string   `form:"id"`
	Name        string   `form:"name"`
	Description string   `form:"description"`
	Permissions []string `form:"permissions"`
}

// Validate validates the form request
func (EditRoleRequest) Validate(context buffalo.Context) (*validate.Validation, error) {
	request := context.Request()
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = false
	})

	data, err := validate.FromRequest(request)
	roleID, _ := data.Get("id")
	if err != nil {
		return nil, err
	}

	validate.AddValidator("exists_for_only_this_id", func(name interface{}) bool {
		roleName := name.(string)
		role := models.Role{}
		models.GormDB.Where("name=?", roleName).First(&role)
		return uuid.FromStringOrNil(roleID.(string)) == role.ID
	})

	validate.AddValidator("exists", func(id interface{}) bool {
		roleID := id.(string)
		role := models.Role{}
		result := models.GormDB.Where("id=?", uuid.FromStringOrNil(roleID)).Limit(1).Find(&role)
		if result.Error != nil {
			panic(result.Error)
		}
		return result.RowsAffected != 0
	})

	validator := data.Create()
	validator.AddRule("id", "required")
	validator.AddRule("id", "exists")
	validator.AddRule("name", "required")
	validator.AddRule("name", "exists_for_only_this_id")

	validator.AddMessages(map[string]string{
		"id.required":                  "Role id is required",
		"id.exists":                    "Role id does not exist",
		"name.required":                "Role name is required",
		"name.exists_for_only_this_id": "A role with the given name already exists",
	})
	return validator, nil
}

// GetBoundValue gets the values in the request
func (EditRoleRequest) GetBoundValue(c buffalo.Context) interface{} {
	request := EditRoleRequest{}
	if err := c.Bind(&request); err == nil {
		return request
	}
	return request
}

// EditPermissionRequest request type
type EditPermissionRequest struct {
	ID          string   `form:"id"`
	Name        string   `form:"name"`
	Description string   `form:"description"`
	Routes      []string `form:"routes"`
}

// Validate validates the form request
func (EditPermissionRequest) Validate(context buffalo.Context) (*validate.Validation, error) {
	request := context.Request()
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = false
	})

	data, err := validate.FromRequest(request)
	permissionID, _ := data.Get("id")
	if err != nil {
		return nil, err
	}

	validate.AddValidator("exists_for_ony_this_id", func(name interface{}) bool {
		permissionName := name.(string)
		permission := models.Permission{}
		models.GormDB.Where("name=?", permissionName).First(&permission)
		return uuid.FromStringOrNil(permissionID.(string)) == permission.ID
	})
	validate.AddValidator("exists", func(id interface{}) bool {
		permissionID := id.(string)
		permission := models.Permission{}
		result := models.GormDB.Where("id=?", uuid.FromStringOrNil(permissionID)).Limit(1).Find(&permission)
		if result.Error != nil {
			panic(result.Error)
		}
		return result.RowsAffected != 0

	})

	validator := data.Create()
	validator.AddRule("id", "required")
	validator.AddRule("id", "exists")
	validator.AddRule("name", "required")
	validator.AddRule("name", "exists_for_ony_this_id")

	validator.AddMessages(map[string]string{
		"id.required":                 "Permission id is required",
		"id.exists":                   "Permission id does not exist",
		"name.required":               "Permission name is required",
		"name.exists_for_ony_this_id": "A permission with the given name already exists",
	})
	return validator, nil
}

// GetBoundValue returns the values submitted
func (EditPermissionRequest) GetBoundValue(c buffalo.Context) interface{} {
	request := EditPermissionRequest{}
	if err := c.Bind(&request); err == nil {
		return request
	}
	return request
}
