package repository

import (
	"Grampus/internal/config"
	"Grampus/internal/ledger"
	"Grampus/pkg"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"github.com/muzhy/lapluma/iterator"
	"github.com/stretchr/testify/assert"
)

func setupSqlite(t *testing.T) ledger.Repo {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "grampus.db")
	dbConfig := config.DatabaseConfig{
		Driver: "sqlite3",
		DSN:    dbPath,
	}

	repo, err := NewSqliteRepo(&dbConfig)
	if err != nil {
		panic(err)
	}
	t.Cleanup(func() {
		if r, ok := repo.(*SqliteRepo); ok {
			r.Close()
		}
	})

	return repo
}

func TestCreateLedger(t *testing.T) {
	repo := setupSqlite(t)

	ledgerName := "Test"
	led := ledger.Ledger{
		Name:        ledgerName,
		DefalutCurr: "XXX",
	}
	assert := assert.New(t)
	// 违反默认货币外键约束
	err := repo.CreateLedger(&led)
	assert.NotNil(err)
	serr, ok := err.(sqlite3.Error)
	assert.True(ok)
	assert.Equal(sqlite3.ErrConstraintForeignKey, serr.ExtendedCode)
	// 符合外键约束
	led.DefalutCurr = string(ledger.CNY)

	err = repo.CreateLedger(&led)
	assert.Nil(err, "Create ledger in sqlite failed")
	assert.NotEmpty(led.ID, "Create ledger but get empty id")
	assert.NotEmpty(led.CreatedAt, "Create ledget but get empty createAt")

	// 违反名字唯一性约束
	led2 := ledger.Ledger{
		Name:        ledgerName,
		DefalutCurr: string(ledger.CNY),
	}
	err = repo.CreateLedger(&led2)
	assert.NotNil(err)
	serr, ok = err.(sqlite3.Error)
	assert.True(ok)
	assert.Equal(sqlite3.ErrConstraintUnique, serr.ExtendedCode)
}

func setupTestLedger(repo ledger.Repo, name string) *ledger.Ledger {
	led := ledger.Ledger{
		Name:        name,
		DefalutCurr: string(ledger.CNY),
	}

	// 这里的正确性应该已经由TestCreateBook保证， 不做错误判断
	repo.CreateLedger(&led)
	return &led
}

func TestGetLedger(t *testing.T) {
	repo := setupSqlite(t)

	l1 := setupTestLedger(repo, "demo1")
	it, err := repo.GetAllLedger()

	assert.Nil(t, err)
	v, ok := it.Next()
	assert.True(t, ok)
	assert.Nil(t, v.Err)
	assert.Equal(t, v.Value.ID, l1.ID)

	v, ok = it.Next()
	assert.False(t, ok)
	assert.Nil(t, it.Close())

	l2 := setupTestLedger(repo, "demo2")

	it, err = repo.GetAllLedger()
	assert.Nil(t, err)
	_, ok = it.Next()
	assert.True(t, ok)
	_, ok = it.Next()
	assert.True(t, ok)
	_, ok = it.Next()
	assert.False(t, ok)
	assert.Nil(t, it.Close())

	demo2, err := repo.GetLedgerById(l2.ID)
	assert.Nil(t, err)
	assert.Equal(t, l2.ID, demo2.ID)

	tmpID := uuid.New().String()
	tmp, err := repo.GetLedgerById(tmpID)
	assert.NotNil(t, err)
	assert.Nil(t, tmp)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestCreateAccount(t *testing.T) {
	repo := setupSqlite(t)
	account := ledger.Account{
		ID:           "",
		LedgerID:     uuid.New().String(),
		ParentID:     "",
		GoodsID:      "",
		Name:         "testAccount",
		Type:         string(ledger.Asset),
		CurrentNum:   0,
		CurrentDenom: 10000,
	}
	// 所属账本不存在
	err := repo.CreateAccount(&account)
	assert.NotNil(t, err)
	assert.Equal(t, pkg.ErrLedgerNotExist, err)
	// 准备账本
	l := setupTestLedger(repo, "test")
	account.LedgerID = l.ID
	// parentID有填, 但是不存在
	account.ParentID = uuid.New().String()
	err = repo.CreateAccount(&account)
	assert.Equal(t, pkg.ErrParentAccountNotExist, err)
	// 插入根科目
	account.ParentID = ""
	err = repo.CreateAccount(&account)
	assert.Nil(t, err)
	assert.Equal(t, l.DefalutCurr, account.GoodsID)
	assert.Equal(t, 0, account.Level)
	assert.NotEmpty(t, account.CreatedAt)
	assert.Equal(t, account.Name, account.Path)
	// 插入子科目
	subAccount := ledger.Account{
		ID:           "",
		LedgerID:     l.ID,
		ParentID:     account.ID,
		GoodsID:      "",
		Name:         "subAccount",
		Type:         string(ledger.Asset),
		CurrentNum:   0,
		CurrentDenom: 10000,
	}
	err = repo.CreateAccount(&subAccount)
	assert.Nil(t, err)
	assert.Equal(t, 1, subAccount.Level)
	assert.NotEmpty(t, subAccount.CreatedAt)
	assert.Equal(t, subAccount.Path, "testAccount/subAccount")

	rootAccount, err := repo.GetAccountById(account.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, rootAccount.Level)
	assert.Equal(t, account.ID, rootAccount.ID)
	assert.Equal(t, account, *rootAccount)
	// 获取账簿下的所有子科目
	accountIt, err := repo.GetLedgerAccounts(l.ID)
	assert.Nil(t, err)
	for account := range iterator.Iter(accountIt) {
		assert.Nil(t, account.Err)
		assert.Equal(t, account.Value.LedgerID, l.ID)
		assert.NotEmpty(t, account.Value.Path)
	}
}
