package ledger

import (
	"testing"

	"github.com/muzhy/lapluma"
	"github.com/stretchr/testify/assert"
)

type TestAccountIt struct {
	data   []Account
	cursor int
}

func (t *TestAccountIt) Next() (lapluma.Result[Account], bool) {
	if t.cursor >= len(t.data) {
		return lapluma.Result[Account]{
			Err: nil,
		}, false
	}
	index := t.cursor
	t.cursor++
	return lapluma.Result[Account]{
		Value: t.data[index],
		Err:   nil,
	}, true
}
func (t *TestAccountIt) Close() error {
	t.cursor = len(t.data)
	return nil
}

func TestAccountTree(t *testing.T) {
	// 直接模拟拿到的account数组, 存储层的正确性由存储层的单元测试负责
	// 避免存储层的问题影响到领域层的单元测试
	ledgerID := "1"
	data := []Account{
		{
			ID:       "1",
			LedgerID: ledgerID,
			ParentID: "",
			Level:    0,
			GoodsID:  string(CNY),
			Name:     "root1",
			Path:     "root1",
		},
		{
			ID:       "1-1",
			LedgerID: ledgerID,
			ParentID: "1",
			Level:    1,
			GoodsID:  string(CNY),
			Name:     "1-1",
			Path:     "root1/1-1",
		},
		{
			ID:       "1-2",
			LedgerID: ledgerID,
			ParentID: "1",
			Level:    1,
			GoodsID:  string(CNY),
			Name:     "1-2",
			Path:     "root1/1-2",
		},
		{
			ID:       "1-1-1",
			LedgerID: ledgerID,
			ParentID: "1-1",
			Level:    2,
			GoodsID:  string(CNY),
			Name:     "1-1-1",
			Path:     "root1/1-1/1-1-1",
		},
	}

	accountIt := TestAccountIt{
		data:   data,
		cursor: 0,
	}
	accounts := buildAccountTrees(&accountIt)
	assert.Equal(t, 1, len(accounts))
	root := accounts[0]
	assert.Equal(t, "1", root.ID)
	assert.Equal(t, 0, root.Level)
	assert.Equal(t, 2, len(root.Childs))
	assert.Equal(t, root.Name, root.Path)
	//
	sub1, ok := root.Childs["1-1"]
	assert.True(t, ok)
	assert.Equal(t, "1-1", sub1.Name)
	assert.Equal(t, "root1/1-1", sub1.Path)
	//
	sub2, ok := root.Childs["1-2"]
	assert.True(t, ok)
	assert.Equal(t, "root1/1-2", sub2.Path)
	//
	sub1_1, ok := sub1.Childs["1-1-1"]
	assert.True(t, ok)
	assert.Equal(t, "root1/1-1/1-1-1", sub1_1.Path)
	assert.Equal(t, "1-1-1", sub1_1.Name)
}
