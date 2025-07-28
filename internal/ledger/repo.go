package ledger

import (
	"github.com/muzhy/lapluma"
	"github.com/muzhy/lapluma/iterator"
)

type RepoIterator[T any] interface {
	Next() (lapluma.Result[T], bool)
	Close() error
}

func UnwrapRepoIt[T any](it RepoIterator[T]) iterator.Iterator[T] {
	return iterator.Map(
		iterator.Filter(it, func(ret lapluma.Result[T]) bool {
			return ret.Err == nil
		}),
		func(ret lapluma.Result[T]) T {
			return ret.Value
		},
	)
}

// 定义需要用到的存储层的接口，只要实现了这些接口的存储层，
// 不管存储后端是Sqlite，PostgreSql, Redis甚至是Json文件，XML都行
// 只要能提供符合接口的查询和存储功能即可。
// 所有需要返回多条数据的场景，均返回对应的迭代器
type Repo interface {
	// commom
	Close() error
	// save Ledger to repository
	CreateLedger(ledger *Ledger) error
	GetAllLedger() (RepoIterator[Ledger], error)
	GetLedgerById(id string) (*Ledger, error)
	GetLedgerByName(name string) (*Ledger, error)

	CreateAccount(account *Account) error
	GetAccountById(id string) (*Account, error)
	// 不支持加载全部科目，只需要加载对应账本下的所有科目即可
	// 科目的树形结构的组装由上层负责
	GetLedgerAccounts(ledgerId string) (RepoIterator[Account], error)
}
