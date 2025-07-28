package pkg

import (
	"github.com/jmoiron/sqlx"
	"github.com/muzhy/lapluma"
)

type RowsIterator[T any] struct {
	rows   *sqlx.Rows
	closed bool
}

func NewRowsIterator[T any](rows *sqlx.Rows) *RowsIterator[T] {
	return &RowsIterator[T]{
		rows:   rows,
		closed: false,
	}
}

func (rit *RowsIterator[T]) Close() error {
	if rit.closed {
		return nil
	}
	if err := rit.rows.Close(); err != nil {
		return err
	}
	rit.closed = true
	return nil
}

func (rit *RowsIterator[T]) Next() (lapluma.Result[T], bool) {
	ret := lapluma.Result[T]{
		Err: nil,
	}
	if rit.closed {
		return ret, false
	}
	if rit.rows.Next() {
		if err := rit.rows.StructScan(&ret.Value); err != nil {
			gerr := ErrDBScanError
			gerr.Errs = []error{err}
			ret.Err = gerr
			return ret, true
		}
		return ret, true
	}

	rit.Close()

	return ret, false
}

type EmptyResultIt[T any] struct{}

func (eit *EmptyResultIt[T]) Next() (lapluma.Result[T], bool) {
	return lapluma.Result[T]{
		Err: nil,
	}, false
}

func (eit *EmptyResultIt[T]) Close() error {
	return nil
}

func NewEmptyItertor[T any]() EmptyResultIt[T] {
	return EmptyResultIt[T]{}
}
