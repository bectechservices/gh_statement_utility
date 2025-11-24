package middlewares

import (
	"net/http"
	"ng-statement-app/models"

	"github.com/gobuffalo/buffalo"
)

// RequiresAValidResetToken redirects if an invalid token is passed
// func RequiresAValidResetToken(next buffalo.Handler) buffalo.Handler {
// 	return func(c buffalo.Context) error {
// 		values, isset := c.Request().URL.Query()["token"]
// 		if isset {
// 			_, ok := c.Value("tx").(*pop.Connection)
// 			if !ok {
// 				panic("no db transaction found")
// 			}
// 			if models.PasswordResetTokenIsValid(values[0], models.GormDB) {
// 				return next(c)
// 			}
// 		}
// 		return c.Redirect(http.StatusFound, "/")
// 	}
// }

// RequiresAValidResetToken redirects if an invalid token is passed
func RequiresAValidResetToken(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		values, isset := c.Request().URL.Query()["token"]
		if isset {
			if models.PasswordResetTokenIsValid(values[0], models.GormDB) {
				return next(c)
			}
		}
		return c.Redirect(http.StatusFound, "/")
	}
}
