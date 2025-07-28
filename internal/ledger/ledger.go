package ledger

import (
	"Grampus/pkg"
	"strings"

	"github.com/muzhy/lapluma"
	"github.com/muzhy/lapluma/iterator"
	"go.uber.org/zap"
)

type CURD interface {
	Create(repo Repo) error
}

const MaxAccountLevel = 10

type AccountType string

const (
	Asset     AccountType = "asset"
	Liability AccountType = "liability"
	Equity    AccountType = "equity"
	Revenue   AccountType = "revenue"
	Expense   AccountType = "expense"
)

func (at AccountType) IsValid() bool {
	switch at {
	case Asset, Liability, Equity, Revenue, Expense:
		return true
	default:
		return false
	}
}

type Ledger struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name" binding:"required"`
	DefalutCurr string `json:"defalut_curr" db:"defalut_curr"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	// 一个账簿可以有多个科目树，记录的视科目树的根节点
	Accounts []Account `json:"accounts,omitempty"`
}

type Account struct {
	ID           string             `json:"id" db:"id"`
	LedgerID     string             `json:"ledger_id" db:"ledger_id" binding:"required"`
	ParentID     string             `json:"parent_id,omitempty" db:"parent_id,omitempty"`
	Level        int                `json:"level" db:"level"`
	GoodsID      string             `json:"goods_id" db:"goods_id"`
	Name         string             `json:"name" db:"name" binding:"required"`
	Path         string             `json:"path" db:"path"`
	Type         string             `json:"type" db:"type" binding:"required"`
	CurrentNum   int64              `json:"current_num" db:"current_num"`
	CurrentDenom int64              `json:"current_denom" db:"current_denom"`
	CreatedAt    string             `json:"created_at" db:"created_at"`
	Childs       map[string]Account `json:"childs,omitempty"`
}

func GetAllLedger(repo Repo) ([]Ledger, error) {
	it, err := repo.GetAllLedger()
	if err != nil {
		return nil, err
	}

	data := iterator.Collect(UnwrapRepoIt(it))
	it.Close()
	return data, nil
}

func (l *Ledger) Create(repo Repo) error {
	if l.Name == "" {
		return pkg.ErrInvalidParam
	}

	if l.DefalutCurr == "" {
		l.DefalutCurr = string(CNY)
	}

	l.Accounts = nil
	err := repo.CreateLedger(l)

	return err
}

func (a *Account) Create(repo Repo) error {
	if a.LedgerID == "" || !AccountType(a.Type).IsValid() || strings.Contains(a.Name, "/") {
		return pkg.ErrInvalidParam
	}

	return repo.CreateAccount(a)
}

func (a *Account) build(datasource map[string][]Account, level int) {
	if level > MaxAccountLevel {
		// 在保存数据时应该已经确保了此约束, 这里多加一层检查,防止递归爆栈
		zap.L().Error(
			"build account exceed max account level",
			zap.Int("max_account_level", MaxAccountLevel),
			zap.Int("current_level", level),
		)
		return
	}
	childs, ok := datasource[a.ID]
	if ok {
		a.Childs = make(map[string]Account)
		for i := range childs {
			subAccount := childs[i]
			subAccount.build(datasource, level+1)
			// 这里需要保证同个账簿中的科目名唯一, 由repo层在插入数据时保证
			a.Childs[subAccount.Name] = subAccount
		}
	} else {
		a.Childs = make(map[string]Account)
	}
}

func buildAccountTrees(accountIt RepoIterator[Account]) []Account {
	accountMap := iterator.Group(
		iterator.Filter(accountIt, func(v lapluma.Result[Account]) bool {
			if v.Err != nil {
				zap.L().Warn("scan account result error", zap.Error(v.Err))
				return false
			}
			return true
		}),
		func(v lapluma.Result[Account]) (string, Account) {
			key := v.Value.ParentID
			return key, v.Value
		},
	)

	rootAccounts, ok := accountMap[""]
	if !ok {
		// 没有根科目
		return make([]Account, 0)
	} else {
		for i := range rootAccounts {
			rootAccounts[i].build(accountMap, 0)
		}
	}
	return rootAccounts
}

// 加载账簿下的所有科目并组成好树结构
func (l *Ledger) loadAccounts(repo Repo) error {
	accountIt, err := repo.GetLedgerAccounts(l.ID)
	if err != nil {
		return err
	}

	defer accountIt.Close()
	l.Accounts = buildAccountTrees(accountIt)

	return nil
}

// 根据id获取包含科目树的账本
// 由于科目树加载比较消耗资源, 不允许批量加载
// 账本名称已经在数据库层做好唯一性约束
func GetLedgerWithAccountsByName(repo Repo, ledgerName string) (*Ledger, error) {
	if ledgerName == "" {
		return nil, pkg.ErrInvalidParam
	}
	l, err := repo.GetLedgerByName(ledgerName)
	if err != nil {
		return nil, err
	}
	err = l.loadAccounts(repo)
	if err != nil {
		return nil, err
	}
	return l, nil
}
