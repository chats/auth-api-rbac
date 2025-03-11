package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/auth-api/internal/service"
	"github.com/yourusername/auth-api/pkg/jwt"
)

// AuthMiddleware ตรวจสอบความถูกต้องของ JWT token
// func AuthMiddleware(jwtService *jwt.JWTService, authService *service.AuthService) gin.HandlerFunc {
func AuthMiddleware(jwtService *jwt.JWTService, authService service.AuthServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// ตรวจสอบรูปแบบ "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer <token>"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// ดึงข้อมูลผู้ใช้จาก token
		user, err := authService.GetUserByID(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// เก็บข้อมูลผู้ใช้ใน context สำหรับใช้ในขั้นตอนต่อไป
		c.Set("user", user)
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
