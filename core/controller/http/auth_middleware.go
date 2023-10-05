package http

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/temukan-co/monolith/config"
	"net/http"
	"time"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			secretKey := []byte(cfg.JWTSecret)
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		exp := claims["exp"].(float64)
		t := time.Unix(int64(exp), 0)
		now := time.Now()
		if t.Before(now) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Expired token claims"})
			c.Abort()
			return
		}

		userID := claims["user_id"].(float64)
		email := claims["email"].(string)
		name := claims["name"].(string)

		// Set the user details in the context for further use
		c.Set("user_id", userID)
		c.Set("email", email)
		c.Set("name", name)

		c.Next()
	}
}

func AuthHomeMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.Next()
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			secretKey := []byte(cfg.JWTSecret)
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			c.Next()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Next()
			return
		}

		exp := claims["exp"].(float64)
		t := time.Unix(int64(exp), 0)
		now := time.Now()
		if t.Before(now) {
			c.Next()
			return
		}

		userID := claims["user_id"].(float64)
		email := claims["email"].(string)
		name := claims["name"].(string)

		// Set the user details in the context for further use
		c.Set("user_id", userID)
		c.Set("email", email)
		c.Set("name", name)

		c.Next()
	}
}
