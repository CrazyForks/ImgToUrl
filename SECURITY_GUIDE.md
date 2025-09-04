# 安全配置指南

本指南将帮助您配置图床系统的安全设置，包括防火墙、访问控制和安全最佳实践。

## 🔥 防火墙配置

### Ubuntu/Debian (UFW)

#### 1. 安装和启用UFW
```bash
# 安装UFW（通常已预装）
sudo apt update
sudo apt install ufw

# 设置默认策略
sudo ufw default deny incoming
sudo ufw default allow outgoing

# 允许SSH（重要：先配置SSH再启用防火墙）
sudo ufw allow 22/tcp

# 允许HTTP和HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# 启用防火墙
sudo ufw enable

# 查看状态
sudo ufw status verbose
```

#### 2. 高级配置
```bash
# 限制SSH连接频率（防止暴力破解）
sudo ufw limit 22/tcp

# 允许特定IP访问特定端口
sudo ufw allow from 192.168.1.100 to any port 3306

# 允许特定网段
sudo ufw allow from 192.168.1.0/24

# 删除规则
sudo ufw delete allow 80/tcp

# 重置所有规则
sudo ufw --force reset
```

### CentOS/RHEL (Firewalld)

#### 1. 安装和配置Firewalld
```bash
# 安装firewalld
sudo yum install firewalld

# 启动并启用服务
sudo systemctl start firewalld
sudo systemctl enable firewalld

# 查看状态
sudo firewall-cmd --state

# 查看当前配置
sudo firewall-cmd --list-all
```

#### 2. 配置规则
```bash
# 添加服务
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https

# 添加端口
sudo firewall-cmd --permanent --add-port=8080/tcp

# 重载配置
sudo firewall-cmd --reload

# 查看开放的端口
sudo firewall-cmd --list-ports

# 删除规则
sudo firewall-cmd --permanent --remove-service=http
sudo firewall-cmd --reload
```

## 🔒 SSH安全配置

### 1. 修改SSH配置
```bash
# 编辑SSH配置文件
sudo nano /etc/ssh/sshd_config
```

推荐的安全配置：
```bash
# 更改默认端口（可选）
Port 2222

# 禁用root登录
PermitRootLogin no

# 禁用密码登录（推荐使用密钥）
PasswordAuthentication no
PubkeyAuthentication yes

# 限制登录用户
AllowUsers your_username

# 设置登录超时
ClientAliveInterval 300
ClientAliveCountMax 2

# 禁用空密码
PermitEmptyPasswords no

# 限制登录尝试
MaxAuthTries 3
MaxStartups 3
```

### 2. 配置SSH密钥认证
```bash
# 在本地生成密钥对
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"

# 复制公钥到服务器
ssh-copy-id -i ~/.ssh/id_rsa.pub username@server_ip

# 或手动复制
cat ~/.ssh/id_rsa.pub | ssh username@server_ip "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys"
```

### 3. 重启SSH服务
```bash
# 重启SSH服务
sudo systemctl restart sshd

# 验证配置
sudo sshd -t
```

## 🛡️ 应用安全配置

### 1. 数据库安全

#### MySQL安全配置
```bash
# 进入MySQL容器
docker-compose exec mysql mysql -u root -p

# 创建专用用户（不使用root）
CREATE USER 'imagehost'@'%' IDENTIFIED BY 'strong_password_here';
GRANT SELECT, INSERT, UPDATE, DELETE ON image_host.* TO 'imagehost'@'%';
FLUSH PRIVILEGES;

# 删除默认用户和数据库
DROP USER IF EXISTS ''@'localhost';
DROP USER IF EXISTS ''@'%';
DROP DATABASE IF EXISTS test;
FLUSH PRIVILEGES;
```

#### 更新环境变量
```bash
# 编辑.env文件，使用强密码
DB_PASSWORD=your_very_strong_password_here_with_special_chars_123!
```

### 2. 应用层安全

#### 限制文件上传
在 `.env` 文件中配置：
```bash
# 文件大小限制（10MB）
MAX_FILE_SIZE=10485760

# 批量上传限制
MAX_BATCH_SIZE=52428800
MAX_BATCH_COUNT=10
```

#### 配置速率限制
后端已内置速率限制中间件，可以通过环境变量调整：
```bash
# 每分钟最大请求数
RATE_LIMIT_PER_MINUTE=60

# 每小时最大上传数
UPLOAD_LIMIT_PER_HOUR=100
```

### 3. Nginx安全配置

编辑 `nginx/conf.d/default.conf`，添加安全配置：

```nginx
server {
    # 隐藏Nginx版本
    server_tokens off;
    
    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self'; img-src 'self' data: https:; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';" always;
    
    # 限制请求方法
    if ($request_method !~ ^(GET|HEAD|POST|PUT|DELETE|OPTIONS)$) {
        return 405;
    }
    
    # 限制文件上传大小
    client_max_body_size 100M;
    client_body_buffer_size 128k;
    
    # 超时设置
    client_body_timeout 12;
    client_header_timeout 12;
    send_timeout 10;
    
    # 限制连接数
    limit_conn_zone $binary_remote_addr zone=conn_limit_per_ip:10m;
    limit_conn conn_limit_per_ip 10;
    
    # 限制请求频率
    limit_req_zone $binary_remote_addr zone=req_limit_per_ip:10m rate=5r/s;
    limit_req zone=req_limit_per_ip burst=10 nodelay;
    
    # 禁止访问敏感文件
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }
    
    location ~ \.(env|log|conf|key|pem)$ {
        deny all;
        access_log off;
        log_not_found off;
    }
    
    # 禁止直接访问PHP文件（如果有）
    location ~ \.php$ {
        deny all;
        access_log off;
        log_not_found off;
    }
}
```

## 🔍 入侵检测和监控

### 1. 安装Fail2Ban
```bash
# Ubuntu/Debian
sudo apt install fail2ban

# CentOS/RHEL
sudo yum install epel-release
sudo yum install fail2ban

# 启动服务
sudo systemctl start fail2ban
sudo systemctl enable fail2ban
```

### 2. 配置Fail2Ban
```bash
# 创建本地配置文件
sudo cp /etc/fail2ban/jail.conf /etc/fail2ban/jail.local

# 编辑配置
sudo nano /etc/fail2ban/jail.local
```

基本配置：
```ini
[DEFAULT]
# 封禁时间（秒）
bantime = 3600

# 查找时间窗口（秒）
findtime = 600

# 最大重试次数
maxretry = 3

# 忽略的IP（白名单）
ignoreip = 127.0.0.1/8 ::1 192.168.1.0/24

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600

[nginx-http-auth]
enabled = true
filter = nginx-http-auth
logpath = /var/log/nginx/error.log
maxretry = 3
bantime = 3600

[nginx-limit-req]
enabled = true
filter = nginx-limit-req
logpath = /var/log/nginx/error.log
maxretry = 10
bantime = 600
```

### 3. 重启Fail2Ban
```bash
sudo systemctl restart fail2ban

# 查看状态
sudo fail2ban-client status

# 查看特定jail状态
sudo fail2ban-client status sshd
```

## 📊 日志监控

### 1. 配置日志轮转
```bash
# 创建日志轮转配置
sudo nano /etc/logrotate.d/imgtourl
```

配置内容：
```bash
/var/log/nginx/*.log {
    daily
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    create 644 nginx nginx
    postrotate
        docker-compose exec nginx nginx -s reload
    endscript
}

/path/to/your/project/backend/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 root root
}
```

### 2. 监控脚本
创建简单的监控脚本：

```bash
# 创建监控脚本
cat > /usr/local/bin/imgtourl-monitor.sh << 'EOF'
#!/bin/bash

# 检查服务状态
check_service() {
    if ! docker-compose ps | grep -q "Up"; then
        echo "[$(date)] 警告: 发现服务异常" >> /var/log/imgtourl-monitor.log
        # 可以添加邮件通知或其他告警
    fi
}

# 检查磁盘空间
check_disk() {
    USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    if [ $USAGE -gt 80 ]; then
        echo "[$(date)] 警告: 磁盘使用率超过80%: ${USAGE}%" >> /var/log/imgtourl-monitor.log
    fi
}

# 检查内存使用
check_memory() {
    USAGE=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
    if [ $USAGE -gt 90 ]; then
        echo "[$(date)] 警告: 内存使用率超过90%: ${USAGE}%" >> /var/log/imgtourl-monitor.log
    fi
}

check_service
check_disk
check_memory
EOF

# 设置执行权限
sudo chmod +x /usr/local/bin/imgtourl-monitor.sh

# 添加到crontab（每5分钟检查一次）
echo "*/5 * * * * /usr/local/bin/imgtourl-monitor.sh" | sudo crontab -
```

## 🔐 数据备份和恢复

### 1. 自动备份脚本
```bash
# 创建备份脚本
cat > /usr/local/bin/imgtourl-backup.sh << 'EOF'
#!/bin/bash

BACKUP_DIR="/backup/imgtourl"
DATE=$(date +%Y%m%d_%H%M%S)

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份数据库
docker-compose exec -T mysql mysqldump -u root -p$DB_PASSWORD image_host > $BACKUP_DIR/database_$DATE.sql

# 备份上传的文件
tar -czf $BACKUP_DIR/uploads_$DATE.tar.gz backend/uploads/

# 备份配置文件
tar -czf $BACKUP_DIR/config_$DATE.tar.gz .env nginx/ docker-compose*.yml

# 删除7天前的备份
find $BACKUP_DIR -name "*.sql" -mtime +7 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete

echo "[$(date)] 备份完成: $DATE" >> /var/log/imgtourl-backup.log
EOF

# 设置执行权限
sudo chmod +x /usr/local/bin/imgtourl-backup.sh

# 添加到crontab（每天凌晨2点备份）
echo "0 2 * * * /usr/local/bin/imgtourl-backup.sh" | sudo crontab -
```

### 2. 恢复数据
```bash
# 恢复数据库
docker-compose exec -T mysql mysql -u root -p$DB_PASSWORD image_host < /backup/imgtourl/database_20240101_020000.sql

# 恢复文件
tar -xzf /backup/imgtourl/uploads_20240101_020000.tar.gz

# 恢复配置
tar -xzf /backup/imgtourl/config_20240101_020000.tar.gz
```

## 🚨 安全检查清单

### 系统安全
- [ ] 防火墙已配置并启用
- [ ] SSH已配置密钥认证
- [ ] SSH已禁用root登录
- [ ] 系统已更新到最新版本
- [ ] Fail2Ban已安装并配置
- [ ] 定期安全更新已配置

### 应用安全
- [ ] 数据库使用强密码
- [ ] 数据库不使用root用户
- [ ] 文件上传大小已限制
- [ ] 速率限制已配置
- [ ] 敏感文件访问已禁止
- [ ] 安全头已配置

### 监控和备份
- [ ] 日志监控已配置
- [ ] 自动备份已设置
- [ ] 监控脚本已部署
- [ ] 告警机制已配置

### SSL和网络安全
- [ ] SSL证书已配置
- [ ] HTTPS重定向已启用
- [ ] 安全的TLS配置
- [ ] HSTS已启用

## 🔧 安全维护

### 定期任务
1. **每周**: 检查系统更新
2. **每月**: 审查访问日志
3. **每季度**: 更新密码和密钥
4. **每年**: 全面安全审计

### 应急响应
1. **发现异常**: 立即隔离受影响系统
2. **分析日志**: 确定攻击来源和方式
3. **修复漏洞**: 更新系统和应用
4. **恢复服务**: 从备份恢复数据
5. **加强防护**: 更新安全策略

---

遵循本指南可以大大提高您的图床系统安全性！ 🛡️