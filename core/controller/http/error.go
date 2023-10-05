package http

import (
	"github.com/gin-gonic/gin"
)

type BaseResponse struct {
	Error      string      `json:"error" example:"message"`
	Code       string      `json:"code"`
	Data       interface{} `json:"data"`
	ServerTime int64       `json:"server_time"`
}

func errorResponse(c *gin.Context, code int, msg string) {
	c.AbortWithStatusJSON(code, BaseResponse{Error: msg})
}
