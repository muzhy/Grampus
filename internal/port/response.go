package port

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Msg  string      `json:"msg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Data: data, Msg: "success"})
}

func Fail(c *gin.Context, err error) {
	c.JSON(http.StatusOK, Response{Msg: err.Error()})
}
