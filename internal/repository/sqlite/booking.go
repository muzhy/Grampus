package sqlite

import (
	"Grampus/internal"
	"Grampus/internal/booking"
	"Grampus/internal/config"
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	sqlite "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type SqliteBooksRepository struct {
	db     *sqlx.DB
	config *config.DatabaseConfig
}

func NewSqliteBooksRepository(db *sqlx.DB, config *config.DatabaseConfig) *SqliteBooksRepository {
	return &SqliteBooksRepository{
		db:     db,
		config: config,
	}
}

// init sqlite datebase
func (r *SqliteBooksRepository) InitBooksDB() error {
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	dbDriver, err := sqlite3.WithInstance(r.db.DB, &sqlite3.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite3", dbDriver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	if err == migrate.ErrNoChange {
		zap.L().Info("Init Books DB: no changes to apply")
	} else {
		zap.L().Info("Init Books DB success!")
	}

	return nil
}

func (r *SqliteBooksRepository) pramaForeignKeys() error {
	_, err := r.db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		zap.L().Error("PRAGMA foreign keys failed", zap.Error(err))
		return internal.Error{Code: internal.ErrDbError, Err: err, Msg: "RAGMA foreign keys failed"}
	}
	return nil
}

func handleInserError(err error, data any) error {
	if serr, ok := err.(sqlite.Error); ok {
		switch serr.ExtendedCode {
		case sqlite.ErrConstraintUnique:
			zap.L().Info("data already exists", zap.Any("data", data))
			return internal.Error{Code: internal.ErrAlreadyExists, Err: err, Msg: "Data already exist"}
		case sqlite.ErrConstraintForeignKey:
			zap.L().Info("Foreign key not exists", zap.Any("data", data))
			return internal.Error{Code: internal.ErrInvalidParam, Err: err, Msg: "Foreign key not exists"}
		}
	}
	return internal.Error{Code: internal.ErrDbError, Err: err, Msg: "insert data to sqlite failed"}
}

func (r *SqliteBooksRepository) CreateBook(book *booking.Book) error {
	if err := r.pramaForeignKeys(); err != nil {
		return err
	}
	const insertSql = `INSERT INTO book (id, name, defalut_curr) VALUES (?, ?, ?) RETURNING created_at`
	book.ID = uuid.New().String()
	err := r.db.QueryRowx(insertSql, book.ID, book.Name, book.DefalutCurr).StructScan(book)
	if err != nil {
		return handleInserError(err, book)
	}
	return nil
}

func (r *SqliteBooksRepository) GetAllBooks() ([]booking.Book, error) {
	const querySql = "SELECT id, name, defalut_curr, created_at FROM book;"
	books := make([]booking.Book, 0)
	err := r.db.Select(&books, querySql)
	if err != nil {
		zap.L().Error("Select books failed.", zap.Error(err))
		return nil, internal.Error{Code: internal.ErrInternal, Err: err, Msg: "select books failed"}
	}
	return books, nil
}

func queryOne[T any](db sqlx.Ext, querySql string, args ...interface{}) (*T, error) {
	var res T
	err := db.QueryRowx(querySql, args...).StructScan(&res)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, internal.Error{Code: internal.ErrDbEmpty, Err: err, Msg: "Query data not exist"}
		}
		return nil, internal.Error{Code: internal.ErrDbError, Err: err, Msg: "Query from database failed"}
	}
	return &res, nil
}

func getBookById(db sqlx.Ext, id string) (*booking.Book, error) {
	const querySql = "SELECT id, name, defalut_curr, created_at FROM book where id = ?"
	return queryOne[booking.Book](db, querySql, id)
}

func (r *SqliteBooksRepository) GetBookById(id string) (*booking.Book, error) {
	return getBookById(r.db, id)
}

func getAccountById(db sqlx.Ext, id string) (*booking.Account, error) {
	const querySql = "SELECT id, book_id, parent_id, comm_id, name, type, current_num, current_denom, created_at FROM account WHERE id = ?"
	return queryOne[booking.Account](db, querySql, id)
}

func (r *SqliteBooksRepository) GetAccountById(id string) (*booking.Account, error) {
	return getAccountById(r.db, id)
}

func (r *SqliteBooksRepository) CreateAccount(account *booking.Account) (retErr error) {
	// 启用外键约束
	if err := r.pramaForeignKeys(); err != nil {
		return err
	}
	// 必须开启事务，涉及到父级accont, book的校验
	// tx, err := r.db.Begin()
	tx, err := r.db.BeginTxx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return internal.Error{Code: internal.ErrDbError, Err: err, Msg: "Begin sqlite transaction failed!"}
	}

	// needRollBack := false
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
		if retErr != nil {
			tx.Rollback()
		}
	}()

	book, err := getBookById(tx, account.BookID)
	if err != nil {
		if gerr, ok := err.(internal.Error); ok {
			if gerr.Code == internal.ErrDbEmpty {
				return internal.Error{Code: internal.ErrInvalidParam, Msg: "Account's BookId not exist"}
			}
		}
		return err
	}
	// 没有指定账户对应的商品，使用默认货币
	if account.CommID == "" {
		account.CommID = book.DefalutCurr
	}
	// 检查CommId是否存在由数据库的外键约束判断，减少一次IO交互

	// 有父级account时，需要查询父级account，检查父级account是否属于同一个book
	if account.ParentID != "" {
		// parentAccount, err := r.GetAccountById(account.ParentID)
		parentAccount, err := getAccountById(tx, account.ParentID)
		if err != nil {
			if gErr, ok := err.(internal.Error); ok {
				if gErr.Code == internal.ErrDbEmpty {
					zap.L().Info("account had parent, but cannot found parent",
						zap.String("parent_id", account.ParentID))
					return internal.Error{Code: internal.ErrInvalidParam, Msg: "not found parent account"}
				}
			}
			return err
		}
		if parentAccount.BookID != account.BookID {
			return internal.Error{Code: internal.ErrInvalidParam,
				Msg: fmt.Sprintf("Current account bookid(%v) is equal to parent account bookid(%v)",
					account.BookID, parentAccount.BookID)}
		}
	}

	account.ID = uuid.New().String()
	const insertSql = `INSERT INTO account ( id, book_id, parent_id, comm_id, name, type, current_num, current_denom )
		VALUES (:id, :book_id, :parent_id, :comm_id, :name, :type, :current_num, :current_denom)
		RETURNING created_at`

	rows, err := tx.NamedQuery(insertSql, account)
	if err != nil {
		return handleInserError(err, account)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&account.CreatedAt); err != nil {
			zap.L().Error("struct scan to account create_at failed.")
			return internal.Error{Code: internal.ErrDbError, Msg: "Scan error", Err: err}
		}
	}

	if cerr := tx.Commit(); cerr != nil {
		zap.L().Error("Commit failed", zap.Error(cerr))
		return internal.Error{Code: internal.ErrDbError, Err: cerr, Msg: "Sqlite commit failed"}
	}

	return nil
}

func (r *SqliteBooksRepository) GetBookAccounts(id string) ([]booking.Account, error) {

	return nil, nil
}
