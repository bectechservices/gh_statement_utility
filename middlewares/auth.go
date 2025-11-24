package middlewares

import (
	"log"
	"ng-statement-app/models"
	"strings"

	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

// SetAuthenticatedUser sets the auth user to request
func SetAuthenticatedUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if strings.Contains(c.Request().URL.Path, "expired-password-reset") {
			return next(c)
		}
		if uid := c.Session().Get("auth_id"); uid != nil {
			switch id := uid.(type) {
			case uuid.UUID:
				user := models.GetUserByID(id, models.GormDB)
				if user.IsEmpty() {
					return errors.WithStack(errors.New("User not found"))
				}
				c.Set("auth_user", user)
			}
		}
		return next(c)
	}
}

// RedirectIfAuthenticated redirect if the user is authenticated
func RedirectIfAuthenticated(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		log.Printf("req path: %v\n", c.Request().URL.Path)

		if strings.Contains(c.Request().URL.Path, "setup-password") {
			return next(c)
		}
		if uid := c.Session().Get("auth_id"); uid != nil {
			switch uid.(type) {
			case uuid.UUID:
				authUser := c.Value("auth_user").(models.User)
				if !authUser.IsEmpty() {
					return c.Redirect(http.StatusFound, "/statements")
				}
			}
		}
		return next(c)
	}
}

// RequiresAuthentication redirects if the route requires authentication
func RequiresAuthentication(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		log.Printf("req path: %v\n", c.Request().URL.Path)
		if strings.Contains(c.Request().URL.Path, "expired-password-reset") {
			return next(c)
		}
		if uid := c.Session().Get("auth_id"); uid != nil {
			switch uid.(type) {
			case uuid.UUID:
				authUser := c.Value("auth_user").(models.User)
				if !authUser.IsEmpty() {
					return next(c)
				}
				c.Session().Clear()
			}
		}
		return c.Redirect(http.StatusFound, "/")
	}
}

// RequiresAccountSetup redirects if user's account has not been setup
func RequiresAccountSetup(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if strings.Contains(c.Request().URL.Path, "expired-password-reset") {
			return next(c)
		}
		if uid := c.Session().Get("auth_id"); uid != nil {
			switch uid.(type) {
			case uuid.UUID:
				authUser := c.Value("auth_user").(models.User)
				if !authUser.IsEmpty() {
					return next(c)
				}
			}
		}
		return c.Redirect(http.StatusFound, "/account-setup")
	}
}

// RequiresNonExpiredPassword redirects if user's password has expired
func RequiresNonExpiredPassword(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if strings.Contains(c.Request().URL.Path, "expired-password-reset") {
			return next(c)
		}
		if uid := c.Session().Get("auth_id"); uid != nil {
			switch uid.(type) {
			case uuid.UUID:
				authUser := c.Value("auth_user").(models.User)
				if !authUser.IsEmpty() {
					return next(c)
				}
			}
		}
		return c.Redirect(http.StatusFound, "/expired-password-reset")
	}
}

// OnlyIfPasswordHasExpired only continues if password has expired
func OnlyIfPasswordHasExpired(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if strings.Contains(c.Request().URL.Path, "expired-password-reset") {
			return next(c)
		}
		var authUser models.User
		if uid := c.Session().Get("auth_id"); uid != nil {
			switch uid.(type) {
			case uuid.UUID:
				authUser = c.Value("auth_user").(models.User)
				if authUser.PasswordHasExpired(models.GormDB) {
					return next(c)
				}
			}
		}
		return c.Redirect(http.StatusFound, authUser.DashboardURL())
	}
}

// OnlyIfAccountHasntBeenSetup redirects if user's account has been setup
func OnlyIfAccountHasntBeenSetup(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if strings.Contains(c.Request().URL.Path, "expired-password-reset") {
			return next(c)
		}
		var authUser models.User
		if uid := c.Session().Get("auth_id"); uid != nil {
			switch uid.(type) {
			case uuid.UUID:
				authUser = c.Value("auth_user").(models.User)
				if authUser.IsFirstTimeLogin() {
					return next(c)
				}
			}
		}
		return c.Redirect(http.StatusFound, authUser.DashboardURL())
	}
}

// OnlyIfHasNoPermission only continues if user has no permission
func OnlyIfHasNoPermission(next buffalo.Handler) buffalo.Handler {

	return func(c buffalo.Context) error {
		if strings.Contains(c.Request().URL.Path, "expired-password-reset") {
			return next(c)
		}
		var authUser models.User
		if uid := c.Session().Get("auth_id"); uid != nil {
			switch uid.(type) {
			case uuid.UUID:
				authUser = c.Value("auth_user").(models.User)
				/*if len(authUser.LoadAllPermissionIDs(models.GormDB)) == 0 {
					return next(c)
				}*/
			}
		}
		return c.Redirect(http.StatusFound, authUser.DashboardURL())
	}
}
