# 图床转换（本地存储版）

基于 Go + Vue3 + MySQL 的图片上传与链接生成系统。当前实现仅使用本地磁盘存储（backend/uploads），不依赖任何云对象存储。后端提供鉴权、上传、列表/删除、统计与系统状态等接口，前端提供可视化上传与管理页面。

备注：代码中的服务名“R2”仅为历史命名，实际实现为本地文件存储。

## 功能概览
- 账号与游客码（管理员登录、游客码登录；创建/查看/删除游客码）
- 图片管理（单图/批量上传、列表、查询、删除、直链生成）
- 统计与系统状态（总量、今日、平均/最大/最小，磁盘/内存/CPU、运行时长）
- 安全与体验（速率限制、类型/大小校验、静态资源映射、统一错误码）

## 界面预览
- 首页  
![](https://qone.kuz7.com/uploads/images/2025/09/07/d284da60-19b2-4682-b695-8a464eb45276.png)

- 上传  
![](https://qone.kuz7.com/uploads/images/2025/09/07/9e63dd82-c1f5-4c79-a5d5-eb457dfc8e7e.png)

- 图库  
![](https://qone.kuz7.com/uploads/images/2025/09/07/803aa069-6d7a-436a-86ad-ab121dd485ae.png)

- 统计  
![](https://qone.kuz7.com/uploads/images/2025/09/07/78916c57-538d-49d8-9b82-8c14d65b4946.png)

- 设置  
![](https://qone.kuz7.com/uploads/images/2025/09/07/e74a7fa2-faf0-44f8-9086-77c151a8c678.png)

## 技术栈
- 前端：Vue 3、TypeScript、Vite、Element Plus、Pinia、Axios
- 后端：Go、Gin、GORM、MySQL
- 存储：本地磁盘（默认 ./uploads）
- 其他：可选 Redis 字段已预留（当前代码未实际启用）

## 目录结构（简化）
```
图床转换/
├── backend/
│   ├── config/            # 配置加载（.env）
│   ├── controllers/       # Auth/Upload/System/GuestCode
│   ├── database/          # MySQL 初始化与迁移/默认管理员
│   ├── middleware/        # 速率限制、鉴权
│   ├── models/            # Image/ImageStats/User/GuestCode
│   ├── routes/            # 路由（/api/v1）
│   ├── services/          # image 处理、本地存储（R2 命名）、游客码清理
│   └── uploads/           # 本地图片目录（映射为 /uploads）
├── frontend/
│   └── src/               # 业务代码（api/views/components 等）
└── nginx/                 # 可选的 Nginx 示例
```

## 快速开始（本地）
### 依赖
- Node.js ≥ 16
- Go ≥ 1.19
- MySQL ≥ 8.0

### 1) 创建数据库
```sql
CREATE DATABASE image_host CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 2) 后端
```bash
cd backend
go mod tidy
```

在 backend 目录下新建 `.env`（示例）：
```env
# 服务
SERVER_PORT=8080

# 鉴权
JWT_SECRET=change_me_secret
JWT_EXPIRE_HOURS=72
DEFAULT_ADMIN=root
DEFAULT_PASSWORD=123456

# 数据库
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=image_host

# 上传
MAX_FILE_SIZE=10485760
# 默认允许类型（建议与处理能力一致，见“上传与格式”）
ALLOWED_TYPES=image/jpeg,image/png
UPLOAD_PATH=./uploads
```

启动：
```bash
go run main.go
```
- 健康检查：GET http://localhost:8080/health
- 静态资源：/uploads 映射到 UPLOAD_PATH

### 3) 前端
```bash
cd frontend
npm install
npm run dev
```
默认地址：http://localhost:3000

## 功能详解
### 1. 账号与游客码
- 登录
  - POST /api/v1/auth/login
  - Body: { "username": string, "password": string }
  - 返回: { success, data: { token, username, expires } }
  - 说明：首次启动会自动创建默认管理员（用户名 DEFAULT_ADMIN，密码 DEFAULT_PASSWORD）。
- 游客码登录
  - POST /api/v1/auth/guest-login
  - Body: { "code": string }（不区分大小写，自动去除空格）
  - 返回: { success, data: { token, username: "guest", expires } }
- 自身信息
  - GET /api/v1/auth/me（受保护）
- 修改密码
  - POST /api/v1/auth/change-password（受保护）
  - Body: { "old_password": string, "new_password": string }

- 游客码管理（仅管理员，受保护）
  - POST /api/v1/guest-codes/        创建游客码（支持永久码或过期时间）
  - GET  /api/v1/guest-codes/        列出游客码
  - DELETE /api/v1/guest-codes/:id   删除游客码
  - 备注：服务包含后台清理任务，定期移除过期游客码。

鉴权方式：除 /health、/api/v1/auth/login、/api/v1/auth/guest-login 外，其余均需在请求头携带
Authorization: Bearer <token>

### 2. 图片上传与管理
- 单图上传（受保护）
  - POST /api/v1/images/upload
  - multipart/form-data：字段名 image
  - 返回：uuid、public_url、尺寸、大小、类型、时间等
- 批量上传（受保护）
  - POST /api/v1/batch-upload
  - multipart/form-data：字段名 images（最多 10 张，支持总大小限制）
- 图片列表（受保护）
  - GET /api/v1/images?page=1&page_size=20
  - 返回：{ items, total, page, page_size }
  - 权限：管理员可查看全部；其他仅查看自己上传的记录
- 获取图片详情（受保护）
  - GET /api/v1/images/:uuid
- 删除图片（受保护）
  - DELETE /api/v1/images/:uuid
  - 逻辑：删除本地文件（忽略不存在错误）→ 硬删数据库记录
- 统计汇总（受保护）
  - GET /api/v1/images/stats/summary
  - 返回：total_images、total_size、today_images 等

存储说明：后端将文件保存到 UPLOAD_PATH 下的 images/yyyy/mm/dd 目录；public_url 为相对路径 /uploads/...，由后端静态映射提供访问。

### 3. 系统状态与健康检查
- 健康检查（无需鉴权）
  - GET /health
- 系统状态（受保护）
  - GET /api/v1/system/status
  - 返回：uptime_seconds、磁盘（total/free/used/used_percent）、内存（total/used/free/used_percent）、cpu_usage、以及图片汇总（总量/今日/平均/最大/最小）

### 4. 速率限制与错误规范
- 速率限制
  - 以 IP 为维度，默认 60 请求/分钟
  - 超限返回 429，包含 { error, code: "RATE_LIMIT_EXCEEDED", retry_after: 60 }
- 错误响应
  - 统一返回 { error, code }，便于前端处理

## 上传与格式
- 建议默认允许：image/jpeg, image/png
- 单文件大小上限：MAX_FILE_SIZE（默认 10MB）
- 图片处理：
  - >1MB 会尝试压缩（JPEG 85% 质量；PNG 使用最佳压缩）
  - 自动获取宽高与缩略图（300x300）
- 若需 GIF/WebP：
  - 在后端增加对应解码器导入（示例：GIF 需 import image/gif，WebP 可用 golang.org/x/image/webp）
  - 同步将 ALLOWED_TYPES 加入相应 MIME 类型

## 前端页面（简要）
- 登录/Login：用户名密码或游客码登录；本地存储 token
- 上传/Upload：单图与批量上传，显示上传进度与结果
- 图库/Gallery：分页查看已上传图片，支持复制直链/Markdown/HTML/BBCode
- 存储/Storage：查看直链与访问路径（/uploads/...）
- 统计/Stats：展示图片统计与系统状态（磁盘/内存/CPU/运行时间）
- 设置/Settings：修改密码等
- 说明：前端通过 /api/v1 前缀请求后端，并在请求头附带 Authorization

## 常见问题
- 直链 404：确认 UPLOAD_PATH 存在且后端路由已映射 /uploads
- 无法写入：检查进程对 UPLOAD_PATH 的写权限
- 跨域：前后端同域最佳；否则在后端 CORS 放行前端域名
- 只看到自己的图片：非管理员账号仅能查看/删除自己上传的数据

## 许可证
MIT