package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJWTService_GenerateToken(t *testing.T) {
	// สร้าง JWTService สำหรับทดสอบ
	secretKey := "test-secret-key"
	issuer := "test-issuer"
	tokenDuration := 1 * time.Hour

	jwtService := NewJWTService(secretKey, issuer, tokenDuration)

	// ทดสอบการสร้าง token
	token, err := jwtService.GenerateToken(1, "test@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTService_ValidateToken(t *testing.T) {
	// สร้าง JWTService สำหรับทดสอบ
	secretKey := "test-secret-key"
	issuer := "test-issuer"
	tokenDuration := 1 * time.Hour

	jwtService := NewJWTService(secretKey, issuer, tokenDuration)

	// สร้าง token
	userID := uint(1)
	email := "test@example.com"
	token, err := jwtService.GenerateToken(userID, email)
	assert.NoError(t, err)

	// ทดสอบการตรวจสอบ token ที่ถูกต้อง
	claims, err := jwtService.ValidateToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, issuer, claims.Issuer)
}

func TestJWTService_ValidateToken_Invalid(t *testing.T) {
	// สร้าง JWTService สำหรับทดสอบ
	secretKey := "test-secret-key"
	issuer := "test-issuer"
	tokenDuration := 1 * time.Hour

	jwtService := NewJWTService(secretKey, issuer, tokenDuration)

	// ทดสอบ token ที่ไม่ถูกต้อง
	invalidToken := "invalid.token.string"
	claims, err := jwtService.ValidateToken(invalidToken)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTService_ValidateToken_Expired(t *testing.T) {
	// สร้าง JWTService ที่มี token หมดอายุอย่างรวดเร็ว
	secretKey := "test-secret-key"
	issuer := "test-issuer"
	tokenDuration := -1 * time.Hour // ทำให้ token หมดอายุทันที

	jwtService := NewJWTService(secretKey, issuer, tokenDuration)

	// สร้าง token ที่หมดอายุ
	token, err := jwtService.GenerateToken(1, "test@example.com")
	assert.NoError(t, err)

	// ทดสอบการตรวจสอบ token ที่หมดอายุ
	claims, err := jwtService.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTService_ValidateToken_DifferentSecret(t *testing.T) {
	// สร้าง JWTService ด้วย secret key หนึ่ง
	secretKey1 := "secret-key-1"
	issuer := "test-issuer"
	tokenDuration := 1 * time.Hour

	jwtService1 := NewJWTService(secretKey1, issuer, tokenDuration)

	// สร้าง token ด้วย service แรก
	token, err := jwtService1.GenerateToken(1, "test@example.com")
	assert.NoError(t, err)

	// สร้าง JWTService อีกตัวด้วย secret key ที่แตกต่าง
	secretKey2 := "secret-key-2"
	jwtService2 := NewJWTService(secretKey2, issuer, tokenDuration)

	// ทดสอบการตรวจสอบ token ด้วย service ที่สอง (ควรล้มเหลว)
	claims, err := jwtService2.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}
