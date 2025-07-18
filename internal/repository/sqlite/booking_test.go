package sqlite_test

import (
	"Grampus/internal"
	"Grampus/internal/booking"
	"Grampus/internal/config"
	"Grampus/internal/repository"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	sqlite "github.com/mattn/go-sqlite3"
)

func setupSqlite(t *testing.T) *repository.Repository {
	// 为每次测试用例创建一个单独的目录，避免干扰
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "Grampus.db")
	dbConfig := config.DatabaseConfig{
		Driver: "sqlite3",
		DSN:    dbPath,
	}
	repo := repository.NewRepository(&dbConfig)

	t.Cleanup(func() {
		repo.Close()
	})

	return repo
}

func TestCreateBook(t *testing.T) {
	repo := setupSqlite(t)
	bookName := "Test"
	bookCurr := string(booking.CNY)
	book := booking.Book{
		Name:        bookName,
		DefalutCurr: bookCurr,
	}

	// err := book.Create(repo.BookRepository)
	err := repo.BookRepository.CreateBook(&book)
	if err != nil {
		t.Errorf("Create book to sqlite failed:%v", book)
	}

	if book.ID == "" {
		t.Errorf("Create book but id is empty:%v", book)
	}

	if book.Name != bookName || book.DefalutCurr != bookCurr {
		t.Error("Create book but name or defaultCurr not expect")
	}

	if book.CreatedAt == "" {
		t.Error("Create book but createAt empty")
	}
}

func TestCreateBookForeignError(t *testing.T) {
	// 创建账簿时，默认货币必须符合外键约束
	repo := setupSqlite(t)
	bookName := "Test"
	bookCurr := "XXX"
	book := booking.Book{
		Name:        bookName,
		DefalutCurr: bookCurr,
	}

	err := repo.BookRepository.CreateBook(&book)
	if err == nil {
		t.Errorf("Create book use not exist currency:%v, but get empty error", book.DefalutCurr)
	}

	if gErr, ok := err.(internal.Error); ok {
		if gErr.Code != internal.ErrInvalidParam {
			t.Errorf("Create book use not exist currency:%v, expect get code:%v, get:%v",
				book.DefalutCurr, internal.ErrInvalidParam, gErr.Code)
		}
		if sErr, ok := gErr.Err.(sqlite.Error); ok {
			if sErr.ExtendedCode != sqlite.ErrConstraintForeignKey {
				t.Errorf("Create book use not exist currency:%v, expect get extendedCode:%v, get:%v",
					book.DefalutCurr, sqlite.ErrConstraintForeignKey, sErr.ExtendedCode)
			}
		}
	}
}

func TestCreateBookUniqueError(t *testing.T) {
	repo := setupSqlite(t)
	bookName := "Test"
	bookCurr := string(booking.CNY)

	book1 := booking.Book{
		Name:        bookName,
		DefalutCurr: bookCurr,
	}
	book2 := booking.Book{
		Name:        bookName,
		DefalutCurr: bookCurr,
	}

	if err := repo.BookRepository.CreateBook(&book1); err != nil {
		t.Errorf("Create book failed. book:%v", book1)
	}

	err := repo.BookRepository.CreateBook(&book2)
	if err == nil {
		t.Errorf("Create book with same name should failed, but err is nil, book1:%v, book2:%v",
			book1, book2)
	}

	if gErr, ok := err.(internal.Error); ok {
		if gErr.Code != internal.ErrAlreadyExists {
			t.Errorf("Create book use sanme book name:%v, expect get code:%v, get:%v",
				bookName, internal.ErrAlreadyExists, gErr.Code)
		}
		if sErr, ok := gErr.Err.(sqlite.Error); ok {
			if sErr.ExtendedCode != sqlite.ErrConstraintUnique {
				t.Errorf("Create book use sanme book name:%v expect get extendedCode:%v, get:%v",
					bookName, sqlite.ErrConstraintUnique, sErr.ExtendedCode)
			}
		}
	}
}

func setupTestBook(t *testing.T, repo booking.BookingRepository, bookName string) booking.Book {
	bookCurr := string(booking.CNY)
	book := booking.Book{
		Name:        bookName,
		DefalutCurr: bookCurr,
	}

	// 这里的正确性应该已经由TestCreateBook保证， 不做错误判断
	repo.CreateBook(&book)
	return book
}

func TestCreateAccoutWithNotExistBook(t *testing.T) {
	repo := setupSqlite(t)
	account := booking.Account{
		ID:           uuid.New().String(),
		BookID:       uuid.New().String(),
		ParentID:     "",
		CommID:       "",
		Name:         "testAccount",
		Type:         string(booking.Asset),
		CurrentNum:   0,
		CurrentDenom: 10000,
	}
	err := repo.BookRepository.CreateAccount(&account)
	if err == nil {
		t.Error("create accout to not exist book but success")
	}

	if gerr, ok := err.(internal.Error); ok {
		if gerr.Code != internal.ErrInvalidParam {
			t.Errorf("should return %v when book not exist", gerr.Code)
		}
	} else {
		t.Error("should return internal.Error")
	}
}

func TestCreateRootAccount(t *testing.T) {
	repo := setupSqlite(t)

	book := setupTestBook(t, repo.BookRepository, "test")

	account := booking.Account{
		ID:           uuid.New().String(),
		BookID:       book.ID,
		ParentID:     "",
		CommID:       "",
		Name:         "testAccount",
		Type:         string(booking.Asset),
		CurrentNum:   0,
		CurrentDenom: 10000,
	}

	// err := account.Create(repo.BookRepository)
	err := repo.BookRepository.CreateAccount(&account)
	if err != nil {
		t.Errorf("Create account failed, account:%v", account)
	}

	if account.CreatedAt == "" {
		t.Errorf("Create account, but not set create_at, account:%v", account)
	}

	if account.CommID != book.DefalutCurr {
		t.Errorf("create account, but bot use book default currency, account:%v, book:%v",
			account, book)
	}

	readAccount, err := repo.BookRepository.GetAccountById(account.ID)
	if err != nil {
		t.Errorf("Get created account failed")
	}

	if readAccount.CreatedAt != account.CreatedAt || readAccount.CommID != account.CommID ||
		readAccount.BookID != account.BookID {
		t.Errorf("readAccount is not equal account, readAccount:%v, insert account %v",
			readAccount, account)
	}
}

func TestCreateSubAccount(t *testing.T) {
	repo := setupSqlite(t)
	book := setupTestBook(t, repo.BookRepository, "test")

	rootAccount := booking.Account{
		ID:           uuid.New().String(),
		BookID:       book.ID,
		ParentID:     "",
		CommID:       "",
		Name:         "rootAccount",
		Type:         string(booking.Asset),
		CurrentNum:   0,
		CurrentDenom: 10000,
	}

	// err := account.Create(repo.BookRepository)
	err := repo.BookRepository.CreateAccount(&rootAccount)
	if err != nil {
		t.Errorf("Create root account failed, account:%v", rootAccount)
	}

	subAccount := booking.Account{
		ID:           uuid.New().String(),
		BookID:       book.ID,
		ParentID:     rootAccount.ID,
		Name:         "subAccount",
		Type:         rootAccount.Type,
		CommID:       "CNY",
		CurrentNum:   0,
		CurrentDenom: 10000,
	}

	err = repo.BookRepository.CreateAccount(&subAccount)
	if err != nil {
		t.Errorf("Create sub account failed, account:%v", subAccount)
	}
	// 属于不同book的account不允许作为父级account
	book2 := setupTestBook(t, repo.BookRepository, "test2")
	subAccount2 := booking.Account{
		ID:           uuid.New().String(),
		BookID:       book2.ID,
		ParentID:     rootAccount.ID,
		Name:         "subAccount",
		Type:         rootAccount.Type,
		CommID:       "CNY",
		CurrentNum:   0,
		CurrentDenom: 10000,
	}
	err = repo.BookRepository.CreateAccount(&subAccount2)
	if err == nil {
		t.Errorf("Create sub account success which parent account and sub account in diff book")
	}
}
