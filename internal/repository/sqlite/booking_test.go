package sqlite_test

import (
	Grampus "Grampus"
	"Grampus/internal/booking"
	"Grampus/internal/config"
	"Grampus/internal/repository"
	"path/filepath"
	"testing"

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

	if gErr, ok := err.(Grampus.Error); ok {
		if gErr.Code != Grampus.ErrInvalidData {
			t.Errorf("Create book use not exist currency:%v, expect get code:%v, get:%v",
				book.DefalutCurr, Grampus.ErrInvalidData, gErr.Code)
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

	if gErr, ok := err.(Grampus.Error); ok {
		if gErr.Code != Grampus.ErrAlreadyExists {
			t.Errorf("Create book use sanme book name:%v, expect get code:%v, get:%v",
				bookName, Grampus.ErrAlreadyExists, gErr.Code)
		}
		if sErr, ok := gErr.Err.(sqlite.Error); ok {
			if sErr.ExtendedCode != sqlite.ErrConstraintUnique {
				t.Errorf("Create book use sanme book name:%v expect get extendedCode:%v, get:%v",
					bookName, sqlite.ErrConstraintUnique, sErr.ExtendedCode)
			}
		}
	}
}
