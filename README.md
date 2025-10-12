# ImgToUrl（图床转换）

基于 Go + Vue3 + MySQL 的图片上传与链接生成系统，支持多存储渠道（本地、Cloudflare R2、阿里云 OSS），并提供“渠道同步”能力，可将对象存储中的已有图片一键同步到站点数据库中。

- 前端（Vue3 + Element Plus + Pinia）经 Nginx 反向代理到后端 API
- 后端（Gin + GORM）提供鉴权、上传、列表、删除、统计、系统状态、渠道同步等接口
- 数据库存储使用 MySQL，图片文件默认落地到本地磁盘（/uploads），可按需切换 R2/OSS
- 渠道同步：从 R2/OSS 读取对象并写入站点图片表，支持前缀过滤、数量限制、覆盖更新

提示：项目内“R2”名称为历史命名；当前实现已同时支持本地存储 / Cloudflare R2 / 阿里云 OSS 三种渠道。

## 功能概览

- 账号与游客码
  - 管理员登录、游客码登录；创建/查看/删除游客码；自动清理过期游客码
- 图片管理
  - 单图与批量上传、分页列表、删除
  - 直链生成（支持本地 / R2 / OSS，不同渠道生成策略不同）
- 统计与系统状态
  - 总数、容量、今日统计、平均/最大/最小
  - 运行时长、磁盘/内存/CPU 占用等
- 安全与体验
  - 速率限制、类型/大小校验、统一错误码
- 存储渠道
  - 设置页切换 local / r2 / oss
  - 同步渠道对象到站点（R2/OSS）——可在“存储内容”页一键同步

## 系统架构

- 前端：Vue 3、TypeScript、Vite、Element Plus、Pinia、Axios（Nginx 提供静态与反代）
- 后端：Go、Gin、GORM、MySQL
- 存储渠道：本地磁盘（默认）、Cloudflare R2（S3 兼容）、阿里云 OSS
- 端口（生产容器）：
  - 宿主机 8080 → web(Nginx) 80
  - web → backend:8080（反向代理 /api 与 /uploads）
- 端口（本地开发）：
  - 前端 dev：3000（Vite），代理 /api、/uploads → 8080
  - 后端 dev：8080

## 一键部署（Docker Compose）

确保已安装 Docker 与 Docker Compose，然后执行：

```bash
# 克隆项目
git clone https://github.com/roseforljh/ImgToUrl.git
cd ImgToUrl

# 启动所有服务（MySQL / 后端 / Nginx 前端）
docker compose up -d
```

启动后：

- 访问地址：http://localhost:8080
- 默认管理员账号：root / 123456
- 健康检查（被 web 反代至后端）：GET http://localhost:8080/health

相关编排与配置：

- Nginx 站点与反代（/api、/uploads）：见 nginx/conf.d/default.conf
- 后端容器环境变量由 docker compose 注入（DB_HOST=mysql 等）
- 静态文件与直链映射：/uploads → 后端 UPLOAD_PATH（容器内 /data/uploads）

## 更新与升级

拉取最新代码并重建容器：

```bash
cd ImgToUrl
git pull
docker compose up -d --build
```

如遇构建缓存导致未更新，可先停止并重建：

```bash
docker compose down
docker compose up -d --build
```

注意：
- 若宿主 8080 端口被占用，需释放后再启动
- 生产构建已禁用前端 Mock（在 nginx/Dockerfile 中设置 VITE_ENABLE_MOCK=false 的构建期变量）

## 存储渠道配置

在“设置”页（仅管理员可见）填写 R2/OSS 的访问参数并保存，后端将动态重载配置。

- R2 参数：account_id、access_key_id、secret_access_key、bucket_name、endpoint、region、public_url（可选）
- OSS 参数：access_key_id、secret_access_key、bucket_name、endpoint、region（可选）、public_url（可选）
- provider：local / r2 / oss

保存后，运行时会重新加载存储客户端。配置重载逻辑参见后端服务层的 ReloadConfig。

## 渠道同步（R2 / OSS → 站点）

场景：对象存储中已存在历史图片；希望统一纳入站点数据库，便于前端列表、删除与链接展示。

- 操作入口：“存储内容”页顶部按钮“同步 R2 / 同步 OSS”
- 默认参数：prefix=''（全部）、limit=-1（全量）、overwrite=true（更新已有记录）
- 同步结果会提示统计：新增 / 更新 / 跳过，并自动刷新列表

同步行为说明：
- 生成的公共 URL 策略：
  - 若配置了 public_url 前缀，则直接使用该前缀拼接对象键
  - 若未配置前缀且桶为私有，后端会生成带有效期的临时签名 URL（R2/OSS）
  - 若无法签名，回退直链（要求公开可读）
- 去重规则：以 r2_key + storage_provider 唯一识别；存在且 overwrite=false 时跳过；overwrite=true 时更新记录

## 环境变量（后端）

后端会从容器环境变量或 .env 文件读取配置，关键项如下：

- 服务
  - PORT / SERVER_PORT：后端服务端口（默认 8080）
- 鉴权
  - JWT_SECRET：签名密钥（默认 change_me_secret）
  - JWT_EXPIRE_HOURS：过期时长（默认 72）
  - DEFAULT_ADMIN / DEFAULT_PASSWORD：默认管理员账号与密码（默认 root / 123456）
- 数据库
  - DB_HOST / DB_PORT / DB_USER / DB_PASSWORD / DB_NAME
- 上传
  - MAX_FILE_SIZE：单文件大小限制（默认 10MB=10485760）
  - ALLOWED_TYPES：允许的 MIME 类型（默认 image/jpeg,image/png,image/gif,image/webp）
  - UPLOAD_PATH：本地存储路径（生产容器映射为 /data/uploads）
- R2（可选）
  - R2_ACCOUNT_ID / R2_ACCESS_KEY_ID / R2_SECRET_ACCESS_KEY / R2_BUCKET_NAME / R2_ENDPOINT / R2_REGION / R2_PUBLIC_URL
- OSS（可选）
  - OSS_ACCESS_KEY_ID / OSS_SECRET_ACCESS_KEY / OSS_BUCKET_NAME / OSS_ENDPOINT / OSS_REGION / OSS_PUBLIC_URL

提示：Compose 已注入必要的生产变量；若需自定义，请修改 docker compose.yml 对应 environment。

## 前端开发与构建

本地开发：

```bash
cd frontend
npm install
npm run dev
# 访问 http://localhost:3000
# /api、/uploads 被代理到 http://localhost:8080
```

生产构建与运行：
- Nginx 多阶段镜像将在构建阶段打包前端产物并复制到 /usr/share/nginx/html
- 生产构建默认禁用 Mock（VITE_ENABLE_MOCK=false）

## 后端开发与运行

本地开发：

```bash
cd backend
go mod tidy

# 创建数据库（首次）
# CREATE DATABASE image_host CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 配置 .env（示例）
# SERVER_PORT=8080
# JWT_SECRET=change_me_secret
# DEFAULT_ADMIN=root
# DEFAULT_PASSWORD=123456
# DB_HOST=localhost
# DB_PORT=3306
# DB_USER=root
# DB_PASSWORD=your_password
# DB_NAME=image_host
# MAX_FILE_SIZE=10485760
# ALLOWED_TYPES=image/jpeg,image/png
# UPLOAD_PATH=./uploads

go run main.go
# 健康检查：GET http://localhost:8080/health
```

启动后：
- 自动迁移表结构（images、image_stats、users、guest_codes、storage_config）
- 自动创建默认管理员（DEFAULT_ADMIN / DEFAULT_PASSWORD）
- 静态映射：/uploads → UPLOAD_PATH

## 常用命令（Docker）

```bash
# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f

# 停止服务
docker compose down

# 重建并启动
docker compose up -d --build
```

## 常见问题

- 8080 端口被占用
  - 修改宿主端口或释放占用进程；或调整 docker compose.yml 的 ports 映射
- 图片直链无法访问（R2/OSS）
  - 若桶为私有且未配置 public_url，后端会生成签名 URL；请确认签名有效期与代理未屏蔽签名参数
  - 若仍无法访问，检查 endpoint/bucket 是否正确、对象键是否存在
- 本地上传 404
  - 确认 UPLOAD_PATH 可写入、后端容器的 /data/uploads 已正确挂载
- 只看到自己的图片
  - 非管理员仅能查看/删除自己上传的图片；管理员可查看全部
- 前端 Mock
  - 生产构建默认禁用；如需在开发启用，请在前端环境中设置 VITE_ENABLE_MOCK=true

## 许可证

MIT

## 界面截图

### 存储渠道配置界面
以下三张图展示了“存储配置渠道”的设置界面，用于切换和配置 Local（本地）、Cloudflare R2、阿里云 OSS 等参数。管理员在“设置”页中填写并保存后，后端会动态重载配置，运行时立即生效。

- 存储配置界面示例 1  
  ![存储配置示例 1](https://qone.kuz7.com/uploads/images/2025/10/12/5d1163fa-8f06-4513-8371-a1ed248af5d6.png)

- 存储配置界面示例 2  
  ![存储配置示例 2](https://qone.kuz7.com/uploads/images/2025/10/12/52095a04-a194-4a12-8431-71d283da2708.png)

- 存储配置界面示例 3  
  ![存储配置示例 3](https://qone.kuz7.com/uploads/images/2025/10/12/1d3d5ec8-b8af-4e74-8134-e9fe237d8c47.png)

说明：
- 填写 R2/OSS 的访问参数（access_key、secret、bucket、endpoint、region 等）
- 可选填写 Public URL 前缀（建议使用 CDN 域名），用于生成更稳定的公共访问链接
- 选择 provider：local / r2 / oss；保存后立即切换

### 渠道同步功能界面
该界面用于“一键同步”对象存储（R2/OSS）中的历史图片到站点数据库，便于前端列表、复制链接、删除等统一管理。

- 渠道同步界面  
  ![渠道同步界面](https://qone.kuz7.com/uploads/images/2025/10/12/640591b2-c722-42d2-9e3d-3b08dd90498d.png)

说明：
- 入口在“存储内容”页顶部：“同步 R2”“同步 OSS”
- 默认参数：prefix=''（全部）、limit=-1（全量）、overwrite=true（更新已有记录）
- 同步完成后显示统计：新增/更新/跳过，并自动刷新列表
- 公共 URL 生成策略：优先 Public URL 前缀；私有桶场景自动生成临时签名 URL；最后回退直链（需公开可读）
