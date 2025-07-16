package internal

import "fmt"

type ErrCode int

const (
	Ok               ErrCode = 0   // 无错误
	ErrInternal      ErrCode = 500 // 内部错误
	ErrAlreadyExists ErrCode = 401 // 数据已存在，不允许重复创建
	ErrInvalidData   ErrCode = 402 // 数据内容错误
	ErrInvalidParam  ErrCode = 403 // 参数错误
	ErrDbError       ErrCode = 600 // 数据库相关的错误
)

type Error struct {
	Code ErrCode
	Msg  string // 应用自定义的错误信息
	Err  error  // 原始错误
}

func (e Error) Error() string {
	// if e.Err != nil {
	// 	return fmt.Sprintf("%d: %s | original error: %v", e.Code, e.Msg, e.Err)
	// }
	return fmt.Sprintf("%d: %s", e.Code, e.Msg)
}

func (e Error) FullError() string {
	if e.Err != nil {
		return fmt.Sprintf("%d: %s | original error: %v", e.Code, e.Msg, e.Err)
	}
	return fmt.Sprintf("%d: %s", e.Code, e.Msg)
}
