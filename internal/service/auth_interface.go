package service

import (
	"github.com/yourusername/auth-api/internal/models"
)

// AuthServiceInterface กำหนด interface สำหรับ AuthService
// เพื่อให้สามารถทำ mock ในการทดสอบได้
type AuthServiceInterface interface {
	GetUserByID(userID uint) (*models.User, error)
	HasPermission(userID uint, resource string, action string) (bool, error)
	Login(req *LoginRequest) (*LoginResponse, error)
}

// ตรวจสอบว่า AuthService เข้ากันได้กับ AuthServiceInterface
var _ AuthServiceInterface = (*AuthService)(nil)
