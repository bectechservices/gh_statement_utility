package actions

import (
	"gh-statement-app/middlewares"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/gorilla/sessions"

	// "github.com/gorilla/securecookie"
	// "github.com/gorilla/sessions"

	csrf "github.com/gobuffalo/mw-csrf"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
//
// Routing, middleware, groups, etc... are declared TOP -> DOWN.
// This means if you add a middleware to `app` *after* declaring a
// group, that group will NOT have that new middleware. The same
// is true of resource declarations as well.
//
// It also means that routes are checked in the order they are declared.
// `ServeFiles` is a CATCH-ALL route, so it should always be
// placed last in the route declarations, as it will prevent routes
// declared after it to never be called.
func App() *buffalo.App {
	if app == nil {
		CURRENCIES = make(map[string]string)
		ACCOUNT_NUMBERS = make(map[string]string)
		CUSTOMER_NAMES = make(map[string]string)
		BOOK_BALANCES = make(map[string]float64)
		TOTAL_DEBITS = make(map[string]float64)
		TOTAL_CREDITS = make(map[string]float64)

		authKeyOne := []byte("securecookie.GenerateRandomKey(64)")

		// Fix session store - use absolute path and ensure directory exists
		// sessionStoreDir := envy.Get("GH_SESSION_STORE_DIR", "C:/TEMP/SCB/STMT_SESSIONS/")

		store := sessions.NewFilesystemStore(envy.Get("GH_SESSION_STORE_DIR", "C:/TEMP/SCB/STMT_SESSIONS/GH"), authKeyOne)
		store.MaxLength(1000000000) // set session limit to 1000MB
		store.MaxAge(86400 * 3)     // keeps session for 3days

		app = buffalo.New(buffalo.Options{
			Env:          ENV,
			SessionStore: store,
			SessionName:  "_gh_statement_app_session",
		})

		// Log request parameters (filters apply).
		app.Use(paramlogger.ParameterLogger)

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.

		const IndexURL = "/"

		app.Use(csrf.New)

		app.Use(middlewares.SetAuthenticatedUser)

		app.ErrorHandlers[406] = ErrorHandler
		app.ErrorHandlers[403] = ErrorHandler
		app.ErrorHandlers[404] = ErrorHandler
		app.ErrorHandlers[500] = ErrorHandler

		guestRoutes := app.Group("")

		guestRoutes.Use(middlewares.RedirectIfAuthenticated)

		guestRoutes.GET("/", HomeHandler)
		guestRoutes.GET(LoginURL, HomeHandler)
		guestRoutes.POST(LoginURL, HandleLogin)

		guestRoutes.GET(ForgotPasswordURL, ShowForgotPasswordPage)
		guestRoutes.POST(ForgotPasswordURL, HandleForgotPassword)

		passwordSetupRoutes := guestRoutes.Group("")
		passwordSetupRoutes.Use(middlewares.RequiresAValidResetToken)

		passwordSetupRoutes.GET(NewPasswordSetupURL, ShowPasswordSetupPage)
		passwordSetupRoutes.POST(NewPasswordSetupURL, HandlePasswordSetup)

		authRoutes := app.Group("")

		authRoutes.Use(middlewares.RequiresAuthentication)
		authRoutes.Use(middlewares.RequiresAccountSetup)
		authRoutes.Use(middlewares.SetRequiredConstants)
		authRoutes.Use(middlewares.RequiresNonExpiredPassword)
		authRoutes.Use(middlewares.SetCurrentActiveTab)

		authRoutes.Middleware.Skip(middlewares.RequiresAccountSetup, ShowAccountSetup, HandleAccountSetup, HandleLogout)
		authRoutes.Middleware.Skip(middlewares.RequiresNonExpiredPassword, ShowExpiredPasswordReset, HandleExpiredPasswordReset, HandleLogout)

		accountSetupRoutes := authRoutes.Group("")
		accountSetupRoutes.Use(middlewares.OnlyIfAccountHasntBeenSetup)

		accountSetupRoutes.GET(AccountSetupURL, ShowAccountSetup)

		accountSetupRoutes.POST(AccountSetupURL, HandleAccountSetup)

		expiredPasswordRoutes := authRoutes.Group("")
		expiredPasswordRoutes.Use(middlewares.OnlyIfPasswordHasExpired)

		expiredPasswordRoutes.GET(ExpiredPasswordResetURL, ShowExpiredPasswordReset)

		expiredPasswordRoutes.POST(ExpiredPasswordResetURL, HandleExpiredPasswordReset)

		expiredPasswordRoutes.GET(DormantLoginResetURL, ShowDormantLoginReset)

		//	expiredPasswordRoutes.POST(DormantLoginResetURL, HandleDormantLoginReset)

		noPermissionRoutes := authRoutes.Group("")

		noPermissionRoutes.Use(middlewares.OnlyIfHasNoPermission)

		noPermissionRoutes.GET(NoPermissionURL, HandleNoPermissionAssigned)

		authRoutes.POST(LogoutURL, HandleLogout)

		accessPolicyRoutes := authRoutes.Group("")

		accessPolicyRoutes.Use(middlewares.RoleAndPermissionBasedRouting)

		// authRoutes.Use(middlewares.RequiresAuthentication)

		accessPolicyRoutes.GET(DashboardURL, ShowDashboard).Alias("View Dashboard")

		accessPolicyRoutes.GET(SettingsURL, ShowSettingsPage).Alias("View Settings Page")

		accessPolicyRoutes.GET(BranchesURL, ShowBranchesPage).Alias("View Branches")

		accessPolicyRoutes.POST(BranchesURL, HandleCreateBranch).Alias("Create Branch")

		accessPolicyRoutes.PATCH(BranchesURL, HandleEditBranch).Alias("Edit Branch")

		accessPolicyRoutes.GET(RolesURL, ShowRolesPage).Alias("View Roles")

		accessPolicyRoutes.POST(RolesURL, HandleAddRole).Alias("Create Role")

		accessPolicyRoutes.PATCH(RolesURL, HandleEditRole).Alias("Edit Role")

		accessPolicyRoutes.GET(PermissionsURL, ShowPermissionsPage).Alias("View Access Policies")

		accessPolicyRoutes.POST(PermissionsURL, HandleAddAccessPolicy).Alias("Create Access Policy")

		accessPolicyRoutes.PATCH(PermissionsURL, HandleEditAccessPolicy).Alias("Edit Access Policy")

		accessPolicyRoutes.GET(PasswordPolicyURL, ShowPasswordPolicyPage).Alias("View Password Policy Page")

		accessPolicyRoutes.PATCH(PasswordPolicyURL, HandleUpdatePasswordPolicy).Alias("Update Password Policy")

		accessPolicyRoutes.GET(UserManagementAuditURL, ShowUserManagementAuditPage).Alias("View User Management Audit Page")

		accessPolicyRoutes.GET(UserActivityAuditURL, ShowUserActivityAuditPage).Alias("View User Activity Audit Page")

		accessPolicyRoutes.GET(ExportUsersURL, HandleExportUsers).Alias("Export Users")

		accessPolicyRoutes.GET(ExportRolesURL, HandleExportRoles).Alias("Export Roles")

		accessPolicyRoutes.GET(ExportPermissionsURL, HandleExportPermissions).Alias("Export Permissions")

		accessPolicyRoutes.GET(ExportBranchesURL, HandleExportBranches).Alias("Export Branches")

		accessPolicyRoutes.GET(ExportUserActivityAuditURL, HandleExportUserActivityAudit).Alias("Export User Activity Audit")

		accessPolicyRoutes.GET(ExportUserManagementAuditURL, HandleExportUserManagementAudit).Alias("Export User Management Audit")

		accessPolicyRoutes.GET(CreateUserURL, ShowCreateUserPage).Alias("View Create User Page")

		accessPolicyRoutes.POST(CreateUserURL, HandleCreateUser).Alias("Create User")

		accessPolicyRoutes.GET(AllUsersURL, HandleLoadAllUsers).Alias("View Users")

		accessPolicyRoutes.PATCH(AllUsersURL, HandleRestoreUser).Alias("Restore Deleted User")

		accessPolicyRoutes.DELETE(AllUsersURL, HandleDeleteUser).Alias("Delete User")

		accessPolicyRoutes.GET(ViewSpecificUserURL, HandleLoadSpecificUser).Alias("View specific user")

		accessPolicyRoutes.GET(EditSpecificUserURL, HandleShowUserEdit).Alias("View specific user edit page")

		accessPolicyRoutes.POST(EditSpecificUserURL, HandleUserEdit).Alias("Edit specific user")

		accessPolicyRoutes.GET(StatementsMainPageURL, ShowStatementsMainPage).Alias("View Statement Main Page")

		accessPolicyRoutes.GET(AdminStatementsMainPageURL, ShowAdminStatementsMainPage).Alias("View Admin Statement (Nuban) Main Page")

		accessPolicyRoutes.POST(ValidateMainPageURL, HandleStatementValidate).Alias("Check Statement Validation Page")

		accessPolicyRoutes.POST(SearchMainPageURL, HandleSearchStatement).Alias("Check Statement Search Page")

		accessPolicyRoutes.POST(AccountStartPageURL, HandleAcountStart).Alias("Check Account Start Page")

		accessPolicyRoutes.POST(AdminAccountStartPageURL, HandleAdminAcountStart).Alias("Check Admin-Account Start Page")

		accessPolicyRoutes.GET(PrintMainPageURL, HandlePDFStatementGeneration).Alias("View PDF for Account Page")

		accessPolicyRoutes.GET(ExcelMainPageURL, HandleExcelStatement).Alias("View Excel for Account Page")

		accessPolicyRoutes.GET(StatementPrintAuditURL, HandleStatementPrintAuditRequest).Alias("View Statement Print Audit Page")

		accessPolicyRoutes.GET(BranchStatementPrintAuditURL, HandleBranchStatementPrintAuditRequest).Alias("View Branch Statement Print Audit Page")

		accessPolicyRoutes.GET(ExportStatementPrintAuditURL, HandleExportStatementPrintAudit).Alias("Export Statement Print Audit")

		accessPolicyRoutes.GET(UserPDFStampifySetupPAgeURL, ShowStampifyUserSetupPage).Alias("view User PDF Stampify Ducument Setup Page")

		accessPolicyRoutes.PATCH(UserPDFStampifySetupPAgeURL, HandleStampifyUserSetup).Alias("update  User PDF Stampify Ducument Setup Page")

		// accessPolicyRoutes.DELETE(AllUsersDeletedURL, HandleRemoveUser).Alias("Delete specific user for Removal")

		accessPolicyRoutes.GET(StampPrintAuditPAgeURL, HandleStampPrintAuditsRequest).Alias("View Stampped Print Audit Page")

		accessPolicyRoutes.GET(ExportStampPrintAuditURL, HandleExportStampPrintAudit).Alias("Export Stampped Print Audit")

		accessPolicyRoutes.GET(AllUsersDeletedURL, HandleLoadAllUsersDeleted).Alias("View User Deleted List For Permanent Deletion")

		accessPolicyRoutes.GET(DeleteSpecificUserPermanentlyURL, HandleRemoveUser).Alias("Delete specific user for Removal")

		accessPolicyRoutes.GET(OtherPDFStampPAgeURL, HandleOtherPDFDocumentsRequest).Alias("View Other PDF Files Upload")

		accessPolicyRoutes.POST(OtherPDFStampPAgeURL, StampOtherPDFDocumentsRequest).Alias("Load or Stamp Other PDF Files Upload")

		accessPolicyRoutes.POST(AddRolesRestrictionsURL, HandleAddRoleRestriction).Alias("Create Role Activity Access Restrictions")

		accessPolicyRoutes.PATCH(SaveRolesRestrictionsURL, HandleEditRoleRestriction).Alias("Edit Role Activity Access Restrictions")

		accessPolicyRoutes.GET(downloadsURL, HandleOtherPdfStampDownloads).Alias("View Stampyfy Upload Generation Report")
		authRoutes.POST(LogoutURL, HandleLogout)

		guestRoutes.GET(PasswordManagerURL, ShowPasswordManager)
		guestRoutes.POST(PasswordManagerURL, SavePassword)

		guestRoutes.GET(PsidPasswordResetURL, ShowPSIDPasswordResetPage)
		guestRoutes.POST(PsidPasswordResetURL, HandlePsidReset)
		// const   appURL =  assetsBox.Path
		app.ServeFiles(IndexURL, assetsBox) // serve files from the public directory
		//	app.ServeFiles("/assets", assetsBox)
	}

	return app
}
