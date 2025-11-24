package actions

import (
	"fmt"
	"gh-statement-app/middlewares"
	"gh-statement-app/models"
	dbPaginator "gh-statement-app/pagination"
	"html/template"
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/plush"
	"github.com/gobuffalo/tags"
)

var r *render.Engine
var assetsBox = packr.New("app:assets", "../public")
var UseRoleAndPermissionBasedRouting = envy.Get("ENABLE_ROLES_AND_PERMISSIONS_BASED_ROUTING", "1")

func init() {
	r = render.New(render.Options{
		// HTML layout to be used for all HTML requests:
		// HTMLLayout: "application.plush.html",

		// Box containing all of the templates:
		TemplatesBox: packr.New("app:templates", "../templates"),
		AssetsBox:    assetsBox,

		// Add template helpers here:
		Helpers: render.Helpers{
			"csrf":                        csrfHelper,
			"old_input":                   oldInput,
			"input_has_error":             inputHasError,
			"input_error_message":         inputErrorMessage,
			"should_show_modal":           shouldShowModal,
			"modal_id":                    modalID,
			"set_active_if_url_is":        setActiveIfURLIs,
			"load_badge_type":             loadBadgeType,
			"fancy_bool":                  FancyBool,
			"format_null_date":            formatNullDate,
			"format_2_date":               formatNullDate2,
			"format_search_date":          formatSearchDate,
			"format_zero_date":            formatZeroDate,
			"format_nil_date":             TimeNOTNIL,
			"user_can_access":             userCanAccess,
			"user_can_access_extended":    userCanAccessExtended,
			"custom_paginator":            customPaginator,
			"load_routes_for_permissions": loadRoutesForPermissions,
		},
	})
}

func csrfHelper(ctx plush.HelperContext) (template.HTML, error) {
	tok, ok := ctx.Value("authenticity_token").(string)
	if !ok {
		return "", fmt.Errorf("expected CSRF token got %T", ctx.Value("authenticity_token"))
	}
	t := tags.New("input", tags.Options{
		"value": tok,
		"type":  "hidden",
		"name":  "authenticity_token",
	})
	return t.HTML(), nil
}

func oldInput(field string, ctx plush.HelperContext) string {
	flash := ctx.Value("flash").(map[string][]string)
	if len(flash[fmt.Sprintf("validation_error_%s_message", field)]) > 0 {
		return flash[fmt.Sprintf("validation_error_%s_value", field)][0]
	}
	return ""
}

func inputHasError(field string, ctx plush.HelperContext) bool {
	flash := ctx.Value("flash").(map[string][]string)
	if len(flash[fmt.Sprintf("validation_error_%s", field)]) > 0 {
		return flash[fmt.Sprintf("validation_error_%s", field)][0] == "true"
	}
	return false
}

func inputErrorMessage(field string, ctx plush.HelperContext) string {
	flash := ctx.Value("flash").(map[string][]string)
	if len(flash[fmt.Sprintf("validation_error_%s_message", field)]) > 0 {
		return flash[fmt.Sprintf("validation_error_%s_message", field)][0]
	}
	return ""
}

func shouldShowModal(field string, ctx plush.HelperContext) bool {
	flash := ctx.Value("flash").(map[string][]string)
	return len(flash["show_modal"]) > 0
}

func modalID(field string, ctx plush.HelperContext) string {
	flash := ctx.Value("flash").(map[string][]string)
	if len(flash["show_modal"]) > 0 {
		return flash["show_modal"][0]
	}
	return ""
}

func setActiveIfURLIs(tab string, ctx plush.HelperContext) string {
	current, _ := ctx.Value("__tab__").(string)
	if current == tab {
		return "link-is-active"
	}
	return ""
}

func loadBadgeType(status bool) string {
	if status {
		return "badge-primary--bg--blue"
	}
	return "badge-error--bg"
}

func FancyBool(status bool) string {
	if status {
		return "YES"
	}
	return "NO"
}

// To disable access for all users
func userCanAccess(url string, ctx plush.HelperContext) bool {
	if UseRoleAndPermissionBasedRouting == "0" {
		return true
	}
	userPermissions := ctx.Value("user_permissions").([]string)
	accessPermissions := ctx.Value("access_permissions").(models.PermissionRoutes)
	//fmt.Println("################## accessPermissions:", accessPermissions)
	for _, pageAccess := range accessPermissions {
		if middlewares.SanitizeURL(pageAccess.Path) == url && middlewares.CanAccessMethod("GET", pageAccess.Alias) {
			for _, permission := range userPermissions {
				if permission == pageAccess.PermissionID.String() {
					return true
				}
			}
		}
	}
	return false
}

// To disable access for all users

func userCanAccessExtended(url string, ctx plush.HelperContext) bool {
	if UseRoleAndPermissionBasedRouting == "0" {
		return true
	}
	userPermissions := ctx.Value("user_permissions").([]string)
	accessPermissions := ctx.Value("access_permissions").(models.PermissionRoutes)
	//fmt.Println("################## accessPermissions:", accessPermissions)
	for _, pageAccess := range accessPermissions {
		if middlewares.SanitizeURL(pageAccess.Path) == url && (middlewares.CanAccessMethod("POST", pageAccess.Alias) || middlewares.CanAccessMethod("PUT", pageAccess.Alias) || middlewares.CanAccessMethod("PATCH", pageAccess.Alias) || middlewares.CanAccessMethod("DELETE", pageAccess.Alias)) {
			for _, permission := range userPermissions {
				if permission == pageAccess.PermissionID.String() {
					return true
				}
			}
		}
	}
	return false
}

func customPaginator(pagination *dbPaginator.Pagination, opts map[string]interface{}, help plush.HelperContext) template.HTML {
	if opts["path"] == nil {
		if req, ok := help.Value("request").(*http.Request); ok {
			opts["path"] = req.URL.String()
		}
	}
	html, err := pagination.Tag(opts)
	if err != nil {
		return ""
	}
	return template.HTML(html)
}

func loadRoutesForPermissions(ctx plush.HelperContext) AccessPolicyRoutes {
	routes := ctx.Value("routes").(buffalo.RouteList)
	routesToDisplay := make(AccessPolicyRoutes, 0)
	//fmt.Println("################## routesToDisplay:", routesToDisplay)
	for _, route := range routes {
		if len(route.Aliases) > 0 {
			routesToDisplay = append(routesToDisplay, AccessPolicyRoute{
				Path:  route.Path,
				Alias: route.Aliases[0],
			})
		}
	}
	return routesToDisplay
}

func formatNullDate(date nulls.Time, format string) string {
	if date.Valid {
		return date.Time.Format(format)

	}
	return "N/A"
}
func formatNullDate2(date nulls.Time, format string) string {
	if date.Valid {
		return date.Time.Format("2006-01-02 15:04")

	}
	return ""
}

func formatZeroDate(date time.Time, format string) string {
	parsedDate, _ := time.Parse(format, date.String())
	//fmt.Println("&&&&&&&&&&&&&&&&&&&", parsedDate)
	if parsedDate.IsZero() {
		return "N/A"
	} else {

		return parsedDate.Format(format)
	}

}
func formatSearchDate(date time.Time, format string) string {
	parsedDate, _ := time.Parse(format, date.String())
	if parsedDate.IsZero() {
		return "N/A"
	} else {

		return parsedDate.Format(format)
	}

}

// TimeOrNA returns time
func TimeNOTNIL(dateTime nulls.Time) string {

	if !dateTime.Valid {
		return "N/A"
	}

	return dateTime.Time.Format("January 02,2006")
	//return dateTime.Format("January 02,2006")
	//nulls.NewTime(time.Time(dateTime.Time.Format("January 02,2006")))
}
