package actions

import (
	"fmt"
	"gh-statement-app/models"

	"github.com/gobuffalo/buffalo"
)

// HandleExportUsers export users
func HandleExportUsers(c buffalo.Context) error {
	filename := AppendTimeToName("Users", "xlsx")
	search := c.Param("search")
	users := models.GetAllUsers(DBConnection(c), search)
	headings := []string{"Name", "User ID", "Email", "Branch Name", "Roles", "Created At", "Deleted At", "Last login Date", " Locked", "User Status"}
	data := make([][]string, 0)
	for _, user := range users {
		status := "Active"
		if user.Locked {
			status = "Locked"
		}
		deleted := "N/A"
		if !user.Deleted.Time.IsZero() {
			deleted = user.Deleted.Time.Format("January 02, 2006 3:04PM")
		}
		data = append(data, []string{
			user.Name(),
			user.ABNumber,
			user.Email,
			user.Branch.Name,
			user.RolesToString(),
			user.CreatedAt.Format("January 02, 2006 3:04PM"),
			deleted,
			formatNullDate2(user.LastLogin, "2006-01-02 15:04"),
			status,
			user.Status.String,
		})
	}
	return ExportToExcel(c, "Sheet1", filename, headings, data)
}

func HandleExportBranches(c buffalo.Context) error {
	filename := AppendTimeToName("Branches", "xlsx")
	branches := models.LoadAllBranches(DBConnection(c))
	headings := []string{"Name", "Code", "Bank Name", "Address/Street Name"}
	data := make([][]string, 0)
	for _, branch := range branches {
		data = append(data, []string{
			branch.Name,
			branch.Code,
			branch.BankName,
			branch.StreetName,
		})
	}
	return ExportToExcel(c, "Sheet1", filename, headings, data)
}

func HandleExportRoles(c buffalo.Context) error {
	filename := AppendTimeToName("Roles", "xlsx")
	roles := models.LoadAllRoles(DBConnection(c))
	headings := []string{"Name", "Description"}
	data := make([][]string, 0)
	for _, role := range roles {
		data = append(data, []string{
			role.Name,
			role.Description.String,
		})
	}
	return ExportToExcel(c, "Sheet1", filename, headings, data)
}

func HandleExportPermissions(c buffalo.Context) error {
	filename := AppendTimeToName("Permissions", "xlsx")
	permissions := models.LoadAllPermissions(DBConnection(c))
	headings := []string{"Name", "Description"}
	data := make([][]string, 0)
	for _, permission := range permissions {
		data = append(data, []string{
			permission.Name,
			permission.Description.String,
		})
	}
	return ExportToExcel(c, "Sheet1", filename, headings, data)
}

func HandleExportUserActivityAudit(c buffalo.Context) error {
	filename := AppendTimeToName("User Activity", "xlsx")
	logs := models.LoadUserActivityAudit(DBConnection(c))
	headings := []string{"User", "Email", "AB-Number", "Activity", "Date"}
	data := make([][]string, 0)
	for _, log := range logs {
		data = append(data, []string{
			log.User.Name(),
			log.User.Email,
			log.User.ABNumber,
			log.Activity,
			log.CreatedAt.Format("02.01.2006 3:04 PM"),
		})
	}
	return ExportToExcel(c, "Sheet1", filename, headings, data)
}

func HandleExportUserManagementAudit(c buffalo.Context) error {
	filename := AppendTimeToName("User Management", "xlsx")
	logs := models.LoadAllAccountAudits(DBConnection(c))
	headings := []string{"Activity By", "Activity For", "Description", "Date"}
	data := make([][]string, 0)
	for _, log := range logs {
		data = append(data, []string{
			fmt.Sprintf("%s (%s)", log.By.Name(), log.By.ABNumber),
			fmt.Sprintf("%s (%s)", log.For.Name(), log.For.ABNumber),
			log.Description,
			log.CreatedAt.Format("02.01.2006 3:04 PM"),
		})
	}
	return ExportToExcel(c, "Sheet1", filename, headings, data)
}

func HandleExportStatementPrintAudit(c buffalo.Context) error {
	filename := AppendTimeToName("Statement Print Audit", "xlsx")
	from := c.Param("from")
	to := c.Param("to")
	search := c.Param("search")
	logs := models.ExportStatementPrintAuditData(from, to, search, DBConnection(c))
	headings := []string{"Date Time", "Account Number", "Customer Name", "Date From", "Date To", "Pages", "Print Type", "Staff Branch", "Staff Name"}
	data := make([][]string, 0)
	for _, log := range logs {

		data = append(data, []string{
			log.CreatedAt.Format("02.01.2006 3:04 PM"),
			log.AccountNumber,
			log.AccountName,
			log.QueryDateFrom.Format("02.01.2006"),
			log.QueryDateTo.Format("02.01.2006"),
			fmt.Sprintf("%d", log.Pages),
			log.PrintType,
			log.RequesterBranch,
			log.RequestedBy,
		})
	}
	return ExportToExcel(c, "Sheet1", filename, headings, data)
}

func HandleExportStampPrintAudit(c buffalo.Context) error {
	filename := AppendTimeToName("Stamped Print Audit", "xlsx")
	from := c.Param("from")
	to := c.Param("to")
	search := c.Param("search")
	logs := models.ExportStampPrintAuditData(from, to, search, DBConnection(c))
	headings := []string{"Date Time", "Account Number", "Account Name", "Date Stamped", "Number of Pages Stamped", "FileName", "Requested User"}
	data := make([][]string, 0)
	for _, log := range logs {

		data = append(data, []string{
			log.CreatedAt.Format("02.01.2006 3:04 PM"),
			log.AccountNumber,
			log.AccountName,
			log.DateStamped.Format("02.01.2006"),
			fmt.Sprintf("%d", log.NumPagesStamped),
			log.FileName,
			log.User.FirstName + " " + log.User.LastName,
		})
	}
	return ExportToExcel(c, "Sheet1", filename, headings, data)
}
