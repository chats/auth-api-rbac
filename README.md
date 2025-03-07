# AUTH-API (RBAC)
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