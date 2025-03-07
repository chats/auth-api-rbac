package config

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config โครงสร้างการตั้งค่าของแอปพลิเคชัน
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

// ServerConfig การตั้งค่าเซิร์ฟเวอร์
type ServerConfig struct {
	Port    string
	Timeout time.Duration
}

// DatabaseConfig การตั้งค่าฐานข้อมูล
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// JWTConfig การตั้งค่า JWT
type JWTConfig struct {
	SecretKey     string
	Issuer        string
	TokenDuration time.Duration
}

// LoadConfig โหลดการตั้งค่าจากไฟล์หรือตัวแปรสภาพแวดล้อม
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// แทนที่ _ ด้วย . สำหรับตัวแปรสภาพแวดล้อม (เช่น SERVER_PORT -> SERVER.PORT)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// อ่านไฟล์การตั้งค่า (ไม่ error ถ้าไม่พบไฟล์)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// ไม่ใช่ error เกี่ยวกับไฟล์ไม่พบ
			return nil, err
		}
		log.Printf("Config file not found. Using environment variables.")
	} else {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}

	// ค่าเริ่มต้น
	// Server config
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.timeout", 10*time.Second)

	// Database config
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "auth_api")
	viper.SetDefault("database.sslmode", "disable")

	// JWT config
	viper.SetDefault("jwt.secretKey", "your-secret-key")
	viper.SetDefault("jwt.issuer", "auth-api")
	viper.SetDefault("jwt.tokenDuration", 24*time.Hour)

	// ตรวจสอบตัวแปรสภาพแวดล้อมโดยตรง (สนับสนุนทั้งรูปแบบพื้นฐานและรูปแบบ Docker Compose)
	checkEnvOverride("SERVER_PORT", "server.port")
	checkEnvOverride("DATABASE_HOST", "database.host")
	checkEnvOverride("DATABASE_PORT", "database.port")
	checkEnvOverride("DATABASE_USER", "database.user")
	checkEnvOverride("DATABASE_PASSWORD", "database.password")
	checkEnvOverride("DATABASE_DBNAME", "database.dbname")
	checkEnvOverride("DATABASE_SSLMODE", "database.sslmode")
	checkEnvOverride("JWT_SECRETKEY", "jwt.secretKey")
	checkEnvOverride("JWT_ISSUER", "jwt.issuer")
	checkEnvOverrideDuration("JWT_TOKENDURATION", "jwt.tokenDuration")

	config := &Config{
		Server: ServerConfig{
			Port:    viper.GetString("server.port"),
			Timeout: viper.GetDuration("server.timeout"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("database.host"),
			Port:     viper.GetString("database.port"),
			User:     viper.GetString("database.user"),
			Password: viper.GetString("database.password"),
			DBName:   viper.GetString("database.dbname"),
			SSLMode:  viper.GetString("database.sslmode"),
		},
		JWT: JWTConfig{
			SecretKey:     viper.GetString("jwt.secretKey"),
			Issuer:        viper.GetString("jwt.issuer"),
			TokenDuration: viper.GetDuration("jwt.tokenDuration"),
		},
	}

	return config, nil
}

// checkEnvOverride ตรวจสอบตัวแปรสภาพแวดล้อมและกำหนดค่าให้ viper ถ้ามี
func checkEnvOverride(envName, configPath string) {
	if val, exists := os.LookupEnv(envName); exists {
		viper.Set(configPath, val)
		log.Printf("Environment override: %s -> %s", envName, configPath)
	}
}

// checkEnvOverrideDuration เหมือน checkEnvOverride แต่สำหรับค่า time.Duration
func checkEnvOverrideDuration(envName, configPath string) {
	if val, exists := os.LookupEnv(envName); exists {
		duration, err := time.ParseDuration(val)
		if err != nil {
			log.Printf("Warning: could not parse duration from %s: %v", envName, err)
			return
		}
		viper.Set(configPath, duration)
		log.Printf("Environment override: %s -> %s", envName, configPath)
	}
}
