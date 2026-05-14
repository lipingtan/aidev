package middleware

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// errorLogFile panic 日志写入路径（固定，方便 AI 读取）
const errorLogFile = "logs/error.log"

// writeErrorLog 将 panic 信息追加写入日志文件
func writeErrorLog(c *gin.Context, err interface{}, stack []byte) {
	_ = os.MkdirAll("logs", 0755)
	f, ferr := os.OpenFile(errorLogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if ferr != nil {
		return
	}
	defer f.Close()
	line := fmt.Sprintf(
		"[%s] PANIC %s %s | err: %v\nstack:\n%s\n---\n",
		time.Now().Format("2006-01-02 15:04:05"),
		c.Request.Method,
		c.Request.RequestURI,
		err,
		string(stack),
	)
	_, _ = f.WriteString(line)
}

func CustomError(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {

			if c.IsAborted() {
				c.Status(200)
			}
			switch errStr := err.(type) {
			case string:
				p := strings.Split(errStr, "#")
				if len(p) == 3 && p[0] == "CustomError" {
					statusCode, e := strconv.Atoi(p[1])
					if e != nil {
						break
					}
					c.Status(statusCode)
					fmt.Println(
						time.Now().Format("2006-01-02 15:04:05"),
						"[ERROR]",
						c.Request.Method,
						c.Request.URL,
						statusCode,
						c.Request.RequestURI,
						c.ClientIP(),
						p[2],
					)
					c.JSON(http.StatusOK, gin.H{
						"code": statusCode,
						"msg":  p[2],
					})
				} else {
					// 写入日志文件
					writeErrorLog(c, err, debug.Stack())
					c.JSON(http.StatusOK, gin.H{
						"code": 500,
						"msg":  errStr,
					})
				}
			case runtime.Error:
				// runtime panic（nil pointer 等）写入日志文件
				writeErrorLog(c, err, debug.Stack())
				c.JSON(http.StatusOK, gin.H{
					"code": 500,
					"msg":  errStr.Error(),
				})
			default:
				// 其他 panic 也写入日志文件再重新抛出
				writeErrorLog(c, err, debug.Stack())
				panic(err)
			}
		}
	}()
	c.Next()
}
