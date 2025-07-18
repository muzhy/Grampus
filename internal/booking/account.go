package booking

import (
	"fmt"

	"Grampus/internal"
)

type AccountType string

const (
	Asset     AccountType = "asset"
	Liability AccountType = "liability"
	Equity    AccountType = "equity"
	Revenue   AccountType = "revenue"
	Expense   AccountType = "expense"
)

func (at AccountType) IsValid() bool {
	switch at {
	case Asset, Liability, Equity, Revenue, Expense:
		return true
	default:
		return false
	}
}

type Account struct {
	ID           string `json:"id" db:"id"`
	BookID       string `json:"book_id" db:"book_id" binding:"required"`
	ParentID     string `json:"parent_id" db:"parent_id,omitempty"`
	CommID       string `json:"comm_id" db:"comm_id"`
	Name         string `json:"name" db:"name" binding:"required"`
	Type         string `json:"type" db:"type" binding:"required"`
	CurrentNum   int64  `json:"current_num" db:"current_num"`
	CurrentDenom int64  `json:"current_denom" db:"current_denom"`
	CreatedAt    string `json:"created_at" db:"created_at"`
}

// 由科目组成的科目树，需要从数据库中读取后在程序进行组装
type AccountTree struct {
	Data   Account
	Childs []AccountTree
}

type Book struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name" binding:"required"`
	DefalutCurr string `json:"defalut_curr" db:"defalut_curr"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	// 一个账簿可以有多个科目树
	Accounts []AccountTree `json:"accounts,omitempty"`
}

func GetAllBooks(repo BookingRepository) ([]Book, error) {
	return repo.GetAllBooks()
}

func (b *Book) Create(repo BookingRepository) error {
	if b.ID != "" {
		return internal.Error{Code: internal.ErrInvalidParam, Msg: "New book id should empty"}
	}

	if b.Name == "" {
		return internal.Error{Code: internal.ErrInvalidParam, Msg: "Book name is required"}
	}
	if b.DefalutCurr == "" {
		// 考虑默认货币设置为CNY
		return internal.Error{Code: internal.ErrInvalidParam, Msg: "Book default currenry is required"}
	}

	b.Accounts = nil
	err := repo.CreateBook(b)

	return err
}

func GetBookByIdWithAccounts(repo BookingRepository, id string) (*Book, error) {
	// accounts, err := repo.GetBookAccounts()
	book, err := repo.GetBookById(id)
	if err != nil {
		return nil, err
	}
	err = book.loadAccounts(repo)
	if err != nil {
		return nil, err
	}

	return book, nil
}

func (b *Book) loadAccounts(repo BookingRepository) error {
	// accounts, err := repo.GetBookAccounts(b.ID)
	// if err != nil {
	// 	return err
	// }
	// var tree AccountTree
	// tree.roots = make([]Account, 0)
	// tree.children = make([]Account, 0)
	// b.Accounts = &tree
	// // 空的科目树
	// if len(accounts) == 0 {
	// 	return nil
	// }

	// accountMap := make(map[string][]Account)
	// for _, account := range accounts {
	// 	if data, ok := accountMap[account.ParentID]; ok {
	// 		accountMap[account.ParentID] = append(data, account)
	// 	} else {
	// 		tmpAccounts := make([]Account, 0)
	// 		accountMap[account.ParentID] = append(tmpAccounts, account)
	// 	}
	// }

	// tree.roots = accountMap[""]
	// if len(tree.roots) == 0 {
	// 	return nil
	// }

	return nil
}

func (at *AccountTree) loadSub(accountMap map[string][]Account) {

}

func (a *Account) Create(repo BookingRepository) error {
	if a.ID != "" {
		return internal.Error{Code: internal.ErrInvalidParam, Msg: "New account id should empty"}
	}
	if !AccountType(a.Type).IsValid() {
		return internal.Error{Code: internal.ErrInvalidParam, Msg: fmt.Sprintf("Account Type:%v valid", a.Type)}
	}

	if a.BookID == "" {
		return internal.Error{Code: internal.ErrInvalidParam, Msg: "Account must belong to book"}
	}
	// bookID是否存在有repo通过数据库判断
	a.CurrentDenom = 0
	a.CurrentNum = 100

	return repo.CreateAccount(a)
}
