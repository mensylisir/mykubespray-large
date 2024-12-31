package aop

import (
	"github.com/gin-gonic/gin"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"gorm.io/gorm"
	"time"
)

type OperationLog struct {
	gorm.Model
	User         string `json:"user" gorm:"index"`
	Timestamp    int64  `json:"timestamp"`
	Resource     string `json:"resource"`
	Endpoint     string `json:"endpoint"`
	Module       string `json:"module"`
	ResourceType string `json:"resource_type"`
	Description  string `json:"description"`
}

type RouteDescription struct {
	Path        string
	Description string
}

var routeDescriptions = []RouteDescription{
	{Path: "/api/v1/users", Description: "Endpoint for managing users."},
	{Path: "/api/v1/orders", Description: "Endpoint for managing orders."},
	// 添加更多描述
}

func getDescription(path string) string {
	for _, route := range routeDescriptions {
		if route.Path == path {
			return route.Description
		}
	}
	return "No description available for this endpoint."
}

func LogRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Capture start time
		startTime := time.Now()

		c.Next()

		endTime := time.Now()

		description := getDescription(c.Request.URL.Path)

		// Log details
		logEntry := OperationLog{
			User:         "example_user", // 在实际应用中，你可能需要从请求上下文中获取用户信息
			Timestamp:    endTime.Unix(),
			Resource:     c.Request.URL.Path,
			Endpoint:     c.Request.RequestURI,
			Module:       "example_module", // 根据实际需要设置模块
			ResourceType: c.Request.Method,
			Description:  description,
			Model: gorm.Model{
				CreatedAt: startTime,
			},
		}
		logger.GetLogger().Infof("%v", logEntry)
		//// Save to database
		//if err := db.Create(&logEntry).Error; err != nil {
		//	logger.GetLogger().Printf("Failed to log operation: %v", err)
		//}
	}
}
