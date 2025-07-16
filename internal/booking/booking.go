package booking

import (
	Grampus "Grampus"
)

type BookingRepository interface {
	CreateBook(book *Book) error
}

func (b *Book) Create(repo BookingRepository) error {
	if b.ID != "" {
		return Grampus.Error{Code: Grampus.ErrInvalidParam, Msg: "New book id should empty"}
	}

	if b.Name == "" {
		return Grampus.Error{Code: Grampus.ErrInvalidParam, Msg: "Book name is required"}
	}
	if b.DefalutCurr == "" {
		// 考虑默认货币设置为CNY
		return Grampus.Error{Code: Grampus.ErrInvalidParam, Msg: "Book default currenry is required"}
	}

	b.Accounts = nil
	err := repo.CreateBook(b)

	return err
}
