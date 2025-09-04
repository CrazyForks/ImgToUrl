# VPS Docker Compose 部署指南

本指南将帮助您在VPS上使用Docker Compose部署图床系统。

## 📋 系统要求

- **操作系统**: Ubuntu 20.04+ / CentOS 7+ / Debian 10+
- **内存**: 至少 2GB RAM
- **存储**: 至少 20GB 可用空间
- **网络**: 公网IP和域名（可选）

## 🛠️ 环境准备

### 1. 安装 Docker 和 Docker Compose

#### Ubuntu/Debian:
```bash
# 更新包管理器
sudo apt update

# 安装必要的包
sudo apt install -y apt-transport-https ca-certificates curl gnupg lsb-release

# 添加Docker官方GPG密钥
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# 添加Docker仓库
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# 安装Docker
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io

# 安装Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 启动Docker服务
sudo systemctl start docker
sudo systemctl enable docker

# 将当前用户添加到docker组
sudo usermod -aG docker $USER
```

#### CentOS/RHEL:
```bash
# 安装Docker
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io

# 安装Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 启动Docker服务
sudo systemctl start docker
sudo systemctl enable docker

# 将当前用户添加到docker组
sudo usermod -aG docker $USER
```

### 2. 验证安装
```bash
# 重新登录或执行以下命令使组权限生效
newgrp docker

# 验证Docker安装
docker --version
docker-compose --version

# 测试Docker运行
docker run hello-world
```

## 🚀 项目部署

### 1. 克隆项目
```bash
# 克隆项目到VPS
git clone https://github.com/roseforljh/ImgToUrl.git
cd ImgToUrl
```

### 2. 配置环境变量
```bash
# 复制环境变量模板
cp .env.example .env

# 编辑环境变量
nano .env
```

**重要配置项说明**:
```bash
# 数据库配置
DB_USER=imagehost
DB_PASSWORD=your_secure_password_here
DB_NAME=image_host

# R2存储配置（可选，如不配置将使用本地存储）
R2_ACCOUNT_ID=your_r2_account_id
R2_ACCESS_KEY_ID=your_access_key
R2_SECRET_ACCESS_KEY=your_secret_key
R2_BUCKET_NAME=your_bucket_name
R2_ENDPOINT=https://your_account_id.r2.cloudflarestorage.com
R2_PUBLIC_URL=https://your_custom_domain.com

# 文件上传限制
MAX_FILE_SIZE=10485760        # 10MB
MAX_BATCH_SIZE=52428800       # 50MB
MAX_BATCH_COUNT=10            # 最多10个文件

# 前端API地址（替换为您的域名或IP）
VITE_API_BASE_URL=http://your_domain_or_ip:8080
```

### 3. 创建必要的目录
```bash
# 创建Nginx配置目录
mkdir -p nginx

# 创建SSL证书目录（如果需要HTTPS）
mkdir -p nginx/ssl

# 创建日志目录
mkdir -p backend/logs
```

### 4. 配置Nginx（可选）
如果您想使用自定义的Nginx配置，创建 `nginx/default.conf`:

```bash
cat > nginx/default.conf << 'EOF'
server {
    listen 80;
    server_name your_domain.com;  # 替换为您的域名
    
    # 前端静态文件
    location / {
        proxy_pass http://frontend:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # 后端API
    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 文件上传相关配置
        client_max_body_size 100M;
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # 静态文件（本地存储的图片）
    location /static/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF
```

## 🏃‍♂️ 启动服务

### 1. 构建和启动所有服务
```bash
# 构建并启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 2. 启动特定服务组合

**基础服务（不包含Nginx代理）**:
```bash
docker-compose up -d mysql redis backend frontend
```

**包含Nginx代理**:
```bash
docker-compose --profile proxy up -d
```

### 3. 验证部署
```bash
# 检查所有容器状态
docker-compose ps

# 检查健康状态
docker-compose exec backend curl -f http://localhost:8080/api/health

# 检查前端
curl -f http://localhost
```

## 🔧 服务管理

### 常用命令
```bash
# 查看服务状态
docker-compose ps

# 查看实时日志
docker-compose logs -f [service_name]

# 重启服务
docker-compose restart [service_name]

# 停止所有服务
docker-compose down

# 停止并删除数据卷（谨慎使用）
docker-compose down -v

# 重新构建服务
docker-compose build [service_name]
docker-compose up -d [service_name]

# 进入容器
docker-compose exec [service_name] /bin/bash
```

### 更新应用
```bash
# 拉取最新代码
git pull origin main

# 重新构建并启动
docker-compose build
docker-compose up -d
```

## 🔒 安全配置

### 1. 防火墙设置
```bash
# Ubuntu/Debian (ufw)
sudo ufw allow 22/tcp      # SSH
sudo ufw allow 80/tcp      # HTTP
sudo ufw allow 443/tcp     # HTTPS
sudo ufw enable

# CentOS/RHEL (firewalld)
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

### 2. SSL证书配置（推荐）

使用Let's Encrypt免费SSL证书:
```bash
# 安装certbot
sudo apt install certbot  # Ubuntu/Debian
# 或
sudo yum install certbot   # CentOS/RHEL

# 获取SSL证书
sudo certbot certonly --standalone -d your_domain.com

# 复制证书到项目目录
sudo cp /etc/letsencrypt/live/your_domain.com/fullchain.pem nginx/ssl/
sudo cp /etc/letsencrypt/live/your_domain.com/privkey.pem nginx/ssl/
sudo chown $USER:$USER nginx/ssl/*
```

## 📊 监控和维护

### 1. 项目更新

当项目有新版本时，按以下步骤更新：

#### 1. 备份数据（重要）

```bash
# 备份数据库
docker-compose exec mysql mysqldump -u root -p image_host > backup_$(date +%Y%m%d_%H%M%S).sql

# 备份上传的文件
sudo cp -r ./public ./backup_public_$(date +%Y%m%d_%H%M%S)
```

#### 拉取最新代码

```bash
# 停止服务
docker-compose down

# 拉取最新代码
git pull origin main

# 或者如果没有使用git，重新下载项目
# wget https://github.com/roseforljh/ImgToUrl/archive/refs/heads/main.zip
# unzip main.zip
# cp -r ImgToUrl-main/* ./
```

#### 3. 更新配置文件

```bash
# 检查是否有新的环境变量
diff .env.example .env

# 根据需要更新 .env 文件
nano .env
```

#### 4. 重新构建和部署

```bash
# 清理旧的镜像和容器
docker system prune -f

# 重新构建并启动
docker-compose up -d --build

# 查看启动状态
docker-compose ps
docker-compose logs -f
```

#### 验证更新

```bash
# 检查服务状态
curl http://localhost/api/health

# 访问前端页面确认功能正常
# http://您的服务器IP
```

#### 快速更新脚本

创建更新脚本 `update.sh`：

```bash
#!/bin/bash
echo "开始更新图床系统..."

# 备份数据库
echo "备份数据库..."
docker-compose exec -T mysql mysqldump -u root -p image_host > backup_$(date +%Y%m%d_%H%M%S).sql

# 停止服务
echo "停止服务..."
docker-compose down

# 拉取最新代码
echo "拉取最新代码..."
git pull origin main

# 重新构建和启动
echo "重新构建和启动服务..."
docker-compose up -d --build

# 等待服务启动
echo "等待服务启动..."
sleep 30

# 检查服务状态
echo "检查服务状态..."
docker-compose ps

echo "更新完成！"
echo "请访问 http://$(curl -s ifconfig.me) 验证服务是否正常"
```

使用更新脚本：

```bash
# 给脚本执行权限
chmod +x update.sh

# 运行更新
./update.sh
```

### 2. 日志管理
```bash
# 查看应用日志
docker-compose logs backend
docker-compose logs frontend

# 查看数据库日志
docker-compose logs mysql

# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f backend
docker-compose logs -f frontend

# 查看最近100行日志
docker-compose logs --tail=100 backend

# 清理日志（定期执行）
docker system prune -f
```

### 3. 数据备份
```bash
# 备份数据库
docker-compose exec mysql mysqldump -u root -p image_host > backup_$(date +%Y%m%d).sql

# 备份上传的文件
tar -czf uploads_backup_$(date +%Y%m%d).tar.gz backend/uploads/
```

### 4. 性能优化
```bash
# 查看资源使用情况
docker stats

# 清理未使用的镜像和容器
docker system prune -a
```

## 🌐 域名和DNS配置

1. **购买域名**并配置DNS记录
2. **A记录**: 将域名指向VPS的公网IP
3. **CNAME记录**（可选): 配置www子域名

## 🚨 故障排除

### 常见问题

1. **前端构建失败：vue-tsc not found**
   ```
   错误信息：sh: vue-tsc: not found
   ```
   
   **解决方案：**
   - 这是因为Dockerfile中使用了`--only=production`参数，跳过了开发依赖
   - 前端构建需要vue-tsc等开发依赖
   - 已在最新版本中修复，如遇到此问题请更新代码
   
   **手动修复：**
   ```bash
   # 编辑 frontend/Dockerfile
   # 将 "RUN npm ci --only=production --silent" 
   # 改为 "RUN npm ci --silent"
   ```

2. **容器启动失败**
   ```bash
   docker-compose logs [service_name]
   ```

3. **数据库连接失败**
   - 检查环境变量配置
   - 确认MySQL容器健康状态

4. **文件上传失败**
   - 检查文件大小限制
   - 确认存储目录权限

5. **R2存储连接失败**
   - 验证R2配置信息
   - 系统会自动回退到本地存储

### 获取帮助
- 查看项目README: `cat README.md`
- 查看R2配置指南: `cat R2_CONFIG_GUIDE.md`
- GitHub Issues: https://github.com/roseforljh/ImgToUrl/issues

---

## 🚀 部署验证

### 1. 服务状态检查

```bash
# 检查所有服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f

# 检查端口监听
sudo netstat -tlnp | grep -E ':(80|443|8080|3306|6379)'
```

### 2. 访问应用

部署完成后，您可以通过以下方式访问应用：

#### 🌐 Web访问

**通过Nginx反向代理访问（推荐）：**
```
http://您的服务器IP地址
# 例如：http://192.168.1.100
# 或者：http://your-server-ip
```

**直接访问前端服务（无Nginx）：**
```
http://您的服务器IP地址:3000
# 例如：http://192.168.1.100:3000
# 注意：需要确保防火墙开放3000端口
```

**HTTPS访问（配置SSL后）：**
```
https://您的域名
# 例如：https://img.yourdomain.com
```

#### 🔌 API访问

**通过Nginx代理访问API：**
```
http://您的服务器IP地址/api/
# 例如：http://192.168.1.100/api/health
```

**直接访问后端服务：**
```
http://您的服务器IP地址:8080/api/
# 例如：http://192.168.1.100:8080/api/health
# 注意：需要确保防火墙开放8080端口
```

**健康检查：**
```bash
# 通过Nginx代理检查
curl http://您的服务器IP/api/health

# 直接检查后端服务
curl http://您的服务器IP:8080/api/health

# 检查前端服务
curl http://您的服务器IP
```

#### 📱 移动端访问

**通过Nginx代理（推荐）：**
- 前端界面：`http://您的服务器IP地址`
- API接口：`http://您的服务器IP地址/api/`

**直接访问服务：**
- 前端界面：`http://您的服务器IP地址:3000`
- API接口：`http://您的服务器IP地址:8080/api/`

在手机浏览器中输入上述任一地址即可访问。

> **注意事项：**
> - 直接访问需要确保防火墙开放对应端口（3000、8080）
> - 推荐使用Nginx代理访问，更安全且便于管理
> - 首次访问可能需要等待1-2分钟服务完全启动

```
http://您的服务器IP地址
```

> **注意事项：**
> - 替换 `您的服务器IP地址` 为实际的VPS公网IP
> - 如果配置了域名，使用域名访问更方便
> - 确保防火墙已开放80和443端口
> - 首次访问可能需要等待1-2分钟服务完全启动

## 🎉 部署完成

享受您的图床服务吧！ 🚀