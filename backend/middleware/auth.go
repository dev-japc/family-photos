package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)


// key for JWT for development - in production, use ENV variables
var jwtSecret = []byte("my-super-secret-key-dev-secure-secret-key-for-development")

func AuthMiddleware() gin.HandlerFunc {
	return func (c *gin.Context) {
		// get token from the header 'Authorization'
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided, UNAUTHORIZED!!!"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format (Must be Bearer <token>)"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 3. Parsing and token validation
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validating signature method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Store user info in context for use in handlers
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Saving UserID and Role in the context
		// Numbers in JWT are parsed as float64 by default
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			// Handle invalid/missing user_id claim
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
		userID := uint(userIDFloat)
		role, _ := claims["role"].(string)

		c.Set("userID", userID)
		c.Set("role", role)

		c.Next()
		
	}
}

// AdminMiddleware block the request if the user doesn't have admin role
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify user role"})
			c.Abort()
			return
		}

		role := roleValue.(string)

		// If is not admin, return a 403 Forbidden error
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: Admin permissions are required"})
			c.Abort()
			return
		}
		c.Next()
	}
}