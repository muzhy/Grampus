package booking

type Trans struct {
	ID          string `json:"id" db:"id"`
	BookID      string `json:"book_id" db:"book_id"`
	TransDate   string `json:"trans_date" db:"trans_date"`
	Description string `json:"description" db:"description"`
	Items       []TransItem
}

type TransItem struct {
	ID          string  `json:"id" db:"id"`
	TransID     string  `json:"trans_id" db:"trans_id"`
	AccountID   string  `json:"account_id" db:"account_id"`
	Desc        *string `json:"desc" db:"desc"`
	AmountNum   int64   `json:"amount_num" db:"amount_num"`
	AmountDenom int64   `json:"amount_denom" db:"amount_denom"`
	CreatedAt   string  `json:"created_at" db:"created_at"`
}
