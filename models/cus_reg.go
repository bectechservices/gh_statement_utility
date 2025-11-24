package models

import (
	"fmt"

	"github.com/gobuffalo/buffalo"
)

type CusReg struct {
	C_Cus_ShortName string `json:"C_Cus_ShortName" gorm:"column:C_Cus_ShortName"`
	C_Addr1         string `json:"C_Addr1" gorm:"column:C_Addr1"`
	C_Addr2         string `json:"C_Addr2" gorm:"column:C_Addr2"`
	C_Addr3         string `json:"C_Addr3" gorm:"column:C_Addr3"`
	I_Cus_Code      string `json:"I_Cus_Code" gorm:"column:I_Cus_Code"`
}

var CusRegs []CusReg

func (cusReg *CusReg) GetAddress(cusId string, gl GeneralLedger) (CusReg, error) {
	cusRegModel := &CusReg{}

	if cusId == "" {
		cusId, _ = gl.MoveForward()
	}

	tableDate := gl.IfStartCusTableExist(gl.StartDate, gl.EndDate, cusId)

	sql := fmt.Sprintf(`SELECT C_Cus_ShortName,C_Addr1,C_Addr2,C_Addr3 FROM %s WHERE I_Cus_Code = ? `, tableDate)

	// Execute query
	tx := GormDB.Raw(sql, cusId).Scan(&cusRegModel)

	if tx.Error != nil {
		return *cusRegModel, tx.Error
	}
	return *cusRegModel, nil
}

func (cusReg CusReg) GetCustomerInfo(buffaloCtx buffalo.Context, accountNum string) (CusReg, error) {
	cusRegModel := CusReg{}

	cusTable := buffaloCtx.Session().Get(fmt.Sprintf("cus_table_%s", accountNum))
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!cusTable1: ", cusTable)

	cusId := buffaloCtx.Session().Get(fmt.Sprintf("cus_id_%s", accountNum))
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!cusId: ", cusId)
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!accountNum: ", accountNum)

	sql := fmt.Sprintf(`
		SELECT C_Addr1,C_Addr2,C_Addr3,C_Cus_ShortName FROM %s where I_Cus_Code= '%s'
		`, cusTable, cusId)

	// Execute query
	tx := GormDB.Raw(sql).Scan(&cusRegModel)

	if tx.Error != nil {
		return cusRegModel, tx.Error
	}
	return cusRegModel, nil
}
