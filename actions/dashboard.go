package actions

import (
	"gh-statement-app/models"
	"net/http"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/gobuffalo/buffalo"
	"github.com/ricochet2200/go-disk-usage/du"
)

// ShowDashboard shows the dashboard
func ShowDashboard(c buffalo.Context) error {
	userID := AuthID(c)
	dbConnection := DBConnection(c)
	branchID := models.UserBranchID(dbConnection, userID)
	formatter := message.NewPrinter(language.English)
	c.Set("user_id", userID)
	c.Set("all_users", formatter.Sprintf("%d", models.AllUsersCount(dbConnection)))
	c.Set("onboarded_branches", formatter.Sprintf("%d", models.CountOnBoardedBranches(dbConnection)))
	diskUsage := du.NewDiskUsage("/")
	free := diskUsage.Free()
	used := diskUsage.Used()
	total := diskUsage.Size()
	c.Set("free_disk_space", PrettyDiskSize(free))
	c.Set("used_disk_space", PrettyDiskSize(used))
	c.Set("percentage_used", (float32(used)/float32(total))*100)
	c.Set("used_disk_space_raw", used/GB)
	c.Set("total_disk_size", PrettyDiskSize(total))
	c.Set("account_log", models.LoadRecentActivities(5, dbConnection))
	c.Set("audits", models.LoadAccountAudits(5, dbConnection))
	c.Set("total_onboardedbranches", formatter.Sprintf("%d", models.CountOnBoardedBranches(dbConnection)))
	c.Set("total_statement_audited", formatter.Sprintf("%d", models.GetTotalStatementPrintAudit(dbConnection)))
	c.Set("total_allbranchusers", formatter.Sprintf("%d", models.AllUsersCountPerBranch(branchID, dbConnection)))
	c.Set("total_allusers", formatter.Sprintf("%d", models.AllUsersCount(dbConnection)))
	return c.Render(http.StatusOK, r.HTML("dashboard.html"))
}
