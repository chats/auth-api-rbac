# ตัวแปรพื้นฐาน
APP_NAME = auth-api
MAIN_PACKAGE = ./cmd/api
DOCKER_IMAGE = auth-api:latest
GO_FILES = $(shell find . -name "*.go" -not -path "./vendor/*")
TEST_PACKAGES = $(shell go list ./... | grep -v /vendor/)

# สั่งเริ่มต้น
.PHONY: all
all: build

# พัฒนา Go
.PHONY: build run clean test test-unit test-integration cover lint vet fmt mod-tidy

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
	rm -rf coverage/
	rm -f coverage.out
	go clean

# รัน unit tests
test-unit:
	go test -v $(TEST_PACKAGES) -short -count=1

# รัน integration tests
test-integration:
	@echo "Running integration tests..."
	@if [ -z "$(SKIP_INTEGRATION_TESTS)" ]; then \
		go test -v -run TestAPIIntegration; \
	else \
		echo "Integration tests skipped"; \
	fi

# รัน tests ทั้งหมด
test: test-unit test-integration

# รันการทดสอบพร้อมสร้างรายงานความครอบคลุมของโค้ด
cover:
	@mkdir -p coverage
	go test -v $(TEST_PACKAGES) -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage/index.html
	@echo "Coverage report generated at coverage/index.html"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

# รัน linter
lint:
	@if ! [ -x "$$(command -v golint)" ]; then \
		echo "Installing golint..."; \
		go get -u golang.org/x/lint/golint; \
	fi
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
.PHONY: docker-compose-up docker-compose-down docker-compose-logs docker-compose-test

# รัน docker-compose
docker-compose-up:
	docker-compose up -d

# หยุด docker-compose
docker-compose-down:
	docker-compose down

# แสดง logs ของ docker-compose
docker-compose-logs:
	docker-compose logs -f

# รัน tests ด้วย docker-compose
docker-compose-test:
	docker-compose -f docker-compose.test.yaml up --build --abort-on-container-exit

# รันแอปพลิเคชันทั้งหมดด้วย Docker Compose
.PHONY: up
up: docker-compose-up

# หยุดแอปพลิเคชันทั้งหมด
.PHONY: down
down: docker-compose-down

# CI/CD
.PHONY: ci-test ci-build

# รัน CI tests
ci-test: test cover

# สร้างและรัน tests ใน CI
ci-build: build ci-test

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
	@echo "  make test               - รัน tests ทั้งหมด"
	@echo "  make test-unit          - รัน unit tests"
	@echo "  make test-integration   - รัน integration tests"
	@echo "  make cover              - รันการทดสอบพร้อมสร้างรายงานความครอบคลุมของโค้ด"
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
	@echo "  make docker-compose-test - รัน tests ด้วย docker-compose"
	@echo "  make up                 - เริ่มแอปพลิเคชันด้วย Docker Compose"
	@echo "  make down               - หยุดแอปพลิเคชัน"
	@echo "  make dev                - เริ่มในโหมดพัฒนา"
	@echo "  make prod               - สร้างและเริ่มในโหมดการผลิต"
	@echo "  make ci-test            - รัน CI tests"
	@echo "  make ci-build           - สร้างและรัน tests ใน CI"