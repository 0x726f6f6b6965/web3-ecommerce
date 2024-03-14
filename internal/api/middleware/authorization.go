package middleware

import (
	"net/http"

	"github.com/0x726f6f6b6965/web3-ecommerce/internal/helper"
	"github.com/gin-gonic/gin"
)

func UserAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := helper.ExtractTokenMetadata(c.Request)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("access_token", token)
		c.Next()
	}
}

func AdminAuthorization(admin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := helper.ExtractTokenMetadata(c.Request)
		if err != nil || token.PublicAddress != admin {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("access_token", token)
		c.Next()
	}
}
