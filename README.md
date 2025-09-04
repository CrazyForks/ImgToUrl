# 高性能图床转换系统

一个基于 Vue3 + Go + MySQL + Cloudflare R2 的高性能图床转换网站，支持图片上传、压缩、存储和链接生成。

## ✨ 特性

- 🚀 **高性能**: Go 后端 + Vue3 前端，响应迅速
- 📁 **多格式支持**: 支持 JPEG、PNG、GIF、WebP、SVG 等主流图片格式
- 🗜️ **智能压缩**: 自动图片压缩和尺寸优化
- 📊 **批量上传**: 支持一次上传多张图片
- 🔗 **多种链接格式**: 直链、Markdown、HTML、BBCode
- 📈 **统计分析**: 上传统计、存储使用情况分析
- 🎨 **现代化UI**: 基于 Element Plus 的美观界面
- 📱 **响应式设计**: 完美适配桌面端和移动端
- ⚡ **CDN加速**: 基于 Cloudflare R2 的全球CDN
- 🔒 **安全可靠**: 文件类型验证、大小限制、速率限制

## 🏗️ 技术栈

### 前端
- **Vue 3** - 渐进式 JavaScript 框架
- **TypeScript** - 类型安全的 JavaScript
- **Vite** - 快速的前端构建工具
- **Element Plus** - Vue 3 组件库
- **Pinia** - Vue 状态管理
- **Vue Router** - Vue 路由管理
- **Axios** - HTTP 客户端

### 后端
- **Go** - 高性能编程语言
- **Gin** - 轻量级 Web 框架
- **GORM** - Go ORM 库
- **MySQL** - 关系型数据库
- **Cloudflare R2** - 对象存储服务
- **Redis** - 缓存和会话存储

## 📦 项目结构

```
图床转换/
├── backend/                 # Go 后端
│   ├── config/             # 配置管理
│   ├── controllers/        # 控制器
│   ├── middleware/         # 中间件
│   ├── models/            # 数据模型
│   ├── routes/            # 路由定义
│   ├── services/          # 业务逻辑
│   ├── utils/             # 工具函数
│   ├── main.go            # 程序入口
│   ├── go.mod             # Go 模块文件
│   └── .env.example       # 环境变量示例
├── frontend/               # Vue3 前端
│   ├── public/            # 静态资源
│   ├── src/               # 源代码
│   │   ├── api/           # API 接口
│   │   ├── components/    # 组件
│   │   ├── router/        # 路由配置
│   │   ├── stores/        # 状态管理
│   │   ├── types/         # 类型定义
│   │   ├── utils/         # 工具函数
│   │   ├── views/         # 页面组件
│   │   ├── App.vue        # 根组件
│   │   ├── main.ts        # 程序入口
│   │   └── style.css      # 全局样式
│   ├── package.json       # 依赖配置
│   ├── tsconfig.json      # TypeScript 配置
│   ├── vite.config.ts     # Vite 配置
│   └── index.html         # HTML 模板
├── docs/                   # 文档
├── scripts/               # 部署脚本
└── README.md              # 项目说明
```

## 🚀 快速开始

### 环境要求

- **Node.js** >= 16.0.0
- **Go** >= 1.19
- **MySQL** >= 8.0
- **Redis** >= 6.0 (可选)

### 1. 克隆项目

```bash
git clone <repository-url>
cd 图床转换
```

### 2. 后端配置

#### 安装依赖

```bash
cd backend
go mod tidy
```

#### 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件，配置以下信息：

```env
# 服务器配置
SERVER_PORT=8080
SERVER_MODE=debug

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=image_host

# Cloudflare R2 配置
R2_ACCOUNT_ID=your_account_id
R2_ACCESS_KEY_ID=your_access_key
R2_SECRET_ACCESS_KEY=your_secret_key
R2_BUCKET_NAME=your_bucket_name
R2_ENDPOINT=your_endpoint
R2_PUBLIC_URL=your_public_url

# Redis 配置 (可选)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# 上传配置
MAX_FILE_SIZE=10485760
MAX_BATCH_SIZE=52428800
MAX_BATCH_COUNT=10
ALLOWED_TYPES=image/jpeg,image/png,image/gif,image/webp,image/svg+xml
```

#### 创建数据库

```sql
CREATE DATABASE image_host CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

#### 启动后端服务

```bash
go run main.go
```

后端服务将在 `http://localhost:8080` 启动。

### 3. 前端配置

#### 安装依赖

```bash
cd frontend
npm install
```

#### 启动开发服务器

```bash
npm run dev
```

前端服务将在 `http://localhost:3000` 启动。

### 4. 访问应用

打开浏览器访问 `http://localhost:3000`，即可使用图床转换系统。

## 📖 API 文档

### 上传接口

#### 单文件上传

```http
POST /api/upload
Content-Type: multipart/form-data

file: <图片文件>
```

**响应示例：**

```json
{
  "success": true,
  "message": "上传成功",
  "data": {
    "id": "uuid",
    "filename": "processed_filename.jpg",
    "originalName": "original.jpg",
    "url": "https://your-domain.com/image.jpg",
    "size": 1024000,
    "mimeType": "image/jpeg",
    "uploadTime": "2024-01-01T12:00:00Z",
    "width": 1920,
    "height": 1080
  }
}
```

#### 批量上传

```http
POST /api/batch-upload
Content-Type: multipart/form-data

files: <图片文件数组>
```

#### 获取图片信息

```http
GET /api/images?page=1&pageSize=20&sortBy=uploadTime&sortOrder=desc
```

#### 获取统计信息

```http
GET /api/stats
```

#### 健康检查

```http
GET /api/health
```

## 🔧 配置说明

### Cloudflare R2 配置

1. 登录 Cloudflare 控制台
2. 创建 R2 存储桶
3. 获取 API 令牌和访问密钥
4. 配置自定义域名（可选）

### MySQL 配置

系统会自动创建所需的数据表：

- `images` - 图片元数据表
- `upload_stats` - 上传统计表

### Redis 配置（可选）

Redis 用于：
- 速率限制
- 会话存储
- 缓存热点数据

## 🚀 部署

### Docker 部署

```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d
```

### 手动部署

#### 后端部署

```bash
# 编译
cd backend
go build -o image-host-backend main.go

# 运行
./image-host-backend
```

#### 前端部署

```bash
# 构建
cd frontend
npm run build

# 部署到 Web 服务器
cp -r dist/* /var/www/html/
```

## 🔒 安全特性

- **文件类型验证**: 严格验证上传文件类型
- **文件大小限制**: 防止大文件攻击
- **速率限制**: 防止恶意请求
- **CORS 配置**: 跨域请求控制
- **输入验证**: 所有输入数据验证
- **错误处理**: 安全的错误信息返回

## 📊 性能优化

- **图片压缩**: 自动压缩减少存储空间
- **并发处理**: Go 协程处理并发请求
- **数据库索引**: 优化查询性能
- **CDN 加速**: Cloudflare 全球 CDN
- **缓存策略**: Redis 缓存热点数据
- **连接池**: 数据库连接池管理

## 🛠️ 开发

### 代码规范

- 后端遵循 Go 官方代码规范
- 前端使用 ESLint + Prettier
- 提交信息遵循 Conventional Commits

### 测试

```bash
# 后端测试
cd backend
go test ./...

# 前端测试
cd frontend
npm run test
```

### 构建

```bash
# 后端构建
cd backend
go build -ldflags "-s -w" -o image-host-backend main.go

# 前端构建
cd frontend
npm run build
```

## 📝 更新日志

### v1.0.0 (2024-01-01)

- ✨ 初始版本发布
- 🚀 支持图片上传和存储
- 📊 统计功能
- 🎨 现代化 UI 界面
- 📱 响应式设计

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Go](https://golang.org/) - 高性能编程语言
- [Element Plus](https://element-plus.org/) - Vue 3 组件库
- [Gin](https://gin-gonic.com/) - Go Web 框架
- [Cloudflare R2](https://developers.cloudflare.com/r2/) - 对象存储服务

## 📞 支持

如果您在使用过程中遇到问题，请：

1. 查看文档和 FAQ
2. 搜索已有的 Issue
3. 创建新的 Issue 描述问题

---

**享受使用高性能图床转换系统！** 🎉