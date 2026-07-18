package handler

import (
	"net/http"

	"go-admin/common/auth/errors"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构体
type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Data:    data,
		Message: "ok",
	})
}

// Error 错误响应，根据 AuthError 自动设置 HTTP 状态码
func Error(c *gin.Context, err error) {
	if authErr, ok := err.(*errors.AuthError); ok {
		httpStatus := mapHTTPStatus(authErr.Code)
		c.JSON(httpStatus, Response{
			Code:    authErr.Code,
			Data:    nil,
			Message: authErr.Message,
		})
		return
	}
	// 处理 handler 内部错误类型
	switch err.(type) {
	case *errBadRequest:
		c.JSON(http.StatusBadRequest, Response{
			Code:    40000,
			Data:    nil,
			Message: err.Error(),
		})
		return
	case *errUnauthorized:
		c.JSON(http.StatusUnauthorized, Response{
			Code:    40101,
			Data:    nil,
			Message: err.Error(),
		})
		return
	}
	// 未知错误
	c.JSON(http.StatusInternalServerError, Response{
		Code:    500,
		Data:    nil,
		Message: "内部服务错误",
	})
}

// mapHTTPStatus 根据业务错误码映射 HTTP 状态码
func mapHTTPStatus(code int) int {
	switch {
	case code >= 40100 && code < 40200:
		return http.StatusUnauthorized
	case code >= 40300 && code < 40400:
		return http.StatusForbidden
	case code >= 40000 && code < 40100:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
