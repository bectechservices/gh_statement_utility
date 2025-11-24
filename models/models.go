package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	b64 "encoding/base64"
	"fmt"
	dbPaginator "gh-statement-app/pagination"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
	"github.com/syyongx/php2go"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is a connection to your database to be used
// throughout your application.
var DB *pop.Connection

// GormDB object
var GormDB *gorm.DB

const TIMELAYOUT = "2006-01-02"
const COOKIE_EXPIRATION = 1 * 24 * time.Hour

// /-----------------------------------------------
// expire := time.Now().Add(20 * time.Minute) // Expires in 20 minutes
// cookie := http.Cookie{Name: "username", Value: "nonsecureuser", Path: "/", Expires: COOKIE_EXPIRATION, MaxAge: 86400}
// http.SetCookie(w, &cookie)
// cookie = http.Cookie{Name: "secureusername", Value: "secureuser", Path: "/", Expires: expire, MaxAge: 86400, HttpOnly: true, Secure: true}
// http.SetCookie(w, &cookie)
// /-----------------------------------------------

func Paginate(value interface{}, pagination *dbPaginator.Pagination, db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var totalRows int64
	db.Model(value).Count(&totalRows)
	pagination.TotalRows = totalRows
	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.Limit)))
	pagination.TotalPages = totalPages
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Order(pagination.GetSort())
	}
}

var PasswordEncryptionKey = "9z$C&F)J@NcRfUjXn2r5u7x!A%D*G-Ka"

func init() {
	var err error
	env := envy.Get("GO_ENV", "development")
	dbUrl := envy.Get("GH_DB_URL", "")
	DB, err = pop.Connect(env)
	if err != nil {
		log.Println(err)
	}

	encryptedPassword, err := b64.StdEncoding.DecodeString(string(ReadPasswordFromFile()))

	if err != nil {
		log.Println(err)
	}
	password := DecryptPassword(encryptedPassword)

	GormDB, err = gorm.Open(sqlserver.Open(strings.ReplaceAll(dbUrl, "__password__", string(password))), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	/// Logger: logger.Default.LogMode(logger.Info)
	if err != nil {
		log.Println(err)
	}

}

func NewUUID() uuid.UUID {
	id, _ := uuid.NewV4()
	return id
}

func EncryptPassword(text []byte) []byte {
	c, err := aes.NewCipher([]byte(PasswordEncryptionKey))
	if err != nil {
		fmt.Println(err)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		panic(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}

	return gcm.Seal(nonce, nonce, text, nil)
}

func DecryptPassword(text []byte) string {
	c, err := aes.NewCipher([]byte(PasswordEncryptionKey))
	if err != nil {
		panic(err)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		panic(err)
	}

	nonceSize := gcm.NonceSize()
	if len(text) < nonceSize {
		panic(err)
	}

	nonce, text := text[:nonceSize], text[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, text, nil)
	if err != nil {
		panic(err)
	}
	return string(plaintext)
}

func ReadPasswordFromFile() []byte {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	password, err := ioutil.ReadFile(filepath.Join(exPath, "ngencrpt.txt"))

	if err != nil {
		panic(err)
		//	fmt.Println(password)
	}
	return password
}

func (gl GeneralLedger) IfStartCusTableExist(start, end, cusId string) string {
	var table string

	date, _ := time.Parse(TIMELAYOUT, start)
	statementStartMonth := date.Month()
	statementStartYear := date.Year()

	date, _ = time.Parse(TIMELAYOUT, end)
	statementEndYear := date.Year()

	i := statementStartYear
	x := int(statementStartMonth)
	y := 12

	for i <= statementEndYear {
		for x <= y {
			table = fmt.Sprintf("Cus_reg_p%d%02d", i, x)

			return table
			x++
		}

	}

	return table
}

func TrimDateFromDb(date string) string {
	return php2go.Explode("T", date)[0]
}

func GetOldAccount(accountNum string) string {
	var masterNum string

	tx := GormDB.Raw("SELECT MASTERNO FROM Nuban_Source_GL WHERE NEWACCOUNTNO = ?", accountNum).Scan(&masterNum)

	if tx.Error != nil {
		return ""
	}
	return masterNum
}

func GetDbEntity(ctx *gorm.DB) {
	return
}

func GetBranch(branchId string) (string, error) {
	var branchName string

	tx := GormDB.Raw("select C_BranchDesc from lkup_branch where I_BranchId = ?", branchId).Scan(&branchName)
	if tx.Error != nil {
		return branchName, tx.Error
	}
	return branchName, nil
}

func List(arr []string, dest ...*string) {
	for i := range dest {
		if len(arr) > i {
			*dest[i] = arr[i]
		}
	}
}

func Explode(delimiter, text string) []string {
	if len(delimiter) > len(text) {
		return strings.Split(delimiter, text)
	} else {
		return strings.Split(text, delimiter)
	}
}

func LogData(data string) {
	url := envy.Get("APP_LOG_URL", "C:/TEMP/LOGS")
	filePath := fmt.Sprintf("%s/%s.log", url, time.Now().Format("2006-01-02"))

	if data == "" {
		data = "No Data"
	}
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	customeData := fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), data)
	if _, err := f.Write([]byte(customeData)); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
