// internal/service/auth_service_test.go
package service

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"github.com/yourusername/auth-api/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type AuthServiceTestSuite struct {
	suite.Suite
	DB          *gorm.DB
	mock        sqlmock.Sqlmock
	authService *AuthService
	jwtService  *jwt.JWTService
}

func (s *AuthServiceTestSuite) SetupTest() {
	var err error

	// สร้าง mock ของฐานข้อมูล
	db, mock, err := sqlmock.New()
	s.NoError(err)

	dialector := postgres.New(postgres.Config{
		DSN:                  "sqlmock_db_0",
		DriverName:           "postgres",
		Conn:                 db,
		PreferSimpleProtocol: true,
	})

	s.DB, err = gorm.Open(dialector, &gorm.Config{})
	s.NoError(err)
	s.mock = mock

	// สร้าง JWT service
	s.jwtService = jwt.NewJWTService("test-secret", "test-issuer", 1*time.Hour)

	// สร้าง auth service
	s.authService = NewAuthService(s.DB, s.jwtService)
}

func (s *AuthServiceTestSuite) AfterTest(_, _ string) {
	// ตรวจสอบว่ามีการเรียก expect ทั้งหมดหรือไม่
	s.NoError(s.mock.ExpectationsWereMet())
}

func TestAuthServiceSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

func (s *AuthServiceTestSuite) TestLogin_UserNotFound() {
	// Mock การค้นหาผู้ใช้ที่ไม่พบในฐานข้อมูล
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE username = \$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs("nonexistentuser", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// ทดสอบ login ด้วยชื่อผู้ใช้ที่ไม่มีอยู่
	loginReq := &LoginRequest{
		Username: "nonexistentuser",
		Password: "anypassword",
	}

	response, err := s.authService.Login(loginReq)

	// ตรวจสอบผลลัพธ์
	s.Error(err)
	s.Nil(response)
	s.Equal("invalid username or password", err.Error())
}

func (s *AuthServiceTestSuite) TestLogin_InvalidCredentials() {
	// สร้างรหัสผ่านที่เข้ารหัสแล้ว
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	s.NoError(err)

	// Mock การค้นหาผู้ใช้
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE username = \$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs("testuser", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "full_name", "created_at", "updated_at"}).
			AddRow(1, "testuser", "test@example.com", string(hashedPassword), "Test User", time.Now(), time.Now()))

	// ไม่ต้อง mock การโหลด roles เพราะจะไม่มีการเรียกถ้ารหัสผ่านไม่ถูกต้อง

	// แม้จะตรวจสอบรหัสผ่านไม่ผ่าน โค้ดจริงยังพยายามโหลด user_roles
	// (พฤติกรรมของ GORM ที่มีการ Preload("Roles.Permissions"))
	s.mock.ExpectQuery(`SELECT \* FROM "user_roles" WHERE "user_roles"\."user_id" = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "role_id"}))

	// ทดสอบ login ด้วยรหัสผ่านที่ไม่ถูกต้อง
	loginReq := &LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	response, err := s.authService.Login(loginReq)

	// ตรวจสอบผลลัพธ์
	s.Error(err)
	s.Nil(response)
	s.Equal("invalid username or password", err.Error())
}

func (s *AuthServiceTestSuite) TestLogin_Success() {
	// สร้างรหัสผ่านที่เข้ารหัสแล้ว
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	s.NoError(err)

	// Mock การค้นหาผู้ใช้
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE username = \$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs("testuser", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "full_name", "created_at", "updated_at"}).
			AddRow(1, "testuser", "test@example.com", string(hashedPassword), "Test User", time.Now(), time.Now()))

	// Mock เมื่อ GORM พยายามดึงข้อมูล user_roles
	s.mock.ExpectQuery(`SELECT \* FROM "user_roles" WHERE "user_roles"\."user_id" = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "role_id"}).
			AddRow(1, 1))

	// Mock เมื่อ GORM พยายามดึงข้อมูล roles
	s.mock.ExpectQuery(`SELECT \* FROM "roles" WHERE "roles"\."id" = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow(1, "admin", "Administrator", time.Now(), time.Now()))

	// Mock เมื่อ GORM พยายามดึงข้อมูล role_permissions
	s.mock.ExpectQuery(`SELECT \* FROM "role_permissions" WHERE "role_permissions"\."role_id" = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"role_id", "permission_id"}).
			AddRow(1, 1).
			AddRow(1, 2))

	// Mock เมื่อ GORM พยายามดึงข้อมูล permissions
	s.mock.ExpectQuery(`SELECT \* FROM "permissions" WHERE "permissions"\."id" IN \(\$1,\$2\)`).
		WithArgs(1, 2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "resource", "action", "description", "created_at", "updated_at"}).
			AddRow(1, "users", "read", "Can read users", time.Now(), time.Now()).
			AddRow(2, "users", "write", "Can write users", time.Now(), time.Now()))

	// ทดสอบ login สำเร็จ
	loginReq := &LoginRequest{
		Username: "testuser",
		Password: "correctpassword",
	}

	response, err := s.authService.Login(loginReq)

	// ตรวจสอบผลลัพธ์
	s.NoError(err)
	s.NotNil(response)
	s.NotEmpty(response.AccessToken)
	s.NotNil(response.User)
	s.Equal("testuser", response.User["username"])
}

func (s *AuthServiceTestSuite) TestGetUserByID_Success() {
	// Mock การค้นหาผู้ใช้จาก ID
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"\."id" = \$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "full_name", "created_at", "updated_at"}).
			AddRow(1, "testuser", "test@example.com", "hashedpassword", "Test User", time.Now(), time.Now()))

	// Mock เมื่อ GORM พยายามดึงข้อมูล user_roles
	s.mock.ExpectQuery(`SELECT \* FROM "user_roles" WHERE "user_roles"\."user_id" = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "role_id"}).
			AddRow(1, 1))

	// Mock เมื่อ GORM พยายามดึงข้อมูล roles
	s.mock.ExpectQuery(`SELECT \* FROM "roles" WHERE "roles"\."id" = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow(1, "admin", "Administrator", time.Now(), time.Now()))

	// Mock เมื่อ GORM พยายามดึงข้อมูล role_permissions
	s.mock.ExpectQuery(`SELECT \* FROM "role_permissions" WHERE "role_permissions"\."role_id" = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"role_id", "permission_id"}).
			AddRow(1, 1))

	// Mock เมื่อ GORM พยายามดึงข้อมูล permissions
	s.mock.ExpectQuery(`SELECT \* FROM "permissions" WHERE "permissions"\."id" = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "resource", "action", "description", "created_at", "updated_at"}).
			AddRow(1, "users", "read", "Can read users", time.Now(), time.Now()))

	// ทดสอบการดึงข้อมูลผู้ใช้
	user, err := s.authService.GetUserByID(1)

	// ตรวจสอบผลลัพธ์
	s.NoError(err)
	s.NotNil(user)
	s.Equal(uint(1), user.ID)
	s.Equal("testuser", user.Username)
	s.Equal("test@example.com", user.Email)
	s.Equal("Test User", user.FullName)
	s.Len(user.Roles, 1)
	s.Equal("admin", user.Roles[0].Name)
	s.Len(user.Roles[0].Permissions, 1)
	s.Equal("users", user.Roles[0].Permissions[0].Resource)
	s.Equal("read", user.Roles[0].Permissions[0].Action)
}

func (s *AuthServiceTestSuite) TestGetUserByID_NotFound() {
	// Mock การค้นหาผู้ใช้ที่ไม่พบในฐานข้อมูล
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"\."id" = \$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// ทดสอบการดึงข้อมูลผู้ใช้ที่ไม่มีอยู่
	user, err := s.authService.GetUserByID(999)

	// ตรวจสอบผลลัพธ์
	s.Error(err)
	s.Nil(user)
	s.Equal(gorm.ErrRecordNotFound, err)
}

func (s *AuthServiceTestSuite) TestHasPermission_Success() {
	// Mock การค้นหาผู้ใช้จาก ID
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"\."id" = \$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "full_name", "created_at", "updated_at"}).
			AddRow(1, "testuser", "test@example.com", "hashedpassword", "Test User", time.Now(), time.Now()))

	// Mock เมื่อ GORM พยายามดึงข้อมูล user_roles
	s.mock.ExpectQuery(`SELECT \* FROM "user_roles" WHERE "user_roles"\."user_id" = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "role_id"}).
			AddRow(1, 1))

	// Mock เมื่อ GORM พยายามดึงข้อมูล roles
	s.mock.ExpectQuery(`SELECT \* FROM "roles" WHERE "roles"\."id" = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow(1, "admin", "Administrator", time.Now(), time.Now()))

	// Mock เมื่อ GORM พยายามดึงข้อมูล role_permissions
	s.mock.ExpectQuery(`SELECT \* FROM "role_permissions" WHERE "role_permissions"\."role_id" = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"role_id", "permission_id"}).
			AddRow(1, 1).
			AddRow(1, 2))

	// Mock เมื่อ GORM พยายามดึงข้อมูล permissions
	s.mock.ExpectQuery(`SELECT \* FROM "permissions" WHERE "permissions"\."id" IN \(\$1,\$2\)`).
		WithArgs(1, 2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "resource", "action", "description", "created_at", "updated_at"}).
			AddRow(1, "users", "read", "Can read users", time.Now(), time.Now()).
			AddRow(2, "users", "write", "Can write users", time.Now(), time.Now()))

	// ทดสอบการตรวจสอบสิทธิ์ที่มี
	hasPermission, err := s.authService.HasPermission(1, "users", "read")

	// ตรวจสอบผลลัพธ์
	s.NoError(err)
	s.True(hasPermission)
}

func (s *AuthServiceTestSuite) TestHasPermission_NoPermission() {
	// Mock การค้นหาผู้ใช้จาก ID
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"\."id" = \$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "full_name", "created_at", "updated_at"}).
			AddRow(1, "testuser", "test@example.com", "hashedpassword", "Test User", time.Now(), time.Now()))

	// Mock เมื่อ GORM พยายามดึงข้อมูล user_roles
	s.mock.ExpectQuery(`SELECT \* FROM "user_roles" WHERE "user_roles"\."user_id" = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "role_id"}).
			AddRow(1, 2))

	// Mock เมื่อ GORM พยายามดึงข้อมูล roles
	s.mock.ExpectQuery(`SELECT \* FROM "roles" WHERE "roles"\."id" = \$1`).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow(2, "viewer", "Viewer", time.Now(), time.Now()))

	// Mock เมื่อ GORM พยายามดึงข้อมูล role_permissions
	s.mock.ExpectQuery(`SELECT \* FROM "role_permissions" WHERE "role_permissions"\."role_id" = \$1`).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"role_id", "permission_id"}).
			AddRow(2, 1))

	// Mock เมื่อ GORM พยายามดึงข้อมูล permissions
	s.mock.ExpectQuery(`SELECT \* FROM "permissions" WHERE "permissions"\."id" = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "resource", "action", "description", "created_at", "updated_at"}).
			AddRow(1, "users", "read", "Can read users", time.Now(), time.Now()))

	// ทดสอบการตรวจสอบสิทธิ์ที่ไม่มี
	hasPermission, err := s.authService.HasPermission(1, "users", "write")

	// ตรวจสอบผลลัพธ์
	s.NoError(err)
	s.False(hasPermission)
}

func (s *AuthServiceTestSuite) TestHasPermission_UserNotFound() {
	// Mock การค้นหาผู้ใช้ที่ไม่พบในฐานข้อมูล
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"\."id" = \$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// ทดสอบการตรวจสอบสิทธิ์กับผู้ใช้ที่ไม่มีอยู่
	hasPermission, err := s.authService.HasPermission(999, "users", "read")

	// ตรวจสอบผลลัพธ์
	s.Error(err)
	s.False(hasPermission)
	s.Equal(gorm.ErrRecordNotFound, err)
}

func (s *AuthServiceTestSuite) TestHasPermission_DatabaseError() {
	// Mock การค้นหาผู้ใช้ที่เกิด error
	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"\."id" = \$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnError(errors.New("database connection error"))

	// ทดสอบการตรวจสอบสิทธิ์ที่เกิด error จากฐานข้อมูล
	hasPermission, err := s.authService.HasPermission(1, "users", "read")

	// ตรวจสอบผลลัพธ์
	s.Error(err)
	s.False(hasPermission)
	s.Equal("database connection error", err.Error())
}
