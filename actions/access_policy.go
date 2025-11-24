package actions

import (
	"fmt"
	"net/http"

	"gh-statement-app/models"
	"gh-statement-app/requests"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
)

// ShowRolesPage shows all roles
func ShowRolesPage(c buffalo.Context) error {
	pagination := models.PaginateRoles(models.GormDB, DBPaginator(c, 10))
	c.Set("permissions", models.LoadAllPermissions(DBConnection(c)))
	c.Set("roles", models.LoadAllRoles(DBConnection(c)))
	c.Set("activity", models.LoadAllRoleRestriction(DBConnection(c)))
	c.Set("pagination", pagination)
	return c.Render(http.StatusOK, r.HTML("roles.html"))
}

// ShowRolesPage shows all roles
// func ShowRolesPage(c buffalo.Context) error {
// 	roles, pagination := models.PaginateRoles(DBPaginator(c, 10))
// 	c.Set("permissions", models.LoadAllPermissions(DBConnection(c)))
// 	c.Set("activity", models.LoadAllRoleRestriction(DBConnection(c)))
// c.Set("roles", roles)
// c.Set("pagination", pagination)
// 	return c.Render(http.StatusOK, r.HTML("roles.html"))
// }

// HandleAddRole creates a new role
func HandleAddRole(c buffalo.Context) error {
	request := requests.CreateRoleRequest{}
	dataIsValid, validator, err := ValidateFormRequest(c, request)
	if err == nil {
		if dataIsValid {
			data := request.GetBoundValue(c).(requests.CreateRoleRequest)
			dbConnection := DBConnection(c)
			role := models.Role{
				ID:          models.NewUUID(),
				Name:        data.Name,
				Description: nulls.NewString(data.Description),
			}.Create(dbConnection)
			role.AddPermissions(data.Permissions, dbConnection)
			return RedirectBack(&c)
		}
		return RedirectWithModalsOpen(&c, validator, request.GetBoundValue(c), "manageRBAC")
	}
	return err
}

// HandleEditRole edits the role
func HandleEditRole(c buffalo.Context) error {
	request := requests.EditRoleRequest{}
	dataIsValid, validator, err := ValidateFormRequest(c, request)
	if err == nil {
		if dataIsValid {
			data := request.GetBoundValue(c).(requests.EditRoleRequest)
			dbConnection := DBConnection(c)
			role := models.GetRoleByID(uuid.FromStringOrNil(data.ID), dbConnection)
			role.Name = data.Name
			role.Description = nulls.NewString(data.Description)
			dbConnection.Save(&role)
			role.DeleteAllPermissions(dbConnection)
			role.AddPermissions(data.Permissions, dbConnection)
			return RedirectBack(&c)
		}
		return RedirectWithModalsOpen(&c, validator, request.GetBoundValue(c), "manageRBAC")
	}
	return err
}

// ShowPermissionsPage shows the permissions page
func ShowPermissionsPage(c buffalo.Context) error {
	pagination := models.PaginatePermissions(models.GormDB, DBPaginator(c, 10))
	c.Set("pagination", pagination)
	return c.Render(http.StatusOK, r.HTML("permissions.html"))
}

// HandleAddAccessPolicy handles adding an access policy
func HandleAddAccessPolicy(c buffalo.Context) error {
	request := requests.CreatePermissionRequest{}
	dataIsValid, validator, err := ValidateFormRequest(c, request)
	if err == nil {
		if dataIsValid {
			data := request.GetBoundValue(c).(requests.CreatePermissionRequest)
			dbConnection := DBConnection(c)
			permission := models.Permission{
				ID:          models.NewUUID(),
				Name:        data.Name,
				Description: nulls.NewString(data.Description),
			}.Create(dbConnection)
			permission.CreateAccessPolicies(data.Routes, dbConnection)
			return RedirectBack(&c)
		}
		return RedirectWithModalsOpen(&c, validator, request.GetBoundValue(c), "manageAP")
	}
	return err
}

// HandleEditAccessPolicy handles editing an access policy
func HandleEditAccessPolicy(c buffalo.Context) error {
	request := requests.EditPermissionRequest{}
	dataIsValid, validator, err := ValidateFormRequest(c, request)
	if err == nil {
		if dataIsValid {
			data := request.GetBoundValue(c).(requests.EditPermissionRequest)
			dbConnection := DBConnection(c)
			permission := models.GetPermissionByID(uuid.FromStringOrNil(data.ID), dbConnection)
			permission.Name = data.Name
			permission.Description = nulls.NewString(data.Description)
			dbConnection.Save(&permission)
			permission.DeleteAllAccessPolicies(dbConnection)
			permission.CreateAccessPolicies(data.Routes, dbConnection)
			return RedirectBack(&c)
		}
		return RedirectWithModalsOpen(&c, validator, request.GetBoundValue(c), "editAP")
	}
	return err
}

// HandleAddRoleRestrictionAccess creates a new role
func HandleAddRoleRestriction(c buffalo.Context) error {
	fmt.Println("########################### ACTION Request  ###########")
	request := requests.CreateRoleRestrictionRequest{}
	fmt.Println("############ HandleAddRoleRestriction request #############", request)
	dataIsValid, validator, err := ValidateFormRequest(c, request)
	fmt.Println("############ HandleAddRoleRestriction dataIsValid #############", dataIsValid)
	if err == nil {
		if dataIsValid {
			data := request.GetBoundValue(c).(requests.CreateRoleRestrictionRequest)
			fmt.Println("############ HandleAddRoleRestriction data #############", data)
			dbConnection := DBConnection(c)
			models.CreateRoleRestriction(data.Monday, data.Tuesday, data.Wednesday, data.Thursday, data.Friday, data.Saturday, data.Sunday, data.StartTime, data.EndTime, dbConnection)
			lastRecord, _ := models.GetLastUpdatedAccessActivity(dbConnection)
			role := models.GetRoleByID(uuid.FromStringOrNil(data.RoleID), dbConnection)
			fmt.Println("############ Last Record #############", lastRecord)
			fmt.Println("############ Role #############", role.ID)
			role.ActivityAccessID = nulls.NewUUID(lastRecord.ID)
			fmt.Println("############ role.ActivityAccessID #############", role.ActivityAccessID)
			dbConnection.Save(&role)
			return c.Redirect(http.StatusFound, RolesURL)
		}
		return RedirectWithModalsOpen(&c, validator, request.GetBoundValue(c), "addRR")
	}
	return err
}

func HandleEditRoleRestriction(c buffalo.Context) error {
	request := requests.EditRoleRestrictionRequest{}
	dataIsValid, validator, err := ValidateFormRequest(c, request)
	if err == nil {
		if dataIsValid {

			request = request.GetBoundValue(c).(requests.EditRoleRestrictionRequest)
			dbConnection := DBConnection(c)
			activity := models.GetRoleRestrictionByID(uuid.FromStringOrNil(request.ID), dbConnection)
			activity.UpdateRoleAccessRestriction(request.Monday, request.Tuesday, request.Wednesday, request.Thursday, request.Friday, request.Saturday, request.Sunday, request.StartTime, request.EndTime, dbConnection)
			return c.Redirect(http.StatusFound, RolesURL)
		}
		return RedirectWithModalsOpen(&c, validator, request.GetBoundValue(c), "saveRR")
	}
	return err
}
