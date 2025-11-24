package models

import (
	"fmt"
	"log"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/syyongx/php2go"
)

type Statement struct {
	AccountNum    string    `json:"account_no" form:"account_no"`
	StartDate     string    `json:"start_date" form:"start_date"`
	EndDate       string    `json:"end_date" form:"end_date"`
	I_Identity    int32     `json:"I_Identity" gorm:"column:I_Identity"`
	D_TxnPostDt   time.Time `json:"D_TxnPostDt" gorm:"column:D_TxnPostDt"`
	D_TxnValueDt  time.Time `json:"D_TxnValueDt" gorm:"column:D_TxnValueDt"`
	C_TxnNar1     string    `json:"C_TxnNar1" gorm:"column:C_TxnNar1"`
	C_TxnNar2     string    `json:"C_TxnNar2" gorm:"column:C_TxnNar2"`
	C_TxnNar3     string    `json:"C_TxnNar3" gorm:"column:C_TxnNar3"`
	C_TxnAccNo    string    `json:"C_TxnAccNo" gorm:"column:C_TxnAccNo"`
	I_Txn_CusCode string    `json:"I_Txn_CusCode" gorm:"column:I_Txn_CusCode"`
	I_Txn_Ccy     string    `json:"I_Txn_Ccy" gorm:"column:I_Txn_Ccy"`
	C_ccy_dsc     string    `json:"c_ccy_dsc" gorm:"column:c_ccy_dsc"`
	M_TxnAmt      float64   `json:"M_TxnAmt" gorm:"column:M_TxnAmt"`
}

type Statements []Statement

func (st Statement) SqlStatement(c buffalo.Context) string {
	var sql, tables, tableType string

	date1, _ := time.Parse(TIMELAYOUT, st.StartDate)
	statementStartMonth := date1.Month()
	statementStartYear := date1.Year()

	date2, _ := time.Parse(TIMELAYOUT, st.EndDate)
	statementEndMonth := date2.Month()
	statementEndYear := date2.Year()

	var monthTable, closeBalTable, glTables, gltables string
	if statementStartMonth == statementEndMonth && statementStartYear == statementEndYear {
		tables = fmt.Sprintf("stats_p%d%02d", statementStartYear, statementEndMonth)

		tableType = "single"
	} else {

		i := statementStartYear
		x := int(statementStartMonth)
		y := 12
		for i <= statementEndYear {

			if i == statementEndYear {
				y = int(date2.Month())
			}
			for x <= y {
				// if missing table
				tbl := fmt.Sprintf("stats_p%d%02d", i, x)
				if !StatsTableExist(tbl) {
					log.Println("Missing Table: ", tbl)
					x++
					continue
				}

				current, _ := time.Parse(TIMELAYOUT, fmt.Sprintf("%d-%02d-01", i, x))
				interval := current.Sub(date1)

				if interval >= 0 {

					tables = tables + fmt.Sprintf(`select I_Identity,D_TxnPostDt,D_TxnValueDt,C_TxnNar1,C_TxnNar2,C_TxnNar3,
									M_TxnAmt,C_TxnAccNo,I_Txn_CusCode,I_Txn_Ccy from stats_p%d%02d UNION ALL `, i, x)

					monthTable = monthTable + fmt.Sprintf(`select I_Identity,D_TxnPostDt,D_TxnValueDt,C_TxnNar1,C_TxnNar2,C_TxnNar3,M_TxnAmt,C_TxnAccNo,I_Txn_CusCode,
					I_Txn_Ccy from stats_p%d%02d AS txn_stats INNER JOIN Ccy_Type ON Ccy_Type.I_CCY_BM_Code = txn_stats.I_Txn_Ccy WHERE C_TxnAccNo = '%s' AND 
					( D_TxnPostDt BETWEEN '%s' AND '%s' ) ORDER BY D_TxnPostDt,I_Identity asc|`, i, x, st.AccountNum, st.StartDate, st.EndDate)

					closeBalTable = closeBalTable + fmt.Sprintf(`select M_TxnAmt,I_Txn_Ccy,C_TxnAccNo,D_TxnPostDt,D_TxnValueDt 
					from stats_p%d%02d  UNION ALL `, i, x)

					glTables = glTables + fmt.Sprintf(`select D_GLOpened from gl_p%d%02d UNION ALL `, i, x)

					closeBalTable = closeBalTable + fmt.Sprintf(`"gl_p%d%02d|`, i, x)
				}
				x++
			}
			i++
			x = 1 // sets to month 1 (Jan)
		}
		// return tables

		tables = php2go.Substr(tables, 0, len(tables)-11)
		// fmt.Println("OG TABLES: ", tables)
		gltables = php2go.Substr(tables, 0, len(glTables)-11)
		tableType = "multi"
		closeBalTable = php2go.Substr(closeBalTable, 0, len(closeBalTable)-11)
		c.Session().Set(fmt.Sprintf("closebal_table_%s", st.AccountNum), closeBalTable)
	}

	if tableType == "multi" {
		tsql := fmt.Sprintf(`SELECT D_GLDateClosed,C_GLAccNo FROM %s as txn_gl
		WHERE C_GLAccNo='%s' AND ( D_GLDateClosed BETWEEN '%s' AND '%s' )`, gltables, st.AccountNum, st.StartDate, st.EndDate)
		c.Session().Set(fmt.Sprintf("gl_table_%s", st.AccountNum), tsql)

		sql = fmt.Sprintf(`SELECT D_TxnValueDt,D_TxnPostDt,C_TxnNar1,C_TxnNar2,C_TxnNar3,M_TxnAmt,c_ccy_dsc,I_Txn_Ccy FROM (%s) as txn_stats
				   INNER JOIN Ccy_Type ON Ccy_Type.I_CCY_BM_Code = txn_stats.I_Txn_Ccy
				   WHERE C_TxnAccNo= '%s' AND ( D_TxnPostDt BETWEEN '%s' AND '%s' )
				   ORDER BY D_TxnPostDt,I_Identity asc`, tables, st.AccountNum, st.StartDate, st.EndDate)

	} else {

		tsql := fmt.Sprintf(`SELECT D_GLDateClosed,C_GLAccNo FROM %s  as txn_gl WHERE C_GLAccNo='%s' AND (D_GLDateClosed BETWEEN '%s' AND '%s')`, gltables, st.AccountNum, st.StartDate, st.EndDate)
		fmt.Println("Single TSQL: ", tsql)
		c.Session().Set(fmt.Sprintf("gl_table_%s", st.AccountNum), tsql)
		sql = fmt.Sprintf(`SELECT D_TxnValueDt,D_TxnPostDt,C_TxnNar1,C_TxnNar2,C_TxnNar3,M_TxnAmt,c_ccy_dsc,I_Txn_Ccy FROM %s as txn_stats
				   INNER JOIN Ccy_Type ON Ccy_Type.I_CCY_BM_Code = txn_stats.I_Txn_Ccy
				   WHERE C_TxnAccNo = '%s' AND ( D_TxnPostDt BETWEEN '%s' AND '%s' )
				   ORDER BY D_TxnPostDt,I_Identity ASC`, tables, st.AccountNum, st.StartDate, st.EndDate)

		monthTable = fmt.Sprintf(`select I_Identity,D_TxnPostDt,D_TxnValueDt,C_TxnNar1,C_TxnNar2,C_TxnNar3,M_TxnAmt,C_TxnAccNo,I_Txn_CusCode,I_Txn_Ccy 
		from %s AS txn_stats INNER JOIN Ccy_Type ON Ccy_Type.I_CCY_BM_Code = txn_stats.I_Txn_Ccy WHERE C_TxnAccNo = '%s' AND ( D_TxnPostDt BETWEEN '%s' AND '%s' ) ORDER BY D_TxnPostDt,I_Identity ASC|`, tables, st.AccountNum, st.StartDate, st.EndDate)
	}

	c.Session().Set(fmt.Sprintf("monthtable_%s", st.AccountNum), monthTable)
	// c.Session().Set(fmt.Sprintf("cust_table_%s", st.AccountNum), cusTable)
	return sql
}

func (st Statement) CheckAmount(amt float64) string {
	var transType string

	if amt < 0 {
		transType = "Withdrawal"
	} else {
		transType = "Deposit"
	}
	return transType
}
