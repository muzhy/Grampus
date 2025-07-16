package port

import (
	"net/http"

	"github.com/gin-gonic/gin"

	Grampus "Grampus"
)

type Response struct {
	Code Grampus.ErrCode `json:"code"`
	Msg  string          `json:"msg,omitempty"`
	Data interface{}     `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: Grampus.Ok, Data: data})
}

func Fail(c *gin.Context, code Grampus.ErrCode, msg string) {
	c.JSON(http.StatusOK, Response{Code: code, Msg: msg})
}
