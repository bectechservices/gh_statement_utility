package actions

import (
	"gh-statement-app/models"
	"gh-statement-app/requests"
	"net/http"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gofrs/uuid"
)

func ShowBranchesPage(c buffalo.Context) error {
	search := c.Param("search")
	pagination := models.PaginateBranches(search, DBConnection(c), DBPaginator(c, 15))
	c.Set("pagination", pagination)
	c.Set("search", search)
	return c.Render(http.StatusOK, r.HTML("branches.html"))
}

func HandleCreateBranch(c buffalo.Context) error {
	request := requests.CreateBranchRequest{}
	dataIsValid, validator, err := ValidateFormRequest(c, request)
	if err == nil {
		if dataIsValid {
			request = request.GetBoundValue(c).(requests.CreateBranchRequest)
			dbConnection := DBConnection(c)
			models.CreateBranch(request.Name, strings.ToUpper(request.Code), request.BankName, request.StreetName, dbConnection)
			return c.Redirect(http.StatusFound, BranchesURL)
		}
		return RedirectWithModalsOpen(&c, validator, request.GetBoundValue(c), "manageBr")
	}
	return err
}

func HandleEditBranch(c buffalo.Context) error {
	request := requests.EditBranchRequest{}
	dataIsValid, validator, err := ValidateFormRequest(c, request)
	if err == nil {
		if dataIsValid {
			request = request.GetBoundValue(c).(requests.EditBranchRequest)
			dbConnection := DBConnection(c)
			branch := models.GetBranchByID(uuid.FromStringOrNil(request.ID), dbConnection)
			branch.EditBranch(request.Name, request.Code, request.BankName, request.StreetName, dbConnection)
			return c.Redirect(http.StatusFound, BranchesURL)
		}
		return RedirectWithModalsOpen(&c, validator, request.GetBoundValue(c), "editBr")
	}
	return err
}
