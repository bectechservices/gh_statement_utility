package models

import (
	"gh-statement-app/pagination"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/now"
	"gorm.io/gorm"
)

// StatementPrintAudit is used by pop to map your statement_print_audits database table to your go code.
type EstampPrintAudit struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	FileName        string    `json:"file_name" db:"file_name"`
	NumPagesStamped int64     `json:"num_pages_stamped" db:"num_pages_stamped"`
	DateStamped     time.Time `json:"date_stamped" db:"date_stamped"`
	AccountNumber   string    `json:"account_number" db:"account_number"`
	AccountName     string    `json:"account_name" db:"account_name"`
	User            User      `gorm:"foreignKey:UserID;references:ID"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// StatementPrintAudits is not required by pop and may be deleted
type EstampPrintAudits []EstampPrintAudit

// LogPrintActivity logs the user print activity
func LogStampPrintActivity(tx *gorm.DB, user uuid.UUID, file_name, accountNumber, account_name string, num_pages_stamped int64, date_stamped time.Time) {

	if err := tx.Create(&EstampPrintAudit{
		ID:              NewUUID(),
		UserID:          user,
		AccountNumber:   accountNumber,
		AccountName:     account_name,
		FileName:        file_name,
		NumPagesStamped: num_pages_stamped,
		DateStamped:     date_stamped,
	}); err != nil {
		//panic(err)
	}
}

func ExportStampPrintAuditData(from, to, search string, tx *gorm.DB) EstampPrintAudits {
	audits := make([]EstampPrintAudit, 0)
	query := "%" + search + "%"
	if from != "" && to != "" {
		parsedFromT, _ := time.Parse("2006-01-02", from)
		parsedFrom := now.With(parsedFromT).BeginningOfDay()
		parsedToT, _ := time.Parse("2006-01-02", to)
		parsedTo := now.With(parsedToT).EndOfDay()
		tx.Where("( account_number like ? or account_name like ?) and created_at between ? and ?", query, query, parsedFrom, parsedTo).Order("created_at desc").Find(&audits)

	} else {
		tx.Where("account_number like ? or account_name like ?", query, query).Order("created_at desc").Find(&audits)

	}
	return audits
}

func GetTotalStampPrintAudit(tx *gorm.DB) int64 {
	var audited int64 = 0
	tx.Model(EstampPrintAudit{}).Count(&audited)
	return audited
}

// PaginateRecordsForAll view all reports without branch user restriction
func PaginateStampPrintAudit(from, to, search string, tx *gorm.DB, pagination pagination.Pagination) *pagination.Pagination {
	audits := make([]EstampPrintAudit, 0)

	query := "%" + search + "%"
	if from != "" && to != "" {
		parsedFromT, _ := time.Parse("2006-01-02", from)
		parsedFrom := now.With(parsedFromT).BeginningOfDay()
		parsedToT, _ := time.Parse("2006-01-02", to)
		parsedTo := now.With(parsedToT).EndOfDay()
		tx.Scopes(Paginate(audits, &pagination, tx)).Where("(account_number like ? or account_name like ?) and created_at between ? and ?", query, query, parsedFrom, parsedTo).Order("created_at desc").Find(&audits)
		pagination.Rows = audits
	} else {
		tx.Scopes(Paginate(audits, &pagination, tx)).Where("account_number like ? or account_name like ?", query, query).Order("created_at desc").Find(&audits)
		pagination.Rows = audits

	}
	return &pagination
}
