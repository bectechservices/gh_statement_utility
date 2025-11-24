package requests

type OnlyID struct {
	ID string `form:"id" json:"id"`
}

type EmailRequest struct {
	Email string `form:"email" json:"email"`
}

// type statementAudit struct {
// 	ID          string `form:"id" json:"id"`
// 	BranchName  string `form:"branch_name" json:"branch_name"`
// 	AccountNum  string `form:"account_no" json:"account_no"`
// 	AccountName string `form:"account_name" json:"account_name"`
// }
