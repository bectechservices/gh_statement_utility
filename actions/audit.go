package actions

import (
	"gh-statement-app/models"
	"net/http"

	"github.com/gobuffalo/buffalo"
)

// ShowUserManagementAuditPage loads the audit page
func ShowUserManagementAuditPage(c buffalo.Context) error {
	pagination := models.PaginateUserAccountAudits(DBConnection(c), DBPaginator(c, 10))
	c.Set("pagination", pagination)
	return c.Render(http.StatusOK, r.HTML("user-management-audit.html"))
}

// ShowUserActivityAuditPage loads the audit page
func ShowUserActivityAuditPage(c buffalo.Context) error {
	pagination := models.PaginateUserActivityAudits(DBConnection(c), DBPaginator(c, 10))
	c.Set("pagination", pagination)
	return c.Render(http.StatusOK, r.HTML("user-activity-audit.html"))
}
