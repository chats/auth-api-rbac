package service

import (
	"errors"

	"github.com/yourusername/auth-api/internal/models"
	"github.com/yourusername/auth-api/pkg/jwt"
	"gorm.io/gorm"
)

type AuthService struct {
	db         *gorm.DB
	jwtService *jwt.JWTService
}

func NewAuthService(db *gorm.DB, jwtService *jwt.JWTService) *AuthService {
	return &AuthService{
		db:         db,
		jwtService: jwtService,
	}
}

// LoginRequest สำหรับรับข้อมูล login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse สำหรับส่งผลลัพธ์ login
type LoginResponse struct {
	AccessToken string                 `json:"access_token"`
	User        map[string]interface{} `json:"user"`
}

// Login ตรวจสอบข้อมูลผู้ใช้และสร้าง JWT token
func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	var user models.User

	// ค้นหาผู้ใช้จาก username
	result := s.db.Where("username = ?", req.Username).Preload("Roles.Permissions").First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, result.Error
	}

	// ตรวจสอบรหัสผ่าน
	if !user.CheckPassword(req.Password) {
		return nil, errors.New("invalid username or password")
	}

	// สร้าง token
	token, err := s.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken: token,
		User:        user.ToResponse(),
	}, nil
}

// GetUserByID ดึงข้อมูลผู้ใช้จาก ID
func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	result := s.db.Preload("Roles.Permissions").First(&user, userID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// HasPermission ตรวจสอบว่าผู้ใช้มีสิทธิ์หรือไม่
func (s *AuthService) HasPermission(userID uint, resource string, action string) (bool, error) {
	var user models.User
	result := s.db.Preload("Roles.Permissions").First(&user, userID)
	if result.Error != nil {
		return false, result.Error
	}

	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			if perm.Resource == resource && perm.Action == action {
				return true, nil
			}
		}
	}

	return false, nil
}
