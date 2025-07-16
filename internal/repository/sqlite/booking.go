package sqlite

import (
	Grampus "Grampus"
	"Grampus/internal/booking"
	"Grampus/internal/config"
	"embed"

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

func (r *SqliteBooksRepository) CreateBook(book *booking.Book) error {
	_, err := r.db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		zap.L().Error("PRAGMA foreign keys failed", zap.Error(err))
		return Grampus.Error{Code: Grampus.ErrDbError, Err: err, Msg: "RAGMA foreign keys failed"}
	}
	const insertSql = `INSERT INTO book (id, name, defalut_curr) VALUES (?, ?, ?) RETURNING created_at`
	book.ID = uuid.New().String()
	// _, err = r.db.Exec(insertSql, book.ID, book.Name, book.DefalutCurr)
	err = r.db.QueryRowx(insertSql, book.ID, book.Name, book.DefalutCurr).StructScan(book)
	if err != nil {
		if sqliteErr, ok := err.(sqlite.Error); ok {
			switch sqliteErr.ExtendedCode {
			case sqlite.ErrConstraintUnique:
				zap.L().Info("Book name already exists", zap.String("name", book.Name))
				return Grampus.Error{Code: Grampus.ErrAlreadyExists, Err: err, Msg: "book name already exists"}
			case sqlite.ErrConstraintForeignKey:
				zap.L().Info("Defalut currency not found", zap.String("currency", book.DefalutCurr))
				return Grampus.Error{Code: Grampus.ErrInvalidData, Err: err, Msg: "default currency not found"}
			}
		}
		zap.L().Error("Create Book failed", zap.Error(err))
		return Grampus.Error{Code: Grampus.ErrInternal, Err: err, Msg: "Create Book failed"}
	}
	return nil
}
