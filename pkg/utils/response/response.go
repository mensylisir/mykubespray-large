package response

import (
	"github.com/gin-gonic/gin"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"net/http"
	"time"
)

var Config = struct {
	LogLevel string
}{
	LogLevel: "INFO", // Default log level
}

func Respond(c *gin.Context, statusCode int, response CustomResponse) {
	c.JSON(statusCode, response)
}

func RespondSuccess(c *gin.Context, data interface{}) {
	Respond(c, http.StatusOK, CustomResponse{
		Status: "success",
		Data:   data,
	})
}

func RespondError(c *gin.Context, statusCode int, message string) {
	Respond(c, statusCode, CustomResponse{
		Status:  "error",
		Message: message,
	})
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.GetLogger().Errorf("Internal server error: %v", err)
				RespondError(c, http.StatusInternalServerError, "An unexpected error occurred")
			}
		}()
		c.Next()
	}
}

func WithLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		if Config.LogLevel == "DEBUG" {
			logger.GetLogger().Errorf("Request %s %s took %v", c.Request.Method, c.Request.URL.Path, duration)
		}
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Dummy authentication check
		token := c.GetHeader("Authorization")
		if token == "" {
			RespondError(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		// Proceed if authenticated
		c.Next()
	}
}
