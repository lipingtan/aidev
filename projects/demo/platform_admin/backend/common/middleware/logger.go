package middleware

import (
	"bufio"
	"bytes"
	"encoding/json"
	"go-admin/common"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/api"
)

// LoggerToFile 日志记录到文件
func LoggerToFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := api.GetRequestLogger(c)
		// 开始时间
		startTime := time.Now()
		// 处理请求
		var body string
		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodGet, http.MethodDelete:
			bf := bytes.NewBuffer(nil)
			wt := bufio.NewWriter(bf)
			_, err := io.Copy(wt, c.Request.Body)
			if err != nil {
				log.Warnf("copy body error, %s", err.Error())
				err = nil
			}
			rb, _ := ioutil.ReadAll(bf)
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(rb))
			body = string(rb)
		}
		// body 变量保留用于日志输出（消除 unused 警告）
		_ = body

		c.Next()
		url := c.Request.RequestURI
		if strings.Index(url, "logout") > -1 ||
			strings.Index(url, "login") > -1 {
			return
		}
		// 结束时间
		endTime := time.Now()
		if c.Request.Method == http.MethodOptions {
			return
		}

		rt, bl := c.Get("result")
		var result = ""
		if bl {
			rb, err := json.Marshal(rt)
			if err != nil {
				log.Warnf("json Marshal result error, %s", err.Error())
			} else {
				result = string(rb)
			}
		}
		_ = result

		// 请求方式
		reqMethod := c.Request.Method
		// 请求路由
		reqUri := c.Request.RequestURI
		// 状态码
		statusCode := c.Writer.Status()
		// 请求IP
		clientIP := common.GetClientIP(c)
		// 执行时间
		latencyTime := endTime.Sub(startTime)
		// 日志格式
		logData := map[string]interface{}{
			"statusCode":  statusCode,
			"latencyTime": latencyTime,
			"clientIP":    clientIP,
			"method":      reqMethod,
			"uri":         reqUri,
		}
		log.WithFields(logData).Info()
		defer func() {
			log.Fields(map[string]interface{}{})
		}()
	}
}
