// ไฟล์นี้จะอยู่ในรูทโฟลเดอร์ของโปรเจค

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/steinfletcher/apitest"
	jsonpath "github.com/steinfletcher/apitest-jsonpath"
	"github.com/stretchr/testify/suite"
	"github.com/yourusername/auth-api/internal/api/handlers"
	"github.com/yourusername/auth-api/internal/api/middlewares"
	"github.com/yourusername/auth-api/internal/service"
	"github.com/yourusername/auth-api/pkg/database"
	"github.com/yourusername/auth-api/pkg/jwt"
	"gorm.io/gorm"
)

type APIIntegrationTestSuite struct {
	suite.Suite
	DB         *gorm.DB
	Router     *gin.Engine
	JWTService *jwt.JWTService
	AdminToken string
}

func (s *APIIntegrationTestSuite) SetupSuite() {
	// ตั้งค่า Gin ให้อยู่ในโหมดทดสอบ
	gin.SetMode(gin.TestMode)

	// สร้างฐานข้อมูลทดสอบชั่วคราว
	dbConfig := &database.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "auth_api_test",
		SSLMode:  "disable",
	}

	// ถ้าต้องการรัน test บน CI/CD ให้ใช้ตัวแปรสภาพแวดล้อม
	if os.Getenv("TEST_DB_HOST") != "" {
		dbConfig.Host = os.Getenv("TEST_DB_HOST")
	}
	if os.Getenv("TEST_DB_PORT") != "" {
		dbConfig.Port = os.Getenv("TEST_DB_PORT")
	}
	if os.Getenv("TEST_DB_USER") != "" {
		dbConfig.User = os.Getenv("TEST_DB_USER")
	}
	if os.Getenv("TEST_DB_PASSWORD") != "" {
		dbConfig.Password = os.Getenv("TEST_DB_PASSWORD")
	}
	if os.Getenv("TEST_DB_NAME") != "" {
		dbConfig.DBName = os.Getenv("TEST_DB_NAME")
	}

	// เชื่อมต่อฐานข้อมูล
	var err error
	s.DB, err = database.NewConnection(dbConfig)
	if err != nil {
		s.T().Fatalf("Failed to connect to test database: %v", err)
	}

	// สร้างโครงสร้างฐานข้อมูล
	err = database.MigrateDB(s.DB)
	if err != nil {
		s.T().Fatalf("Failed to migrate database: %v", err)
	}

	// สร้างข้อมูลเริ่มต้น
	err = database.SeedDefaultData(s.DB)
	if err != nil {
		s.T().Fatalf("Failed to seed initial data: %v", err)
	}

	// สร้าง JWT service
	s.JWTService = jwt.NewJWTService(
		"test-secret-key",
		"test-issuer",
		24*time.Hour,
	)

	// สร้าง services
	authService := service.NewAuthService(s.DB, s.JWTService)

	// สร้าง handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(s.DB)
	roleHandler := handlers.NewRoleHandler(s.DB)
	permissionHandler := handlers.NewPermissionHandler(s.DB)

	// สร้าง middlewares
	authMiddleware := middlewares.AuthMiddleware(s.JWTService, authService)

	// สร้าง router
	s.Router = gin.Default()

	// ตั้งค่า API routes
	s.Router.POST("/api/login", authHandler.Login)

	// กลุ่ม routes ที่ต้องการการยืนยันตัวตน
	authorized := s.Router.Group("/api")
	authorized.Use(authMiddleware)

	// User routes
	authorized.GET("/users", middlewares.RequirePermission(authService, "users", "read"), userHandler.GetUsers)
	authorized.GET("/users/:id", middlewares.RequirePermission(authService, "users", "read"), userHandler.GetUser)
	authorized.POST("/users", middlewares.RequirePermission(authService, "users", "write"), userHandler.CreateUser)
	authorized.PUT("/users/:id", middlewares.RequirePermission(authService, "users", "write"), userHandler.UpdateUser)
	authorized.DELETE("/users/:id", middlewares.RequirePermission(authService, "users", "write"), userHandler.DeleteUser)
	authorized.POST("/users/:id/roles", middlewares.RequirePermission(authService, "users", "write"), userHandler.AddRoleToUser)
	authorized.DELETE("/users/:id/roles/:roleId", middlewares.RequirePermission(authService, "users", "write"), userHandler.RemoveRoleFromUser)

	// Role routes
	authorized.GET("/roles", middlewares.RequirePermission(authService, "roles", "read"), roleHandler.GetRoles)
	authorized.GET("/roles/:id", middlewares.RequirePermission(authService, "roles", "read"), roleHandler.GetRole)
	authorized.POST("/roles", middlewares.RequirePermission(authService, "roles", "write"), roleHandler.CreateRole)
	authorized.PUT("/roles/:id", middlewares.RequirePermission(authService, "roles", "write"), roleHandler.UpdateRole)
	authorized.DELETE("/roles/:id", middlewares.RequirePermission(authService, "roles", "write"), roleHandler.DeleteRole)
	authorized.POST("/roles/:id/permissions", middlewares.RequirePermission(authService, "roles", "write"), roleHandler.AddPermissionToRole)
	authorized.DELETE("/roles/:id/permissions/:permissionId", middlewares.RequirePermission(authService, "roles", "write"), roleHandler.RemovePermissionFromRole)

	// Permission routes
	authorized.GET("/permissions", middlewares.RequirePermission(authService, "permissions", "read"), permissionHandler.GetPermissions)
	authorized.GET("/permissions/:id", middlewares.RequirePermission(authService, "permissions", "read"), permissionHandler.GetPermission)
	authorized.POST("/permissions", middlewares.RequirePermission(authService, "permissions", "write"), permissionHandler.CreatePermission)
	authorized.PUT("/permissions/:id", middlewares.RequirePermission(authService, "permissions", "write"), permissionHandler.UpdatePermission)
	authorized.DELETE("/permissions/:id", middlewares.RequirePermission(authService, "permissions", "write"), permissionHandler.DeletePermission)

	// เข้าสู่ระบบด้วยผู้ใช้ admin เพื่อให้ได้ token สำหรับการทดสอบ
	loginReq := service.LoginRequest{
		Username: "admin",
		Password: "adminpassword",
	}
	jsonData, _ := json.Marshal(loginReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	s.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		s.T().Fatalf("Failed to login with admin user: %v", w.Body.String())
	}

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		s.T().Fatalf("Failed to parse login response: %v", err)
	}

	s.AdminToken = response["access_token"].(string)
}

func (s *APIIntegrationTestSuite) TearDownSuite() {
	// ลบฐานข้อมูลทดสอบหลังจากเสร็จสิ้นการทดสอบ
	db, err := s.DB.DB()
	if err != nil {
		s.T().Fatalf("Failed to get DB instance: %v", err)
	}
	db.Close()
}

func (s *APIIntegrationTestSuite) TestLoginEndpoint() {
	// ทดสอบเข้าสู่ระบบสำเร็จ
	apitest.New().
		Handler(s.Router).
		Post("/api/login").
		JSON(map[string]interface{}{
			"username": "admin",
			"password": "adminpassword",
		}).
		Expect(s.T()).
		Status(http.StatusOK).
		Assert(jsonpath.Equal("$.user.username", "admin")).
		Assert(jsonpath.Present("$.access_token")).
		End()

	// ทดสอบเข้าสู่ระบบล้มเหลว - รหัสผ่านไม่ถูกต้อง
	apitest.New().
		Handler(s.Router).
		Post("/api/login").
		JSON(map[string]interface{}{
			"username": "admin",
			"password": "wrongpassword",
		}).
		Expect(s.T()).
		Status(http.StatusUnauthorized).
		Assert(jsonpath.Present("$.error")).
		End()

	// ทดสอบเข้าสู่ระบบล้มเหลว - ชื่อผู้ใช้ไม่ถูกต้อง
	apitest.New().
		Handler(s.Router).
		Post("/api/login").
		JSON(map[string]interface{}{
			"username": "nonexistentuser",
			"password": "anypassword",
		}).
		Expect(s.T()).
		Status(http.StatusUnauthorized).
		Assert(jsonpath.Present("$.error")).
		End()
}

func (s *APIIntegrationTestSuite) TestUsersEndpoints() {
	// ตรวจสอบการรับรายการผู้ใช้ทั้งหมด
	apitest.New().
		Handler(s.Router).
		Get("/api/users").
		Header("Authorization", "Bearer "+s.AdminToken).
		Expect(s.T()).
		Status(http.StatusOK).
		Assert(jsonpath.Present("$[0].id")).
		Assert(jsonpath.Present("$[0].username")).
		End()

	// สร้างผู้ใช้ใหม่
	newUserResp := apitest.New().
		Handler(s.Router).
		Post("/api/users").
		Header("Authorization", "Bearer "+s.AdminToken).
		JSON(map[string]interface{}{
			"username":  "testuser",
			"email":     "test@example.com",
			"password":  "password123",
			"full_name": "Test User",
		}).
		Expect(s.T()).
		Status(http.StatusCreated).
		Assert(jsonpath.Equal("$.username", "testuser")).
		Assert(jsonpath.Equal("$.email", "test@example.com")).
		End().Response.Body

	// แปลงผลลัพธ์เป็น map เพื่อให้ได้ ID ของผู้ใช้ใหม่
	var newUser map[string]interface{}
	respBody, _ := io.ReadAll(newUserResp)
	json.Unmarshal(respBody, &newUser)
	newUserID := fmt.Sprintf("%.0f", newUser["id"].(float64))

	// ตรวจสอบการรับข้อมูลผู้ใช้ตาม ID
	apitest.New().
		Handler(s.Router).
		Get("/api/users/"+newUserID).
		Header("Authorization", "Bearer "+s.AdminToken).
		Expect(s.T()).
		Status(http.StatusOK).
		Assert(jsonpath.Equal("$.username", "testuser")).
		End()

	// อัปเดตข้อมูลผู้ใช้
	apitest.New().
		Handler(s.Router).
		Put("/api/users/"+newUserID).
		Header("Authorization", "Bearer "+s.AdminToken).
		JSON(map[string]interface{}{
			"full_name": "Updated Test User",
		}).
		Expect(s.T()).
		Status(http.StatusOK).
		Assert(jsonpath.Equal("$.full_name", "Updated Test User")).
		End()

	// เพิ่มบทบาทให้กับผู้ใช้
	apitest.New().
		Handler(s.Router).
		Post("/api/users/"+newUserID+"/roles").
		Header("Authorization", "Bearer "+s.AdminToken).
		JSON(map[string]interface{}{
			"role_id": 2, // role_id 2 น่าจะเป็น supervisor จาก seed data
		}).
		Expect(s.T()).
		Status(http.StatusOK).
		End()

	// ลบบทบาทออกจากผู้ใช้
	apitest.New().
		Handler(s.Router).
		Delete("/api/users/"+newUserID+"/roles/2").
		Header("Authorization", "Bearer "+s.AdminToken).
		Expect(s.T()).
		Status(http.StatusOK).
		End()

	// ลบผู้ใช้
	apitest.New().
		Handler(s.Router).
		Delete("/api/users/"+newUserID).
		Header("Authorization", "Bearer "+s.AdminToken).
		Expect(s.T()).
		Status(http.StatusOK).
		End()
}

func (s *APIIntegrationTestSuite) TestRolesEndpoints() {
	// ตรวจสอบการรับรายการบทบาททั้งหมด
	apitest.New().
		Handler(s.Router).
		Get("/api/roles").
		Header("Authorization", "Bearer "+s.AdminToken).
		Expect(s.T()).
		Status(http.StatusOK).
		Assert(jsonpath.Present("$[0].id")).
		Assert(jsonpath.Present("$[0].name")).
		End()

	// สร้างบทบาทใหม่
	newRoleResp := apitest.New().
		Handler(s.Router).
		Post("/api/roles").
		Header("Authorization", "Bearer "+s.AdminToken).
		JSON(map[string]interface{}{
			"name":        "testrole",
			"description": "Test Role",
		}).
		Expect(s.T()).
		Status(http.StatusCreated).
		Assert(jsonpath.Equal("$.name", "testrole")).
		End().Response.Body

	// แปลงผลลัพธ์เป็น map เพื่อให้ได้ ID ของบทบาทใหม่
	var newRole map[string]interface{}
	respBody, _ := io.ReadAll(newRoleResp)
	json.Unmarshal(respBody, &newRole)
	newRoleID := fmt.Sprintf("%.0f", newRole["id"].(float64))

	// ตรวจสอบการรับข้อมูลบทบาทตาม ID
	apitest.New().
		Handler(s.Router).
		Get("/api/roles/"+newRoleID).
		Header("Authorization", "Bearer "+s.AdminToken).
		Expect(s.T()).
		Status(http.StatusOK).
		Assert(jsonpath.Equal("$.name", "testrole")).
		End()

	// อัปเดตข้อมูลบทบาท
	apitest.New().
		Handler(s.Router).
		Put("/api/roles/"+newRoleID).
		Header("Authorization", "Bearer "+s.AdminToken).
		JSON(map[string]interface{}{
			"description": "Updated Test Role",
		}).
		Expect(s.T()).
		Status(http.StatusOK).
		Assert(jsonpath.Equal("$.description", "Updated Test Role")).
		End()

	// เพิ่มสิทธิ์ให้กับบทบาท
	apitest.New().
		Handler(s.Router).
		Post("/api/roles/"+newRoleID+"/permissions").
		Header("Authorization", "Bearer "+s.AdminToken).
		JSON(map[string]interface{}{
			"permission_id": 1, // permission_id 1 น่าจะเป็น "users:read" จาก seed data
		}).
		Expect(s.T()).
		Status(http.StatusOK).
		End()

	// ลบสิทธิ์ออกจากบทบาท
	apitest.New().
		Handler(s.Router).
		Delete("/api/roles/"+newRoleID+"/permissions/1").
		Header("Authorization", "Bearer "+s.AdminToken).
		Expect(s.T()).
		Status(http.StatusOK).
		End()

	// ลบบทบาท
	apitest.New().
		Handler(s.Router).
		Delete("/api/roles/"+newRoleID).
		Header("Authorization", "Bearer "+s.AdminToken).
		Expect(s.T()).
		Status(http.StatusOK).
		End()
}

func (s *APIIntegrationTestSuite) TestPermissionsEndpoints() {
	// ตรวจสอบการรับรายการสิทธิ์ทั้งหมด
	apitest.New().
		Handler(s.Router).
		Get("/api/permissions").
		Header("Authorization", "Bearer "+s.AdminToken).
		Expect(s.T()).
		Status(http.StatusOK).
		Assert(jsonpath.Present("$[0].id")).
		Assert(jsonpath.Present("$[0].resource")).
		Assert(jsonpath.Present("$[0].action")).
		End()

	// สร้างสิทธิ์ใหม่
	newPermissionResp := apitest.New().
		Handler(s.Router).
		Post("/api/permissions").
		Header("Authorization", "Bearer "+s.AdminToken).
		JSON(map[string]interface{}{
			"resource":    "articles",
			"action":      "read",
			"description": "Can read articles",
		}).
		Expect(s.T()).
		Status(http.StatusCreated).
		Assert(jsonpath.Equal("$.resource", "articles")).
		Assert(jsonpath.Equal("$.action", "read")).
		End().Response.Body

	// แปลงผลลัพธ์เป็น map เพื่อให้ได้ ID ของสิทธิ์ใหม่
	var newPermission map[string]interface{}
	respBody, _ := io.ReadAll(newPermissionResp)
	json.Unmarshal(respBody, &newPermission)
	newPermissionID := fmt.Sprintf("%.0f", newPermission["id"].(float64))

	// ตรวจสอบการรับข้อมูลสิทธิ์ตาม ID
	apitest.New().
		Handler(s.Router).
		Get("/api/permissions/"+newPermissionID).
		Header("Authorization", "Bearer "+s.AdminToken).
		Expect(s.T()).
		Status(http.StatusOK).
		Assert(jsonpath.Equal("$.resource", "articles")).
		Assert(jsonpath.Equal("$.action", "read")).
		End()

	// อัปเดตข้อมูลสิทธิ์
	apitest.New().
		Handler(s.Router).
		Put("/api/permissions/"+newPermissionID).
		Header("Authorization", "Bearer "+s.AdminToken).
		JSON(map[string]interface{}{
			"description": "Updated can read articles",
		}).
		Expect(s.T()).
		Status(http.StatusOK).
		Assert(jsonpath.Equal("$.description", "Updated can read articles")).
		End()

	// ลบสิทธิ์
	apitest.New().
		Handler(s.Router).
		Delete("/api/permissions/"+newPermissionID).
		Header("Authorization", "Bearer "+s.AdminToken).
		Expect(s.T()).
		Status(http.StatusOK).
		End()
}

func (s *APIIntegrationTestSuite) TestUnauthorizedAccess() {
	// ทดสอบเข้าถึง API โดยไม่มี token
	apitest.New().
		Handler(s.Router).
		Get("/api/users").
		Expect(s.T()).
		Status(http.StatusUnauthorized).
		End()

	// ทดสอบเข้าถึง API ด้วย token ที่ไม่ถูกต้อง
	apitest.New().
		Handler(s.Router).
		Get("/api/users").
		Header("Authorization", "Bearer invalid.token.string").
		Expect(s.T()).
		Status(http.StatusUnauthorized).
		End()
}

func TestAPIIntegration(t *testing.T) {
	// ข้ามการทดสอบถ้าไม่ได้ตั้งค่าให้ทำ integration test
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration tests")
	}
	suite.Run(t, new(APIIntegrationTestSuite))
}
