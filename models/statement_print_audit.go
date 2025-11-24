package models

import (
	"gh-statement-app/pagination"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/now"
	"gorm.io/gorm"
)

// StatementPrintAudit is used by pop to map your statement_print_audits database table to your go code.
type StatementPrintAudit struct {
	ID              uuid.UUID `json:"id" db:"id"`
	AccountNumber   string    `json:"account_number" db:"account_number"`
	AccountName     string    `json:"account_name" db:"account_name"`
	Pages           int64     `json:"pages" db:"pages"`
	PrintType       string    `json:"print_type" db:"print_type"`
	QueryDateFrom   time.Time `json:"query_date_from" db:"query_date_from"`
	QueryDateTo     time.Time `json:"query_date_to" db:"query_date_to"`
	RequestedBy     string    `json:"requested_by" db:"requested_by"`
	RequesterBranch string    `json:"requester_branch" db:"requester_branch"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// StatementPrintAudits is not required by pop and may be deleted
type StatementPrintAudits []StatementPrintAudit

// LogPrintActivity logs the user print activity
func LogPrintActivity(tx *gorm.DB, user, branch_name, accountNumber, account_name string, pages int64, print_type string, dateFrom, dateTo time.Time) {
	if err := tx.Create(&StatementPrintAudit{
		ID:              NewUUID(),
		RequestedBy:     user,
		RequesterBranch: branch_name,
		AccountNumber:   accountNumber,
		AccountName:     account_name,
		Pages:           pages,
		PrintType:       print_type,
		QueryDateFrom:   dateFrom,
		QueryDateTo:     dateTo,
	}); err != nil {
		//panic(err)
	}
}

// PaginateRecordsForAll view all reports without branch user restriction
func PaginateStatementPrintAudit(from, to, search string, tx *gorm.DB, pagination pagination.Pagination) *pagination.Pagination {
	audits := make([]StatementPrintAudit, 0)

	query := "%" + search + "%"
	if from != "" && to != "" {
		parsedFromT, _ := time.Parse("2006-01-02", from)
		parsedFrom := now.With(parsedFromT).BeginningOfDay()
		parsedToT, _ := time.Parse("2006-01-02", to)
		parsedTo := now.With(parsedToT).EndOfDay()
		tx.Scopes(Paginate(audits, &pagination, tx)).Where("(requested_by like ? or requester_branch like ? or account_number like ? or account_name like ?) and created_at between ? and ?", query, query, query, query, parsedFrom, parsedTo).Order("created_at desc").Find(&audits)
		pagination.Rows = audits
	} else {
		tx.Scopes(Paginate(audits, &pagination, tx)).Where("requested_by like ? or requester_branch like ? or account_number like ? or account_name like ?", query, query, query, query).Order("created_at desc").Find(&audits)
		pagination.Rows = audits

	}
	return &pagination
}

// PaginateBranchStatementPrintAudit view all reports without branch user restriction
func PaginateBranchStatementPrintAudit(from, to, search string, branch uuid.UUID, tx *gorm.DB, pagination pagination.Pagination) *pagination.Pagination {
	audits := make([]StatementPrintAudit, 0)
	//branches := make([]Branches, 0)

	//branch1 = tx.Raw(`select code from branches where id = ?`, branch).Scan(&branches)
	//branch1 := tx.Exec(`select name from branches where id = ?`, branch)

	//fmt.Println("##############", branch1, "###############")
	query := "%" + search + "%"
	//query1 := "%" + branch + "%"
	if from != "" && to != "" {
		parsedFromT, _ := time.Parse("2006-01-02", from)
		parsedFrom := now.With(parsedFromT).BeginningOfDay()
		parsedToT, _ := time.Parse("2006-01-02", to)
		parsedTo := now.With(parsedToT).EndOfDay()
		tx.Scopes(Paginate(audits, &pagination, tx)).Where("(requested_by like ? or requester_branch like ? or account_number like ? or account_name like ?) and created_at between ? and ?  and branch_requested like ?", query, query, query, query, parsedFrom, parsedTo, branch).Order("created_at desc").Find(&audits)
		pagination.Rows = audits
	} else {
		tx.Scopes(Paginate(audits, &pagination, tx)).Where("requester_branch like ?", query).Order("created_at desc").Find(&audits)
		pagination.Rows = audits

	}
	return &pagination
}

func ExportStatementPrintAuditData(from, to, search string, tx *gorm.DB) StatementPrintAudits {
	audits := make([]StatementPrintAudit, 0)
	query := "%" + search + "%"
	if from != "" && to != "" {
		parsedFromT, _ := time.Parse("2006-01-02", from)
		parsedFrom := now.With(parsedFromT).BeginningOfDay()
		parsedToT, _ := time.Parse("2006-01-02", to)
		parsedTo := now.With(parsedToT).EndOfDay()
		tx.Where("(requested_by like ? or requester_branch like ? or account_number like ? or account_name like ?) and created_at between ? and ?", query, query, query, query, parsedFrom, parsedTo).Order("created_at desc").Find(&audits)

	} else {
		tx.Where("requested_by like ? or requester_branch like ? or account_number like ? or account_name like ?", query, query, query, query).Order("created_at desc").Find(&audits)

	}
	return audits
}

func GetTotalStatementPrintAudit(tx *gorm.DB) int64 {
	var audited int64 = 0
	tx.Model(StatementPrintAudit{}).Count(&audited)
	return audited
}
