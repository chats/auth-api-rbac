package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/auth-api/internal/models"
	"github.com/yourusername/auth-api/internal/service"
)

// RequirePermission ตรวจสอบว่าผู้ใช้มีสิทธิ์ที่ต้องการหรือไม่
func RequirePermission(authService *service.AuthService, resource string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ดึง userID จาก context ที่ถูกตั้งค่าโดย AuthMiddleware
		userIDValue, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		userID := userIDValue.(uint)
		hasPermission, err := authService.HasPermission(userID, resource, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole ตรวจสอบว่าผู้ใช้มีบทบาทที่ต้องการหรือไม่
func RequireRole(roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userValue, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		user := userValue.(*models.User)
		hasRole := false

		for _, role := range user.Roles {
			if role.Name == roleName {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role required: " + roleName})
			c.Abort()
			return
		}

		c.Next()
	}
}
