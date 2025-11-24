package models

type NubanSourceGl struct {
	Masterno     string `json:"masterno" gorm:"masterno"`
	OldAccountNo string `json:"oldaccountno" gorm:"oldaccountno"`
	NewAccountNo string `json:"newaccountno" gorm:"newaccountno"`
}

type NubanSourceGls []NubanSourceGl
