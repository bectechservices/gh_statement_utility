package middlewares

import (
	"github.com/gobuffalo/buffalo"
)

// SetCurrentActiveTab sets the current active tab based on the url
func SetCurrentActiveTab(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		currentURL := SanitizeURL(c.Request().URL.Path)
		tab := ""
		switch currentURL {
		case "dashboard":
			tab = "dashboard"
		case "statements":
			tab = "statements"
		case "admin-statements":
			tab = "admin-statements"
		case "stampify-user-profile":
			tab = "stampify-user-profile"
		case "other-pdf-stampify":
			tab = "other-pdf-stampify"
		default:
			tab = "settings"
		}
		c.Set("__tab__", tab)
		return next(c)
	}
}
