package actions

import (
	"net/http"
	"ng-statement-app/models"

	"github.com/gobuffalo/buffalo"
)

// HandleStatementPrintAuditRequest
func HandleStatementPrintAuditRequest(c buffalo.Context) error {
	search := c.Param("search")
	from := c.Param("from")
	to := c.Param("to")
	pagination := models.PaginateStatementPrintAudit(from, to, search, DBConnection(c), DBPaginator(c, 15))
	c.Set("pagination", pagination)
	c.Set("search", search)
	c.Set("from", from)
	c.Set("to", to)
	return c.Render(http.StatusOK, r.HTML("statement-audit.html"))
}

// HandleBranchStatementPrintAuditRequest
func HandleBranchStatementPrintAuditRequest(c buffalo.Context) error {
	search := c.Param("search")
	from := c.Param("from")
	to := c.Param("to")
	pagination := models.PaginateBranchStatementPrintAudit(from, to, search, AuthUser(c).BranchID, DBConnection(c), DBPaginator(c, 15))
	c.Set("pagination", pagination)
	c.Set("search", search)
	c.Set("from", from)
	c.Set("to", to)
	//return c.Render(200, r.JSON(pagination))
	return c.Render(http.StatusOK, r.HTML("branch-statement-audit.html"))
}

// HandleStampPrintAuditsRequest
func HandleStampPrintAuditsRequest(c buffalo.Context) error {
	search := c.Param("search")
	from := c.Param("from")
	to := c.Param("to")
	pagination := models.PaginateStampPrintAudit(from, to, search, DBConnection(c), DBPaginator(c, 15))
	c.Set("pagination", pagination)
	c.Set("search", search)
	c.Set("from", from)
	c.Set("to", to)
	return c.Render(http.StatusOK, r.HTML("stamp-print-audit.html"))
}
