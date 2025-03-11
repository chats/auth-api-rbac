package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/yourusername/auth-api/internal/models"
	"github.com/yourusername/auth-api/internal/service"
)

// MockAuthServiceRBAC เป็น mock ของ AuthService สำหรับการทดสอบ RBAC
type MockAuthServiceRBAC struct {
	HasPermissionFunc func(userID uint, resource string, action string) (bool, error)
}

// GetUserByID implements AuthServiceInterface
func (m *MockAuthServiceRBAC) GetUserByID(userID uint) (*models.User, error) {
	// ไม่จำเป็นต้องใช้ในการทดสอบนี้
	return nil, nil
}

// HasPermission implements AuthServiceInterface
func (m *MockAuthServiceRBAC) HasPermission(userID uint, resource string, action string) (bool, error) {
	return m.HasPermissionFunc(userID, resource, action)
}

// Login implements AuthServiceInterface
func (m *MockAuthServiceRBAC) Login(_ *service.LoginRequest) (*service.LoginResponse, error) {
	// ไม่จำเป็นต้องใช้ในการทดสอบนี้
	return nil, nil
}

func setupRBACTest() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestRequirePermission_Success(t *testing.T) {
	r := setupRBACTest()

	// สร้าง mock AuthService
	mockAuthService := &MockAuthServiceRBAC{
		HasPermissionFunc: func(userID uint, resource string, action string) (bool, error) {
			// ตรวจสอบว่าพารามิเตอร์ที่ส่งมาถูกต้อง
			assert.Equal(t, uint(1), userID)
			assert.Equal(t, "users", resource)
			assert.Equal(t, "read", action)
			return true, nil
		},
	}

	// เพิ่ม handler ที่ตั้งค่า userID ใน context (จำลองการทำงานของ AuthMiddleware)
	r.Use(func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Next()
	})

	// เพิ่ม middleware ที่ต้องการทดสอบ
	r.Use(RequirePermission(mockAuthService, "users", "read"))

	// เพิ่ม handler สุดท้าย
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// ทดสอบ request ที่ผู้ใช้มีสิทธิ์
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// ตรวจสอบผลลัพธ์
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePermission_NoPermission(t *testing.T) {
	r := setupRBACTest()

	// สร้าง mock AuthService
	mockAuthService := &MockAuthServiceRBAC{
		HasPermissionFunc: func(userID uint, resource string, action string) (bool, error) {
			return false, nil
		},
	}

	// เพิ่ม handler ที่ตั้งค่า userID ใน context (จำลองการทำงานของ AuthMiddleware)
	r.Use(func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Next()
	})

	// เพิ่ม middleware ที่ต้องการทดสอบ
	r.Use(RequirePermission(mockAuthService, "users", "write"))

	// เพิ่ม handler สุดท้าย
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// ทดสอบ request ที่ผู้ใช้ไม่มีสิทธิ์
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// ตรวจสอบผลลัพธ์
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequirePermission_DatabaseError(t *testing.T) {
	r := setupRBACTest()

	// สร้าง mock AuthService
	mockAuthService := &MockAuthServiceRBAC{
		HasPermissionFunc: func(userID uint, resource string, action string) (bool, error) {
			return false, errors.New("database error")
		},
	}

	// เพิ่ม handler ที่ตั้งค่า userID ใน context (จำลองการทำงานของ AuthMiddleware)
	r.Use(func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Next()
	})

	// เพิ่ม middleware ที่ต้องการทดสอบ
	r.Use(RequirePermission(mockAuthService, "users", "read"))

	// เพิ่ม handler สุดท้าย
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// ทดสอบ request ที่เกิด error
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// ตรวจสอบผลลัพธ์
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRequirePermission_NoUser(t *testing.T) {
	r := setupRBACTest()

	// สร้าง mock AuthService
	mockAuthService := &MockAuthServiceRBAC{}

	// ไม่ตั้งค่า userID ใน context

	// เพิ่ม middleware ที่ต้องการทดสอบ
	r.Use(RequirePermission(mockAuthService, "users", "read"))

	// เพิ่ม handler สุดท้าย
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// ทดสอบ request ที่ไม่มี userID ใน context
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// ตรวจสอบผลลัพธ์
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireRole_Success(t *testing.T) {
	r := setupRBACTest()

	// เพิ่ม handler ที่ตั้งค่า user ใน context (จำลองการทำงานของ AuthMiddleware)
	r.Use(func(c *gin.Context) {
		// สร้าง user ที่มี role "admin"
		user := &models.User{
			ID:       1,
			Username: "testuser",
			Roles: []models.Role{
				{
					ID:   1,
					Name: "admin",
				},
			},
		}
		c.Set("user", user)
		c.Next()
	})

	// เพิ่ม middleware ที่ต้องการทดสอบ
	r.Use(RequireRole("admin"))

	// เพิ่ม handler สุดท้าย
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// ทดสอบ request ที่ผู้ใช้มี role ที่ต้องการ
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// ตรวจสอบผลลัพธ์
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_NoRole(t *testing.T) {
	r := setupRBACTest()

	// เพิ่ม handler ที่ตั้งค่า user ใน context (จำลองการทำงานของ AuthMiddleware)
	r.Use(func(c *gin.Context) {
		// สร้าง user ที่มี role "viewer"
		user := &models.User{
			ID:       1,
			Username: "testuser",
			Roles: []models.Role{
				{
					ID:   2,
					Name: "viewer",
				},
			},
		}
		c.Set("user", user)
		c.Next()
	})

	// เพิ่ม middleware ที่ต้องการทดสอบ
	r.Use(RequireRole("admin"))

	// เพิ่ม handler สุดท้าย
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// ทดสอบ request ที่ผู้ใช้ไม่มี role ที่ต้องการ
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// ตรวจสอบผลลัพธ์
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRole_NoUser(t *testing.T) {
	r := setupRBACTest()

	// ไม่ตั้งค่า user ใน context

	// เพิ่ม middleware ที่ต้องการทดสอบ
	r.Use(RequireRole("admin"))

	// เพิ่ม handler สุดท้าย
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// ทดสอบ request ที่ไม่มี user ใน context
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// ตรวจสอบผลลัพธ์
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireRole_MultipleRoles(t *testing.T) {
	r := setupRBACTest()

	// เพิ่ม handler ที่ตั้งค่า user ใน context (จำลองการทำงานของ AuthMiddleware)
	r.Use(func(c *gin.Context) {
		// สร้าง user ที่มีหลาย roles
		user := &models.User{
			ID:       1,
			Username: "testuser",
			Roles: []models.Role{
				{
					ID:   2,
					Name: "editor",
				},
				{
					ID:   3,
					Name: "supervisor",
				},
			},
		}
		c.Set("user", user)
		c.Next()
	})

	// เพิ่ม middleware ที่ต้องการทดสอบ
	r.Use(RequireRole("supervisor"))

	// เพิ่ม handler สุดท้าย
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// ทดสอบ request ที่ผู้ใช้มีหลาย roles และมี role ที่ต้องการ
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// ตรวจสอบผลลัพธ์
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_NoRoles(t *testing.T) {
	r := setupRBACTest()

	// เพิ่ม handler ที่ตั้งค่า user ใน context (จำลองการทำงานของ AuthMiddleware)
	r.Use(func(c *gin.Context) {
		// สร้าง user ที่ไม่มี roles
		user := &models.User{
			ID:       1,
			Username: "testuser",
			Roles:    []models.Role{}, // ไม่มี roles
		}
		c.Set("user", user)
		c.Next()
	})

	// เพิ่ม middleware ที่ต้องการทดสอบ
	r.Use(RequireRole("admin"))

	// เพิ่ม handler สุดท้าย
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// ทดสอบ request ที่ผู้ใช้ไม่มี roles เลย
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// ตรวจสอบผลลัพธ์
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRole_NilRoles(t *testing.T) {
	r := setupRBACTest()

	// เพิ่ม handler ที่ตั้งค่า user ใน context (จำลองการทำงานของ AuthMiddleware)
	r.Use(func(c *gin.Context) {
		// สร้าง user ที่มี roles เป็น nil
		user := &models.User{
			ID:       1,
			Username: "testuser",
			Roles:    nil, // roles เป็น nil
		}
		c.Set("user", user)
		c.Next()
	})

	// เพิ่ม middleware ที่ต้องการทดสอบ
	r.Use(RequireRole("admin"))

	// เพิ่ม handler สุดท้าย
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// ทดสอบ request ที่ผู้ใช้มี roles เป็น nil
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// ตรวจสอบผลลัพธ์
	assert.Equal(t, http.StatusForbidden, w.Code)
}
