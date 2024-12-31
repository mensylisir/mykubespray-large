package aop

import (
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
	"github.com/whoisfisher/mykubespray/pkg/jwt"
	"net/http"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {

		profile, err := jwt.ExtractTokenMetadata(c.Request)
		if err != nil {
			ginx.Bomb(http.StatusUnauthorized, "unauthorized")
		}

		c.Set("user", profile)
		c.Next()

	}
}
