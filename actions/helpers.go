package actions

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"ng-statement-app/models"
	"ng-statement-app/pagination"
	"ng-statement-app/requests"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/envy"
	"github.com/gofrs/uuid"
	"github.com/gookit/validate"
	"gorm.io/gorm"
)

// GB size in GB
var GB = uint64(1024 * 1024 * 1024)

// AccessPolicyRoute access ploicy route
type AccessPolicyRoute struct {
	Path  string `json:"path"`
	Alias string `json:"alias"`
}

// AccessPolicyRoutes plural form
type AccessPolicyRoutes []AccessPolicyRoute

// ValidationErrorWithData the structure of a validation error with its initial data
type ValidationErrorWithData struct {
	Field   string
	Error   string
	Value   string
	IsError bool
}

// PrettyDiskSize returns the disk size in a string format
func PrettyDiskSize(size uint64) string {
	sizeInGB := float32(size / GB)
	if sizeInGB > 1 {
		return fmt.Sprintf("%.2f GB", sizeInGB)
	}
	return fmt.Sprintf("%.2f MB", sizeInGB*1024)
}

// CustomError ValidationErrorWithData type
type CustomError = ValidationErrorWithData

// GetStructNameFromTag returns the struct field name given the tag
func GetStructNameFromTag(t reflect.Type, key, tag string) string {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		v := strings.Split(f.Tag.Get(key), ",")[0]
		if v == tag {
			return f.Name
		}
	}
	return ""
}

// GetAllStructTags returns all the tags for a given key for a struct
func GetAllStructTags(t reflect.Type, key string) []string {
	tags := make([]string, 0)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := strings.Split(f.Tag.Get(key), ",")[0]
		tags = append(tags, tag)
	}
	return tags
}

// ValidationErrorsWithData returns the validation errors with default data
func ValidationErrorsWithData(validator *validate.Validation, data interface{}) []ValidationErrorWithData {
	validationErrors := make([]ValidationErrorWithData, 0)
	t := reflect.TypeOf(data)
	value := reflect.ValueOf(data)
	for field := range validator.Errors {
		validationErrors = append(validationErrors, ValidationErrorWithData{
			Field:   field,
			IsError: true,
			Error:   validator.Errors.FieldOne(field),
			Value:   reflect.Indirect(value).FieldByName(GetStructNameFromTag(t, "form", field)).String(),
		})
	}
	tags := GetAllStructTags(t, "form")
	dataToResponse := make([]ValidationErrorWithData, 0)
	for _, tag := range tags {
		exists := false
		for _, err := range validationErrors {
			if err.Field == tag {
				exists = true
				break
			}
		}
		if !exists {
			dataToResponse = append(dataToResponse, ValidationErrorWithData{
				Field:   tag,
				IsError: false,
				Error:   "",
				Value:   reflect.Indirect(value).FieldByName(GetStructNameFromTag(t, "form", tag)).String(),
			})
		}
	}
	return append(validationErrors, dataToResponse...)
}

// CustomValidationErrorsWithData returns the custom validation errors with default data
func CustomValidationErrorsWithData(customError CustomError, data interface{}) []ValidationErrorWithData {
	t := reflect.TypeOf(data)
	value := reflect.ValueOf(data)
	tags := GetAllStructTags(t, "form")
	dataToResponse := make([]ValidationErrorWithData, 0)
	for _, tag := range tags {
		if customError.Field != tag {
			dataToResponse = append(dataToResponse, ValidationErrorWithData{
				Field:   tag,
				IsError: false,
				Error:   "",
				Value:   reflect.Indirect(value).FieldByName(GetStructNameFromTag(t, "form", tag)).String(),
			})
		}
	}
	return append(dataToResponse, customError)
}

// RedirectWithErrors redirects the request with validation errors
func RedirectWithErrors(c *buffalo.Context, validator *validate.Validation, request interface{}) error {
	validationErrors := ValidationErrorsWithData(validator, request)
	for _, err := range validationErrors {
		(*c).Flash().Add(fmt.Sprintf("validation_error_%s", err.Field), strconv.FormatBool(err.IsError))
		(*c).Flash().Add(fmt.Sprintf("validation_error_%s_message", err.Field), err.Error)
		(*c).Flash().Add(fmt.Sprintf("validation_error_%s_value", err.Field), err.Value)
	}
	return RedirectBack(c)
}

// ValidateFormRequest validates a form request
func ValidateFormRequest(context buffalo.Context, formRequest requests.FormRequest) (bool, *validate.Validation, error) {
	validator, err := formRequest.Validate(context)
	if err == nil {
		if validator.Validate() {
			return true, validator, nil
		}
		return false, validator, nil
	}
	return false, validator, err
}

// RedirectWithCustomError redirects the request with custom error
func RedirectWithCustomError(c *buffalo.Context, err CustomError, request interface{}) error {
	validationErrors := CustomValidationErrorsWithData(err, request)
	for _, err := range validationErrors {
		(*c).Flash().Add(fmt.Sprintf("validation_error_%s", err.Field), strconv.FormatBool(err.IsError))
		(*c).Flash().Add(fmt.Sprintf("validation_error_%s_message", err.Field), err.Error)
		(*c).Flash().Add(fmt.Sprintf("validation_error_%s_value", err.Field), err.Value)
	}
	return RedirectBack(c)
}

// AuthUser returns the authenticated user
func AuthUser(c buffalo.Context) models.User {
	if uid := c.Session().Get("auth_id"); uid != nil {
		switch id := uid.(type) {
		case uuid.UUID:
			return models.GetUserByID(id, DBConnection(c))
		}
	}
	panic("user not authenticated")
}

// AuthID returns the ID of the authenticated user
func AuthID(c buffalo.Context) uuid.UUID {
	if uid := c.Session().Get("auth_id"); uid != nil {
		switch id := uid.(type) {
		case uuid.UUID:
			return models.GetUserByID(id, DBConnection(c)).ID
		}
	}
	panic("user not authenticated")
}

// RedirectWithModalsOpen redirects with modals open
func RedirectWithModalsOpen(c *buffalo.Context, validator *validate.Validation, request interface{}, modal string) error {
	validationErrors := ValidationErrorsWithData(validator, request)
	for _, err := range validationErrors {
		(*c).Flash().Add(fmt.Sprintf("validation_error_%s", err.Field), strconv.FormatBool(err.IsError))
		(*c).Flash().Add(fmt.Sprintf("validation_error_%s_message", err.Field), err.Error)
		(*c).Flash().Add(fmt.Sprintf("validation_error_%s_value", err.Field), err.Value)
	}
	(*c).Flash().Add("show_modal", modal)
	return RedirectBack(c)
}

// RedirectBack redirects to previous page
func RedirectBack(c *buffalo.Context) error {
	return (*c).Redirect(302, (*c).Request().Referer())
}

// DBConnection returns an instance of the database connection
func DBConnection(c buffalo.Context) *gorm.DB {
	return models.GormDB
}

// RandomBytes generates a random bytes
func RandomBytes(n int) []byte {
	var letter = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]byte, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return b
}

// RandomString generates a random string
func RandomString(n int) string {
	return string(RandomBytes(n))
}

// APPURL return the app url
func APPURL() string {
	//return "localhost:3000"
	return envy.Get("GH_APP_URL", "localhost:4050")
}

// PasswordResetTokenURL returns the url plus the token
func PasswordResetTokenURL(token string) string {
	return APPURL() + NewPasswordSetupURL + "/?token=" + token
}

// GetResetTokenFromURL retreives the token from the url
func GetResetTokenFromURL(c buffalo.Context) string {
	values, isset := c.Request().URL.Query()["token"]
	if isset {
		return values[0]
	}
	return ""
}

func DBPaginator(c buffalo.Context, perPage int) pagination.Pagination {
	page, _ := strconv.Atoi(c.Param("page"))
	return pagination.Pagination{
		Limit: perPage,
		Page:  page,
	}
}

// AppendTimeToName appends the time to a string
func AppendTimeToName(name, ext string) string {
	return name + "-" + time.Now().Format("20060102150405") + "." + ext
}

func makeExcelIndex(x, y int) string {
	return fmt.Sprintf("%c%d", 65+y, x)
}

// ExportToExcel exports to excel
func ExportToExcel(c buffalo.Context, sheet, filename string, headings []string, data [][]string) error {
	file := excelize.NewFile()
	sheetIndex := file.NewSheet(sheet)
	for index, header := range headings {
		file.SetCellValue(sheet, makeExcelIndex(1, index), header)
	}
	for xIndex, datum := range data {
		for index, each := range datum {
			file.SetCellValue(sheet, makeExcelIndex(xIndex+2, index), each)
		}
	}
	file.SetActiveSheet(sheetIndex)
	excelBytes, err := file.WriteToBuffer()
	if err != nil {
		return err
	}

	c.Response().Header().Set("Content-Disposition", "attachment; filename="+filename)
	c.Response().Header().Set("Content-Transfer-Encoding", "binary")
	return c.Render(http.StatusOK, r.Func("application/octet-stream", func(w io.Writer, d render.Data) error {
		_, writeError := w.Write(excelBytes.Bytes())
		return writeError
	}))
}

func Strtotime(str string) (int64, error) {
	layout := "2006-01-02"
	t, err := time.Parse(layout, str)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// get current auth user
// func getAuthenticatedUser(ctx buffalo.Context) models.User {
// 	if uid := ctx.Session().Get("auth_id"); uid != nil {
// 		user := models.GetUserByStaffid(fmt.Sprintf("%s", uid), models.GormDB)
// 		return user
// 	}

//		return models.User{}
//	}
func SliceContainsSameElementsWithoutOrder(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}

	itemAppearsTimes := make(map[string]int, len(x))
	for _, i := range x {
		itemAppearsTimes[i]++
	}

	for _, i := range y {
		if _, ok := itemAppearsTimes[i]; !ok {
			return false
		}

		itemAppearsTimes[i]--

		if itemAppearsTimes[i] == 0 {
			delete(itemAppearsTimes, i)
		}
	}

	if len(itemAppearsTimes) == 0 {
		return true
	}

	return false
}
