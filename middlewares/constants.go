package middlewares

import (
	"gh-statement-app/constants"

	"github.com/gobuffalo/buffalo"
)

// SetRequiredConstants sets the constants in the request
func SetRequiredConstants(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		c.Set("USER_LOGOUT", constants.UserLogout)
		c.Set("SYSTEM_LOGOUT", constants.SystemLogout)
		return next(c)
	}
}
