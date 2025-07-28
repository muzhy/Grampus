package port

import (
	"Grampus/internal/ledger"
	"Grampus/pkg"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AddLedgerRoutter(router *gin.Engine, repo ledger.Repo) {
	ledgerRouter := router.Group("/ledger")
	ledgerRouter.GET("/ledger", GetAllLedger(repo))
	ledgerRouter.POST("/ledger", Create(repo, func() *ledger.Ledger { return new(ledger.Ledger) }))
	ledgerRouter.POST("/account", Create(repo, func() *ledger.Account { return new(ledger.Account) }))
	ledgerRouter.GET("/account/:ledger_name", GetLedgerAccounts(repo))
}

func GetAllLedger(repo ledger.Repo) func(c *gin.Context) {
	return func(c *gin.Context) {
		ledgers, err := ledger.GetAllLedger(repo)
		if err != nil {
			Fail(c, err)
			return
		}
		Success(c, ledgers)
	}
}

func Create[T ledger.CURD](repo ledger.Repo, factory func() T) func(c *gin.Context) {
	return func(c *gin.Context) {
		data := factory()
		if err := c.ShouldBindJSON(data); err != nil {
			zap.L().Debug("invalid param", zap.Error(err))
			Fail(c, pkg.ErrInvalidParam)
			return
		}
		err := data.Create(repo)
		if err != nil {
			zap.L().Error("Create failed", zap.Any("data", data), zap.Error(err))
			Fail(c, err)
			return
		}
		zap.L().Debug("Create successed", zap.Any("data", data))
		Success(c, data)
	}
}

func GetLedgerAccounts(repo ledger.Repo) func(c *gin.Context) {
	return func(c *gin.Context) {
		if ledgerName, ok := c.Params.Get("ledger_name"); ok {
			l, err := ledger.GetLedgerWithAccountsByName(repo, ledgerName)
			if err != nil {
				Fail(c, err)
				return
			}
			Success(c, l)
		}
		Fail(c, pkg.ErrInvalidParam)
	}
}
