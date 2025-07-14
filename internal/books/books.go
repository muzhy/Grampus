package books

import (
	"fmt"
	"time"
)

type AccountType string

const (
	Asset     AccountType = "asset"
	Liability AccountType = "liability"
	Equity    AccountType = "equity"
	Revenue   AccountType = "revenue"
	Expense   AccountType = "expense"
	Root      AccountType = "root"
)

// book represents a financial book that contains accounts and transactions.
type Book struct {
	ID          int32  `json:"id" db:"id"`
	Name        string `json:"name" db:"name" binding:"required"`
	CreateAt    string `json:"create_at" db:"create_at"`
	AccountTree *AccountTree
}

type Account struct {
	ID       int32  `json:"id" db:"id"`
	ParentID string `json:"parent_id" db:"parent_id"`
	BookID   int32  `json:"book_id" db:"book_id""`
	Name     string `json:"name" db:"name"`
	Type     string `json:"type" db:"type"`
	CreateAt string `json:"create_at" db:"create_at"`
}

type AccountTree struct {
	Data     Account
	Children []Account
}

type Entry struct {
	ID        int32   `json:"id" db:"id"`
	AccountID int32   `json:"account_id" db:"account_id"`
	Amount    float64 `json:"amount" db:"amount"`
	CreateAt  string  `json:"create_at" db:"create_at"`
	Account   Account
}

type Transaction struct {
	ID               int32   `json:"id" db:"id"`
	Description      string  `json:"description" db:"description"`
	Transaction_date string  `json:"transaction_date" db:"transaction_date"`
	Entries          []Entry `json:"entries" db:"entries"`
}

type BookRepository interface {
	CreateBook(book *Book) error
}

func (b *Book) Create(repo BookRepository) error {
	if b.Name == "" {
		return fmt.Errorf("book name is required")
	}

	if b.CreateAt == "" {
		b.CreateAt = time.Now().Format(time.RFC3339)
	}
	b.AccountTree = nil

	if err := repo.CreateBook(b); err != nil {
		// TODO 检查错误类型，判断是否为唯一约束冲突
		return fmt.Errorf("failed to create book: %w", err)
	}

	return nil
}
