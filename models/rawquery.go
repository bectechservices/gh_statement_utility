package models

import "gorm.io/gorm"

type AccountDates struct {
	From string
	To   string
}

func GetAccountDatesForStatement(account string, db *gorm.DB) AccountDates {
	return AccountDates{
		From: "2020-02-02",
		To:   "2030-01-01",
	}
}
