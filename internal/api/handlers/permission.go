package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/auth-api/internal/models"
	"gorm.io/gorm"
)

type PermissionHandler struct {
	db *gorm.DB
}

func NewPermissionHandler(db *gorm.DB) *PermissionHandler {
	return &PermissionHandler{
		db: db,
	}
}

// GetPermissions รับรายการสิทธิ์ทั้งหมด
func (h *PermissionHandler) GetPermissions(c *gin.Context) {
	var permissions []models.Permission
	result := h.db.Find(&permissions)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch permissions"})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

// GetPermission รับข้อมูลสิทธิ์ตาม ID
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	id := c.Param("id")
	permissionID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	var permission models.Permission
	result := h.db.First(&permission, permissionID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
		return
	}

	c.JSON(http.StatusOK, permission)
}

// CreatePermission สร้างสิทธิ์ใหม่
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var permission models.Permission
	if err := c.ShouldBindJSON(&permission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ตรวจสอบว่ามีสิทธิ์ซ้ำหรือไม่
	var existingPermission models.Permission
	if result := h.db.Where("resource = ? AND action = ?", permission.Resource, permission.Action).First(&existingPermission); result.RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Permission already exists"})
		return
	}

	// บันทึกสิทธิ์ใหม่
	result := h.db.Create(&permission)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create permission"})
		return
	}

	c.JSON(http.StatusCreated, permission)
}

// UpdatePermission อัปเดตข้อมูลสิทธิ์
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id := c.Param("id")
	permissionID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	var permission models.Permission
	if result := h.db.First(&permission, permissionID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
		return
	}

	var updateData struct {
		Resource    string `json:"resource"`
		Action      string `json:"action"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// อัปเดตข้อมูลที่ไม่ใช่ค่าว่าง
	updates := make(map[string]interface{})
	if updateData.Resource != "" {
		updates["resource"] = updateData.Resource
	}

	if updateData.Action != "" {
		updates["action"] = updateData.Action
	}

	// ตรวจสอบว่ามีสิทธิ์ซ้ำหรือไม่หากมีการเปลี่ยนแปลง resource หรือ action
	if (updateData.Resource != "" && updateData.Resource != permission.Resource) ||
		(updateData.Action != "" && updateData.Action != permission.Action) {

		resourceToCheck := permission.Resource
		if updateData.Resource != "" {
			resourceToCheck = updateData.Resource
		}

		actionToCheck := permission.Action
		if updateData.Action != "" {
			actionToCheck = updateData.Action
		}

		var existingPermission models.Permission
		if result := h.db.Where("resource = ? AND action = ? AND id != ?", resourceToCheck, actionToCheck, permission.ID).
			First(&existingPermission); result.RowsAffected > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Permission already exists"})
			return
		}
	}

	if updateData.Description != "" {
		updates["description"] = updateData.Description
	}

	// อัปเดตข้อมูล
	if result := h.db.Model(&permission).Updates(updates); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update permission"})
		return
	}

	// ดึงข้อมูลสิทธิ์ที่อัปเดตแล้ว
	h.db.First(&permission, permissionID)

	c.JSON(http.StatusOK, permission)
}

// DeletePermission ลบสิทธิ์
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id := c.Param("id")
	permissionID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	// ลบสิทธิ์
	result := h.db.Delete(&models.Permission{}, permissionID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete permission"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission deleted successfully"})
}
