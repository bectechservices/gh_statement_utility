package actions

import (
	"encoding/base64"
	"fmt"
	"gh-statement-app/models"
	"gh-statement-app/requests"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gobuffalo/envy"

	"github.com/gobuffalo/buffalo"
)

func ShowStampifyUserSetupPage(c buffalo.Context) error {

	uid := AuthID(c)
	models.UserBranchIDUpdate(uid, DBConnection(c))
	userStamp := models.GetStampifyUserID(uid, DBConnection(c))
	contents, _ := os.ReadFile(userStamp.SignatureLocation)
	data := base64.StdEncoding.EncodeToString(contents)
	c.Set("stampify", userStamp)
	c.Set("imageBytes", data)

	return c.Render(http.StatusOK, r.HTML("stampify-user-profile.html"))
}

func HandleStampifyUserSetup(c buffalo.Context) error {

	request := requests.StampifyUserSetupRequest{}
	dataIsValid, validator, err := ValidateFormRequest(c, request)
	if err == nil {
		if dataIsValid {
			request = request.GetBoundValue(c).(requests.StampifyUserSetupRequest)
			uid := AuthID(c)
			dbConnection := DBConnection(c)
			stampifyuser := models.GetStampifyUserID(uid, dbConnection)
			stampifyuser.UpdateStampifyUserDetails(request.Position, request.BranchStamp, request.SignatureLocation, dbConnection)
			return c.Redirect(http.StatusFound, UserPDFStampifySetupPAgeURL)
		}
		return RedirectWithErrors(&c, validator, request.GetBoundValue(c))
	}
	return err
}

//-----------------------------------------------------------------------------------------------------------------

// HandleOtherPDFDocumentsRequest
func HandleOtherPDFDocumentsRequest(c buffalo.Context) error {

	return c.Render(http.StatusOK, r.HTML("other-pdf-stampify.html"))
}

// StampOtherPDFDocumentsRequest
func StampOtherPDFDocumentsRequest(c buffalo.Context) error {
	//return c.Render(http.StatusOK, r.JSON("Auto Stamped through other PDF Documents Successfully"))
	uid := AuthID(c)
	dbConnection := DBConnection(c)
	user := models.GetUserByID(uid, dbConnection)
	file, err := c.File("file")
	if err != nil {
		return c.Render(http.StatusBadRequest, r.JSON("no file uploaded"))
	}

	useStampApi, _ := strconv.ParseBool(envy.Get("GH_STAMPIFYPDF_API", "false"))

	if !useStampApi {
		return c.Render(http.StatusUnauthorized, r.JSON("Cannot generate stamp on pdf. Please ask the administrator to check whether this service is authorize to communicate with the stampify api service"))
	}

	responseChan, errChan := make(chan []byte, 1), make(chan error, 1)

	go MakeRestCallToStampifyAPI(&user, file.Filename, &file, responseChan, errChan)
	stampifyResponse := <-responseChan
	err = <-errChan

	if err != nil {
		return err
	}

	tempDir := envy.Get("GH_ACCOUNT_STATEMENT_TEMP", "./")

	//f, err := os.WriteFile(filepath.Join(tempDir, file.Filename), response, 0644)
	f, err := os.Create(filepath.Join(tempDir, file.Filename))
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.Write(stampifyResponse)
	if err != nil {
		return err
	}

	f.Sync()

	response := make(map[string]string)
	response["message"] = "Stamp added successfully!"
	response["url"] = fmt.Sprintf("/downloads?file_name=%s", file.Filename)

	return c.Render(http.StatusOK, r.JSON(response))
}
