package booking

type AccountType string

const (
	Asset     AccountType = "asset"
	Liability AccountType = "liability"
	Equity    AccountType = "equity"
	Revenue   AccountType = "revenue"
	Expense   AccountType = "expense"
)

type Account struct {
	ID           string `json:"id" db:"id"`
	BookID       string `json:"book_id" db:"book_id"`
	ParentID     string `json:"parent_id" db:"parent_id"`
	CommID       string `json:"comm_id" db:"comm_id"`
	Name         string `json:"name" db:"name"`
	Type         string `json:"type" db:"type"`
	CurrentNum   int64  `json:"current_num" db:"current_num"`
	CurrentDenom int64  `json:"current_denom" db:"current_denom"`
	CreatedAt    string `json:"created_at" db:"created_at"`
}

// 由科目组成的科目树，需要从数据库中读取后在程序进行组装
type AccountTree struct {
	Data     Account
	Children []Account
}

type Book struct {
	ID          string       `json:"id" db:"id"`
	Name        string       `json:"name" db:"name" binding:"required"`
	DefalutCurr string       `json:"defalut_curr" db:"defalut_curr"`
	CreatedAt   string       `json:"created_at" db:"created_at"`
	Accounts    *AccountTree `json:"accounts"`
}
