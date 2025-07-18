package booking

type BookingRepository interface {
	CreateBook(book *Book) error
	GetAllBooks() ([]Book, error)
	GetBookById(id string) (*Book, error)

	CreateAccount(account *Account) error
	GetAccountById(id string) (*Account, error)
	GetBookAccounts(id string) ([]Account, error)
}
