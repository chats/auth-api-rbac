package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/yourusername/auth-api/internal/models"
	"github.com/yourusername/auth-api/internal/service"
	"github.com/yourusername/auth-api/pkg/jwt"
)

// MockAuthService เป็น mock ของ AuthService
type MockAuthService struct {
	GetUserByIDFunc func(userID uint) (*models.User, error)
}

// GetUserByID implements AuthServiceInterface
func (m *MockAuthService) GetUserByID(userID uint) (*models.User, error) {
	return m.GetUserByIDFunc(userID)
}

// HasPermission implements AuthServiceInterface
func (m *MockAuthService) HasPermission(userID uint, resource string, action string) (bool, error) {
	// ไม่จำเป็นต้องใช้ในการทดสอบนี้
	return false, nil
}

// Login implements AuthServiceInterface
func (m *MockAuthService) Login(_ *service.LoginRequest) (*service.LoginResponse, error) {
	// ไม่จำเป็นต้องใช้ในการทดสอบนี้
	return nil, nil
}

func setupAuthTest() (*gin.Engine, *jwt.JWTService) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	jwtService := jwt.NewJWTService("test-secret", "test-issuer", 1*time.Hour)

	return r, jwtService
}

func TestAuthMiddleware_Success(t *testing.T) {
	r, jwtService := setupAuthTest()

	// สร้าง mock AuthService
	mockAuthService := &MockAuthService{
		GetUserByIDFunc: func(userID uint) (*models.User, error) {
			return &models.User{
				ID:       userID,
				Username: "testuser",
				Email:    "test@example.com",
			}, nil
		},
	}

	// สร้าง token ที่ถูกต้อง
	token, err := jwtService.GenerateToken(1, "test@example.com")
	assert.NoError(t, err)

	// เพิ่ม middleware และ handler
	r.Use(AuthMiddleware(jwtService, mockAuthService))
	r.GET("/test", func(c *gin.Context) {
		// ตรวจสอบว่ามีข้อมูลผู้ใช้ใน context
		user, exists := c.Get("user")
		assert.True(t, exists)
		assert.NotNil(t, user)

		userID, exists := c.Get("userID")
		assert.True(t, exists)
		assert.Equal(t, uint(1), userID)

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// ทดสอบ request ที่มี token ที่ถูกต้อง
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	// ตรวจสอบผลลัพธ์
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_NoAuthHeader(t *testing.T) {
	r, jwtService := setupAuthTest()

	// สร้าง mock AuthService
	mockAuthService := &MockAuthService{}

	// เพิ่ม middleware และ handler
	r.Use(AuthMiddleware(jwtService, mockAuthService))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// ทดสอบ request ที่ไม่มี Authorization header
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// ตรวจสอบผลลัพธ์
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidAuthFormat(t *testing.T) {
	r, jwtService := setupAuthTest()

	// สร้าง mock AuthService
	mockAuthService := &MockAuthService{}

	// เพิ่ม middleware และ handler
	r.Use(AuthMiddleware(jwtService, mockAuthService))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// ทดสอบ request ที่มี Authorization header ในรูปแบบที่ไม่ถูกต้อง
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	r.ServeHTTP(w, req)

	// ตรวจสอบผลลัพธ์
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	r, jwtService := setupAuthTest()

	// สร้าง mock AuthService
	mockAuthService := &MockAuthService{}

	// เพิ่ม middleware และ handler
	r.Use(AuthMiddleware(jwtService, mockAuthService))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// ทดสอบ request ที่มี token ที่ไม่ถูกต้อง
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.string")
	r.ServeHTTP(w, req)

	// ตรวจสอบผลลัพธ์
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_UserNotFound(t *testing.T) {
	r, jwtService := setupAuthTest()

	// สร้าง mock AuthService ที่คืนค่า user not found
	mockAuthService := &MockAuthService{
		GetUserByIDFunc: func(userID uint) (*models.User, error) {
			return nil, errors.New("user not found")
		},
	}

	// สร้าง token ที่ถูกต้อง
	token, err := jwtService.GenerateToken(1, "test@example.com")
	assert.NoError(t, err)

	// เพิ่ม middleware และ handler
	r.Use(AuthMiddleware(jwtService, mockAuthService))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// ทดสอบ request ที่มี token ที่ถูกต้องแต่ผู้ใช้ไม่มีอยู่
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	// ตรวจสอบผลลัพธ์
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
