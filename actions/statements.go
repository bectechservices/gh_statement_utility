package actions

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"ng-statement-app/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/envy"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/syyongx/php2go"
)

// var CURRENCY, CUS_NAME, ACCOUNTNUM string
var CURRENCIES, CUSTOMER_NAMES, ACCOUNT_NUMBERS map[string]string

// var BOOK_BALANCE, TOTAL_CREDIT, TOTAL_DEBIT float64
var BOOK_BALANCES, TOTAL_CREDITS, TOTAL_DEBITS map[string]float64

func ShowStatementsMainPage(c buffalo.Context) error {
	// c.Set("auth_user", getAuthenticatedUser(c))
	return c.Render(http.StatusOK, r.HTML("statements.html"))
}

func ShowAdminStatementsMainPage(c buffalo.Context) error {
	// c.Set("auth_user", getAuthenticatedUser(c))
	return c.Render(http.StatusOK, r.HTML("admin-statements.html"))
}

// Handles when customer account was opened
func HandleAdminAcountStart(c buffalo.Context) error {
	gl := &models.GeneralLedger{}
	err := c.Bind(gl)
	if err != nil {
		return err
	}

	data := gl.GetAdminAccountStartDates()

	return c.Render(200, r.JSON(data))
}

// Handles when customer account was opened
func HandleAcountStart(c buffalo.Context) error {
	gl := &models.GeneralLedger{}
	err := c.Bind(gl)
	if err != nil {
		return err
	}

	data := gl.GetAccountStartDates()

	return c.Render(200, r.JSON(data))
}

// Handle validation statement logic
func HandleStatementValidate(c buffalo.Context) error {
	gl := &models.GeneralLedger{}
	err := c.Bind(gl)
	if err != nil {
		// return err
	}

	gl.GetAccStartDate(c)
	data := gl.MarchAllGLTransactions()
	return c.Render(200, r.JSON(data))
}

// handle search statement
func HandleSearchStatement(c buffalo.Context) error {
	gl := &models.GeneralLedger{}
	err := c.Bind(gl)
	if err != nil {
		return err
	}

	currentUser := GetCurrentUser(c)

	gl.GetAccStartDate(c) // get account start date
	gl.IfCusTableEmpty(c) // check if cus table is empty

	isGLTableMissing, _ := gl.MissingGLTable()

	log.Printf("Acc %s isGLTableMissing = %s", gl.AccountNum, isGLTableMissing)

	if isGLTableMissing != "All tables exist" {
		// create date out of miss GL table
		y := php2go.Substr(isGLTableMissing, 4, 4)
		m := php2go.Substr(isGLTableMissing, 8, 2)

		intM, _ := strconv.Atoi(m)
		date := fmt.Sprintf("%s-%02d-01", y, intM)

		// create new date object out of GL opening balance date
		missingTableDate, err := time.Parse(models.TIMELAYOUT, date)
		if err != nil {
			return err
		}
		data := make(map[string]interface{})

		data["has_transaction"] = false
		data["transaction"] = fmt.Sprintf(`GL data for %s does not exist for this date range.
		Please select another date range with end date before %s`, missingTableDate.Format("2006-01-02"), missingTableDate.Format("2006-01-02"))
		data["currency"] = nil
		data["total_debits"] = 0.00
		data["total_credits"] = 0.00
		data["cleared_balance"] = 0.00

		result := make(map[string]interface{})
		result["has_transaction"] = false
		result["data"] = data

		return c.Render(200, r.JSON(result))
	}

	var cusBranch string

	openBalData, _ := gl.FindOpeningBalance(c) // find opening balance
	val := fmt.Sprintf("%f", openBalData["bookBal"])
	bkBal, _ := strconv.ParseFloat(val, 64)
	// BOOK_BALANCE = bkBal
	BOOK_BALANCES[fmt.Sprintf("%s_%s", gl.AccountNum, currentUser.ID)] = bkBal
	c.Session().Set(fmt.Sprintf("book_balance_%s", gl.AccountNum), fmt.Sprintf("%f", bkBal))
	// booKBalance := c.Session().Get(fmt.Sprintf("book_balance_%s", gl.AccountNum))
	fmt.Println("##### Book Balance: ", BOOK_BALANCES[fmt.Sprintf("%s_%s", gl.AccountNum, currentUser.ID)])
	cusBranch = fmt.Sprintf("%s", openBalData["cusBranch"])

	fmt.Println("openBalData: ", openBalData)

	if bkBal == 0 || bkBal == -0 {
		glTransData, err := gl.MatchGLTransactions()

		fmt.Println(" Test 1 glTransData: ", glTransData)

		if err != nil {
			return err
		}
		customerName := openBalData["cusName"].(string)

		// if customer name is empty
		if len(customerName) == 0 {
			log.Printf("Before Date: %s", gl.StartDate)
			tempDate, _ := time.Parse(time.DateOnly, gl.StartDate)
			tempDate = tempDate.AddDate(0, 1, 0)
			tempGL := models.GeneralLedger{
				StartDate:  tempDate.Format(time.DateOnly),
				EndDate:    gl.EndDate,
				AccountNum: gl.AccountNum,
			}
			log.Printf("tempGL: %v", tempGL)

			tempOpenBal, _ := tempGL.FindOpeningBalance(c)
			log.Printf("tempOpenBal: %v", tempOpenBal)

			if customerName = tempOpenBal["cusName"].(string); len(customerName) != 0 {
				openBalData["cusName"] = customerName
				log.Printf("New openBalData: %v", openBalData)
			}
		}

		if glTransData["match"] == false {
			var transaction string

			tr, err := gl.GetFirstTransaction()

			if err != nil {
				return err
			}

			if tr != "" {
				transaction = fmt.Sprintf(`
				The start date for the search should start from %s or the beginning of the following month, 
				because there no transactions on the account prior to that.`, php2go.Explode("T", tr)[0])
			} else {
				transaction = "There are no transactions for the account specified."
			}

			data := make(map[string]interface{})

			data["has_transaction"] = false
			data["customer_name"] = openBalData["cusName"]
			data["currency"] = openBalData["Ccy"]
			data["transaction"] = transaction
			data["total_debits"] = "0.00"
			data["total_credits"] = "0.00"
			data["cleared_balance"] = "0.00"

			result := make(map[string]interface{})

			d := make(map[string]interface{})
			d["has_transaction"] = false
			result["has_transaction"] = false
			result["data"] = data

			return c.Render(200, r.JSON(result))
		} else if glTransData["match"] == true && openBalData["openBalanceDate"] == "0001-01-01" {
			date, _ := time.Parse(time.DateOnly, gl.StartDate)
			openBalData["openBalanceDate"] = models.GetLastDateOfPreviousMonth(date.Year(), int(date.Month())).Format(time.DateOnly)
			// var transaction string

			// tr, err := gl.GetFirstTransaction()

			// if err != nil {
			// 	return err
			// }

			// if tr != "" {
			// 	transaction = fmt.Sprintf(`
			// 	The start date for the search should start from %s or the beginning of the following month,
			// 	because there no transactions on the account prior to that.`, php2go.Explode("T", tr)[0])
			// } else {
			// 	transaction = "There are no transactions for the account specified."
			// }

			// data := make(map[string]interface{})

			// data["has_transaction"] = false
			// data["customer_name"] = openBalData["cusName"]
			// data["currency"] = openBalData["Ccy"]
			// data["transaction"] = transaction
			// data["total_debits"] = "0.00"
			// data["total_credits"] = "0.00"
			// data["cleared_balance"] = "0.00"

			// result := make(map[string]interface{})

			// d := make(map[string]interface{})
			// d["has_transaction"] = false
			// result["has_transaction"] = false
			// result["data"] = data

			// return c.Render(200, r.JSON(result))
		}
	}

	// check if openBalanceDate is greater than glStartDate
	// BBF date will be last month
	glStartDate, _ := time.Parse(time.DateOnly, gl.StartDate)
	openBalanceDate, _ := time.Parse(time.DateOnly, openBalData["openBalanceDate"].(string))

	if openBalanceDate.After(glStartDate) || openBalanceDate.Equal(glStartDate) {
		// openBalData["openBalanceDate"] = models.GetLastDateOfPreviousMonth(glStartDate.Year(), int(glStartDate.Month())).Format(time.DateOnly)
		openBalData["openBalanceDate"] = models.GetLastDateOfPreviousMonth(openBalanceDate.Year(), int(openBalanceDate.Month())).Format(time.DateOnly)
	}

	// get customer branch
	branchName, err := models.GetBranch(cusBranch)
	if err != nil {
		return err
	}

	// c.Session.Set("branch_name", branchName)
	// c.Session.Set("openbal", openBalData["bookBal"])
	// c.Session.Set("org_openbal", openBalData["bookBal"])
	// c.Session.Set("open_clrbal", openBalData["clrBal"])

	c.Session().Set(fmt.Sprintf("branch_name_%s", gl.AccountNum), branchName)
	c.Session().Set(fmt.Sprintf("openbal_%s", gl.AccountNum), fmt.Sprintf("%f", openBalData["bookBal"]))
	c.Session().Set(fmt.Sprintf("org_openbal_%s", gl.AccountNum), fmt.Sprintf("%f", openBalData["bookBal"]))
	c.Session().Set(fmt.Sprintf("open_clrbal_%s", gl.AccountNum), fmt.Sprintf("%f", openBalData["clrBal"]))

	fmt.Println("@@@@@@@@ CLR_BAL: ", openBalData["clrBal"])

	// fmt.Println("openbal: ", c.Session().Get("openbal"))
	// er := c.Session().Save()
	// fmt.Println("SESSION ERR: ", er)

	if openBalData["Cus_id"] != "" {
		cusId := fmt.Sprintf("%s", openBalData["Cus_id"])
		cusReg := &models.CusReg{}
		cReg, _ := cusReg.GetAddress(cusId, *gl)

		// c.Session.Set("address1", cReg.C_Addr1)
		// c.Session.Set("address2", cReg.C_Addr2)
		// c.Session.Set("address3", cReg.C_Addr3)

		c.Session().Set(fmt.Sprintf("address1_%s", gl.AccountNum), cReg.C_Addr1)
		c.Session().Set(fmt.Sprintf("address2_%s", gl.AccountNum), cReg.C_Addr2)
		c.Session().Set(fmt.Sprintf("address3_%s", gl.AccountNum), cReg.C_Addr3)

	}

	str := fmt.Sprintf("%f", openBalData["bookBal"])

	openBal, _ := strconv.ParseFloat(str, 64)
	// c.Session.Set("cus_name", openBalData["cusName"])
	c.Session().Set(fmt.Sprintf("cus_name_%s", gl.AccountNum), fmt.Sprintf("%s", openBalData["cusName"]))
	// CUS_NAME = fmt.Sprintf("%s", openBalData["cusName"])
	CUSTOMER_NAMES[fmt.Sprintf("%s_%s", gl.AccountNum, currentUser.ID)] = fmt.Sprintf("%s", openBalData["cusName"])
	fmt.Println("~~~~Customer: ", CUSTOMER_NAMES[fmt.Sprintf("%s_%s", gl.AccountNum, currentUser.ID)])
	balBroughtForward := make(map[string]interface{})

	balBroughtForward["date"] = openBalData["openBalanceDate"]
	balBroughtForward["title"] = "Balance Brought Forward"
	balBroughtForward["open_balance"] = php2go.NumberFormat(php2go.Round(openBal, 2), 2, ".", ",")

	st := &models.Statement{
		AccountNum: gl.AccountNum,
		StartDate:  gl.StartDate,
		EndDate:    gl.EndDate,
	}
	sql := st.SqlStatement(c)

	fmt.Println("Search SQL Statement: ", sql)

	// get statement per month
	startDate, _ := time.Parse(models.TIMELAYOUT, gl.StartDate)
	endDate, _ := time.Parse(models.TIMELAYOUT, gl.EndDate)

	// get months between start and end date
	diff := endDate.Sub(startDate)
	months := diff.Hours() / 24 / 30

	fmt.Println("~~~~~~~~~~~~~ Testing listing Months: ", math.Round(months))
	i := 1

	count := 1

	var balanceSummary []map[string]interface{}

	var narration string
	var crAmount, dbAmount string
	var cr, dr int64
	var totalCredit, totalDebit float64

	data := make(map[string]interface{})

	var statementRows []models.Statement
	models.GormDB.Raw(sql).Scan(&statementRows)

	// Check if start and end date have same month and year
	if startDate.Month() == endDate.Month() && startDate.Year() == endDate.Year() {
		months = 1
	}

	if statementRows != nil {
		for i <= int(months) {
			// first day of current month
			sdate := fmt.Sprintf("%d-%02d-01", startDate.Year(), startDate.Month())
			firstDayofTheMonth, _ := time.Parse(models.TIMELAYOUT, sdate)

			// last day of the current month
			endOfThisMonth := time.Date(startDate.Year(), startDate.Month()+1, 0, 0, 0, 0, 0, startDate.Location())
			if i == int(months) {
				endOfThisMonth = endDate
			}

			st := &models.Statement{
				AccountNum: gl.AccountNum,
				StartDate:  sdate,
				EndDate:    endOfThisMonth.Format(models.TIMELAYOUT),
			}

			sql := st.SqlStatement(c)

			stModel := &models.Statement{}
			var statementRows []models.Statement
			models.GormDB.Raw(sql).Scan(&statementRows)

			// no transaction for that month
			if len(statementRows) < 1 {
				closingBal, _ := gl.FindOpenBalance(firstDayofTheMonth)
				openBal = closingBal * -1

				transType := st.CheckAmount(openBal)

				if transType == "Deposit" {

					crAmount = php2go.NumberFormat(php2go.Round(php2go.Abs(openBal), 2), 2, ".", ",")
					dbAmount = "-"
					// openBal = openBal
					// totalCredit = totalCredit + openBal
					cr++
				} else {
					crAmount = "-"
					dbAmount = php2go.NumberFormat(php2go.Round(php2go.Abs(openBal), 2), 2, ".", ",")
					// openBal = openBal
					// totalDebit = totalDebit + openBal
					dr++
				}

				// mapData := make(map[string]interface{})
				// mapData["entry_date"] = endOfThisMonth.Format(models.TIMELAYOUT)
				// mapData["value_date"] = endOfThisMonth.Format(models.TIMELAYOUT)
				// mapData["narration"] = strings.ToUpper(fmt.Sprintf("LAST DAY OF THE MONTH (%s %d)", startDate.Month().String(), startDate.Year()))
				// mapData["debit_amt"] = dbAmount
				// mapData["credit_amt"] = crAmount
				// mapData["opening_balance"] = php2go.NumberFormat(php2go.Round(php2go.Abs(openBal), 2), 2, ".", ",")
				//
				// balanceSummary = append(balanceSummary, mapData)

			} else {
				for _, row := range statementRows {
					narration = fmt.Sprintf("%s %s %s", row.C_TxnNar1, row.C_TxnNar2, row.C_TxnNar3)
					amt := row.M_TxnAmt * -1

					stModel.D_TxnPostDt = row.D_TxnPostDt
					stModel.I_Txn_Ccy = row.I_Txn_Ccy

					transType := st.CheckAmount(amt)

					if transType == "Deposit" {

						// floatAmt, _ := strconv.ParseFloat(amount, 64)
						crAmount = php2go.NumberFormat(php2go.Round(php2go.Abs(amt), 2), 2, ".", ",")
						dbAmount = "-"
						openBal = openBal + amt
						totalCredit = totalCredit + amt
						cr++
					} else {
						// floatAmt, _ := strconv.ParseFloat(amount, 64)
						dbAmount = php2go.NumberFormat(php2go.Round(php2go.Abs(amt), 2), 2, ".", ",")
						crAmount = "-"
						openBal = openBal + amt
						totalDebit = totalDebit + amt
						dr++
					}

					mapData := make(map[string]interface{})
					mapData["entry_date"] = row.D_TxnPostDt.Format(models.TIMELAYOUT)
					mapData["value_date"] = row.D_TxnValueDt.Format(models.TIMELAYOUT)
					mapData["narration"] = narration
					mapData["debit_amt"] = dbAmount
					mapData["credit_amt"] = crAmount
					mapData["opening_balance"] = php2go.NumberFormat(php2go.Round(php2go.Abs(openBal), 2), 2, ".", ",")

					balanceSummary = append(balanceSummary, mapData)
				}

				clrBal, err := gl.GetClearBalance(stModel.D_TxnPostDt)
				if err != nil {
					return err
				}

				currency, err := gl.GetCurrency(stModel.I_Txn_Ccy)
				if err != nil {
					return err
				}

				CURRENCIES[fmt.Sprintf("%s_%s", st.AccountNum, currentUser.ID)] = currency
				c.Session().Set(fmt.Sprintf("currency_%s", st.AccountNum), currency)
				fmt.Println("Current User: ", currentUser.ID)
				fmt.Println("@@@@@Line 353")

				fmt.Println("@@@@@Line 359")
				c.Session().Set(fmt.Sprintf("currency_%s", gl.AccountNum), currency)
				c.Session().Set(fmt.Sprintf("close_clearedbal_%s", gl.AccountNum), fmt.Sprintf("%f", clrBal.M_GL_ClrBal))
				c.Session().Set(fmt.Sprintf("close_bkbal_%s", gl.AccountNum), fmt.Sprintf("%f", clrBal.M_GL_BkBal))
				c.Session().Set(fmt.Sprintf("totaldebit_%s", gl.AccountNum), fmt.Sprintf("%f", totalDebit))
				c.Session().Set(fmt.Sprintf("totalcredit_%s", gl.AccountNum), fmt.Sprintf("%f", totalCredit))

				// fmt.Println("Currency: ", c.Session().Get(""))
				fmt.Println("@@@@@Line 367")
				data["has_transaction"] = true
				data["customer_name"] = openBalData["cusName"]
				data["currency"] = currency
				data["transaction"] = balanceSummary
				data["total_debits"] = php2go.NumberFormat(php2go.Round(php2go.Abs(totalDebit), 2), 2, ".", ",")
				data["total_credits"] = php2go.NumberFormat(php2go.Round(php2go.Abs(totalCredit), 2), 2, ".", ",")
				data["cleared_balance"] = php2go.NumberFormat(php2go.Round(php2go.Abs(clrBal.M_GL_ClrBal), 2), 2, ".", ",")

			}

			i++
			count++
			startDate = time.Date(startDate.Year(), startDate.Month()+1, 1, 0, 0, 0, 0, startDate.Location())

		}

		TOTAL_CREDITS[fmt.Sprintf("%s_%s", st.AccountNum, currentUser.ID)] = totalCredit
		TOTAL_DEBITS[fmt.Sprintf("%s_%s", st.AccountNum, currentUser.ID)] = totalDebit
		c.Session().Set(fmt.Sprintf("total_credit_%s", st.AccountNum), fmt.Sprintf("%f", totalCredit))
		c.Session().Set(fmt.Sprintf("total_debit_%s", st.AccountNum), fmt.Sprintf("%f", totalDebit))

		result := make(map[string]interface{})
		result["has_transaction"] = true
		result["data"] = data
		result["bal_brought_forward"] = balBroughtForward

		return c.Render(200, r.JSON(result))
	}

	data = make(map[string]interface{})
	data["has_transaction"] = false
	data["customer_name"] = openBalData["cusName"]
	data["currency"] = "N/A"
	data["transaction"] = "Account/Transactions Does Not exist for this Date Range"
	data["total_debits"] = 0.00
	data["total_credits"] = 0.00
	data["cleared_balance"] = 0.00

	result := make(map[string]interface{})
	result["has_transaction"] = false
	result["data"] = data

	return c.Render(200, r.JSON(result))
}

// handle pdf  statement geberation
func HandlePDFStatementGeneration(c buffalo.Context) error {
	_err := models.DeleteOldPrintedFiles()
	if _err != nil {
		fmt.Println("&&&&&&&&&&&&&&&& Error: ", _err.Error())
	}
	// bind values to models
	pdf := &models.PDFGenerator{}
	// statement := &models.Statement{}
	err := c.Bind(pdf)
	if err != nil {
		return err
	}

	// get auth user
	user := GetCurrentUser(c)
	hashKey := fmt.Sprintf("%s_%s", pdf.AccountNum, user.ID)

	gl := models.GeneralLedger{
		AccountNum: pdf.AccountNum,
		StartDate:  pdf.StartDate,
		EndDate:    pdf.EndDate,
	}

	openBalData, _ := gl.FindOpeningBalance(c) // find opening balance

	fmt.Println("openBalData: ", openBalData)

	if openBalData["cusName"] == "" {
		startDate, _ := time.Parse(time.DateOnly, pdf.StartDate)

		nextMonthDate := startDate.AddDate(0, 1, 0)

		tempGL := models.GeneralLedger{
			StartDate:  nextMonthDate.Format(time.DateOnly),
			EndDate:    gl.EndDate,
			AccountNum: gl.AccountNum,
		}

		tempOpenBal, _ := tempGL.FindOpeningBalance(c)

		if customerName := tempOpenBal["cusName"].(string); len(customerName) != 0 {
			openBalData["cusName"] = customerName
		}
	}

	pdf.New(c)
	headerPath := pdf.WriteHeaderHtml(CURRENCIES[hashKey], user.ID)
	footerPath := pdf.WriteFooterHtml(user.ABNumber, user.Branch.Name, user.ID)
	fnData, pdfPath := pdf.CreateDocument(openBalData, user, headerPath, footerPath)

	// get temp statement file
	file, err := os.Open(pdfPath)
	if err != nil {
		return err
	}

	fmt.Println("@@@@@@@@@@@@@@@@PDF Path", pdfPath)
	TOTAL_CREDITS[hashKey] = fnData["totalCredit"]
	TOTAL_DEBITS[hashKey] = fnData["totalDebit"]

	useStampifyAPI, _ := strconv.ParseBool(envy.Get("USE_STAMPIFYPDF_API", "false"))
	// userPosition := "Dummy Role"

	if useStampifyAPI {
		// make rest call to stampifyPdf web api
		responseChan, errChan := make(chan []byte, 1), make(chan error, 1)

		go makeRestCallToStampifyPdfAPI(&user, "", pdf.AccountNum, file, responseChan, errChan)
		response := <-responseChan
		err = <-errChan

		if err != nil {
			return err
		}
		// return response based on the rest call
		// Display PDF For Statement Generated

		return c.Render(200, r.Func("application/pdf", func(w io.Writer, d render.Data) error {
			w.Write(response)
			return nil
		}))
	}

	return c.Render(200, r.Func("application/pdf", func(w io.Writer, d render.Data) error {
		b, _ := ioutil.ReadFile(pdfPath)
		w.Write(b)
		return nil
	}))
}

func HandleExcelStatement(c buffalo.Context) error {
	models.DeleteOldPrintedFiles()
	excel := &models.Excel{}
	// stModel := &models.Statement{}
	err := c.Bind(excel)

	if err != nil {
		return errors.WithStack(errors.New("Failed to bind excel data"))
	}

	gl := models.GeneralLedger{
		AccountNum: excel.AccountNum,
		StartDate:  excel.StartDate,
		EndDate:    excel.EndDate,
	}

	openBalData, _ := gl.FindOpeningBalance(c) // find opening balance
	currency, _ := gl.GetCurrency(fmt.Sprintf("%s", openBalData["Ccy"]))
	// get auth user
	user := GetCurrentUser(c)
	hashKey := fmt.Sprintf("%s_%s", excel.AccountNum, user.ID)

	if openBalData["cusName"] == "" {
		startDate, _ := time.Parse(time.DateOnly, excel.StartDate)

		nextMonthDate := startDate.AddDate(0, 1, 0)

		tempGL := models.GeneralLedger{
			StartDate:  nextMonthDate.Format(time.DateOnly),
			EndDate:    gl.EndDate,
			AccountNum: gl.AccountNum,
		}

		tempOpenBal, _ := tempGL.FindOpeningBalance(c)

		if customerName := tempOpenBal["cusName"].(string); len(customerName) != 0 {
			openBalData["cusName"] = customerName
		}
	}
	excel.New(c)
	// currency := fmt.Sprintf("%s", openBalData["Ccy"])    // CURRENCIES[hashKey]
	cusName := fmt.Sprintf("%s", openBalData["cusName"]) // CUSTOMER_NAMES[hashKey]
	bookBalance := openBalData["bookBal"].(float64)      // BOOK_BALANCES[hashKey]
	totalCredit := TOTAL_CREDITS[hashKey]
	totalDebit := TOTAL_DEBITS[hashKey]
	filePath, err := excel.ExportWithQuery(currency, cusName, bookBalance, totalCredit, totalDebit, user)

	if err != nil {
		return err
	}

	f, err := os.Open(filePath)
	if err != nil {

		return err
	}
	fmt.Println("~~~~~ Excel File Name: ", filePath)
	return c.Render(200, r.Download(c, fmt.Sprintf("%s.xls", excel.AccountNum), f))
}

func GetCurrentUser(c buffalo.Context) models.User {
	// get auth user
	user := models.User{}
	if uid := c.Session().Get("auth_id"); uid != nil {
		switch id := uid.(type) {
		case uuid.UUID:
			user = models.GetUserByID(id, models.GormDB)
			user.Branch = models.GetBranchByID(user.BranchID, models.GormDB)
		}
	}

	return user
}

func makeRestCallToStampifyPdfAPI(user *models.User, position, fileName string, file *os.File, responseChan chan []byte, errChan chan error) {
	defer file.Close()
	// Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer

	writer := multipart.NewWriter(&requestBody)
	// Create a form file field for the file
	fileFormField, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		responseChan <- nil
		errChan <- errors.New("failed to create form file")
		return
	}

	// Copy the file content to the form file field
	_, err = io.Copy(fileFormField, file)
	if err != nil {
		log.Println(err)
		responseChan <- nil
		errChan <- errors.New("failed to copy file content")
		return
	}

	// writer.WriteField("user", fmt.Sprintf("%s %s", user.FirstName, user.LastName))
	// writer.WriteField("role", position)
	writer.WriteField("file_name", fileName)
	writer.WriteField("file_type", "PDF")
	writer.WriteField("user_id", user.ID.String())

	err = writer.Close()
	if err != nil {
		responseChan <- nil
		errChan <- errors.New("failed to close writer")
		return
	}

	// Create a new request using http with multipart form data
	stampifyPdfAPIHost := envy.Get("STAMPIFYPDF_API_HOST", "")
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/generate-stamps", stampifyPdfAPIHost), &requestBody)
	if err != nil {
		log.Println(err)
		responseChan <- nil
		errChan <- errors.New("failed to create request")
		return
	}
	log.Println("Request: ", req.URL)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	// Make the request
	client := &http.Client{}
	log.Println("making rest call to stampifypdf api....")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		responseChan <- nil
		errChan <- errors.New("failed to generate stamp on file")
		return
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		log.Println(bodyString)
		responseChan <- nil
		errChan <- errors.New("failed to generate stamp on file")
		return
	}

	responseContents, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	if err != nil || responseContents == nil {
		log.Println(err)
		responseChan <- nil
		errChan <- errors.New("failed to generate stamp on file")
		return
	}

	responseChan <- responseContents
	errChan <- nil
}
