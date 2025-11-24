package models

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/syyongx/php2go"
	"gorm.io/gorm"
)

type GeneralLedgerStart struct {
	AccountNum string `json:"account_no" form:"account_no" gorm:"account_no"`
	StartDate  string `json:"start_date" form:"start_date" gorm:"start_date"`
	EndDate    string `json:"end_date" form:"end_date" gorm:"end_date"`
}

type GeneralLedger struct {
	AccountNum      string    `json:"account_no" form:"account_no"`
	StartDate       string    `json:"start_date" form:"start_date"`
	EndDate         string    `json:"end_date" form:"end_date"`
	D_GLOpened      time.Time `json:"D_GLOpened" gorm:"column:D_GLOpened"`
	C_GLAccNo       string    `json:"C_GLAccNo" gorm:"column:C_GLAccNo"`
	C_GLCusCode     string    `json:"C_GLCusCode" gorm:"column:C_GLCusCode"`
	C_Cus_Shortname string    `json:"C_Cus_Shortname" gorm:"column:C_Cus_Shortname"`
	C_GLAccDesc     string    `json:"C_GLAccDesc" gorm:"column:C_GLAccDesc"`
	M_GL_BkBal      float64   `json:"M_GL_BkBal" gorm:"column:M_GL_BkBal"`
	D_lastchanged   time.Time `json:"D_lastchanged" gorm:"column:D_lastchanged"`
	D_GL_Stdate     time.Time `json:"D_GL_Stdate" gorm:"column:D_GL_Stdate"`
	I_GL_Ccy        string    `json:"I_GL_Ccy" gorm:"column:I_GL_Ccy"`
	M_GL_ClrBal     float64   `json:"M_GL_ClrBal" gorm:"column:M_GL_ClrBal"`
	C_GLBranchID    string    `json:"C_GLBranchID" gorm:"column:C_GLBranchID"`
	M_TxnAmt        float64   `json:"M_TxnAmt" gorm:"column:M_TxnAmt"`
}

type GeneralLedgerStarts []GeneralLedgerStart
type GeneralLedgers []GeneralLedger

// GetAdminAccountStartDates This is for Nuban Accounts (2006 - 2010)

func (gl GeneralLedger) GetAdminAccountStartDates() map[string]string {
	fmt.Println("######## This is admin Statement #######")
	result := make(map[string]string, 0)

	tablePrefix := "gl_p______"

	var firstTable, lastTable string

	wg := &sync.WaitGroup{}

	// get last table
	lastTableSql := "select TOP 1 table_name from information_schema.tables where table_name like ? order by TABLE_NAME DESC"
	wg.Add(1)
	// get last table name
	go func(string, string, string, *gorm.DB, *sync.WaitGroup) {
		defer wg.Done()
		GormDB.Raw(lastTableSql, tablePrefix).Find(&lastTable)
	}(lastTableSql, tablePrefix, lastTable, GormDB, wg)

	// get first table
	sql := "select TOP 1 table_name from information_schema.tables where table_name like ? order by TABLE_NAME ASC"

	wg.Add(1)
	// get last table name
	go func(string, string, string, *gorm.DB, *sync.WaitGroup) {
		defer wg.Done()
		GormDB.Raw(sql, tablePrefix).Find(&firstTable)
	}(sql, tablePrefix, firstTable, GormDB, wg)

	wg.Wait()

	log.Println("lastTable; ", lastTable)

	// set year and  month for last table
	lastTableYearStr := php2go.Substr(lastTable, 4, 4)
	lastTableMonthStr := php2go.Substr(lastTable, 8, 2)

	lastTableYear, _ := strconv.Atoi(lastTableYearStr)
	lastTableMonth, _ := strconv.Atoi(lastTableMonthStr)

	lastTableDate := time.Date(lastTableYear, time.Month(lastTableMonth), 1, 0, 0, 0, 0, time.UTC)
	lastTableDate = time.Date(lastTableDate.Year(), lastTableDate.Month()+1, 0, 0, 0, 0, 0, lastTableDate.Location())
	// dayOfLastDayOfthisMonth := lastDayOfthisMonth.Day()

	tableNames := make([]string, 0)
	sql = "select table_name from information_schema.tables WHERE table_name like ? ORDER BY table_name ASC"

	// get all table names with gl_p_____
	GormDB.Raw(sql, tablePrefix).Find(&tableNames)
	var rowsAffected int64
	if len(tableNames) > 0 {
		var sqlUnion string
		for _, table := range tableNames {
			sqlUnion = sqlUnion + "select D_GLOpened,C_GLAccNo,M_GL_BkBal,M_GL_ClrBal,D_lastchanged from " + table + " UNION ALL "
		}

		sqlUnion = php2go.Substr(sqlUnion, 0, len(sqlUnion)-11)

		sql = fmt.Sprintf("SELECT TOP 1 D_GLOpened,C_GLAccNo,M_GL_BkBal,M_GL_ClrBal,D_lastchanged FROM (" + sqlUnion + ") AS gl WHERE C_GLAccNo= ? ")

		wg.Add(1)
		go func(string, GeneralLedger, *gorm.DB, *sync.WaitGroup) {
			defer wg.Done()
			rowsAffected = GormDB.Raw(sql, gl.AccountNum).Find(&gl).RowsAffected
		}(sql, gl, GormDB, wg)

		wg.Wait()

	}

	var firstTableDate time.Time

	accountStartDate := gl.D_GLOpened

	if rowsAffected < 1 {
		result["Year"] = ""
		result["Month"] = ""
		result["Day"] = ""
		result["ltable_y"] = fmt.Sprintf("%d", lastTableDate.Year())
		result["ltable_m"] = fmt.Sprintf("%02d", lastTableDate.Month())
		result["ltable_d"] = fmt.Sprintf("%02d", lastTableDate.Day())
		return result
	}

	if !accountStartDate.IsZero() {
		firstTableYearStr := php2go.Substr(firstTable, 4, 4)
		firstTableMonthStr := php2go.Substr(firstTable, 8, 2)
		firstTableDate, _ = time.Parse(TIMELAYOUT, fmt.Sprintf("%s-%s-01", firstTableYearStr, firstTableMonthStr))

		if accountStartDate.Before(firstTableDate) || accountStartDate.Equal(firstTableDate) {
			accountStartDate = firstTableDate
		}
	} else {
		result["Year"] = ""
		result["Month"] = ""
		result["Day"] = ""
		result["ltable_y"] = fmt.Sprintf("%d", lastTableDate.Year())
		result["ltable_m"] = fmt.Sprintf("%02d", lastTableDate.Month())
		result["ltable_d"] = fmt.Sprintf("%02d", lastTableDate.Day())
		return result
	}

	fmt.Println("accountStartDate: ", accountStartDate)
	fmt.Println("firstTableDate: ", firstTableDate)

	accountType := IsAccountOldOrNew(gl.AccountNum, GormDB)
	log.Printf("Checking Account AC %s is an %s accType", gl.AccountNum, accountType)
	if accountType == "old" {
		// old accounts start from 2006 feb per the restored date
		accountStartDate, _ = time.Parse(time.DateOnly, "2006-02-01")
		lastTableDate, _ = time.Parse(time.DateOnly, "2010-12-31")
	} else {
		// new accounts start from 2011 jan per the restored date
		defaultSlstmtDate, _ := time.Parse(TIMELAYOUT, "2011-01-01")
		if accountStartDate.Before(defaultSlstmtDate) {
			accountStartDate = time.Date(defaultSlstmtDate.Year(), gl.D_GLOpened.Month(), defaultSlstmtDate.Day(), 0, 0, 0, 0, time.UTC)
		}
	}

	result["Year"] = fmt.Sprintf("%d", accountStartDate.Year())
	// result["Month"] = fmt.Sprintf("%02d", accountStartDate.Month()+1)
	result["Month"] = fmt.Sprintf("%02d", accountStartDate.Month())
	result["Day"] = fmt.Sprintf("%02d", 01)
	//	result["Day"] = fmt.Sprintf("%02d", accountStartDate.Day())
	result["ltable_y"] = fmt.Sprintf("%d", lastTableDate.Year())
	result["ltable_m"] = fmt.Sprintf("%02d", lastTableDate.Month())
	result["ltable_d"] = fmt.Sprintf("%02d", lastTableDate.Day())

	return result
}

// Get first transaction logic
func (gl GeneralLedger) GetFirstTransaction() (string, error) {
	date, _ := time.Parse(TIMELAYOUT, gl.StartDate)
	statementStartMonth := date.Month()
	statementStartYear := date.Year()

	date, _ = time.Parse(TIMELAYOUT, gl.EndDate)
	statementEndMonth := date.Month()
	statementEndYear := date.Year()
	var transStartDate string

	i := statementStartYear
	x := int(statementStartMonth)
	y := 1

	for i >= statementStartYear {
		for x >= y {
			table := fmt.Sprintf("stats_p%d%02d", i, x)
			sql := fmt.Sprintf("select top 1 D_TxnPostDt from %s where C_TxnAccNo = ?", table)

			GormDB.Raw(sql, gl.AccountNum).Scan(&transStartDate)
			x++
			return php2go.Explode("T", transStartDate)[0], nil
		}
		i++
		x = 12

		if i == statementEndYear {
			y = int(statementEndMonth)
		}
	}

	return "", nil
}

func (gl GeneralLedger) GetAccStartDate(c buffalo.Context) string {
	accstartdate := "null"
	custcode := GetOldAccount(gl.AccountNum) // get old account

	newDate, _ := time.Parse(TIMELAYOUT, gl.StartDate)

	previous_month := time.Date(newDate.Year(), newDate.Month(), 0, 0, 0, 0, 0, newDate.Location())

	y := previous_month.Year()
	m := previous_month.Month()
	var sql string

	if custcode != "" {
		sql = fmt.Sprintf("SELECT D_GLOpened,C_GLCusCode FROM gl_p%d%02d WHERE C_GLCusCode='%s'", y, m, custcode)
	} else {

		sql = fmt.Sprintf("SELECT D_GLOpened,C_GLCusCode FROM gl_p%d%02d WHERE C_GLAccNo='%s'", y, m, gl.AccountNum)
	}

	// Execute query
	GormDB.Raw(sql).Scan(&gl)

	if !gl.D_GLOpened.IsZero() {
		accstartdate = gl.D_GLOpened.Format("2006-01-02 15:04:05.999999999")
	}
	// c.Session().Set("cus_reg_date", accstartdate)
	// c.Session().Set("D_DteOpened", accstartdate)
	// c.Session().Set("org_cus_reg_date", accstartdate)
	// c.Session().Set("cus_id", gl.C_GLCusCode)

	c.Session().Set(fmt.Sprintf("cus_reg_date_%s", gl.AccountNum), accstartdate)
	c.Session().Set(fmt.Sprintf("D_DteOpened_%s", gl.AccountNum), accstartdate)
	c.Session().Set(fmt.Sprintf("org_cus_reg_date_%s", gl.AccountNum), accstartdate)
	c.Session().Set(fmt.Sprintf("cus_id_%s", gl.AccountNum), gl.C_GLCusCode)

	c.Session().Set(fmt.Sprintf("cus_id_%s", gl.AccountNum), gl.C_GLCusCode)

	return accstartdate

}

func (gl GeneralLedger) MarchAllGLTransactions() []map[string]string {
	noTrans := ""
	var totalTnx float64

	var result []map[string]string
	strStartDate, _ := php2go.Strtotime(TIMELAYOUT, gl.StartDate)

	strEndDate, _ := php2go.Strtotime(TIMELAYOUT, gl.EndDate)

	var tempCB float64
	var addOneMonth time.Time
	tempCB = 0
	i := 0
	addOneMonth, _ = time.Parse(TIMELAYOUT, gl.StartDate)
	for strStartDate < strEndDate {
		m := addOneMonth.Month()
		y := addOneMonth.Year()
		f := addOneMonth.Month().String()
		finalEndDate := time.Date(addOneMonth.Year(), addOneMonth.Month()+1, 0, 0, 0, 0, 0, time.UTC).Format(TIMELAYOUT)

		newDate := fmt.Sprintf("%d-%02d-01", y, m)
		fmt.Println("NEWDATE: ", newDate)

		// loop thru the start and end date, and compare all the sum of transaction for
		// each month plus previous month GL with the current GL balance for that month.
		// if the two match continue to loop till end. the event that two do not match the
		// function exits and returns the previous date that the GL and Stats marched

		var table string
		if m < 10 {
			table = fmt.Sprintf("stats_p%d0%d", y, m)
		} else {
			table = fmt.Sprintf("stats_p%d%d", y, m)
		}

		sql := fmt.Sprintf("Select ISNULL(SUM(cast(M_TxnAmt as decimal(18,2))),100000000000000000000000000000) AS totaltnx from %s where C_TxnAccNo = ?", table)
		// Execute query
		GormDB.Raw(sql, gl.AccountNum).Scan(&totalTnx)

		if totalTnx == 100000000000000000000000000000 {
			fmt.Println("NOTRANS CURRENT MONTH: ", addOneMonth)
			noTrans = "yes"
			totalTnx = 0
		} else {
			totalTnx = totalTnx * -1
			noTrans = "no"
		}

		// validityTotalTnx := totalTnx

		preMonth, _ := time.Parse(TIMELAYOUT, newDate)

		preMonth = time.Date(preMonth.Year(), preMonth.Month()-1, preMonth.Day(), 0, 0, 0, 0, preMonth.Location())

		openingBal, tableExists := gl.FindOpenBalance(preMonth) // * -1
		openingBal = openingBal * -1

		// has no transaction
		closingBal := 0.0
		if totalTnx == 0 {
			currentMonth := time.Date(preMonth.Year(), preMonth.Month()+1, preMonth.Day(), 0, 0, 0, 0, preMonth.Location())
			fmt.Println("CURRENT MONTH: ", currentMonth)
			closingBal, _ = gl.FindOpenBalance(currentMonth) // * -1
			closingBal = closingBal * -1
			// validity = ""
		} else {
			closingBal = totalTnx + openingBal
			// fmt.Printf("TOTALTNX: %f + OPENINGBAL: %f = CLOSINGBAL: %f\n", totalTnx, openingBal, closingBal)
		}

		// For some reson assigning closingbal to tempcb gives adds additional value
		// recommended using numberformat for comparism

		// var tempCBNumFormat, openingBalNumFormat string
		// tempCBNumFormat = php2go.NumberFormat(php2go.Round(closingBal, 2), 2, ".", ",")
		// openingBalNumFormat = php2go.NumberFormat(php2go.Round(closingBal, 2), 2, ".", ",")
		//
		// if noTrans == "yes" {
		//	if tempCBNumFormat != openingBalNumFormat {
		//		validity = "FALSE"
		//	} else {
		//		validity = ""
		//	}
		// }

		var validity string

		// table does exist and opening bal and total transaction equals closing balance
		fmt.Println("Date: ", preMonth)
		fmt.Println("tableExists: ", tableExists)
		fmt.Println("(openingBal+totalTnx): ", (openingBal + totalTnx))
		fmt.Println("closingBal: ", closingBal)

		if tableExists == "" && (openingBal+totalTnx) == closingBal {
			validity = ""
			// } else {
			// 	validity = "FALSE"
			// }
		}
		if tableExists == "" && (openingBal+totalTnx) != closingBal {
			validity = "FALSE"
		}

		if tableExists != "" {
			if openingBal == 0 && closingBal == 0 {
				tempCB = 0.0
				fmt.Println("Closing Balance TEMPCB: ", tempCB)
				validity = ""

				addOneMonth = time.Date(addOneMonth.Year(), addOneMonth.Month()+1, addOneMonth.Day(), 0, 0, 0, 0, addOneMonth.Location())
				strStartDate, _ = php2go.Strtotime(TIMELAYOUT, addOneMonth.Format(TIMELAYOUT))

				i += 1
				totalTnx = 0.0

				continue
			}
			// get opening balance of current month and make it as the total tnx and closing balance
			// make the opening balance as zero since the account num has no transaction on the previous month
			currentMonthDate, _ := time.Parse(TIMELAYOUT, newDate)
			closingBal, _ = gl.FindOpenBalance(currentMonthDate)
			closingBal = closingBal * -1
			log.Printf("closing balance for false table exist: %f\n", closingBal)

			log.Printf("totalTnx (%f) != closingBal (%f)\n", totalTnx, closingBal)
			log.Printf("(openingBal (%f)+totalTnx (%f)) != closingBal (%f)", openingBal, totalTnx, closingBal)
			if (openingBal + totalTnx) != closingBal {
				validity = "FALSE"
			}
		}

		currentMonth := fmt.Sprintf("%s %d", f, y)

		dt, _ := time.Parse(TIMELAYOUT, finalEndDate)
		finalStartDate := time.Date(dt.Year(), dt.Month(), 1, 0, 0, 0, 0, dt.Location()).Format(TIMELAYOUT)

		if noTrans == "yes" {
			validateInfo := make(map[string]string)
			validateInfo["current_month"] = currentMonth
			validateInfo["has_transaction"] = "true"
			validateInfo["opening_balance"] = php2go.NumberFormat(php2go.Round(openingBal, 2), 2, ".", ",")
			validateInfo["total_tnx"] = "0.00"
			validateInfo["closing_balance"] = php2go.NumberFormat(php2go.Round(closingBal, 2), 2, ".", ",")
			validateInfo["validity"] = validity
			validateInfo["finalStartDate"] = finalStartDate
			validateInfo["finalEndDate"] = finalEndDate

			result = append(result, validateInfo)

		} else {
			validateInfo := make(map[string]string)
			validateInfo["current_month"] = currentMonth
			validateInfo["has_transaction"] = "true"
			validateInfo["opening_balance"] = php2go.NumberFormat(php2go.Round(openingBal, 2), 2, ".", ",")
			validateInfo["total_tnx"] = php2go.NumberFormat(php2go.Round(totalTnx, 2), 2, ".", ",")
			validateInfo["closing_balance"] = php2go.NumberFormat(php2go.Round(closingBal, 2), 2, ".", ",")
			validateInfo["validity"] = validity
			validateInfo["finalStartDate"] = finalStartDate
			validateInfo["finalEndDate"] = finalEndDate

			result = append(result, validateInfo)
		}

		tempCB = closingBal
		fmt.Println("Closing Balance TEMPCB: ", tempCB)
		validity = ""

		addOneMonth = time.Date(addOneMonth.Year(), addOneMonth.Month()+1, addOneMonth.Day(), 0, 0, 0, 0, addOneMonth.Location())
		strStartDate, _ = php2go.Strtotime(TIMELAYOUT, addOneMonth.Format(TIMELAYOUT))

		i += 1
		totalTnx = 0.0

		// break initiatlise
	}

	return result
}

func (gl GeneralLedger) FindOpenBalance(startDate time.Time) (float64, string) {
	year := startDate.Year()
	month := startDate.Month()
	var bookBal float64

	sql := fmt.Sprintf("Select M_GL_BkBal from gl_p%d%02d where C_GLAccNo=?", year, month)
	// fmt.Println("OPENBAL SQL: ", sql)
	// Execute query
	tx := GormDB.Raw(sql, gl.AccountNum).First(&bookBal)
	// fmt.Println("OPENBAL SQL RESULT: ", bookBal)

	if tx.Error != nil {
		return 0, "FALSE"
	}

	return bookBal, ""
}

func (gl GeneralLedger) IfCusTableEmpty(c buffalo.Context) string {
	var sql, table string
	newDate, _ := time.Parse(TIMELAYOUT, gl.StartDate)
	endDate, _ := time.Parse(TIMELAYOUT, gl.EndDate)

	date := time.Date(newDate.Year(), newDate.Month(), newDate.Day(), newDate.Hour(), newDate.Minute(), newDate.Second(), newDate.Nanosecond(), newDate.Location())
	statement_start_Year := date.Year()

	date = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), endDate.Hour(), endDate.Minute(), endDate.Second(), endDate.Nanosecond(), endDate.Location())
	statement_end_Month := date.Month()
	statement_end_Year := date.Year()

	i := statement_end_Year
	x := statement_end_Month
	y := 1

	for i >= statement_start_Year {
		intX := int(x)
		strX := strconv.Itoa(intX)
		val, err := strconv.Atoi(strX)
		if err != nil {
			return "eer1"
		}
		for val >= y {

			sql = fmt.Sprintf("select top 1 I_Cus_Code from cus_reg_p%d%02d", i, int(x))
			table = fmt.Sprintf("cus_reg_p%d%02d", i, x)

			// Execute query
			var cusCode string
			res := GormDB.Raw(sql).First(&cusCode)

			if res.Error != nil {
				return table
			}
			// c.Session().Set("cus_table", table)
			c.Session().Set(fmt.Sprintf("cus_table_%s", gl.AccountNum), table)
			c.Session().Set(fmt.Sprintf("cus_table_%s", gl.AccountNum), table)
			fmt.Println("CusYable: ", table)
			er := c.Session().Save()
			fmt.Println("SESSION ERR: ", er)
			t_date := fmt.Sprintf("%d%02d-01", i, x)
			// c.Session().Set("cus_reg_date_1", t_date)
			c.Session().Set(fmt.Sprintf("cus_reg_date_1_%s", gl.AccountNum), t_date)

			x = x - 1
			return table

		}

		i = i - 1
		x = 12
		if i == statement_end_Year {
			y = int(statement_end_Month)
		}
	}

	return table
}

func (gl GeneralLedger) MissingGLTable() (string, error) {
	date, _ := time.Parse(TIMELAYOUT, gl.StartDate)
	statementStartMonth := int(date.Month())
	statementStartYear := int(date.Year())

	date, _ = time.Parse(TIMELAYOUT, gl.EndDate)
	statementEndMonth := int(date.Month())
	statementEndYear := int(date.Year())

	december := statementEndMonth

	// loop thru start year and end year
	for statementStartYear <= statementEndYear {
		for statementStartMonth <= december {
			table := fmt.Sprintf("gl_p%d%02d", statementStartYear, statementStartMonth)

			if !GLTableExist(table) {
				log.Println("Missing Table: ", table)
				statementStartMonth = statementStartMonth + 1
				continue
			}

			sql := fmt.Sprintf("Select C_GLAccNo from %s", table)

			rows, err := GormDB.Raw(sql).Rows()
			if err != nil {
				return table, nil
			}
			defer rows.Close()

			statementStartMonth = statementStartMonth + 1
		}

		statementStartYear = statementStartYear + 1
		statementStartMonth = 1
		if statementStartYear == statementEndYear {
			december = statementEndMonth
		}
	}

	table := "All tables exist"

	return table, nil
}

func (gl GeneralLedger) FindOpeningBalance(c buffalo.Context) (map[string]interface{}, error) {

	glModel := &GeneralLedger{}
	result := make(map[string]interface{})

	sDate, _ := time.Parse(TIMELAYOUT, gl.StartDate)
	sDate = time.Date(sDate.Year(), sDate.Month(), sDate.Day(), 0, 0, 0, 0, sDate.Location())
	sdDate := sDate
	// c.Session().Set("clr_bal_Start_date", sdDate.Format(TIMELAYOUT))
	c.Session().Set(fmt.Sprintf("clr_bal_Start_date_%s", gl.AccountNum), sdDate.Format(TIMELAYOUT))
	sessionErr := c.Session().Save()

	if sessionErr != nil {
		c.Logger().Error(sessionErr)
	}

	// preMonth := time.Date(sDate.Year(), sDate.Month()-1, sDate.Day(), 0, 0, 0, 0, sDate.Location())
	preMonth := time.Date(sDate.Year(), sDate.Month()-1, sDate.Day(), 0, 0, 0, 0, sDate.Location())
	y := preMonth.Year()
	m := int(preMonth.Month())

	// if y == 2010 && m == 12 {
	// 	if accType, _ := c.Session().Get("acc_type"); accType == "new" {
	// 		oldAcc, _ := c.Session().Get("oldacc")
	// 		gl.AccountNum = fmt.Sprintf("%s", oldAcc)
	// 	}
	// }

	table := fmt.Sprintf("gl_p%d%02d", y, m)

	sql := fmt.Sprintf(`Select C_Cus_Shortname,C_GLAccDesc,M_GL_BkBal,D_lastchanged,D_GL_Stdate,I_GL_Ccy,C_GLCusCode,
		M_GL_ClrBal,C_GLBranchID from %s where C_GLAccNo = ?`, table)

	// Execute query
	GormDB.Raw(sql, gl.AccountNum).Scan(&glModel)

	result["cusName"] = glModel.C_Cus_Shortname
	result["accDesc"] = glModel.C_GLAccDesc
	result["bookBal"] = glModel.M_GL_BkBal * -1
	result["clrBal"] = glModel.M_GL_ClrBal * -1
	result["glStartdate"] = glModel.D_GL_Stdate.Format(TIMELAYOUT)
	result["Ccy"] = glModel.I_GL_Ccy
	result["Cus_id"] = glModel.C_GLCusCode
	result["cusBranch"] = glModel.C_GLBranchID
	result["openBalanceDate"] = glModel.D_lastchanged.Format(TIMELAYOUT)

	return result, nil
}

// match General Ledger transactions
func (gl GeneralLedger) MatchGLTransactions() (map[string]interface{}, error) {
	match := false
	// var totalTnx, closingBal string
	data := make(map[string]interface{})
	glModel := &GeneralLedger{}

	nDate, _ := time.Parse(TIMELAYOUT, gl.StartDate)
	y := nDate.Year()
	m := nDate.Month()

	table := fmt.Sprintf("stats_p%d%02d", y, m)

	sql := fmt.Sprintf("Select SUM(cast(M_TxnAmt as decimal(18,2))) AS totaltnx from %s where C_TxnAccNo = ?", table)
	GormDB.Raw(sql, gl.AccountNum).Scan(&glModel.M_TxnAmt)

	closingBal, err := gl.FindClosingBalance()
	if err != nil {
		return nil, err
	}

	if closingBal == glModel.M_TxnAmt {
		match = true
	}

	data["match"] = match
	data["closingBal"] = closingBal
	return data, nil
}

func (gl GeneralLedger) FindClosingBalance() (float64, error) {
	// var bookBal string
	glModel := &GeneralLedger{}
	nDate, _ := time.Parse(TIMELAYOUT, gl.StartDate)
	y := nDate.Year()
	m := nDate.Month()

	table := fmt.Sprintf("gl_p%d%02d", y, m)
	sql := fmt.Sprintf("select M_GL_BkBal from %s where C_GLAccNo = ?", table)
	tx := GormDB.Raw(sql, gl.AccountNum).Scan(&glModel)

	if tx.Error != nil {
		return 0, nil
	}

	return glModel.M_GL_BkBal, nil
}

func (gl GeneralLedger) MoveForward() (string, error) {
	var tableType, sql, tables string
	glModel := &GeneralLedger{}
	date, _ := time.Parse(TIMELAYOUT, gl.StartDate)
	statementStartMonth := date.Month()
	statementStartYear := date.Year()

	date, _ = time.Parse(TIMELAYOUT, gl.EndDate)
	statementEndMonth := date.Month()
	statementEndYear := date.Year()

	if statementStartMonth == statementEndMonth && statementStartYear == statementEndYear {
		tables = fmt.Sprintf("gl_p%d%02d", statementStartYear, statementEndMonth)
		tableType = "single"
	} else {
		i := statementStartYear
		x := int(statementStartMonth)
		y := 12

		for i <= statementEndYear {
			for x <= y {
				tables = tables + fmt.Sprintf(`select C_Cus_Shortname,C_GLAccDesc,M_GL_BkBal,
				D_GL_Stdate,C_GLAccNo,D_GLDateClosed,I_GL_Ccy,C_GLCusCode,M_GL_ClrBal,C_GLBranchID
				from gl_p%d%02d UNION ALL `, i, x)

				x++
			}
			i++
			x = 1
			if i == statementEndYear {
				y = int(statementEndMonth)
			}
		}

		tables = php2go.Substr(tables, 0, len(tables)-11)
		tableType = "multi"
	}

	// openBal := 0
	if tableType == "multi" {
		sql = fmt.Sprintf(`SELECT TOP 1 C_Cus_Shortname,C_GLAccDesc,M_GL_BkBal,
		D_GL_Stdate,I_GL_Ccy,C_GLCusCode,M_GL_ClrBal,C_GLBranchID
		FROM (%s)
		AS gl WHERE C_GLAccNo= ?`, tables)
	} else {
		sql = fmt.Sprintf(`SELECT TOP 1 C_Cus_Shortname,C_GLAccDesc,M_GL_BkBal,
		D_GL_Stdate,I_GL_Ccy,C_GLCusCode,M_GL_ClrBal,C_GLBranchID
		FROM %s AS gl WHERE C_GLAccNo= ?`, tables)
	}

	// Execute query
	tx := GormDB.Raw(sql, gl.AccountNum).Scan(&glModel)

	if tx.Error != nil {
		return "", nil
	}

	return glModel.C_GLCusCode, nil

}

func (gl GeneralLedger) GetClearBalance(postDate time.Time) (GeneralLedger, error) {
	// clrBal := 0.00
	// bkBal := 0.00
	glModel := &GeneralLedger{}
	month := postDate.Month()
	year := postDate.Year()

	// fmt.Printf("wells    %s\n\n\n\n", postDate)
	sql := fmt.Sprintf("SELECT M_GL_ClrBal,M_GL_BkBal FROM gl_p%d%02d WHERE C_GLAccNo= ?", year, month)

	GormDB.Raw(sql, gl.AccountNum).Scan(&glModel)

	glModel.M_GL_ClrBal = glModel.M_GL_ClrBal * -1
	glModel.M_GL_BkBal = glModel.M_GL_BkBal * -1

	return *glModel, nil
}

func (gl GeneralLedger) GetCurrency(ccy string) (string, error) {
	ccy_name := ""

	// sql := fmt.Sprintf("SELECT C_CCY_Dsc FROM Ccy_Type WHERE I_CCY_BM_Code='%s'", ccy)

	GormDB.Raw("SELECT C_CCY_Dsc FROM Ccy_Type WHERE I_CCY_BM_Code= ?", ccy).Scan(&ccy_name)

	return ccy_name, nil
}

func (gl GeneralLedger) TrimTotalTnx(totaltnx float64) float64 {
	strTotalTnx := fmt.Sprintf("%.2f", totaltnx)

	tnx, _ := strconv.ParseFloat(strTotalTnx, 64)

	return tnx
}

// GetAccountStartDates this account check for start date and enddate is without Pre-Nuban (2011 t0 date)
func (gl GeneralLedger) GetAccountStartDates() map[string]string {
	fmt.Println("######## This is branch User Statement #######")
	result := make(map[string]string, 0)

	tablePrefix := "gl_p______"

	var firstTable, lastTable string

	wg := &sync.WaitGroup{}

	// get last table
	lastTableSql := "select TOP 1 table_name from information_schema.tables where table_name like ? order by TABLE_NAME DESC"
	wg.Add(1)
	// get last table name
	go func(string, string, string, *gorm.DB, *sync.WaitGroup) {
		defer wg.Done()
		GormDB.Raw(lastTableSql, tablePrefix).Find(&lastTable)
	}(lastTableSql, tablePrefix, lastTable, GormDB, wg)

	// get first table
	sql := "select TOP 1 table_name from information_schema.tables where table_name like ? order by TABLE_NAME ASC"

	wg.Add(1)
	// get last table name
	go func(string, string, string, *gorm.DB, *sync.WaitGroup) {
		defer wg.Done()
		GormDB.Raw(sql, tablePrefix).Find(&firstTable)
	}(sql, tablePrefix, firstTable, GormDB, wg)

	wg.Wait()

	log.Println("lastTable; ", lastTable)

	// set year and  month for last table
	lastTableYearStr := php2go.Substr(lastTable, 4, 4)
	lastTableMonthStr := php2go.Substr(lastTable, 8, 2)

	lastTableYear, _ := strconv.Atoi(lastTableYearStr)
	lastTableMonth, _ := strconv.Atoi(lastTableMonthStr)

	lastTableDate := time.Date(lastTableYear, time.Month(lastTableMonth), 1, 0, 0, 0, 0, time.UTC)
	lastTableDate = time.Date(lastTableDate.Year(), lastTableDate.Month()+1, 0, 0, 0, 0, 0, lastTableDate.Location())
	// dayOfLastDayOfthisMonth := lastDayOfthisMonth.Day()

	tableNames := make([]string, 0)
	sql = "select table_name from information_schema.tables WHERE table_name like ? ORDER BY table_name ASC"

	// get all table names with gl_p_____
	GormDB.Raw(sql, tablePrefix).Find(&tableNames)
	var rowsAffected int64
	if len(tableNames) > 0 {
		var sqlUnion string
		for _, table := range tableNames {
			sqlUnion = sqlUnion + "select D_GLOpened,C_GLAccNo,M_GL_BkBal,M_GL_ClrBal,D_lastchanged from " + table + " UNION ALL "
		}

		sqlUnion = php2go.Substr(sqlUnion, 0, len(sqlUnion)-11)

		sql = fmt.Sprintf("SELECT TOP 1 D_GLOpened,C_GLAccNo,M_GL_BkBal,M_GL_ClrBal,D_lastchanged FROM (" + sqlUnion + ") AS gl WHERE C_GLAccNo= ?")

		wg.Add(1)
		go func(string, GeneralLedger, *gorm.DB, *sync.WaitGroup) {
			defer wg.Done()
			rowsAffected = GormDB.Raw(sql, gl.AccountNum).Find(&gl).RowsAffected
		}(sql, gl, GormDB, wg)

		wg.Wait()

	}

	var firstTableDate time.Time

	accountStartDate := gl.D_GLOpened

	// account does not exist
	if rowsAffected < 1 {
		result["Year"] = ""
		result["Month"] = ""
		result["Day"] = ""
		result["ltable_y"] = fmt.Sprintf("%d", lastTableDate.Year())
		result["ltable_m"] = fmt.Sprintf("%02d", lastTableDate.Month())
		result["ltable_d"] = fmt.Sprintf("%02d", lastTableDate.Day())
		return result
	}

	if !accountStartDate.IsZero() {
		firstTableYearStr := php2go.Substr(firstTable, 4, 4)
		firstTableMonthStr := php2go.Substr(firstTable, 8, 2)
		firstTableDate, _ = time.Parse(TIMELAYOUT, fmt.Sprintf("%s-%s-01", firstTableYearStr, firstTableMonthStr))

		if accountStartDate.Before(firstTableDate) || accountStartDate.Equal(firstTableDate) {
			accountStartDate = firstTableDate
		}
	} else {
		result["Year"] = php2go.Substr(firstTable, 4, 4)
		result["Month"] = php2go.Substr(firstTable, 8, 2)
		result["Day"] = "01"
		result["ltable_y"] = fmt.Sprintf("%d", lastTableDate.Year())
		result["ltable_m"] = fmt.Sprintf("%02d", lastTableDate.Month())
		result["ltable_d"] = fmt.Sprintf("%02d", lastTableDate.Day())
		return result
	}

	fmt.Println("accountStartDate: ", accountStartDate)
	fmt.Println("firstTableDate: ", firstTableDate)

	accountType := IsAccountOldOrNew(gl.AccountNum, GormDB)
	log.Printf("Checking Account AC %s is an %s accType", gl.AccountNum, accountType)
	if accountType == "old" {
		//
		result["Year"] = ""
		result["Month"] = ""
		result["Day"] = ""
		result["ltable_y"] = fmt.Sprintf("%d", lastTableDate.Year())
		result["ltable_m"] = fmt.Sprintf("%02d", lastTableDate.Month())
		result["ltable_d"] = fmt.Sprintf("%02d", lastTableDate.Day())
		return result
	} else {
		// new accounts start from 2011 jan per the restored date
		defaultSlstmtDate, _ := time.Parse(TIMELAYOUT, "2011-01-01")
		if accountStartDate.Before(defaultSlstmtDate) {
			// accountStartDate = time.Date(defaultSlstmtDate.Year(), gl.D_GLOpened.Month(), defaultSlstmtDate.Day(), 0, 0, 0, 0, time.UTC)
			accountStartDate = defaultSlstmtDate
		}
	}

	result["Year"] = fmt.Sprintf("%d", accountStartDate.Year())
	// result["Month"] = fmt.Sprintf("%02d", accountStartDate.Month()+1)
	result["Month"] = fmt.Sprintf("%02d", accountStartDate.Month())
	result["Day"] = fmt.Sprintf("%02d", 01)
	//	result["Day"] = fmt.Sprintf("%02d", accountStartDate.Day())
	result["ltable_y"] = fmt.Sprintf("%d", lastTableDate.Year())
	result["ltable_m"] = fmt.Sprintf("%02d", lastTableDate.Month())
	result["ltable_d"] = fmt.Sprintf("%02d", lastTableDate.Day())

	return result
}
