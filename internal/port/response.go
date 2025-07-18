package port

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"Grampus/internal"
)

type Response struct {
	Code internal.ErrCode `json:"code"`
	Msg  string           `json:"msg,omitempty"`
	Data interface{}      `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: internal.Ok, Data: data})
}

func Fail(c *gin.Context, code internal.ErrCode, msg string) {
	c.JSON(http.StatusOK, Response{Code: code, Msg: msg})
}
