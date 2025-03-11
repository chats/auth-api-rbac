# AUTH-API (Sample RBAC)
## API Endpoints
### การ Authentication
- POST /api/login: เข้าสู่ระบบและรับ JWT token
```
{
  "username": "admin",
  "password": "adminpassword"
}
```
### การจัดการผู้ใช้ (User Management)
- ```GET /api/users```: รับรายการผู้ใช้ทั้งหมด
- ```GET /api/users/:id```: รับข้อมูลผู้ใช้ตาม ID
- ```POST /api/users```: สร้างผู้ใช้ใหม่
- ```PUT /api/users/:id```: อัปเดตข้อมูลผู้ใช้
- ```DELETE /api/users/:id```: ลบผู้ใช้
- ```POST /api/users/:id/roles```: เพิ่มบทบาทให้กับผู้ใช้
- ```DELETE /api/users/:id/roles/:roleId```: ลบบทบาทออกจากผู้ใช้
### การจัดการบทบาท (Role Management)
- ```GET /api/roles```: รับรายการบทบาททั้งหมด
- ```GET /api/roles/:id```: รับข้อมูลบทบาทตาม ID
- ```POST /api/roles```: สร้างบทบาทใหม่
- ```PUT /api/roles/:id```: อัปเดตข้อมูลบทบาท
- ```DELETE /api/roles/:id```: ลบบทบาท
- ```POST /api/roles/:id/permissions```: เพิ่มสิทธิ์ให้กับบทบาท
- ```DELETE /api/roles/:id/permissions/:permissionId```: ลบสิทธิ์ออกจากบทบาท
### การจัดการสิทธิ์ (Permission Management)
- ```GET /api/permissions```: รับรายการสิทธิ์ทั้งหมด
- ```GET /api/permissions/:id```: รับข้อมูลสิทธิ์ตาม ID
- ```POST /api/permissions```: สร้างสิทธิ์ใหม่
- ```PUT /api/permissions/:id```: อัปเดตข้อมูลสิทธิ์
- ```DELETE /api/permissions/:id```: ลบสิทธิ์

<br>

## ตัวอย่างการใช้งาน
1. การเข้าสู่ระบบ (Login)
```
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "adminpassword"}'
```

2. การเรียกดูข้อมูลผู้ใช้ทั้งหมด (Get all users)
```
curl -X GET http://localhost:8080/api/users \
  -H "Authorization: Bearer <your_access_token>"
```

3. การสร้างผู้ใช้ใหม่ (Create a new user)
```
curl -X POST http://localhost:8080/api/users \
  -H "Authorization: Bearer <your_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"username": "newuser", "email": "newuser@example.com", 
	"password": "password123", "full_name": "New User"}'
```

4. การเพิ่มบทบาทให้กับผู้ใช้ (Add role to user)
```
curl -X POST http://localhost:8080/api/users/2/roles \
  -H "Authorization: Bearer <your_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"role_id": 2}'
```

<br>

## Tests

### หลักการทดสอบที่ใช้

1. Unit Testing: ทดสอบส่วนประกอบย่อยๆ แยกจากกัน เช่น:
	- ทดสอบโมเดล User และการเข้ารหัสรหัสผ่าน
	- ทดสอบระบบ JWT
	- ทดสอบ Service และ Middleware

2. Integration Testing: ทดสอบการทำงานร่วมกันของหลายส่วนประกอบ:
	- ทดสอบ API endpoints ทั้งระบบ

3. Mock Testing: จำลองส่วนประกอบบางอย่างในระบบ เช่น ฐานข้อมูล เพื่อให้สามารถทดสอบได้อย่างแยกส่วน
	- จำลองฐานข้อมูลโดยไม่ต้องใช้ฐานข้อมูลจริง
	- จำลอง interfaces ต่างๆ เพื่อควบคุมพฤติกรรมในการทดสอบ


4. Behavior-Driven Development (BDD): ใช้แนวคิดการทดสอบตามพฤติกรรมที่คาดหวัง
	- ใช้รูปแบบ Test Suite



### Framework และ Libraries ที่ใช้

1. testing: แพ็คเกจมาตรฐานของ Go สำหรับการทดสอบ
2. stretchr/testify:
	- ```assert```: ใช้สำหรับการตรวจสอบค่าที่คาดหวัง
	- ```suite```: ใช้สำหรับจัดกลุ่มการทดสอบและจัดการ setup/teardown

3. DATA-DOG/go-sqlmock: ใช้จำลองฐานข้อมูล SQL
	- จำลองการเรียกใช้ SQL queries และกำหนดผลลัพธ์
	- ทดสอบโดยไม่ต้องเชื่อมต่อกับฐานข้อมูลจริง

4. golang/mock/mockgen: สร้าง mock objects จาก interfaces
	- ใช้จำลองพฤติกรรมของ interfaces ต่างๆ

5. steinfletcher/apitest: ใช้ทดสอบ API endpoints
	- สร้าง HTTP requests ไปยัง endpoints
	- ตรวจสอบ HTTP responses

6. steinfletcher/apitest-jsonpath: ใช้ตรวจสอบข้อมูล JSON ใน response
	- ตรวจสอบข้อมูลใน JSON ด้วย JSONPath expressions


### ประเภทของการทดสอบ

1. โมเดลทดสอบ:
	- ทดสอบการเข้ารหัสและตรวจสอบรหัสผ่าน
	- ทดสอบ hooks ต่างๆ เช่น BeforeCreate


2. ทดสอบ JWT:
	- ทดสอบการสร้างและตรวจสอบ tokens
	- ทดสอบ tokens ที่หมดอายุหรือไม่ถูกต้อง


3. ทดสอบ Service:
	- ทดสอบการ login สำเร็จและไม่สำเร็จ
	- ทดสอบการดึงข้อมูลผู้ใช้
	- ทดสอบการตรวจสอบสิทธิ์


4. ทดสอบ Middleware:
	- ทดสอบการตรวจสอบ JWT tokens
	- ทดสอบการตรวจสอบสิทธิ์และบทบาท


5. ทดสอบ API Integration:
	- ทดสอบการเรียกใช้ API endpoints ทั้งหมด
	- ทดสอบการทำงานร่วมกันของทั้งระบบ

### Run Tests
```
# รัน unit tests
make test-unit

# รัน integration tests
make test-integration

# รัน tests ทั้งหมด
make test

# รัน tests พร้อมสร้างรายงานความครอบคลุมของโค้ด
make cover

# รัน tests ด้วย Docker
make docker-compose-test
```