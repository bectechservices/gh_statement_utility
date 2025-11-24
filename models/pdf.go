package models

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gofrs/uuid"
	"github.com/syyongx/php2go"
)

type PDFGenerator struct {
	AccountNum string `json:"account_no" form:"account_no"`
	StartDate  string `json:"start_date" form:"start_date"`
	EndDate    string `json:"end_date" form:"end_date"`
}

// var newBalance, _totalCredits, _totalDebits, noTransactionBBF float64
// var hasTransaction bool

// var allCredits

// init
func (pdfGen *PDFGenerator) New(c buffalo.Context) {
	buffaloCtx = c
}

// set pdf attributes
func (pdfGen *PDFGenerator) SetPdfAttributes(pdf wkhtmltopdf.PDFGenerator) wkhtmltopdf.PDFGenerator {
	pdf.Title.Set("SCB - Print Statement")
	pdf.TOC.HeaderFontName.Set("helvetica")
	pdf.TOC.HeaderFontSize.Set(8)
	pdf.TOC.FooterFontName.Set("helvetica")
	// pdf.TOC.FooterFontSize.Set(3)

	pdf.MarginLeft.Set(8)
	pdf.MarginTop.Set(45)
	pdf.MarginRight.Set(8)

	pdf.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdf.Dpi.Set(300)

	return pdf
}

func (pdfGen *PDFGenerator) CreateDocument(openBalData map[string]interface{}, user User, headerPath string, footerPath string) (map[string]float64, string) {
	pdfg, _ := wkhtmltopdf.NewPDFGenerator()
	pdf := pdfGen.SetPdfAttributes(*pdfg)

	var pdfPages []string
	var CUSTOMER_NAME string
	var NUM_PAGES int64

	var newBalance, _totalCredits, _totalDebits, noTransactionBBF float64
	var hasTransaction bool

	resultData := make(map[string]float64)
	balSummary := make(map[string]string)
	_totalCredits, _totalDebits = 0, 0

	var openBal float64
	bookBalance := openBalData["bookBal"].(float64)
	newBalance = bookBalance
	totaldebit := 0.0
	totalcredit := 0.0

	AV_Close_Bal := ""

	// get statement per month
	startDate, _ := time.Parse(TIMELAYOUT, pdfGen.StartDate)
	endDate, _ := time.Parse(TIMELAYOUT, pdfGen.EndDate)

	// get months between start and end date
	diff := endDate.Sub(startDate)
	months := diff.Hours() / 24 / 30
	months = math.Round(months)
	i := 1
	k := 1

	count := 1
	// Check if start and end date have same month and year
	if startDate.Month() == endDate.Month() && startDate.Year() == endDate.Year() {
		months = 1
	}

	for i <= int(months) {
		OPENING_BALANCE := "OPENING BALANCE"
		CLOSING_BALANCE := "CLOSING BALANCE"
		TOTALDEBITTEXT := "TOTAL DEBIT"
		TOTALCREDITTEXT := "TOTAL CREDIT"
		AVAILABLE := "AVAILABLE"
		BOOK := "BOOK"

		transTable := fmt.Sprintf("stats_p%d%02d", startDate.Year(), startDate.Month())

		// init statement model
		var openClearingBalance, closeClearedBalance, closeAvailableBalance float64
		var closeClearedBal, closedAvailBal float64
		// buffaloCtx.Session().Set(fmt.Sprintf()("loop", 1)
		buffaloCtx.Session().Set(fmt.Sprintf("loop_%s", pdfGen.AccountNum), "1")

		// closed clear balance
		closeClearedBal = pdfGen.GetClosedClearedBal(transTable)

		noTransactionBBF = openBal
		openBal = newBalance
		fmt.Printf("========== Account: %s newBalance: %f openBal: %f noTransactionBBF: %f\n", pdfGen.AccountNum, newBalance, openBal, noTransactionBBF)

		// closed clear balance result
		res := (openBal * -1) + closeClearedBal

		res = res * -1
		// buffaloCtx.Session().Set(fmt.Sprintf("close_clearedbal_%s", pdfGen.AccountNum), fmt.Sprintf("%f", php2go.Round(res, 2)))
		closeClearedBalance = res
		closedAvailBal = pdfGen.GetClosedAvailableBal(transTable)

		res = (openBal * -1) + closedAvailBal

		res = res * -1
		// buffaloCtx.Session().Set(fmt.Sprintf("close_avail_bal_%s", pdfGen.AccountNum), fmt.Sprintf("%f", res))
		closeAvailableBalance = res
		// customer info
		cusRegModel := CusReg{}
		cusReg, _ := cusRegModel.GetCustomerInfo(buffaloCtx, pdfGen.AccountNum)

		buffaloCtx.Session().Set(fmt.Sprintf("address1_%s", pdfGen.AccountNum), cusReg.C_Addr1)
		buffaloCtx.Session().Set(fmt.Sprintf("address2_%s", pdfGen.AccountNum), cusReg.C_Addr2)
		buffaloCtx.Session().Set(fmt.Sprintf("address3_%s", pdfGen.AccountNum), cusReg.C_Addr3)
		buffaloCtx.Session().Set(fmt.Sprintf("cus_name_%s", pdfGen.AccountNum), cusReg.C_Cus_ShortName)

		if openBal == 0 {
			openClearingBalance = 0
			// buffaloCtx.Session().Set(fmt.Sprintf("openbal_%s", pdfGen.AccountNum), "0")
		} else {
			// buffaloCtx.Session().Set(fmt.Sprintf()("openbal", openBalance)
			increment := k
			// date, _ := time.Parse(TIMELAYOUT, fmt.Sprintf("%s", buffaloCtx.Session().Get(fmt.Sprintf("clr_bal_Start_date_%s", pdfGen.AccountNum))))
			date, _ := time.Parse(TIMELAYOUT, pdfGen.StartDate)
			date = date.AddDate(0, increment-1, 0)
			month := date.Month()
			year := date.Year()

			table := fmt.Sprintf("gl_p%d%02d", year, month)

			sql := fmt.Sprintf("SELECT M_GL_ClrBal,D_lastchanged FROM %s WHERE C_GLAccNo= '%s'", table, pdfGen.AccountNum)

			glModel := GeneralLedger{}
			GormDB.Raw(sql).Scan(&glModel)

			if !glModel.D_lastchanged.IsZero() {
				openClearingBalance = glModel.M_GL_ClrBal * -1
				// buffaloCtx.Session().Set(fmt.Sprintf("open_clrbal_%s", pdfGen.AccountNum), fmt.Sprintf("%v", (glModel.M_GL_ClrBal*-1)))
			}
		}

		sdate := fmt.Sprintf("%d-%02d-01", startDate.Year(), startDate.Month())
		firstDayofTheMonth, _ := time.Parse(TIMELAYOUT, sdate)

		// last day of the current month
		endOfThisMonth := time.Date(startDate.Year(), startDate.Month()+1, 0, 0, 0, 0, 0, startDate.Location())

		Total_Debits := 0.0
		Total_Credits := 0.0
		total_debit := 0.0
		total_credit := 0.0

		x := 1
		y := 0
		z := x

		var cus_name, BK_Open_Bal, AV_Open_Bal, BK_Close_Bal string
		var Enter_date1, Enter_date2, Enter_date3 string
		var Enter_date4, Enter_date5, Enter_date6, Enter_date7, Enter_date8 string
		var Enter_date9, Enter_date10, Enter_date11, Enter_date12 string

		var value_date1, value_date2, value_date3 string
		var value_date4, value_date5, value_date6, value_date7 string
		var value_date8, value_date9, value_date10, value_date11, value_date12 string
		var Debit1, Debit2, Debit3, Debit4, Debit5, Debit6 string
		var Debit7, Debit8, Debit9, Debit10, Debit11, Debit12 string
		var Credit1, Credit2, Credit3, Credit4 string
		var Credit5, Credit6, Credit7, Credit8, Credit9, Credit10, Credit11 string
		var Credit12, balance1, balance2, balance3, balance4, balance5, balance6, balance7, balance8 string
		var balance9, balance10, balance11, balance12 string
		var Description1, Description2 string
		var Description3, Description4, Description5, Description6, Description7, Description8 string
		var Description9, Description10, Description11, Description12 string

		first := true

		stModel := &Statement{
			AccountNum: pdfGen.AccountNum,
			StartDate:  sdate,
			EndDate:    endOfThisMonth.Format(TIMELAYOUT),
		}
		sql := stModel.SqlStatement(buffaloCtx)

		// fmt.Printf("STMTSQL1: %s\n\n\n", sql)
		var data []map[string]interface{}

		var statmentRows []Statement
		GormDB.Raw(sql).Scan(&statmentRows)

		// no transaction for that month
		if len(statmentRows) < 1 {
			fmt.Println("firstDayofTheMonth2: ", firstDayofTheMonth)
			hasTransaction = false
			gl := &GeneralLedger{
				AccountNum: pdfGen.AccountNum,
				StartDate:  pdfGen.StartDate,
				EndDate:    pdfGen.EndDate,
			}
			tranAmt, _ := gl.FindOpenBalance(firstDayofTheMonth)

			dd := make(map[string]interface{})
			dd["I_Identity"] = ""
			dd["D_TxnPostDt"] = ""  // endOfThisMonth.Format(TIMELAYOUT)
			dd["D_TxnValueDt"] = "" // endOfThisMonth.Format(TIMELAYOUT)
			dd["C_TxnNar1"] = ""    // strings.ToUpper(fmt.Sprintf("LAST DAY OF THE MONTH (%s %d)", startDate.Month().String(), startDate.Year()))
			dd["C_TxnNar2"] = ""
			dd["C_TxnNar3"] = ""
			dd["M_TxnAmt"] = tranAmt
			dd["C_TxnAccNo"] = pdfGen.AccountNum
			dd["I_Txn_CusCode"] = ""
			dd["I_Txn_Ccy"] = ""
			dd["IsNoTransMonth"] = true

			data = append(data, dd)

		} else {
			hasTransaction = true
			fmt.Println(hasTransaction)
			for _, row := range statmentRows {
				// fmt.Println(row)
				dd := make(map[string]interface{})

				dd["I_Identity"] = row.I_Identity
				dd["D_TxnPostDt"] = row.D_TxnPostDt.Format(TIMELAYOUT)
				dd["D_TxnValueDt"] = row.D_TxnValueDt.Format(TIMELAYOUT)
				dd["C_TxnNar1"] = row.C_TxnNar1
				dd["C_TxnNar2"] = row.C_TxnNar2
				dd["C_TxnNar3"] = row.C_TxnNar3
				dd["M_TxnAmt"] = row.M_TxnAmt
				dd["C_TxnAccNo"] = row.C_TxnAccNo
				dd["I_Txn_CusCode"] = row.I_Txn_CusCode
				dd["I_Txn_Ccy"] = row.I_Txn_Ccy
				dd["IsNoTransMonth"] = false
				data = append(data, dd)

			}
		}

		v := float64(len(data)) / 12
		loop := php2go.Ceil(v)

		AccountNumber := pdfGen.AccountNum
		StartDate := ""
		Enddate := ""

		dr, cr, page := 1, 1, 1

		var D_postdate string

		// fmt.Printf("Date: %s Data Count: %d\n\n\n", startDate.Format(TIMELAYOUT), len(data))
		if len(data) > 0 {
			// fmt.Printf("Date: %s Data Count: %d\n\n\n", startDate.Format(TIMELAYOUT), len(data))
			for _, row := range data {
				// check if row has no transactions in the month
				// noTransMonth, ok := row["IsNoTransMonth"].(bool)
				// if !ok {
				//	noTransMonth = false
				// }
				//
				postDate := fmt.Sprintf("%s", row["D_TxnPostDt"])
				valueDate := fmt.Sprintf("%s", row["D_TxnValueDt"])

				narration := fmt.Sprintf("%s %s %s", row["C_TxnNar1"], row["C_TxnNar2"], row["C_TxnNar3"])

				amount, _ := strconv.ParseFloat(fmt.Sprintf("%f", row["M_TxnAmt"]), 64)

				amount = amount * -1

				if amount < 0 {
					Total_Debits = Total_Debits + 1
					total_debit = total_debit + amount
				} else {
					Total_Credits = Total_Credits + 1
					total_credit = total_credit + amount
				}

				stModel := &Statement{
					AccountNum: pdfGen.AccountNum,
					StartDate:  pdfGen.StartDate,
					EndDate:    pdfGen.EndDate,
				}

				// get transaction type
				transType := stModel.CheckAmount(amount)

				var cr_amount, dr_amount string

				if hasTransaction {
					if transType == "Deposit" {
						cr_amount = php2go.NumberFormat(php2go.Round(php2go.Abs(amount), 2), 2, ".", ",")
						dr_amount = "-"
						openBal = openBal + amount
						totalcredit = totalcredit + amount
						_totalCredits += amount
						cr = cr + 1

					} else {
						dr_amount = php2go.NumberFormat(php2go.Round(php2go.Abs(amount), 2), 2, ".", ",")
						cr_amount = "-"
						openBal = openBal + amount
						totaldebit = totaldebit + amount
						_totalDebits += amount
						dr = dr + 1

					}
				} else {
					continue
					if transType == "Deposit" {
						dr_amount = "-"
						cr_amount = php2go.NumberFormat(php2go.Round(php2go.Abs(amount), 2), 2, ".", ",")
						openBal = amount
						totalcredit = totalcredit + amount
						_totalCredits += amount
						cr = cr + 1

					} else {
						dr_amount = php2go.NumberFormat(php2go.Round(php2go.Abs(amount), 2), 2, ".", ",")
						cr_amount = "-"
						openBal = amount
						totaldebit = totaldebit + amount
						_totalDebits += amount
						dr = dr + 1

					}
				}

				runningbal := php2go.NumberFormat(php2go.Round(openBal, 2), 2, ".", ",")
				newBalance = openBal

				fmt.Printf("Date: %s newBalance: %f\n\n\n", startDate.Format(TIMELAYOUT), newBalance)

				D_postdate = postDate

				if x == 1 {
					Enter_date1 = postDate
					value_date1 = valueDate
					Debit1 = dr_amount
					Credit1 = cr_amount
					balance1 = runningbal
					Description1 = narration
				}

				if x == 2 {
					Enter_date2 = postDate
					value_date2 = valueDate
					Debit2 = dr_amount
					Credit2 = cr_amount
					balance2 = runningbal
					Description2 = narration
				}

				if x == 3 {
					Enter_date3 = postDate
					value_date3 = valueDate
					Debit3 = dr_amount
					Credit3 = cr_amount
					balance3 = runningbal
					Description3 = narration
				}

				if x == 4 {
					Enter_date4 = postDate
					value_date4 = valueDate
					Debit4 = dr_amount
					Credit4 = cr_amount
					balance4 = runningbal
					Description4 = narration
				}

				if x == 5 {
					Enter_date5 = postDate
					value_date5 = valueDate
					Debit5 = dr_amount
					Credit5 = cr_amount
					balance5 = runningbal
					Description5 = narration
				}

				if x == 6 {
					Enter_date6 = postDate
					value_date6 = valueDate
					Debit6 = dr_amount
					Credit6 = cr_amount
					balance6 = runningbal
					Description6 = narration
				}

				if x == 7 {
					Enter_date7 = postDate
					value_date7 = valueDate
					Debit7 = dr_amount
					Credit7 = cr_amount
					balance7 = runningbal
					Description7 = narration
				}

				if x == 8 {
					Enter_date8 = postDate
					value_date8 = valueDate
					Debit8 = dr_amount
					Credit8 = cr_amount
					balance8 = runningbal
					Description8 = narration
				}

				if x == 9 {
					Enter_date9 = postDate
					value_date9 = valueDate
					Debit9 = dr_amount
					Credit9 = cr_amount
					balance9 = runningbal
					Description9 = narration
				}

				if x == 10 {
					Enter_date10 = postDate
					value_date10 = valueDate
					Debit10 = dr_amount
					Credit10 = cr_amount
					balance10 = runningbal
					Description10 = narration
				}

				if x == 11 {
					Enter_date11 = postDate
					value_date11 = valueDate
					Debit11 = dr_amount
					Credit11 = cr_amount
					balance11 = runningbal
					Description11 = narration
				}

				if x == 12 {
					Enter_date12 = postDate
					value_date12 = valueDate
					Debit12 = dr_amount
					Credit12 = cr_amount
					balance12 = runningbal
					Description12 = narration
				}

				// fmt.Printf("Date: %s DEBIT1: %s CREDIT1: %s BALANCE1: %s\n\n\n", startDate.Format(TIMELAYOUT), Debit1, Credit1, balance1)

				// prevBalance = runningbal

				// cus_name, _ = buffaloCtx.Session().Get(fmt.Sprintf("cus_name_%s", pdfGen.AccountNum))
				cus_name = openBalData["cusName"].(string)
				fmt.Println("----------CUSTOMER NAME--------", cus_name)
				CUSTOMER_NAME = cus_name

				if i == 0 {
					// session, _ := buffaloCtx.Session().Get(fmt.Sprintf("org_openbal_%s", pdfGen.AccountNum))
					// org_openbal, _ := strconv.ParseFloat(fmt.Sprintf("%v", session), 64)

					// session, _ = buffaloCtx.Session().Get(fmt.Sprintf("open_clrbal_%s", pdfGen.AccountNum))
					// open_clrbal, _ := strconv.ParseFloat(fmt.Sprintf("%v", session), 64)

					org_openbal := openBalData["bookBal"].(float64)
					open_clrbal := openBalData["clrBal"].(float64)

					BK_Open_Bal = php2go.NumberFormat(php2go.Round(org_openbal, 2), 2, ".", ",")
					AV_Open_Bal = php2go.NumberFormat(php2go.Round(open_clrbal, 2), 2, ".", ",")

				} else {
					open_clrbal := openBalData["clrBal"].(float64)
					BK_Close_Bal = runningbal
					AV_Open_Bal = php2go.NumberFormat(php2go.Round(open_clrbal, 2), 2, ".", ",")
				}

				// fmt.Println("totalCredit3: ", _totalCredits)
				// fmt.Println("totalDebit3: ", _totalDebits)

				if x == 12 {

					if first {
						first = false
					}

					// openBal, _ = pdfGen.GetOpenBalance()

					// session, _ := buffaloCtx.Session().Get(fmt.Sprintf("open_clrbal_%s", pdfGen.AccountNum))
					// open_clrbal, _ := strconv.ParseFloat(fmt.Sprintf("%v", session), 64)

					open_clrbal := openClearingBalance

					// session, _ = buffaloCtx.Session().Get(fmt.Sprintf("close_clearedbal_%s", pdfGen.AccountNum))
					// close_clearedbal, _ := strconv.ParseFloat(fmt.Sprintf("%v", session), 64)

					close_clearedbal := closeClearedBalance

					// session, _ = buffaloCtx.Session().Get(fmt.Sprintf("close_avail_bal_%s", pdfGen.AccountNum))
					// close_avail_bal, _ := strconv.ParseFloat(fmt.Sprintf("%v", session), 64)

					close_avail_bal := closeAvailableBalance

					BK_Open_Bal = php2go.NumberFormat(php2go.Round(openBal, 2), 2, ".", ",")
					AV_Open_Bal = php2go.NumberFormat(php2go.Round(open_clrbal, 2), 2, ".", ",")
					BK_Close_Bal = php2go.NumberFormat(php2go.Round(close_clearedbal, 2), 2, ".", ",")
					AV_Close_Bal = php2go.NumberFormat(php2go.Round(close_avail_bal, 2), 2, ".", ",")
					// fmt.Println("1BK_Close_Bal", BK_Close_Bal)
					var argv []string
					argv = append(argv, BOOK, AVAILABLE, TOTALCREDITTEXT, TOTALDEBITTEXT, OPENING_BALANCE, CLOSING_BALANCE,
						fmt.Sprintf("%f", totalcredit), fmt.Sprintf("%f", totaldebit), cus_name, BK_Open_Bal, AV_Open_Bal, BK_Close_Bal,
						AV_Close_Bal, Enter_date1, Enter_date2, Enter_date3,
						Enter_date4, Enter_date5, Enter_date6, Enter_date7, Enter_date8,
						Enter_date9, Enter_date10, Enter_date11, Enter_date12,
						value_date1, value_date2, value_date3, value_date4, value_date5, value_date6, value_date7,
						value_date8, value_date9, value_date10, value_date11, value_date12,
						Debit1, Debit2, Debit3, Debit4, Debit5, Debit6,
						Debit7, Debit8, Debit9, Debit10, Debit11, Debit12, Credit1, Credit2,
						Credit3, Credit4, Credit5, Credit6, Credit7, Credit8, Credit9, Credit10, Credit11, Credit12,
						balance1, balance2, balance3, balance4, balance5, balance6,
						balance7, balance8, balance9, balance10, balance11, balance12, Description1, Description2,
						Description3, Description4, Description5, Description6, Description7, Description8,
						Description9, Description10, Description11, Description12,
						fmt.Sprintf("%.2f", Total_Debits), fmt.Sprintf("%.2f", Total_Credits), fmt.Sprintf("%f", openBal), AccountNumber, StartDate, Enddate, D_postdate)

					html := pdfGen.OutputContent(argv, int64(loop), int64(page), count, openBalData, hasTransaction, noTransactionBBF)
					// fmt.Println("eeee")
					count++
					// fmt.Println("HTML: ", html)
					pdfPages = append(pdfPages, html)
					// fmt.Println("PAGE: ", page)

					resultData["totalDebit"] = _totalDebits
					resultData["totalCredit"] = _totalCredits

					y = y + 1
					page = page + 1
					x = 0

					BOOK = ""
					AVAILABLE = ""
					TOTALCREDITTEXT = ""
					TOTALDEBITTEXT = ""
					OPENING_BALANCE = ""
					CLOSING_BALANCE = ""
					cus_name = ""
					BK_Open_Bal = ""
					AV_Open_Bal = ""
					BK_Close_Bal = ""
					Enter_date1 = ""
					Enter_date2 = ""
					Enter_date3 = ""
					Enter_date4 = ""
					Enter_date5 = ""
					Enter_date6 = ""
					Enter_date7 = ""
					Enter_date8 = ""
					Enter_date9 = ""
					Enter_date10 = ""
					Enter_date11 = ""
					Enter_date12 = ""
					value_date1 = ""
					value_date2 = ""
					value_date3 = ""
					value_date4 = ""
					value_date5 = ""
					value_date6 = ""
					value_date7 = ""
					value_date8 = ""
					value_date9 = ""
					value_date10 = ""
					value_date11 = ""
					value_date12 = ""
					Debit1 = ""
					Debit2 = ""
					Debit3 = ""
					Debit4 = ""
					Debit5 = ""
					Debit6 = ""
					Debit7 = ""
					Debit8 = ""
					Debit9 = ""
					Debit10 = ""
					Debit11 = ""
					Debit12 = ""
					Credit1 = ""
					Credit2 = ""
					Credit3 = ""
					Credit4 = ""
					Credit5 = ""
					Credit6 = ""
					Credit7 = ""
					Credit8 = ""
					Credit9 = ""
					Credit10 = ""
					Credit11 = ""
					Credit12 = ""
					balance1 = ""
					balance2 = ""
					balance3 = ""
					balance4 = ""
					balance5 = ""
					balance6 = ""
					balance7 = ""
					balance8 = ""
					balance9 = ""
					balance10 = ""
					balance11 = ""
					balance12 = ""
					Description1 = ""
					Description2 = ""
					Description3 = ""
					Description4 = ""
					Description5 = ""
					Description6 = ""
					Description7 = ""
					Description8 = ""
					Description9 = ""
					Description10 = ""
					Description11 = ""
					Description12 = ""
					Total_Debits = 0.00
					Total_Credits = 0.00
				} else {

					if len(data) == z {

						// session, _ := buffaloCtx.Session().Get(fmt.Sprintf("open_clrbal_%s", pdfGen.AccountNum))
						// open_clrbal, _ := strconv.ParseFloat(fmt.Sprintf("%v", session), 64)

						open_clrbal := openClearingBalance

						// session, _ = buffaloCtx.Session().Get(fmt.Sprintf("close_clearedbal_%s", pdfGen.AccountNum))
						// close_clearedbal, _ := strconv.ParseFloat(fmt.Sprintf("%v", session), 64)
						close_clearedbal := closeClearedBalance

						// session, _ = buffaloCtx.Session().Get(fmt.Sprintf("close_avail_bal_%s", pdfGen.AccountNum))
						// close_avail_bal, _ := strconv.ParseFloat(fmt.Sprintf("%v", session), 64)
						close_avail_bal := closeAvailableBalance

						BK_Open_Bal = php2go.NumberFormat(php2go.Round(bookBalance, 2), 2, ".", ",")
						AV_Open_Bal = php2go.NumberFormat(php2go.Round(open_clrbal, 2), 2, ".", ",")
						BK_Close_Bal = php2go.NumberFormat(php2go.Round(close_clearedbal, 2), 2, ".", ",")
						AV_Close_Bal = php2go.NumberFormat(php2go.Round(close_avail_bal, 2), 2, ".", ",")

						var argv []string
						argv = append(argv, BOOK, AVAILABLE, TOTALCREDITTEXT, TOTALDEBITTEXT, OPENING_BALANCE, CLOSING_BALANCE,
							fmt.Sprintf("%f", totalcredit), fmt.Sprintf("%f", totaldebit), cus_name, BK_Open_Bal, AV_Open_Bal, BK_Close_Bal,
							AV_Close_Bal, Enter_date1, Enter_date2, Enter_date3,
							Enter_date4, Enter_date5, Enter_date6, Enter_date7, Enter_date8,
							Enter_date9, Enter_date10, Enter_date11, Enter_date12,
							value_date1, value_date2, value_date3, value_date4, value_date5, value_date6, value_date7,
							value_date8, value_date9, value_date10, value_date11, value_date12,
							Debit1, Debit2, Debit3, Debit4, Debit5, Debit6,
							Debit7, Debit8, Debit9, Debit10, Debit11, Debit12, Credit1, Credit2,
							Credit3, Credit4, Credit5, Credit6, Credit7, Credit8, Credit9, Credit10, Credit11, Credit12,
							balance1, balance2, balance3, balance4, balance5, balance6,
							balance7, balance8, balance9, balance10, balance11, balance12, Description1, Description2,
							Description3, Description4, Description5, Description6, Description7, Description8,
							Description9, Description10, Description11, Description12,
							fmt.Sprintf("%.2f", Total_Debits), fmt.Sprintf("%.2f", Total_Credits), fmt.Sprintf("%f", openBal), AccountNumber, StartDate, Enddate, D_postdate)

						html := pdfGen.OutputContent(argv, int64(loop), int64(page), count, openBalData, hasTransaction, noTransactionBBF)
						count++
						pdfPages = append(pdfPages, html)

						resultData["totalDebit"] = _totalDebits
						resultData["totalCredit"] = _totalCredits

						balSummary["BOOK"] = "BOOK"
						balSummary["AVAILABLE"] = "AVAILABLE"
						balSummary["OPENING_BALANCE"] = "OPENING BALANCE"
						balSummary["BK_Open_Bal"] = BK_Open_Bal
						balSummary["AV_Open_Bal"] = AV_Open_Bal
						balSummary["CLOSING_BALANCE"] = "CLOSING BALANCE"
						balSummary["BK_Close_Bal"] = BK_Close_Bal
						balSummary["AV_Close_Bal"] = AV_Close_Bal
						balSummary["TOTALDEBITTEXT"] = "TOTAL DEBIT"
						balSummary["totalDb"] = php2go.NumberFormat(php2go.Round(php2go.Abs(_totalDebits), 2), 2, ".", ",")
						balSummary["TOTALCREDITTEXT"] = "TOTAL CREDIT"
						balSummary["totalCr"] = php2go.NumberFormat(php2go.Round(php2go.Abs(_totalCredits), 2), 2, ".", ",")

						// fmt.Println("balSummary: ", balSummary)
						page = page + 1
						// fmt.Println("PAGE: ", page)
					}
				}
				x++
				z++
			}
		}

		// fmt.Println("TOTAL DEBIT3: ", totaldebit)

		i++
		k++
		totaldebit = 0.0
		totalcredit = 0.0
		OPENING_BALANCE = ""
		CLOSING_BALANCE = ""
		TOTALDEBITTEXT = ""
		TOTALCREDITTEXT = ""
		BOOK = ""
		AVAILABLE = ""
		AVAILABLE = ""
		BOOK = ""

		session := buffaloCtx.Session().Get(fmt.Sprintf("cus_reg_date_1_%s", pdfGen.AccountNum))
		if session == nil {
			session = ""
		}
		cusRegDate, _ := time.Parse(TIMELAYOUT, session.(string))
		endDate, _ := time.Parse(TIMELAYOUT, pdfGen.EndDate)

		if endDate.Sub(cusRegDate) <= 0 {

			date := time.Date(cusRegDate.Year(), cusRegDate.Month()+1, cusRegDate.Day(), cusRegDate.Hour(), cusRegDate.Minute(), 0, 0, cusRegDate.Location())
			buffaloCtx.Session().Set(fmt.Sprintf("cus_reg_date_1_%s", pdfGen.AccountNum), date.Format(TIMELAYOUT))
			buffaloCtx.Session().Set(fmt.Sprintf("cus_table_%s", pdfGen.AccountNum), fmt.Sprintf("cus_reg_p%d%02d", date.Year(), date.Month()))
		}

		session = buffaloCtx.Session().Get(fmt.Sprintf("sdate_%s", pdfGen.AccountNum))
		if session == nil {
			session = ""
		}
		d, _ := time.Parse(TIMELAYOUT, session.(string))
		d = time.Date(d.Year(), d.Month()+1, 1, 0, 0, 0, 0, d.Location())
		buffaloCtx.Session().Set(fmt.Sprintf("cus_reg_date_1_%s", pdfGen.AccountNum), d.Format(TIMELAYOUT))

		buffaloCtx.Session().Set(fmt.Sprintf("trans_table_%s", pdfGen.AccountNum), fmt.Sprintf("stats_p%d%02d", d.Year(), d.Month()))
		if session == nil {
			session = ""
		}
		d, _ = time.Parse(TIMELAYOUT, session.(string))
		d = time.Date(d.Year(), d.Month()+1, 0, 0, 0, 0, 0, d.Location())
		buffaloCtx.Session().Set(fmt.Sprintf("edate_%s", pdfGen.AccountNum), d.Format(TIMELAYOUT))

		startDate = time.Date(startDate.Year(), startDate.Month()+1, 1, 0, 0, 0, 0, startDate.Location())

		// fmt.Printf("Date: %s EOL\n\n\n", startDate.Format(TIMELAYOUT))
	}

	NUM_PAGES = int64(len(pdfPages))
	var pdfPath string
	// generate pdf
	buff := new(bytes.Buffer)
	if len(pdfPages) > 0 {
		for i := 0; i < len(pdfPages); i++ {

			buff.WriteString(pdfGen.GetBodyHtmlStyle())
			buff.WriteString(pdfGen.GetCustomerTransSummary(balSummary, openBalData["cusName"].(string), openBalData["Cus_id"].(string)))
			buff.WriteString(pdfPages[i])
			buff.WriteString(`<P style="page-break-before: always">`)
			buff.WriteString(pdfGen.GetBodyHtmlStyle())
		}
		page := wkhtmltopdf.NewPageReader(buff)
		page.MinimumFontSize.Set(12)
		page.HeaderHTML.Set(headerPath)
		page.FooterHTML.Set(footerPath)
		// page.FooterFontSize.Set(3)
		page.Zoom.Set(1.2)
		pdf.AddPage(page)
		pdf.TOC.FooterRight.Set("[page]")
		err := pdf.Create()
		if err != nil {
			fmt.Println("@@@ Error: ", err.Error())
		}

		fmt.Println("@@@ HeaderPath: ", headerPath)

		// write to file
		dirPath := envy.Get("GH_ACCOUNT_STATEMENT_TEMP", "./")
		filePath := fmt.Sprintf("%s_%s_account_statement.pdf", pdfGen.AccountNum, user.ID)
		pdfPath = fmt.Sprintf("%s/%s", dirPath, filePath)
		err = pdf.WriteFile(pdfPath)
		if err != nil {
			fmt.Println("@@@ Error: ", err.Error())
		}

		// DeleteFile(headerPath)
	}

	// resultData["pdf_buffer"] = pdf.Buffer().Bytes()

	// PDF Audit
	sdate, _ := time.Parse("2006-01-02", pdfGen.StartDate)
	edate, _ := time.Parse("2006-01-02", pdfGen.EndDate)
	PrintType := "PDF"
	LogPrintActivity(GormDB, fmt.Sprintf("%s %s", user.FirstName, user.LastName), user.Branch.Name, pdfGen.AccountNum, CUSTOMER_NAME, NUM_PAGES, PrintType, sdate, edate)

	return resultData, pdfPath
}

// func (pdfGen *PDFGenerator) GetOpenBalance() (float64, error) {
// 	//ss := fmt.Sprintf("%f", buffaloCtx.Session().Get(fmt.Sprintf("openbal_%s", pdfGen.AccountNum)))
// 	//_ss, _ := buffaloCtx.Session().Get(fmt.Sprintf("openbal_%s", pdfGen.AccountNum))
// 	ss, _ := buffaloCtx.Session().Get(fmt.Sprintf("openbal_%s", pdfGen.AccountNum))
// 	return strconv.ParseFloat(ss, 64)
// }

func (pdfGen *PDFGenerator) GetClosedAvailableBal(transTable string) float64 {
	var closeClearedBal float64
	sql := fmt.Sprintf("select sum(M_TxnAmt) as bal from %s WHERE  C_TxnAccNo = '%s' AND ( D_TxnPostDt BETWEEN '%s' AND '%s' )",
		transTable, pdfGen.AccountNum, pdfGen.StartDate, pdfGen.EndDate)

	// get query result
	GormDB.Raw(sql).First(&closeClearedBal)

	return closeClearedBal
}

func (pdfGen *PDFGenerator) GetClosedClearedBal(transTable string) float64 {
	var closedAvailBal float64
	sql := fmt.Sprintf("select sum(M_TxnAmt) as closedavailbal from %s WHERE C_TxnAccNo= '%s' AND (D_TxnValueDt BETWEEN '%s' AND '%s' )", transTable, pdfGen.AccountNum, pdfGen.StartDate, pdfGen.EndDate)

	// get query result
	GormDB.Raw(sql).First(&closedAvailBal)

	return closedAvailBal
}

func (pdfGen *PDFGenerator) OutputContent(arrayValues []string, loop int64, pageNum int64, count int, openBalData map[string]interface{},
	hasTransaction bool, noTransactionBBF float64) string {
	var BOOK, AVAILABLE, TOTALCREDITTEXT, TOTALDEBITTEXT, OPENING_BALANCE, CLOSING_BALANCE,
		totalcredit, totaldebit, cus_name, BK_Open_Bal, AV_Open_Bal, BK_Close_Bal,
		AV_Close_Bal, Enter_date1, Enter_date2, Enter_date3,
		Enter_date4, Enter_date5, Enter_date6, Enter_date7, Enter_date8,
		Enter_date9, Enter_date10, Enter_date11, Enter_date12,
		value_date1, value_date2, value_date3,
		value_date4, value_date5, value_date6, value_date7,
		value_date8, value_date9, value_date10, value_date11, value_date12,
		Debit1, Debit2, Debit3, Debit4, Debit5, Debit6,
		Debit7, Debit8, Debit9, Debit10, Debit11, Debit12, Credit1, Credit2, Credit3, Credit4,
		Credit5, Credit6, Credit7, Credit8, Credit9, Credit10, Credit11, Credit12,
		Balance1, Balance2, Balance3, Balance4, Balance5, Balance6, Balance7, Balance8,
		Balance9, Balance10, Balance11, Balance12, Description1, Description2,
		Description3, Description4,
		Description5, Description6, Description7, Description8,
		Description9, Description10, Description11, Description12,
		Total_Debits, Total_Credits, openBal, AccountNumber, StartDate, Enddate, D_postdate string

	List(arrayValues, &BOOK, &AVAILABLE, &TOTALCREDITTEXT, &TOTALDEBITTEXT, &OPENING_BALANCE, &CLOSING_BALANCE,
		&totalcredit, &totaldebit, &cus_name, &BK_Open_Bal, &AV_Open_Bal, &BK_Close_Bal,
		&AV_Close_Bal, &Enter_date1, &Enter_date2, &Enter_date3,
		&Enter_date4, &Enter_date5, &Enter_date6, &Enter_date7, &Enter_date8,
		&Enter_date9, &Enter_date10, &Enter_date11, &Enter_date12,
		&value_date1, &value_date2, &value_date3,
		&value_date4, &value_date5, &value_date6, &value_date7,
		&value_date8, &value_date9, &value_date10, &value_date11, &value_date12,
		&Debit1, &Debit2, &Debit3, &Debit4, &Debit5, &Debit6,
		&Debit7, &Debit8, &Debit9, &Debit10, &Debit11, &Debit12, &Credit1, &Credit2, &Credit3, &Credit4,
		&Credit5, &Credit6, &Credit7, &Credit8, &Credit9, &Credit10, &Credit11, &Credit12,
		&Balance1, &Balance2, &Balance3, &Balance4, &Balance5, &Balance6, &Balance7, &Balance8,
		&Balance9, &Balance10, &Balance11, &Balance12, &Description1, &Description2,
		&Description3, &Description4,
		&Description5, &Description6, &Description7, &Description8,
		&Description9, &Description10, &Description11, &Description12,
		&Total_Debits, &Total_Credits, &openBal, &AccountNumber, &StartDate, &Enddate, &D_postdate)

	Description13 := ""

	Debit13 := ""
	Credit13 := ""

	// cus_name, _ = buffaloCtx.Session().Get(fmt.Sprintf("cus_name_%s", pdfGen.AccountNum))
	cus_name = openBalData["cusName"].(string)

	td, _ := strconv.ParseFloat(totaldebit, 64)
	tc, _ := strconv.ParseFloat(totalcredit, 64)
	total_debit := php2go.NumberFormat(php2go.Round(php2go.Abs(td), 2), 2, ".", ",")
	total_credit := php2go.NumberFormat(php2go.Round(php2go.Abs(tc), 2), 2, ".", ",")
	if php2go.Empty(Enter_date1) {
		if !hasTransaction {
			Description1 = "----------------- END OF STATEMENT --------------"
			Debit1 = "_______________"
			Credit1 = "_______________"
			Balance1 = ""
		} else {
			Description1 = "----------------- END OF STATEMENT --------------"
			Debit1 = "_______________"
			Credit1 = "_______________"
			Debit2 = total_debit
		}

	} else if php2go.Empty(Enter_date2) {
		Description2 = "----------------- END OF STATEMENT --------------"
		Debit2 = "_______________"
		Credit2 = "_______________"
		Debit3 = total_debit
		Credit3 = total_credit
	} else if php2go.Empty(Enter_date3) {
		Description3 = "----------------- END OF STATEMENT --------------"
		Debit3 = "_______________"
		Credit3 = "_______________"
		Debit4 = total_debit
		Credit4 = total_credit
	} else if php2go.Empty(Enter_date4) {
		Description4 = "----------------- END OF STATEMENT --------------"
		Debit4 = "_______________"
		Credit4 = "_______________"
		Debit5 = total_debit
		Credit5 = total_credit
	} else if php2go.Empty(Enter_date5) {
		Description5 = "----------------- END OF STATEMENT --------------"
		Debit5 = "_______________"
		Credit5 = "_______________"
		Debit6 = total_debit
		Credit6 = total_credit
	} else if php2go.Empty(Enter_date6) {
		Description6 = "----------------- END OF STATEMENT --------------"
		Debit6 = "_______________"
		Credit6 = "_______________"
		Debit7 = total_debit
		Credit7 = total_credit
	} else if php2go.Empty(Enter_date7) {
		Description7 = "----------------- END OF STATEMENT --------------"
		Debit7 = "_______________"
		Credit7 = "_______________"
		Debit8 = total_debit
		Credit8 = total_credit
	} else if php2go.Empty(Enter_date8) {
		Description8 = "----------------- END OF STATEMENT --------------"
		Debit8 = "_______________"
		Credit8 = "_______________"
		Debit9 = total_debit
		Credit9 = total_credit
	} else if php2go.Empty(Enter_date9) {
		Description9 = "----------------- END OF STATEMENT --------------"
		Debit9 = "_______________"
		Credit9 = "_______________"
		Debit10 = total_debit
		Credit10 = total_credit
	} else if php2go.Empty(Enter_date10) {
		Description10 = "----------------- END OF STATEMENT --------------"
		Debit10 = "_______________"
		Credit10 = "_______________"
		Debit11 = total_debit
		Credit11 = total_credit
	} else if php2go.Empty(Enter_date11) {
		Description11 = "----------------- END OF STATEMENT --------------"
		Debit11 = "_______________"
		Credit11 = "_______________"
		Debit12 = total_debit
		Credit12 = total_credit
	} else {
		if loop == pageNum {
			Description13 = "----------------- END OF STATEMENT --------------"
			Debit13 = "_______________" + total_debit
			Credit13 = "_______________" + total_credit
		}
	}

	branch_name := buffaloCtx.Session().Get(fmt.Sprintf("branch_name_%s", pdfGen.AccountNum))
	date, _ := time.Parse(TIMELAYOUT, pdfGen.StartDate)
	endOfLastMonth := time.Date(date.Year(), date.Month(), 0, 0, 0, 0, 0, date.Location())
	OPEN_BAL_DATE := endOfLastMonth.Format(TIMELAYOUT)

	if OPENING_BALANCE == "" {
		BK_Open_Bal = ""
		AV_Open_Bal = ""
		BK_Close_Bal = ""
		AV_Close_Bal = ""
		Total_Debits = ""
		Total_Credits = ""
	}

	html := ""
	html += `
	<tr>
	<td valign="top"><table height="61" border="0">
    <tr>
        <td class="td_4"  height="40" valign="middle">
            <div align="center">
            ENTRY DATE</div></td>
        <td class="td_2" valign="middle">
            <div align="center">
                VALUE DATE </div></td>
        <td class="td_2" width="10" >&nbsp;</td>
        <td class="td_2"><br><div align="center">DESCRIPTION</div></td>
        <td class="td_2" ><br><div align="center">DEBITS</div></td>
        <td class="td_2" ><br><div align="center">CREDITS</div></td>
        <td class="td_2" ><br><div align="center">BALANCE</div></td>
    </tr>
	`

	OPEN_BAL2 := php2go.NumberFormat(php2go.Round(php2go.Abs(pdfGen.GetBalanceBroughtForward(Debit1, Credit1, Balance1)), 2), 2, ".", ",")

	if count == 1 {
		html += pdfGen.BalanceBroughtForwardHtml(OPEN_BAL_DATE, OPEN_BAL2, hasTransaction, noTransactionBBF)
	} else {
		html += pdfGen.BalanceBroughtForwardHtml("", OPEN_BAL2, hasTransaction, noTransactionBBF)
	}

	html += fmt.Sprintf(`
	<tr>
        <td class="td" height="45"><div align="center">%s</div></td>
        <td class="td"><div align="center">%s</div></td>
        <td class="td_1">&nbsp;</td>
        <td class="td_3"><div align="left">%s</div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td_1"><div align="right">%s &nbsp;</div></td>
    </tr>
    <tr>
        <td class="td" height="45"><div align="center">%s</div></td>
        <td class="td" style="text-align:center;border-top: 0px solid white;border-bottom: 0px solid white;"><div align="center">%s</div></td>
        <td class="td_1">&nbsp;</td>
        <td class="td_3"><div align="left">%s</div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td_1"><div align="right">%s &nbsp;</div></td>
    </tr>
    <tr>
        <td class="td" height="45"><div align="center">%s</div></td>
        <td class="td"><div align="center">%s</div></td>
        <td class="td_1">&nbsp;</td>
        <td class="td_3"><div align="left">%s</div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td_1" ><div align="right">%s &nbsp;</div></td>
    </tr>
    <tr>
        <td class="td" height="45"><div align="center">%s</div></td>
        <td class="td"><div align="center">%s</div></td>
        <td class="td_1">&nbsp;</td>
        <td class="td_3"><div align="left">%s</div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td_1"><div align="right">%s &nbsp;</div></td>
    </tr>
    <tr>
        <td class="td" height="45"><div align="center">%s </div></td>
        <td class="td" ><div align="center">%s </div></td>
        <td class="td_1" >&nbsp;</td>
        <td class="td_3" ><div align="left">%s </div></td>
        <td class="td" ><div align="right">%s &nbsp;</div></td>
        <td class="td" ><div align="right">%s &nbsp;</div></td>
        <td class="td_1" ><div align="right">%s &nbsp;</div></td>
    </tr>
    <tr>
        <td class="td" height="45"><div align="center">%s </div></td>
        <td class="td"><div align="center">%s </div></td>
        <td class="td_1">&nbsp;</td>
        <td class="td_3"><div align="left">%s </div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td_1"><div align="right">%s &nbsp;</div></td>
    </tr>
                <tr>
        <td class="td" height="45"><div align="center">%s</div></td>
        <td class="td"><div align="center">%s </div></td>
        <td class="td_1">&nbsp;</td>
        <td class="td_3"><div align="left">%s </div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td_1"><div align="right">%s &nbsp;</div></td>
    </tr>
    <tr>
        <td class="td" height="45"><div align="center">%s </div></td>
        <td class="td"><div align="center">%s </div></td>
        <td class="td_1">&nbsp;</td>
        <td class="td_3"><div align="left">%s </div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td_1"><div align="right">%s &nbsp;</div></td>
    </tr>
    <tr>
        <td class="td" height="45"><div align="center">%s </div></td>
        <td class="td"><div align="center">%s </div></td>
        <td class="td_1">&nbsp;</td>
        <td class="td_3"><div align="left">%s </div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td_1"><div align="right">%s &nbsp;</div></td>
    </tr>
    <tr>
        <td class="td" height="45"><div align="center">%s </div></td>
        <td class="td"><div align="center">%s </div></td>
        <td class="td_1">&nbsp;</td>
        <td class="td_3"><div align="left">%s </div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td_1"><div align="right">%s &nbsp;</div></td>
    </tr>

    <tr>
        <td class="td" height="45"><div align="center">%s </div></td>
        <td class="td"><div align="center">%s </div></td>
        <td class="td_1">&nbsp;</td>
        <td class="td_3"><div align="left">%s </div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td_1"><div align="right">%s &nbsp;</div></td>
    </tr>
    <tr>
        <td class="td" height="45"><div align="center">%s </div></td>
        <td class="td"><div align="center">%s </div></td>
        <td class="td_1">&nbsp;</td>
        <td class="td_3"><div align="left">%s </div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td"><div align="right">%s &nbsp;</div></td>
        <td class="td_1"><div align="right">%s &nbsp;</div></td>
    </tr>
    <tr>
        <td class="td"><div align="center"></div></td>
        <td class="td"><div align="center"></div></td>
        <td class="td_1">&nbsp;</td>
        <td class="td_3"><div align="left">%s</div></td>
        <td class="td" valign="top"><div align="right">%s</div></td>
        <td class="td" valign="top"><div align="right">%s</div></td>
        <td class="td_1"><div align="right">&nbsp;</div></td>
    </tr>
    </table>
	</td>
	</tr>
	`, Enter_date1, value_date1, Description1, Debit1, Credit1, Balance1,
		Enter_date2, value_date2, Description2, Debit2, Credit2, Balance2,
		Enter_date3, value_date3, Description3, Debit3, Credit3, Balance3,
		Enter_date4, value_date4, Description4, Debit4, Credit4, Balance4,
		Enter_date5, value_date5, Description5, Debit5, Credit5, Balance5,
		Enter_date6, value_date6, Description6, Debit6, Credit6, Balance6,
		Enter_date7, value_date7, Description7, Debit7, Credit7, Balance7,
		Enter_date8, value_date8, Description8, Debit8, Credit8, Balance8,
		Enter_date9, value_date9, Description9, Debit9, Credit9, Balance9,
		Enter_date10, value_date10, Description10, Debit10, Credit10, Balance10,
		Enter_date11, value_date11, Description11, Debit11, Credit11, Balance11,
		Enter_date12, value_date12, Description12, Debit12, Credit12, Balance12,
		Description13, Debit13, Credit13)

	// adding footer
	html += fmt.Sprintf(`
		<tr>
		<td style="font-size:8px">Important Notice: Outward Telegraphic Transfer
		<p>Effective Monday August 11, 2008, you will now be required to attach a duly
		signed cheque in favour of Standard Chatered Bank for all outward telegraphic
		transfers on domiciliary current accounts <br>For domiciliary saving accounts, the account signatory should be present to
		effect the transfer</p>
		<p>This is for your added security.</p>
		<p>Deposits and Payments are governed by the laws in effect from time to time
		in the Republic of Ghana and are payable only at the branch of
		Standard Chartered Bank (GH) Ltd in Ghana where the deposits were made.
		Standard Chartered Bank (GH) Ltd has a discretion to allow withdrawal at
		other branches in Ghana. The items and branches on this statement should
		be verified and the bank notified of any discrepencies.</p>
		</td>
        </tr>
        <tr>
            <td>
            <table class="footer-table">
                    <tr>
                            <td >
                                %s
                            </td>
                            <td width="400">
                                <p>** This is a re-generated copy and not a copy of the original**</p>
                            </td>
                    </tr>
            </table>
            </td>
        </tr>
        </table>
        <div style="text-align:center;">Page %d</div>
	`, branch_name, count)
	// fmt.Println(branch_name)

	html += "</table>"

	// orgOpenBal, _ := buffaloCtx.Session().Get(fmt.Sprintf("org_openbal_%s", pdfGen.AccountNum))
	// orgOpenBal:=openBalData["bookBal"].(float64)
	// buffaloCtx.Session().Set(fmt.Sprintf("openbal_%s", pdfGen.AccountNum), orgOpenBal)
	// buffaloCtx.Session().Set(fmt.Sprintf("totalcredit_%s", pdfGen.AccountNum), totalcredit)
	// buffaloCtx.Session().Set(fmt.Sprintf("totaldebit_%s", pdfGen.AccountNum), totaldebit)
	return html

}

func (pdfGen *PDFGenerator) WriteHeaderHtml(currency string, currentUserId uuid.UUID) string {
	hostName := envy.Get("GH_APP_URL", buffaloCtx.Request().Host)
	if !strings.Contains(hostName, "http") {
		hostName = "http://" + hostName
	}
	dirPath := envy.Get("GH_ACCOUNT_STATEMENT_TEMP", "./templates")
	style := `
	<style>
			tr{
				display: flex;
				justify-content: space-between;
			}
			p{
				font-size: medium;
			}
			table{
				width: 100%;
			}
		</style>
	`
	html :=
		fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="en">
		<body>
		%s
		<table style="height:120;" >
		<tr>
			<td>&nbsp;</td>
			<td style="text-align: center;float: right;">
				<h3>STATEMENT OF ACCOUNT</h4>
				<p>FOR ACCOUNT NUMBER %s</p>
				<p >From %s To %s </p>
				<p><strong>CURRENCY %s</strong></p>
			</td>
			<td>
				<img src="%s/assets/images/logo.jpg" height="130" width="150"  alt="" style="float: right;">
			</td>
		</tr>
	</table>
		</body>
	</html>
	`, style, pdfGen.AccountNum, pdfGen.StartDate, pdfGen.EndDate, currency, hostName)

	filePath := fmt.Sprintf("%s_%s_header.html", pdfGen.AccountNum, currentUserId)

	WriteToFile(html, fmt.Sprintf("%s/%s", dirPath, filePath))
	return fmt.Sprintf("%s/%s", dirPath, filePath)

}

func (pdfGen *PDFGenerator) WriteFooterHtml(abNumber, branchName string, currentUserId uuid.UUID) string {

	dirPath := envy.Get("GH_ACCOUNT_STATEMENT_TEMP", "./templates")
	branchName = strings.ReplaceAll(branchName, " ", "-")
	uniqueIdentifier := fmt.Sprintf("%s-%s-%s", strings.ToUpper(branchName), time.Now().Format("20060102150405"), abNumber)
	html := fmt.Sprintf(`<div style="text-align:right;font-size:10px;">%s</div>`, uniqueIdentifier)

	filePath := fmt.Sprintf("%s_%s_footer.html", pdfGen.AccountNum, currentUserId)

	WriteToFile(html, fmt.Sprintf("%s/%s", dirPath, filePath))
	return fmt.Sprintf("%s/%s", dirPath, filePath)

}
func (pdfGen *PDFGenerator) GetCustomerTransSummary(balSum map[string]string, cusName, cusId string) string {
	cus_name := cusName
	cusReg := &CusReg{}
	cReg, _ := cusReg.GetAddress(cusId, GeneralLedger{AccountNum: pdfGen.AccountNum, StartDate: pdfGen.StartDate, EndDate: pdfGen.EndDate})
	addr1 := cReg.C_Addr1
	addr2 := cReg.C_Addr2
	addr3 := cReg.C_Addr3

	// adding Balance Summary to html
	html := fmt.Sprintf(`
	<table height="903" border="1"><tr>
	<td valign="top">
	<table width="900" cellpadding="2" border="0">
	<tr>
		<td width="332" rowspan="6">
			<div class="details">
			%s <br>
			%s<br>
			%s<br>
			%s<br>
			</div>
		</td>
		<td width="101">&nbsp;</td>
		<td width="119"><div align="right">%s</div></td>
		<td width="113"><div align="right">%s</div></td>
	</tr>
	<tr>
		<td><div align="right">%s</div></td>
		<td><div align="right">%s&nbsp;&nbsp;</div></td>
		<td><div align="right">%s&nbsp;&nbsp;</div></td>
	</tr>
	<tr>
		<td><div align="right">%s</div></td>
		<td><div align="right">%s&nbsp;&nbsp;</div></td>
		<td><div align="right">%s&nbsp;&nbsp;</div></td>
	</tr>
	<tr>
		<td><div align="right">%s</div></td>
		<td><div align="right">%s&nbsp;&nbsp;</div></td>
		<td></td>
	</tr>
	<tr>
		<td><div align="right">%s</div></td>
		<td><div align="right">%s&nbsp;&nbsp;</div></td>
		<td align=center></td>
	</tr>
	<tr>
		<td height="10">&nbsp;</td>
		<td>&nbsp;</td>
		<td>&nbsp;</td>
	</tr>
	</table>
		</td>
	</tr>
`, cus_name, addr1, addr2, addr3, balSum["BOOK"], balSum["AVAILABLE"],
		balSum["OPENING_BALANCE"], balSum["BK_Open_Bal"], balSum["AV_Open_Bal"], balSum["CLOSING_BALANCE"], balSum["BK_Close_Bal"], balSum["AV_Close_Bal"],
		balSum["TOTALDEBITTEXT"], balSum["totalDb"], balSum["TOTALCREDITTEXT"], balSum["totalCr"])

	return html
}

func (pdfGen *PDFGenerator) GetBodyHtmlStyle() string {
	style := `
	<style>
        .td{
            border-left:1px solid black;
            border-bottom: 0px solid white;
            border-top: 0px solid white;
            }
        .td_1{
            border-left:1px solid black;
            border-bottom: 0px solid white;
            border-top: 0px solid white;
            border-right: 0px solid white;
            }
        .td_2{
            border-left:0px solid white;
            border-right: 0px solid white;
            border-bottom: 1px solid black;
            vertical-align:middle;
            }

            .td_3{
            border-left:0px solid white;
            border-bottom: 0px solid white;
            border-top: 0px solid white;
			width:100%;
            }

        .td_4{
            border-left:1px solid black;
            border-right: 0px solid white;
            border-bottom: 1px solid black;
            vertical-align:middle;
        }
		.details{
            font-size: 12px;
            font-weight: Bold;
        }

		.footer-table{
			font-size:7px;
			border:0;
			width:100%;	
		}
            
        </style>
	`

	return style
}

func (pdfGen *PDFGenerator) BalanceBroughtForwardHtml(openBalDate, balance string, hasTransaction bool, noTransactionBBF float64) string {
	if !hasTransaction {
		balance = php2go.NumberFormat(php2go.Round(php2go.Abs(noTransactionBBF), 2), 2, ".", ",")
	}
	html := fmt.Sprintf(`
	<tr>
                <td height="30" class="td"><div align="center">%s</div></td>
                <td class="td"><div align="center"></div></td>
                <td class="td_1">&nbsp;</td>
                <td class="td_3">BALANCE BROUGHT FORWARD</td>
                <td class="td"><div align="right"></div></td>
                <td class="td"><div align="right"></div></td>
                <td class="td_1"><div align="right">%s&nbsp;</div></td>
            </tr>
	`, openBalDate, balance)

	return html
}

func (pdfGen *PDFGenerator) DisplayEndOfStatement() string {
	return ""
}

func (pdfGen *PDFGenerator) GetBalanceBroughtForward(debit, credit, balance string) float64 {
	var cr, db, bal float64

	strSplit := php2go.Explode(",", credit)
	str := strings.Join(strSplit[:], "")
	if credit == "-" {
		cr = 0
	} else {
		cr, _ = strconv.ParseFloat(str, 64)
	}

	strSplit = php2go.Explode(",", debit)
	str = strings.Join(strSplit[:], "")
	if debit == "-" {
		db = 0
	} else {
		db, _ = strconv.ParseFloat(str, 64)
	}

	strSplit = php2go.Explode(",", balance)
	str = strings.Join(strSplit[:], "")
	if balance == "-" {
		bal = 0
	} else {
		bal, _ = strconv.ParseFloat(str, 64)
	}

	bbf := db - cr + bal

	return bbf
}

func WriteToFile(contents, path string) {
	file, _ := os.Create(path)
	file.WriteString(contents)
	file.Sync()
}

func DeleteFile(path string) {
	os.Remove(path)
}

func DeleteOldPrintedFiles() error {
	// get location
	tempDir := envy.Get("GH_ACCOUNT_STATEMENT_TEMP", "")

	if tempDir != "" {
		files, err := ioutil.ReadDir(tempDir)
		if err != nil {
			return err
		}

		fmt.Println("((((((((((())))))))))) Files: ", len(files))

		for _, file := range files {
			modificationTime := file.ModTime()

			fmt.Println("~~~~~~~~~~~~` ", modificationTime.Before(time.Now()))
			if modificationTime.Before(time.Now()) {
				os.Remove(fmt.Sprintf("%s/%s", tempDir, file.Name()))

			}
		}
	}

	return nil
}
