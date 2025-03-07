package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/auth-api/internal/models"
	"gorm.io/gorm"
)

type RoleHandler struct {
	db *gorm.DB
}

func NewRoleHandler(db *gorm.DB) *RoleHandler {
	return &RoleHandler{
		db: db,
	}
}

// GetRoles รับรายการบทบาททั้งหมด
func (h *RoleHandler) GetRoles(c *gin.Context) {
	var roles []models.Role
	result := h.db.Preload("Permissions").Find(&roles)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch roles"})
		return
	}

	c.JSON(http.StatusOK, roles)
}

// GetRole รับข้อมูลบทบาทตาม ID
func (h *RoleHandler) GetRole(c *gin.Context) {
	id := c.Param("id")
	roleID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var role models.Role
	result := h.db.Preload("Permissions").First(&role, roleID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	c.JSON(http.StatusOK, role)
}

// CreateRole สร้างบทบาทใหม่
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var role models.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ตรวจสอบว่ามีชื่อบทบาทซ้ำหรือไม่
	var existingRole models.Role
	if result := h.db.Where("name = ?", role.Name).First(&existingRole); result.RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role name already exists"})
		return
	}

	// บันทึกบทบาทใหม่
	result := h.db.Create(&role)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// UpdateRole อัปเดตข้อมูลบทบาท
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	roleID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var role models.Role
	if result := h.db.First(&role, roleID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	var updateData struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// อัปเดตข้อมูลที่ไม่ใช่ค่าว่าง
	updates := make(map[string]interface{})
	if updateData.Name != "" {
		// ตรวจสอบว่ามีชื่อบทบาทซ้ำหรือไม่
		if updateData.Name != role.Name {
			var existingRole models.Role
			if result := h.db.Where("name = ?", updateData.Name).First(&existingRole); result.RowsAffected > 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Role name already exists"})
				return
			}
		}
		updates["name"] = updateData.Name
	}

	if updateData.Description != "" {
		updates["description"] = updateData.Description
	}

	// อัปเดตข้อมูล
	if result := h.db.Model(&role).Updates(updates); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	// ดึงข้อมูลบทบาทที่อัปเดตแล้ว
	h.db.Preload("Permissions").First(&role, roleID)

	c.JSON(http.StatusOK, role)
}

// DeleteRole ลบบทบาท
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	roleID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	// ลบบทบาท
	result := h.db.Delete(&models.Role{}, roleID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete role"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

// AddPermissionToRole เพิ่มสิทธิ์ให้กับบทบาท
func (h *RoleHandler) AddPermissionToRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var requestData struct {
		PermissionID uint `json:"permission_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role models.Role
	if result := h.db.First(&role, roleID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	var permission models.Permission
	if result := h.db.First(&permission, requestData.PermissionID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
		return
	}

	// เพิ่มสิทธิ์ให้กับบทบาท
	if err := h.db.Model(&role).Association("Permissions").Append(&permission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add permission to role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission added to role successfully"})
}

// RemovePermissionFromRole ลบสิทธิ์ออกจากบทบาท
func (h *RoleHandler) RemovePermissionFromRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	permissionID, err := strconv.ParseUint(c.Param("permissionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	var role models.Role
	if result := h.db.First(&role, roleID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	var permission models.Permission
	if result := h.db.First(&permission, permissionID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
		return
	}

	// ลบสิทธิ์ออกจากบทบาท
	if err := h.db.Model(&role).Association("Permissions").Delete(&permission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove permission from role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission removed from role successfully"})
}
