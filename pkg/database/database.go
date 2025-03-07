package database

import (
	"fmt"
	"log"

	"github.com/yourusername/auth-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config สำหรับการเชื่อมต่อฐานข้อมูล
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewConnection สร้างการเชื่อมต่อฐานข้อมูลใหม่
func NewConnection(config *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// MigrateDB สร้างหรืออัปเดตโครงสร้างฐานข้อมูล
func MigrateDB(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
	)
	if err != nil {
		return err
	}

	return nil
}

// SeedDefaultData สร้างข้อมูลเริ่มต้นในฐานข้อมูล
func SeedDefaultData(db *gorm.DB) error {
	// สร้าง permissions
	permissions := []models.Permission{
		{Resource: "users", Action: "read", Description: "อ่านข้อมูลผู้ใช้"},
		{Resource: "users", Action: "write", Description: "แก้ไขข้อมูลผู้ใช้"},
		{Resource: "roles", Action: "read", Description: "อ่านข้อมูลบทบาท"},
		{Resource: "roles", Action: "write", Description: "แก้ไขข้อมูลบทบาท"},
		{Resource: "permissions", Action: "read", Description: "อ่านข้อมูลสิทธิ์"},
		{Resource: "permissions", Action: "write", Description: "แก้ไขข้อมูลสิทธิ์"},
	}

	for _, perm := range permissions {
		var existingPerm models.Permission
		result := db.Where("resource = ? AND action = ?", perm.Resource, perm.Action).First(&existingPerm)
		if result.RowsAffected == 0 {
			if err := db.Create(&perm).Error; err != nil {
				log.Printf("Failed to create permission %s:%s: %v", perm.Resource, perm.Action, err)
			}
		}
	}

	// สร้าง roles พื้นฐาน
	adminRole := models.Role{Name: "admin", Description: "ผู้ดูแลระบบ"}
	supervisorRole := models.Role{Name: "supervisor", Description: "ผู้ควบคุม"}
	editorRole := models.Role{Name: "editor", Description: "ผู้แก้ไข"}
	viewerRole := models.Role{Name: "viewer", Description: "ผู้ดู"}

	// ตรวจสอบและสร้าง role ถ้ายังไม่มี
	var existingRole models.Role

	if db.Where("name = ?", adminRole.Name).First(&existingRole).RowsAffected == 0 {
		if err := db.Create(&adminRole).Error; err != nil {
			log.Printf("Failed to create admin role: %v", err)
		} else {
			// ให้สิทธิ์ทั้งหมดกับ admin
			var allPermissions []models.Permission
			db.Find(&allPermissions)
			db.Model(&adminRole).Association("Permissions").Append(allPermissions)
		}
	}

	if db.Where("name = ?", supervisorRole.Name).First(&existingRole).RowsAffected == 0 {
		if err := db.Create(&supervisorRole).Error; err != nil {
			log.Printf("Failed to create supervisor role: %v", err)
		} else {
			// ให้สิทธิ์อ่านทั้งหมดและแก้ไขบางส่วนกับ supervisor
			var readPermissions []models.Permission
			db.Where("action = ?", "read").Find(&readPermissions)
			db.Model(&supervisorRole).Association("Permissions").Append(readPermissions)
		}
	}

	if db.Where("name = ?", editorRole.Name).First(&existingRole).RowsAffected == 0 {
		if err := db.Create(&editorRole).Error; err != nil {
			log.Printf("Failed to create editor role: %v", err)
		} else {
			// ให้สิทธิ์อ่านและแก้ไขบางส่วนกับ editor
			var userPermissions []models.Permission
			db.Where("resource = ? AND action = ?", "users", "read").Or("resource = ? AND action = ?", "users", "write").Find(&userPermissions)
			db.Model(&editorRole).Association("Permissions").Append(userPermissions)
		}
	}

	if db.Where("name = ?", viewerRole.Name).First(&existingRole).RowsAffected == 0 {
		if err := db.Create(&viewerRole).Error; err != nil {
			log.Printf("Failed to create viewer role: %v", err)
		} else {
			// ให้สิทธิ์อ่านอย่างเดียวกับ viewer
			var readPermissions []models.Permission
			db.Where("action = ?", "read").Find(&readPermissions)
			db.Model(&viewerRole).Association("Permissions").Append(readPermissions)
		}
	}

	// สร้าง admin user เริ่มต้น
	adminUser := models.User{
		Username: "admin",
		Email:    "admin@example.com",
		Password: "adminpassword", // จะถูกเข้ารหัสโดย BeforeCreate hook
		FullName: "System Administrator",
	}

	var existingUser models.User
	if db.Where("username = ?", adminUser.Username).First(&existingUser).RowsAffected == 0 {
		if err := db.Create(&adminUser).Error; err != nil {
			log.Printf("Failed to create admin user: %v", err)
		} else {
			// กำหนด role admin ให้กับ user admin
			var adminRoleModel models.Role
			db.Where("name = ?", "admin").First(&adminRoleModel)
			db.Model(&adminUser).Association("Roles").Append(&adminRoleModel)
		}
	}

	return nil
}
