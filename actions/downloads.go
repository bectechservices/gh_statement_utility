package actions

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
)

func HandleOtherPdfStampDownloads(c buffalo.Context) error {
	fileName := c.Param("file_name")
	if fileName == "" {
		return c.Redirect(302, "/")
	}
	tempDir := envy.Get("GH_ACCOUNT_STATEMENT_TEMP", "./")
	path := fmt.Sprintf("%s/%s", tempDir, fileName)
	file, err := getRequestedFile(path)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, r.Download(c, fileName, file))
}

func getRequestedFile(path string) (*os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return file, err
}
