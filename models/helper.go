package models

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
)

var GLTables, StatsPTables []*string

var (
	glPrefix    = "gl_p______"
	statsPrefix = "stats_p______"
)

type AccountType string

const (
	AccountType_NEW = "New Account"
	AccountType_OLD = "Old Account"
)

// GetAccountType checks if account type is old or new
func GetAccountType(year, month int) AccountType {
	if year == 2010 && month == 12 {
		return AccountType_OLD
	}

	return AccountType_NEW
}

func GetFloatFromString(str string) float64 {
	floatValue, _ := strconv.ParseFloat(str, 64)
	return floatValue
}

// IsAccountOldOrNew checks whether account number is old or new
func IsAccountOldOrNew(accountNum string, db *gorm.DB) string {
	var oldAccountNumber string
	db.Raw("SELECT OLDACCOUNTNO FROM Nuban_Source_GL WHERE OLDACCOUNTNO= ? ", accountNum).Find(&oldAccountNumber)

	if len(oldAccountNumber) != 0 {
		return "old"
	} else {
		return "new"
	}
}

// IsloggedInResetForAllUsers reset all users with is_logged_in true
func IsloggedInResetForAllUsers() {
	log.Printf("starting cron at %s", time.Now().Format(time.DateTime))
	GormDB.Exec("Update users set is_logged_in = 'false' where is_logged_in = 'true'")
	log.Printf("cron completed at %s", time.Now().Format(time.DateTime))
}

func GetAllGeneralLedgerTables() []*string {
	dataChan := make(chan []*string, 1)

	go func(*gorm.DB, chan []*string) {
		data := make([]*string, 0)

		GormDB.Raw("select table_name from information_schema.tables where table_name like ? order by TABLE_NAME ASC", glPrefix).Find(&data)
		dataChan <- data
	}(GormDB, dataChan)

	return <-dataChan
}

func GetAllStatsPTables() []*string {
	dataChan := make(chan []*string, 1)

	go func(*gorm.DB, chan []*string) {
		data := make([]*string, 0)

		GormDB.Raw("select table_name from information_schema.tables where table_name like ? order by TABLE_NAME ASC", statsPrefix).Find(&data)
		dataChan <- data
	}(GormDB, dataChan)

	return <-dataChan
}

func GetLastDateOfPreviousMonth(year, month int) time.Time {
	date, _ := time.Parse(TIMELAYOUT, fmt.Sprintf("%d-%02d-01", year, month))
	return time.Date(date.Year(), date.Month(), 0, 0, 0, 0, 0, time.UTC)
}

func GLTableExist(glTable string) bool {
	for _, table := range GLTables {
		if *table == glTable {
			return true
		}
	}

	return false
}
func StatsTableExist(statsTable string) bool {
	for _, table := range StatsPTables {
		if *table == statsTable {
			return true
		}
	}

	return false
}
