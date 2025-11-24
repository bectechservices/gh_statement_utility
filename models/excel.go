package models

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/syyongx/php2go"
)

type Excel struct {
	AccountNum string `json:"account_no" form:"account_no"`
	StartDate  string `json:"start_date" form:"start_date"`
	EndDate    string `json:"end_date" form:"end_date"`
}

type Excels []Excel

var buffaloCtx buffalo.Context

func (excel *Excel) New(c buffalo.Context) {
	buffaloCtx = c
}

func (excel Excel) ExportWithQuery(currency, customerName string, bookBalance, tc, td float64, user User) (string, error) {
	c_ccy := currency

	cusName := customerName

	openBal := bookBalance
	totalDebit := td
	totalCredit := tc

	TOTALDEBITS := php2go.NumberFormat(php2go.Round(php2go.Abs(totalDebit), 2), 2, ".", ",")

	TOTALCREDITS := php2go.NumberFormat(php2go.Round(php2go.Abs(totalCredit), 2), 2, ".", ",")

	OPENBAL := php2go.NumberFormat(php2go.Round(php2go.Abs(openBal), 2), 2, ".", ",")

	startDate, _ := time.Parse(TIMELAYOUT, excel.StartDate)
	openBalDate := time.Date(startDate.Year(), startDate.Month(), 0, 0, 0, 0, 0, startDate.Location())

	table := "<center><table border=1px><th>Statement</th>"

	table = table +
		`<table width='100%' cellspacing='15px'>` +
		"<tr><td>Name:</td><td>" + cusName + "</td></tr>" +
		"<tr><td>Currency:</td><td>" + c_ccy + "</td></tr>" +
		"<tr> <td class='dt'>Total Credit:</td><td id='totalcredits'>" + TOTALCREDITS + "</td></tr>" +
		"<tr> <td class='dt'>Total Debit:</td><td id='totaldebits'>" + TOTALDEBITS + "</td></tr></table><br>" +
		`<table>
		<tr>
			<td colspan=5></td>
		</tr>
		<tr>
			<td>ENTRY DATE</td>
			<td>VALUE DATE</td>
			<td>PARTICULARS</td>
			<td>WITHDRAWAL</td>
			<td>DEPOSIT</td>
			<td>BALANCE</td>
		</tr>
	</table>`

	table = table +
		`
	<table width="1032px"><tr><td colspan='2'
	style ="width:88px;text-align:center;padding:7px;" class="tr_td">` +
		openBalDate.Format(TIMELAYOUT) + "</td>" +
		` <td style = 'width:505px;text-align:left;padding:7px;'
	class='tr_td'>Balance Brought Forward </td>
	<td style = 'width:139px;text-align:right;padding:7px;' class='tr_td'>
	</td><td style = 'width:139px;text-align:right;padding:7px;' class='tr_td'>
	</td><td style = 'width:140px;text-align:right;padding:7px;' class='tr_td'>` +
		OPENBAL + "</td></tr>"

	cr, dr, totalCredit, totalDebit := 0, 0, 0, 0

	//st := Statement{}
	var postDate, valueDate, narration, dbAmount, crAmount string
	var amount float64

	endDate, _ := time.Parse(TIMELAYOUT, excel.EndDate)

	//get months between start and end date
	diff := endDate.Sub(startDate)
	months := diff.Hours() / 24 / 30
	months = math.Round(months)

	i := 1

	rowCount := 0

	//Check if start and end date have same month and year
	if startDate.Month() == endDate.Month() && startDate.Year() == endDate.Year() {
		months = 1
	}

	for i <= int(months) {
		var statmentRows []Statement
		sdate := fmt.Sprintf("%d-%02d-01", startDate.Year(), startDate.Month())
		firstDayofTheMonth, _ := time.Parse(TIMELAYOUT, sdate)
		//last day of the current month
		endOfThisMonth := time.Date(startDate.Year(), startDate.Month()+1, 0, 0, 0, 0, 0, startDate.Location())

		st := &Statement{
			AccountNum: excel.AccountNum,
			StartDate:  sdate,
			EndDate:    endOfThisMonth.Format(TIMELAYOUT),
		}
		sql := st.SqlStatement(buffaloCtx)

		GormDB.Raw(sql).Scan(&statmentRows)
		fmt.Printf("DATE: %s ROWS: %d\n\n\n", startDate.Format(TIMELAYOUT), len(statmentRows))

		//no transaction for that month
		if len(statmentRows) < 1 {

			fmt.Printf("DATE: %s Notrans\n\n\n", startDate.Format(TIMELAYOUT))
			gl := &GeneralLedger{
				AccountNum: excel.AccountNum,
				StartDate:  excel.StartDate,
				EndDate:    excel.EndDate,
			}
			tranAmt, _ := gl.FindOpenBalance(firstDayofTheMonth)

			postDate = endOfThisMonth.Format(TIMELAYOUT)
			valueDate = endOfThisMonth.Format(TIMELAYOUT)
			narration = strings.ToUpper(fmt.Sprintf("LAST DAY OF THE MONTH (%s %d)", startDate.Month().String(), startDate.Year()))
			amount = tranAmt * -1
			transType := st.CheckAmount(amount)

			if transType == "Deposit" {
				crAmount = php2go.NumberFormat(php2go.Round(php2go.Abs(amount), 2), 2, ".", ",")
				dbAmount = "-"
				openBal = amount
				totalCredit += amount
				cr++

			} else {
				crAmount = "-"
				dbAmount = php2go.NumberFormat(php2go.Round(php2go.Abs(amount), 2), 2, ".", ",")
				openBal = amount
				totalDebit += amount
				dr++
			}

			//table = table + "<tr><td style='width:80px;text-align:center;padding:7px;' class='tr_td'>" + postDate + "</td>" +
			//	"<td style='width:80px;text-align:center;padding:7px;' class='tr_td'>" + valueDate + "</td>" +
			//	"<td style='width:510px;text-align:left;padding:7px;' class='tr_td'>" + narration + "</td>" +
			//	"<td style='width:146px;text-align:right;padding:7px;' class='tr_td'>" + dbAmount + "</td>" +
			//	"<td style='width:139px;text-align:right;padding:7px;' class='tr_td'>" + crAmount + "</td>" +
			//	"<td style='width:140px;text-align:right;padding:7px;' class='tr_td'>" + php2go.NumberFormat(php2go.Round(openBal, 2), 2, ".", ",") + "</td></tr>"

			rowCount++
		} else {
			for _, row := range statmentRows {
				//c_ccy = row.C_ccy_dsc
				postDate = row.D_TxnPostDt.Format(TIMELAYOUT)
				valueDate = row.D_TxnValueDt.Format(TIMELAYOUT)
				narration = fmt.Sprintf("%s %s %s", row.C_TxnNar1, row.C_TxnNar2, row.C_TxnNar3)
				amount = row.M_TxnAmt * -1
				transType := st.CheckAmount(amount)

				if transType == "Deposit" {
					crAmount = php2go.NumberFormat(php2go.Round(php2go.Abs(amount), 2), 2, ".", ",")
					dbAmount = "-"
					openBal += amount
					totalCredit += amount
					cr++

				} else {
					dbAmount = php2go.NumberFormat(php2go.Round(php2go.Abs(amount), 2), 2, ".", ",")
					crAmount = "-"
					openBal += amount
					totalDebit += amount
					dr++
				}

				table = table + "<tr><td style='width:80px;text-align:center;padding:7px;' class='tr_td'>" + postDate + "</td>" +
					"<td style='width:80px;text-align:center;padding:7px;' class='tr_td'>" + valueDate + "</td>" +
					"<td style='width:510px;text-align:left;padding:7px;' class='tr_td'>" + narration + "</td>" +
					"<td style='width:146px;text-align:right;padding:7px;' class='tr_td'>" + dbAmount + "</td>" +
					"<td style='width:139px;text-align:right;padding:7px;' class='tr_td'>" + crAmount + "</td>" +
					"<td style='width:140px;text-align:right;padding:7px;' class='tr_td'>" + php2go.NumberFormat(php2go.Round(openBal, 2), 2, ".", ",") + "</td></tr>"

				rowCount++
			}
		}

		startDate = time.Date(startDate.Year(), startDate.Month()+1, 1, 0, 0, 0, 0, startDate.Location())

		i++
	}

	table = table + "</table>"

	//filePath := "./account statement.xls"
	// filePath := fmt.Sprintf("./%s.xls", excel.AccountNum)
	// f, err := os.Create(filePath)
	// f.WriteString(table)
	// f.Sync()

	dirPath := envy.Get("GH_ACCOUNT_STATEMENT_TEMP", "./")
	filePath := fmt.Sprintf("%s_%s_account_statement.xls", excel.AccountNum, user.ID)
	excelPath := fmt.Sprintf("%s/%s", dirPath, filePath)

	WriteToFile(table, excelPath)

	fmt.Println("------------Row Count: ", rowCount)
	numPages := math.Round(float64(rowCount / 10))

	//Excel Audit
	sdate, _ := time.Parse("2006-01-02", excel.StartDate)
	edate, _ := time.Parse("2006-01-02", excel.EndDate)
	PrintType := "Excel"
	LogPrintActivity(GormDB, fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		user.Branch.Name, excel.AccountNum, customerName, int64(numPages), PrintType, sdate, edate)

	return excelPath, nil
}
