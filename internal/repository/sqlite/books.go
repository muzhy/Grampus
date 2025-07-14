package sqlite

import (
	"Grampus/internal/books"
	"Grampus/internal/config"
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

//go:embed *.sql
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
	sourceDriver, err := iofs.New(migrationsFS, r.config.DSN)
	if err != nil {
		return err
	}

	dbDriver, err := sqlite3.WithInstance(r.db.DB, &sqlite3.Config{})
	if err != nil {
		return err
	}

	// 5. 使用新的源驱动来创建 migrate 实例
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite3", dbDriver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil {
		return err
	}

	zap.L().Info("Init Books DB success!")

	return nil
}

func (r *SqliteBooksRepository) CreateBook(book *books.Book) error {
	return nil
}
