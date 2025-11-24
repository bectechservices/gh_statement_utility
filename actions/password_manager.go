package actions

import (
	"gh-statement-app/models"
	"gh-statement-app/requests"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	b64 "encoding/base64"

	"github.com/gobuffalo/buffalo"
)

func ShowPasswordManager(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("password-manager.html"))
}

func SavePassword(c buffalo.Context) error {
	request := requests.AccountSetupRequest{}
	dataIsValid, validator, err := ValidateFormRequest(c, request)
	if err == nil {
		if dataIsValid {
			data := validator.SafeData()
			password := data["password"].(string)
			ex, err := os.Executable()
			if err != nil {
				panic(err)
			}
			exPath := filepath.Dir(ex)

			err = ioutil.WriteFile(filepath.Join(exPath, "ngencrpt.txt"), []byte(b64.StdEncoding.EncodeToString(models.EncryptPassword([]byte(password)))), 0644)
			if err != nil {
				log.Fatal(err)
			}
		}
		return c.Redirect(http.StatusFound, IndexURL)
	}
	return err
}
