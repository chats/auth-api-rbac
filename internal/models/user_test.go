package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUser_SetPassword(t *testing.T) {
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	// ทดสอบการเข้ารหัสรหัสผ่าน
	err := user.SetPassword("password123")
	assert.NoError(t, err)
	assert.NotEqual(t, "password123", user.Password)
	assert.NotEmpty(t, user.Password)

	// ทดสอบการตรวจสอบรหัสผ่าน
	assert.True(t, user.CheckPassword("password123"))
	assert.False(t, user.CheckPassword("wrongpassword"))
}

func TestUser_CheckPassword(t *testing.T) {
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	// ทดสอบรหัสผ่านที่ถูกต้อง
	err := user.SetPassword("securepwd")
	assert.NoError(t, err)
	assert.True(t, user.CheckPassword("securepwd"))

	// ทดสอบรหัสผ่านที่ไม่ถูกต้อง
	assert.False(t, user.CheckPassword("hackattempt"))
	assert.False(t, user.CheckPassword(""))
}

func TestUser_BeforeCreate(t *testing.T) {
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "rawpassword",
	}

	// จำลอง GORM DB transaction
	mockDB := &gorm.DB{}

	// ทดสอบ hook BeforeCreate
	err := user.BeforeCreate(mockDB)
	assert.NoError(t, err)
	assert.NotEqual(t, "rawpassword", user.Password)

	// ตรวจสอบว่ารหัสผ่านถูกเข้ารหัสอย่างถูกต้อง
	assert.True(t, user.CheckPassword("rawpassword"))
}

func TestUser_ToResponse(t *testing.T) {
	now := time.Now()
	user := &User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		FullName:  "Test User",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// ทดสอบการแปลงเป็น response
	response := user.ToResponse()

	// ตรวจสอบข้อมูลที่ถูกต้อง
	assert.Equal(t, uint(1), response["id"])
	assert.Equal(t, "testuser", response["username"])
	assert.Equal(t, "test@example.com", response["email"])
	assert.Equal(t, "Test User", response["full_name"])

	// ตรวจสอบว่าไม่มีข้อมูลที่ sensitive
	_, hasPassword := response["password"]
	assert.False(t, hasPassword)
	_, hasCreatedAt := response["created_at"]
	assert.False(t, hasCreatedAt)
	_, hasUpdatedAt := response["updated_at"]
	assert.False(t, hasUpdatedAt)
}
