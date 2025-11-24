package middlewares

import (
	"errors"
	"fmt"
	"gh-statement-app/models"
	"net/http"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
)

var UseRoleAndPermissionBasedRouting = envy.Get("ENABLE_ROLES_AND_PERMISSIONS_BASED_ROUTING", "1")

// SanitizeURL cleans the URL
func SanitizeURL(url string) string {
	return strings.Trim(url, "/")
}

// CanAccessMethod redirect to appropriate page
func CanAccessMethod(method, alias string) bool {
	alias = strings.ToLower(alias)
	switch method {
	case "GET":
		return strings.Contains(alias, "view") || strings.Contains(alias, "export") || strings.Contains(alias, "print")
	case "POST":
		return strings.Contains(alias, "create") || strings.Contains(alias, "edit") || strings.Contains(alias, "search") || strings.Contains(alias, "convert") || strings.Contains(alias, "load") || strings.Contains(alias, "run") || strings.Contains(alias, "reset") || strings.Contains(alias, "send") || strings.Contains(alias, "check")

	case "DELETE":
		return strings.Contains(alias, "delete")

	case "PATCH", "PUT":
		return strings.Contains(alias, "edit") || strings.Contains(alias, "restore") || strings.Contains(alias, "update")
	}
	return false
}

// RoleAndPermissionBasedRouting displays pages based on roles and permissions
func RoleAndPermissionBasedRouting(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		//use based on env value
		if UseRoleAndPermissionBasedRouting == "0" {
			return next(c)
		}

		currentURL := SanitizeURL(c.Request().URL.Path)
		currentURLMethod := c.Request().Method
		user, found := c.Value("auth_user").(models.User)
		if !found {
			return errors.New("user not found in context")
		}

		dbConnection := models.GormDB
		// dbConnection := c.Value("tx").(*pop.Connection)
		userPermissions := user.LoadAllPermissionIDs(dbConnection)
		accessPermissions := models.LoadAllAccessPermissions(dbConnection)

		c.Set("user_permissions", userPermissions)
		c.Set("access_permissions", accessPermissions)

		for _, pageAccess := range accessPermissions {
			if SanitizeURL(pageAccess.Path) == currentURL && CanAccessMethod(currentURLMethod, pageAccess.Alias) {
				fmt.Printf("%+v\n", userPermissions)
				for _, permission := range userPermissions {
					if permission == pageAccess.PermissionID.String() {
						return next(c)
					}
				}
			}
		}
		return c.Error(http.StatusForbidden, errors.New("you do not have the permissions required to view this page"))
	}
}
