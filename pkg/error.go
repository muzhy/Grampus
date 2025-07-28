package pkg

import (
	"fmt"
)

type GrampusError struct {
	Code int
	Msg  string
	Errs []error // 记录原始错误栈，使用`Error`时不打印错误栈
}

func (e GrampusError) Error() string {
	return fmt.Sprintf("%d -> %s", e.Code, e.Msg)
}

var (
	// 参数错误，
	ErrInvalidParam           = GrampusError{Code: 403001, Msg: "invalid param"}
	ErrInvalidAccount         = GrampusError{Code: 403002, Msg: "account data is invalid"}
	ErrExceedAccountNestLevel = GrampusError{Code: 403003, Msg: "exceed Account max nest level"}
	// 请求数据不存在 404
	ErrEmptyResult           = GrampusError{Code: 404000, Msg: "empty result"}
	ErrLedgerNotExist        = GrampusError{Code: 404001, Msg: "ledger not exist"}
	ErrAccountNotExist       = GrampusError{Code: 404002, Msg: "account not exist"}
	ErrParentAccountNotExist = GrampusError{Code: 404003, Msg: "parent account not exist"}
	// 内部错误 500
	// 100 数据库错误
	ErrDBInsertFailed = GrampusError{Code: 500101, Msg: "insert data to database failed"}
	ErrDBScanError    = GrampusError{Code: 500102, Msg: "Scan data to struct failed"}
)
