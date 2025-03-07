package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/auth-api/internal/api/handlers"
	"github.com/yourusername/auth-api/internal/api/middlewares"
	"github.com/yourusername/auth-api/internal/config"
	"github.com/yourusername/auth-api/internal/service"
	"github.com/yourusername/auth-api/pkg/database"
	"github.com/yourusername/auth-api/pkg/jwt"
)

func main() {
	// โหลดการตั้งค่า
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// เชื่อมต่อฐานข้อมูล
	dbConfig := &database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// สร้างโครงสร้างฐานข้อมูล
	if err := database.MigrateDB(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// สร้างข้อมูลเริ่มต้น
	if err := database.SeedDefaultData(db); err != nil {
		log.Printf("Warning: Failed to seed initial data: %v", err)
	}

	// สร้าง JWT service
	jwtService := jwt.NewJWTService(
		cfg.JWT.SecretKey,
		cfg.JWT.Issuer,
		cfg.JWT.TokenDuration,
	)

	// สร้าง services
	authService := service.NewAuthService(db, jwtService)

	// สร้าง handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(db)
	roleHandler := handlers.NewRoleHandler(db)
	permissionHandler := handlers.NewPermissionHandler(db)

	// สร้าง middlewares
	authMiddleware := middlewares.AuthMiddleware(jwtService, authService)

	// สร้าง Gin router
	r := gin.Default()

	// API routes
	r.POST("/api/login", authHandler.Login)

	// กลุ่ม routes ที่ต้องการการยืนยันตัวตน
	authorized := r.Group("/api")
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

	// เริ่มต้นเซิร์ฟเวอร์
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", serverAddr)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
