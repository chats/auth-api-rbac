# ตัวแปรพื้นฐาน
APP_NAME = auth-api
MAIN_PACKAGE = ./cmd/api
DOCKER_IMAGE = auth-api:latest
GO_FILES = $(shell find . -name "*.go" -not -path "./vendor/*")

# สั่งเริ่มต้น
.PHONY: all
all: build

# พัฒนา Go
.PHONY: build run clean test lint vet fmt mod-tidy

# สร้างแอปพลิเคชัน
build:
	go build -o bin/$(APP_NAME) $(MAIN_PACKAGE)

# รันแอปพลิเคชันในโหมดพัฒนา
run:
	go run $(MAIN_PACKAGE)

# ลบไฟล์ที่สร้างขึ้น
clean:
	rm -rf bin/
	rm -rf dist/
	go clean

# รัน unit tests
test:
	go test -v ./...

# รัน linter
lint:
	golint ./...

# ตรวจสอบปัญหาทั่วไปในโค้ด
vet:
	go vet ./...

# จัดรูปแบบโค้ด
fmt:
	gofmt -s -w $(GO_FILES)

# จัดการ dependencies
mod-tidy:
	go mod tidy

# Docker
.PHONY: docker-build docker-run docker-push

# สร้าง Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# รัน Docker container
docker-run:
	docker run --rm -p 8080:8080 $(DOCKER_IMAGE)

# ส่ง Docker image ไปยัง registry
docker-push:
	docker push $(DOCKER_IMAGE)

# Docker Compose
.PHONY: docker-compose-up docker-compose-down docker-compose-logs

# รัน docker-compose
docker-compose-run:
	docker-compose up

# รัน docker-compose background
docker-compose-up:
	docker-compose up -d

# หยุด docker-compose
docker-compose-down:
	docker-compose down

# แสดง logs ของ docker-compose
docker-compose-logs:
	docker-compose logs -f

# รันแอปพลิเคชันทั้งหมดด้วย Docker Compose
.PHONY: up
up: docker-compose-up

# หยุดแอปพลิเคชันทั้งหมด
.PHONY: down
down: docker-compose-down

# อีกทางเลือกในการสร้างและรัน
.PHONY: dev prod

# เริ่มในโหมดพัฒนา
dev: build run

# สร้างและเริ่มในโหมดการผลิต
prod: docker-build docker-compose-up

# ช่วยเหลือ
.PHONY: help
help:
	@echo "สามารถใช้คำสั่งต่อไปนี้:"
	@echo "  make build              - สร้างแอปพลิเคชัน"
	@echo "  make run                - รันแอปพลิเคชันในโหมดพัฒนา"
	@echo "  make clean              - ลบไฟล์ที่สร้างขึ้น"
	@echo "  make test               - รัน unit tests"
	@echo "  make lint               - รัน linter"
	@echo "  make vet                - ตรวจสอบปัญหาทั่วไปในโค้ด"
	@echo "  make fmt                - จัดรูปแบบโค้ด"
	@echo "  make mod-tidy           - จัดการ dependencies"
	@echo "  make docker-build       - สร้าง Docker image"
	@echo "  make docker-run         - รัน Docker container"
	@echo "  make docker-push        - ส่ง Docker image ไปยัง registry"
	@echo "  make docker-compose-up  - รัน docker-compose"
	@echo "  make docker-compose-down - หยุด docker-compose"
	@echo "  make docker-compose-logs - แสดง logs ของ docker-compose"
	@echo "  make up                 - เริ่มแอปพลิเคชันด้วย Docker Compose"
	@echo "  make down               - หยุดแอปพลิเคชัน"
	@echo "  make dev                - เริ่มในโหมดพัฒนา"
	@echo "  make prod               - สร้างและเริ่มในโหมดการผลิต"