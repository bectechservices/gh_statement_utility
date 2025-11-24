package requests

import (
	"database/sql"
	"fmt"
	"gh-statement-app/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
	"github.com/gookit/validate"
)

type CreateRoleRestrictionRequest struct {
	Monday    bool   `form:"monday"`
	Tuesday   bool   `form:"tuesday"`
	Wednesday bool   `form:"wednesday"`
	Thursday  bool   `form:"thursday"`
	Friday    bool   `form:"friday"`
	Saturday  bool   `form:"saturday"`
	Sunday    bool   `form:"sunday"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
	RoleID    string `form:"add_role_id"`
	// models.Role
}

// Validate validates the form request
func (CreateRoleRestrictionRequest) Validate(context buffalo.Context) (*validate.Validation, error) {
	request := context.Request()
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = false
	})
	// fmt.Println("########################### Request CreateRoleRestrictionRequest ###########")
	// validate.AddValidator("does_not_exist", func(id interface{}) bool {
	// 	roleID := id.(string)
	// 	db := context.Value("tx").(*pop.Connection)
	// 	role := models.Role{}
	// 	var exists bool
	// 	exists, err := db.Where("id=?", roleID).Exists(&role)
	// 	fmt.Println("########################### Request Data ---- roleID ###########", roleID)
	// 	if err != nil {
	// 		if err == sql.ErrNoRows {
	// 			return true
	// 		}
	// 		panic(err)
	// 	}
	// 	return !exists
	// })

	data, err := validate.FromRequest(request)
	if err != nil {
		return nil, err
	}
	fmt.Println("########################### Request Data ###########", data)
	validator := data.Create()
	fmt.Println("########################### Request validator ###########", validator)
	validator.AddRule("start_time", "required")
	validator.AddRule("end_time", "required")

	validator.AddMessages(map[string]string{
		"start_time.required": " A start date is required",
		"end_time.required":   " A end date is required",
	})
	return validator, nil
}

// GetBoundValue gets the values in the request
func (CreateRoleRestrictionRequest) GetBoundValue(c buffalo.Context) interface{} {
	request := CreateRoleRestrictionRequest{}

	if err := c.Bind(&request); err != nil {
		fmt.Println("########################### Request request err ###########", err)
		panic(err)
	}
	fmt.Println("########################### Request request ###########", request)
	return request
}

//--------------------------------------------------------------------------------------------------------------------//

// EditRoleRequest request type
type EditRoleRestrictionRequest struct {
	ID        string `form:"id"`
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

// Validate validates the form request
func (EditRoleRestrictionRequest) Validate(context buffalo.Context) (*validate.Validation, error) {
	request := context.Request()
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = false
	})

	data, err := validate.FromRequest(request)
	activityID, _ := data.Get("id")
	if err != nil {
		return nil, err
	}

	validate.AddValidator("exists_for_only_this_id", func(id interface{}) bool {
		roleid := id.(string)
		db := context.Value("tx").(*pop.Connection)
		activity := models.ActivityAccess{}
		err := db.Where("id=?", roleid).First(&activity)
		if err != nil {
			if err == sql.ErrNoRows {
				return true
			}
			panic(err)
		}
		return uuid.FromStringOrNil(activityID.(string)) == activity.ID
	})

	validate.AddValidator("exists", func(id interface{}) bool {
		activityID := id.(string)
		db := context.Value("tx").(*pop.Connection)
		role := models.ActivityAccess{}
		var exists bool
		exists, err := db.Where("id=?", uuid.FromStringOrNil(activityID)).Exists(&role)
		if err != nil {
			if err == sql.ErrNoRows {
				return false
			}
			panic(err)
		}
		return exists
	})

	validator := data.Create()
	// validator.AddRule("id", "required")
	// validator.AddRule("id", "exists")

	// validator.AddMessages(map[string]string{
	// 	"id.required": "activity id is required",
	// 	"id.exists":   "activity id does not exist",
	// })
	return validator, nil
}

// GetBoundValue gets the values in the request
func (EditRoleRestrictionRequest) GetBoundValue(c buffalo.Context) interface{} {
	request := EditRoleRestrictionRequest{}
	if err := c.Bind(&request); err == nil {
		return request
	}
	return request
}
