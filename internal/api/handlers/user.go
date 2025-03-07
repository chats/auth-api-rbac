package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/auth-api/internal/models"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{
		db: db,
	}
}

// GetUsers รับรายการผู้ใช้ทั้งหมด
func (h *UserHandler) GetUsers(c *gin.Context) {
	var users []models.User
	result := h.db.Preload("Roles").Find(&users)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// แปลงเป็น response ที่ไม่มีข้อมูล sensitive
	var usersResponse []map[string]interface{}
	for _, user := range users {
		usersResponse = append(usersResponse, user.ToResponse())
	}

	c.JSON(http.StatusOK, usersResponse)
}

// GetUser รับข้อมูลผู้ใช้ตาม ID
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	result := h.db.Preload("Roles").First(&user, userID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// CreateUser สร้างผู้ใช้ใหม่
func (h *UserHandler) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ตรวจสอบว่ามี username หรือ email ซ้ำหรือไม่
	var existingUser models.User
	if result := h.db.Where("username = ? OR email = ?", user.Username, user.Email).First(&existingUser); result.RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email already exists"})
		return
	}

	// บันทึกผู้ใช้ใหม่
	result := h.db.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user.ToResponse())
}

// UpdateUser อัปเดตข้อมูลผู้ใช้
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if result := h.db.First(&user, userID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var updateData struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// อัปเดตข้อมูลที่ไม่ใช่ค่าว่าง
	updates := make(map[string]interface{})
	if updateData.Username != "" {
		// ตรวจสอบว่ามี username ซ้ำหรือไม่
		if updateData.Username != user.Username {
			var existingUser models.User
			if result := h.db.Where("username = ?", updateData.Username).First(&existingUser); result.RowsAffected > 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
				return
			}
		}
		updates["username"] = updateData.Username
	}

	if updateData.Email != "" {
		// ตรวจสอบว่ามี email ซ้ำหรือไม่
		if updateData.Email != user.Email {
			var existingUser models.User
			if result := h.db.Where("email = ?", updateData.Email).First(&existingUser); result.RowsAffected > 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
				return
			}
		}
		updates["email"] = updateData.Email
	}

	if updateData.Password != "" {
		// เข้ารหัสรหัสผ่านใหม่
		if err := user.SetPassword(updateData.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}
		updates["password"] = user.Password
	}

	if updateData.FullName != "" {
		updates["full_name"] = updateData.FullName
	}

	// อัปเดตข้อมูล
	if result := h.db.Model(&user).Updates(updates); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// ดึงข้อมูลผู้ใช้ที่อัปเดตแล้ว
	h.db.Preload("Roles").First(&user, userID)

	c.JSON(http.StatusOK, user.ToResponse())
}

// DeleteUser ลบผู้ใช้
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// ลบผู้ใช้
	result := h.db.Delete(&models.User{}, userID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// AddRoleToUser เพิ่มบทบาทให้กับผู้ใช้
func (h *UserHandler) AddRoleToUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var requestData struct {
		RoleID uint `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if result := h.db.First(&user, userID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var role models.Role
	if result := h.db.First(&role, requestData.RoleID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	// เพิ่มบทบาทให้กับผู้ใช้
	if err := h.db.Model(&user).Association("Roles").Append(&role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add role to user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role added to user successfully"})
}

// RemoveRoleFromUser ลบบทบาทออกจากผู้ใช้
func (h *UserHandler) RemoveRoleFromUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	roleID, err := strconv.ParseUint(c.Param("roleId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var user models.User
	if result := h.db.First(&user, userID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var role models.Role
	if result := h.db.First(&role, roleID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	// ลบบทบาทออกจากผู้ใช้
	if err := h.db.Model(&user).Association("Roles").Delete(&role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove role from user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role removed from user successfully"})
}
