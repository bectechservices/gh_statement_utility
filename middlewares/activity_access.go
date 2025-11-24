package middlewares

import (
	"fmt"
	"ng-statement-app/constants"
	"ng-statement-app/models"

	"github.com/gobuffalo/buffalo"
)

func MustBeActiveWorkingTimeFrame(next buffalo.Handler) buffalo.Handler {

	return func(c buffalo.Context) error {
		//is in range with active hours
		authUser := c.Value("auth_user").(models.User)
		fmt.Printf("######## AuthUser #########%s", authUser)
		if models.GrantRolesAccessActivity(authUser.ABNumber, models.GormDB) {
			fmt.Println("################### Is within active hours")
			c.Set("is_active", true)
			return next(c)
		}
		authUser.LogAccountActivity(constants.UnauthorizedTimeAndDay, models.GormDB)
		c.Set("is_active", false)
		c.Session().Clear()
		return c.Redirect(302, "/")
		//return c.Redirect(301, "/")
	}
}
