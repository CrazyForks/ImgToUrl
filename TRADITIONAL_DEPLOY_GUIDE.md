# 传统部署指南

本指南适用于不使用Docker的传统部署方式，直接在VPS上安装和配置各个组件。

## 📋 系统要求

### 最低配置
- **CPU**: 1核心
- **内存**: 1GB RAM
- **存储**: 20GB 可用空间
- **操作系统**: Ubuntu 20.04+ / CentOS 8+ / Debian 11+
- **网络**: 公网IP地址

### 推荐配置
- **CPU**: 2核心
- **内存**: 2GB RAM
- **存储**: 50GB+ SSD
- **带宽**: 10Mbps+

## 🛠️ 环境准备

### 1. 更新系统

#### Ubuntu/Debian
```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl wget git unzip software-properties-common
```

#### CentOS/RHEL
```bash
sudo yum update -y
sudo yum install -y curl wget git unzip epel-release
sudo yum groupinstall -y "Development Tools"
```

### 2. 安装Node.js

#### 使用NodeSource仓库（推荐）
```bash
# Ubuntu/Debian
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# CentOS/RHEL
curl -fsSL https://rpm.nodesource.com/setup_18.x | sudo bash -
sudo yum install -y nodejs

# 验证安装
node --version
npm --version
```

#### 使用NVM（备选）
```bash
# 安装NVM
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
source ~/.bashrc

# 安装Node.js
nvm install 18
nvm use 18
nvm alias default 18
```

### 3. 安装Go

```bash
# 下载Go
wget https://golang.org/dl/go1.21.0.linux-amd64.tar.gz

# 解压安装
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# 配置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

### 4. 安装MySQL

#### Ubuntu/Debian
```bash
# 安装MySQL
sudo apt install -y mysql-server

# 启动服务
sudo systemctl start mysql
sudo systemctl enable mysql

# 安全配置
sudo mysql_secure_installation
```

#### CentOS/RHEL
```bash
# 安装MySQL
sudo yum install -y mysql-server

# 启动服务
sudo systemctl start mysqld
sudo systemctl enable mysqld

# 获取临时密码
sudo grep 'temporary password' /var/log/mysqld.log

# 安全配置
sudo mysql_secure_installation
```

#### 配置数据库
```bash
# 登录MySQL
mysql -u root -p

# 创建数据库和用户
CREATE DATABASE image_host CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'imagehost'@'localhost' IDENTIFIED BY 'your_strong_password';
GRANT ALL PRIVILEGES ON image_host.* TO 'imagehost'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

### 5. 安装Redis

#### Ubuntu/Debian
```bash
# 安装Redis
sudo apt install -y redis-server

# 启动服务
sudo systemctl start redis-server
sudo systemctl enable redis-server

# 测试连接
redis-cli ping
```

#### CentOS/RHEL
```bash
# 安装Redis
sudo yum install -y redis

# 启动服务
sudo systemctl start redis
sudo systemctl enable redis

# 测试连接
redis-cli ping
```

#### 配置Redis
```bash
# 编辑配置文件
sudo nano /etc/redis/redis.conf

# 修改以下配置
bind 127.0.0.1
port 6379
requirepass your_redis_password
maxmemory 256mb
maxmemory-policy allkeys-lru

# 重启Redis
sudo systemctl restart redis
```

### 6. 安装Nginx

#### Ubuntu/Debian
```bash
# 安装Nginx
sudo apt install -y nginx

# 启动服务
sudo systemctl start nginx
sudo systemctl enable nginx
```

#### CentOS/RHEL
```bash
# 安装Nginx
sudo yum install -y nginx

# 启动服务
sudo systemctl start nginx
sudo systemctl enable nginx
```

## 📦 部署应用

### 1. 克隆项目

```bash
# 创建应用目录
sudo mkdir -p /opt/imgtourl
sudo chown $USER:$USER /opt/imgtourl
cd /opt/imgtourl

# 克隆项目
git clone https://github.com/roseforljh/ImgToUrl.git .
```

### 2. 配置环境变量

```bash
# 创建环境配置文件
cp .env.example .env
nano .env
```

配置内容：
```bash
# 服务器配置
PORT=8080
GIN_MODE=release

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=imagehost
DB_PASSWORD=your_strong_password
DB_NAME=image_host

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0

# 文件存储配置
UPLOAD_PATH=./uploads
MAX_FILE_SIZE=10485760
MAX_BATCH_SIZE=52428800
MAX_BATCH_COUNT=10

# R2存储配置（可选）
# R2_ACCOUNT_ID=your_account_id
# R2_ACCESS_KEY_ID=your_access_key
# R2_SECRET_ACCESS_KEY=your_secret_key
# R2_BUCKET_NAME=your_bucket_name
# R2_ENDPOINT=your_endpoint
# R2_PUBLIC_URL=your_public_url

# 前端配置
VITE_API_BASE_URL=http://your-domain.com:8080
```

### 3. 构建后端

```bash
# 进入后端目录
cd backend

# 下载依赖
go mod download

# 构建应用
go build -o imgtourl-backend .

# 创建上传目录
mkdir -p uploads

# 设置权限
chmod 755 imgtourl-backend
chmod 755 uploads
```

### 4. 构建前端

```bash
# 进入前端目录
cd ../frontend

# 安装依赖
npm install

# 构建生产版本
npm run build

# 复制构建文件到Nginx目录
sudo cp -r dist/* /var/www/html/
```

### 5. 配置系统服务

#### 创建后端服务
```bash
# 创建服务文件
sudo nano /etc/systemd/system/imgtourl-backend.service
```

服务配置：
```ini
[Unit]
Description=ImgToUrl Backend Service
After=network.target mysql.service redis.service
Wants=mysql.service redis.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/imgtourl/backend
ExecStart=/opt/imgtourl/backend/imgtourl-backend
Restart=always
RestartSec=5
Environment=GIN_MODE=release
EnvironmentFile=/opt/imgtourl/.env

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/imgtourl/backend/uploads

# 日志设置
StandardOutput=journal
StandardError=journal
SyslogIdentifier=imgtourl-backend

[Install]
WantedBy=multi-user.target
```

#### 启动后端服务
```bash
# 重载systemd配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start imgtourl-backend
sudo systemctl enable imgtourl-backend

# 检查状态
sudo systemctl status imgtourl-backend

# 查看日志
sudo journalctl -u imgtourl-backend -f
```

### 6. 配置Nginx

```bash
# 创建站点配置
sudo nano /etc/nginx/sites-available/imgtourl
```

Nginx配置：
```nginx
server {
    listen 80;
    server_name your-domain.com www.your-domain.com;
    
    # 安全设置
    server_tokens off;
    client_max_body_size 100M;
    
    # 前端静态文件
    location / {
        root /var/www/html;
        index index.html;
        try_files $uri $uri/ /index.html;
        
        # 缓存设置
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
    }
    
    # 后端API代理
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # 文件上传设置
        client_max_body_size 100M;
        proxy_request_buffering off;
    }
    
    # 静态文件服务（上传的图片）
    location /uploads/ {
        alias /opt/imgtourl/backend/uploads/;
        expires 1y;
        add_header Cache-Control "public";
        
        # 安全设置
        location ~ \.(php|jsp|asp|sh|py|pl|exe)$ {
            deny all;
        }
    }
    
    # 健康检查
    location /health {
        proxy_pass http://127.0.0.1:8080/health;
        access_log off;
    }
    
    # 安全设置
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }
    
    location ~ \.(env|log|conf)$ {
        deny all;
        access_log off;
        log_not_found off;
    }
}
```

#### 启用站点
```bash
# 创建软链接
sudo ln -s /etc/nginx/sites-available/imgtourl /etc/nginx/sites-enabled/

# 删除默认站点
sudo rm -f /etc/nginx/sites-enabled/default

# 测试配置
sudo nginx -t

# 重启Nginx
sudo systemctl restart nginx
```

## 🔒 SSL证书配置

### 使用Let's Encrypt（推荐）

```bash
# 安装Certbot
# Ubuntu/Debian
sudo apt install -y certbot python3-certbot-nginx

# CentOS/RHEL
sudo yum install -y certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d your-domain.com -d www.your-domain.com

# 设置自动续期
sudo crontab -e
# 添加以下行
0 12 * * * /usr/bin/certbot renew --quiet
```

### 手动配置HTTPS

如果需要手动配置HTTPS，编辑Nginx配置：

```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com www.your-domain.com;
    
    # SSL证书配置
    ssl_certificate /path/to/your/certificate.crt;
    ssl_certificate_key /path/to/your/private.key;
    
    # SSL安全配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    
    # HSTS
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    
    # 其他配置同HTTP版本...
}

# HTTP重定向到HTTPS
server {
    listen 80;
    server_name your-domain.com www.your-domain.com;
    return 301 https://$server_name$request_uri;
}
```

## 🔥 防火墙配置

### Ubuntu/Debian (UFW)
```bash
# 启用UFW
sudo ufw enable

# 允许SSH
sudo ufw allow 22

# 允许HTTP和HTTPS
sudo ufw allow 80
sudo ufw allow 443

# 查看状态
sudo ufw status
```

### CentOS/RHEL (Firewalld)
```bash
# 启动firewalld
sudo systemctl start firewalld
sudo systemctl enable firewalld

# 添加服务
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https

# 重载配置
sudo firewall-cmd --reload
```

## 📊 监控和日志

### 1. 日志配置

```bash
# 创建日志目录
sudo mkdir -p /var/log/imgtourl
sudo chown www-data:www-data /var/log/imgtourl

# 配置日志轮转
sudo nano /etc/logrotate.d/imgtourl
```

日志轮转配置：
```bash
/var/log/imgtourl/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 www-data www-data
    postrotate
        systemctl reload imgtourl-backend
    endscript
}

/var/log/nginx/*.log {
    daily
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    create 644 nginx nginx
    postrotate
        systemctl reload nginx
    endscript
}
```

### 2. 监控脚本

```bash
# 创建监控脚本
sudo nano /usr/local/bin/imgtourl-monitor.sh
```

监控脚本内容：
```bash
#!/bin/bash

LOG_FILE="/var/log/imgtourl/monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

# 检查后端服务
if ! systemctl is-active --quiet imgtourl-backend; then
    echo "[$DATE] 错误: 后端服务未运行" >> $LOG_FILE
    systemctl restart imgtourl-backend
fi

# 检查Nginx
if ! systemctl is-active --quiet nginx; then
    echo "[$DATE] 错误: Nginx服务未运行" >> $LOG_FILE
    systemctl restart nginx
fi

# 检查MySQL
if ! systemctl is-active --quiet mysql; then
    echo "[$DATE] 错误: MySQL服务未运行" >> $LOG_FILE
    systemctl restart mysql
fi

# 检查Redis
if ! systemctl is-active --quiet redis; then
    echo "[$DATE] 错误: Redis服务未运行" >> $LOG_FILE
    systemctl restart redis
fi

# 检查磁盘空间
USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $USAGE -gt 80 ]; then
    echo "[$DATE] 警告: 磁盘使用率 ${USAGE}%" >> $LOG_FILE
fi

# 检查内存使用
MEM_USAGE=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
if [ $MEM_USAGE -gt 90 ]; then
    echo "[$DATE] 警告: 内存使用率 ${MEM_USAGE}%" >> $LOG_FILE
fi
```

```bash
# 设置执行权限
sudo chmod +x /usr/local/bin/imgtourl-monitor.sh

# 添加到crontab
echo "*/5 * * * * /usr/local/bin/imgtourl-monitor.sh" | sudo crontab -
```

## 🔄 备份和恢复

### 自动备份脚本

```bash
# 创建备份脚本
sudo nano /usr/local/bin/imgtourl-backup.sh
```

备份脚本：
```bash
#!/bin/bash

BACKUP_DIR="/backup/imgtourl"
DATE=$(date +%Y%m%d_%H%M%S)
DB_PASSWORD="your_db_password"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份数据库
mysqldump -u imagehost -p$DB_PASSWORD image_host > $BACKUP_DIR/database_$DATE.sql

# 备份上传文件
tar -czf $BACKUP_DIR/uploads_$DATE.tar.gz -C /opt/imgtourl/backend uploads/

# 备份配置文件
tar -czf $BACKUP_DIR/config_$DATE.tar.gz -C /opt/imgtourl .env
tar -czf $BACKUP_DIR/nginx_$DATE.tar.gz -C /etc/nginx/sites-available imgtourl

# 删除7天前的备份
find $BACKUP_DIR -name "*.sql" -mtime +7 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete

echo "$(date): 备份完成" >> /var/log/imgtourl/backup.log
```

```bash
# 设置权限和定时任务
sudo chmod +x /usr/local/bin/imgtourl-backup.sh
echo "0 2 * * * /usr/local/bin/imgtourl-backup.sh" | sudo crontab -
```

## 🚀 部署验证

### 1. 服务状态检查

```bash
# 检查所有服务状态
sudo systemctl status imgtourl-backend
sudo systemctl status nginx
sudo systemctl status mysql
sudo systemctl status redis

# 检查端口监听
sudo netstat -tlnp | grep -E ':(80|443|8080|3306|6379)'
```

### 2. 功能测试

```bash
# 测试后端API
curl -X GET http://localhost:8080/health

# 测试前端访问
curl -I http://your-domain.com

# 测试HTTPS（如果配置了）
curl -I https://your-domain.com
```

### 3. 访问应用

部署完成后，您可以通过以下方式访问应用：

#### 🌐 Web访问

**HTTP访问（默认）：**
```
http://您的服务器IP地址
# 例如：http://192.168.1.100
# 或者：http://your-server-ip
```

**域名访问（配置域名后）：**
```
http://您的域名
# 例如：http://img.yourdomain.com
```

**HTTPS访问（配置SSL后）：**
```
https://您的域名
# 例如：https://img.yourdomain.com
```

#### 🔌 API访问

**后端API接口：**
```
http://您的服务器IP地址:8080/api/
# 例如：http://192.168.1.100:8080/api/health
```

**通过Nginx代理访问API：**
```
http://您的服务器IP地址/api/
# 例如：http://192.168.1.100/api/health
```

**健康检查：**
```bash
# 直接访问后端
curl http://您的服务器IP:8080/api/health

# 通过Nginx代理访问
curl http://您的服务器IP/api/health

# 检查前端
curl http://您的服务器IP
```

#### 📱 移动端访问

在手机浏览器中输入：
```
http://您的服务器IP地址
```

> **注意事项：**
> - 替换 `您的服务器IP地址` 为实际的VPS公网IP
> - 如果配置了域名，使用域名访问更方便
> - 确保防火墙已开放80、443和8080端口
> - 后端服务运行在8080端口，前端通过Nginx在80端口提供服务
> - 首次访问可能需要等待服务完全启动

### 3. 日志检查

```bash
# 查看后端日志
sudo journalctl -u imgtourl-backend -f

# 查看Nginx日志
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log

# 查看MySQL日志
sudo tail -f /var/log/mysql/error.log
```

## 🔧 故障排除

### 常见问题

1. **后端服务启动失败**
   ```bash
   # 检查日志
   sudo journalctl -u imgtourl-backend -n 50
   
   # 检查配置文件
   cat /opt/imgtourl/.env
   
   # 手动启动测试
   cd /opt/imgtourl/backend
   ./imgtourl-backend
   ```

2. **数据库连接失败**
   ```bash
   # 测试数据库连接
   mysql -u imagehost -p -h localhost image_host
   
   # 检查MySQL状态
   sudo systemctl status mysql
   ```

3. **文件上传失败**
   ```bash
   # 检查上传目录权限
   ls -la /opt/imgtourl/backend/uploads/
   
   # 修复权限
   sudo chown -R www-data:www-data /opt/imgtourl/backend/uploads/
   sudo chmod -R 755 /opt/imgtourl/backend/uploads/
   ```

4. **Nginx配置错误**
   ```bash
   # 测试Nginx配置
   sudo nginx -t
   
   # 重新加载配置
   sudo systemctl reload nginx
   ```

## 📝 维护清单

### 日常维护
- [ ] 检查服务状态
- [ ] 查看错误日志
- [ ] 监控磁盘空间
- [ ] 检查备份状态

### 定期维护
- [ ] 更新系统包
- [ ] 更新应用代码
- [ ] 清理日志文件
- [ ] 检查SSL证书有效期
- [ ] 审查访问日志

### 安全维护
- [ ] 更新密码
- [ ] 检查防火墙规则
- [ ] 审查用户权限
- [ ] 更新安全补丁

---

传统部署方式提供了更多的控制权和灵活性，但需要更多的维护工作。建议在生产环境中使用Docker部署方式以获得更好的可维护性。 🚀