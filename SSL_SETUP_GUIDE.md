# SSL证书配置指南

本指南将帮助您为图床系统配置SSL证书，实现HTTPS访问。

## 🔒 SSL证书获取方式

### 方式一：Let's Encrypt 免费证书（推荐）

Let's Encrypt提供免费的SSL证书，有效期90天，支持自动续期。

#### 1. 安装Certbot

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install certbot
```

**CentOS/RHEL:**
```bash
sudo yum install epel-release
sudo yum install certbot
```

#### 2. 获取证书

**方法A: 独立模式（推荐）**
```bash
# 停止当前的Web服务
docker-compose down

# 获取证书
sudo certbot certonly --standalone -d your_domain.com -d www.your_domain.com

# 重新启动服务
docker-compose up -d
```

**方法B: Webroot模式**
```bash
# 创建webroot目录
mkdir -p /var/www/certbot

# 获取证书
sudo certbot certonly --webroot -w /var/www/certbot -d your_domain.com -d www.your_domain.com
```

#### 3. 复制证书到项目目录
```bash
# 复制证书文件
sudo cp /etc/letsencrypt/live/your_domain.com/fullchain.pem nginx/ssl/
sudo cp /etc/letsencrypt/live/your_domain.com/privkey.pem nginx/ssl/

# 修改文件权限
sudo chown $USER:$USER nginx/ssl/*
chmod 600 nginx/ssl/privkey.pem
chmod 644 nginx/ssl/fullchain.pem
```

#### 4. 设置自动续期
```bash
# 添加续期任务到crontab
echo "0 12 * * * /usr/bin/certbot renew --quiet --deploy-hook 'docker-compose restart nginx'" | sudo crontab -
```

### 方式二：自签名证书（仅用于测试）

⚠️ **警告**: 自签名证书仅适用于测试环境，浏览器会显示安全警告。

```bash
# 生成私钥
openssl genrsa -out nginx/ssl/privkey.pem 2048

# 生成证书
openssl req -new -x509 -key nginx/ssl/privkey.pem -out nginx/ssl/fullchain.pem -days 365 -subj "/C=CN/ST=State/L=City/O=Organization/CN=your_domain.com"
```

### 方式三：商业证书

如果您购买了商业SSL证书，请将证书文件放置到以下位置：
- 证书文件: `nginx/ssl/fullchain.pem`
- 私钥文件: `nginx/ssl/privkey.pem`

## ⚙️ 配置HTTPS

### 1. 修改Nginx配置

编辑 `nginx/conf.d/default.conf` 文件，取消HTTPS配置的注释：

```bash
# 编辑配置文件
nano nginx/conf.d/default.conf
```

找到HTTPS配置部分，取消注释并修改域名：

```nginx
server {
    listen 443 ssl http2;
    server_name your_domain.com;  # 替换为您的实际域名
    
    # SSL证书配置
    ssl_certificate /etc/nginx/ssl/fullchain.pem;
    ssl_certificate_key /etc/nginx/ssl/privkey.pem;
    
    # ... 其他配置
}
```

### 2. 启用HTTP到HTTPS重定向

在HTTP配置中启用重定向：

```nginx
server {
    listen 80;
    server_name your_domain.com;
    
    # 重定向到HTTPS
    return 301 https://$server_name$request_uri;
}
```

### 3. 更新环境变量

修改 `.env` 文件，将API地址改为HTTPS：

```bash
# 将HTTP改为HTTPS
VITE_API_BASE_URL=https://your_domain.com/api
```

### 4. 重新部署

```bash
# 重新构建前端（因为API地址变更）
docker-compose build frontend

# 重启服务
docker-compose restart
```

## 🔧 使用Nginx代理部署

### 1. 启用Nginx代理

```bash
# 使用包含Nginx代理的配置启动
docker-compose --profile proxy up -d
```

### 2. 访问地址

启用Nginx代理后，访问地址为：
- HTTP: `http://your_domain.com:8000`
- HTTPS: `https://your_domain.com:8443`

### 3. 修改端口映射（可选）

如果希望使用标准端口（80/443），可以修改 `docker-compose.prod.yml` 中的端口映射：

```yaml
nginx:
  ports:
    - "80:80"    # HTTP
    - "443:443"  # HTTPS
```

## 🛡️ 安全最佳实践

### 1. SSL安全配置

确保Nginx配置包含以下安全设置：

```nginx
# 仅使用安全的TLS版本
ssl_protocols TLSv1.2 TLSv1.3;

# 使用安全的加密套件
ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;
ssl_prefer_server_ciphers off;

# 启用HSTS
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

# 其他安全头
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header X-Content-Type-Options "nosniff" always;
```

### 2. 防火墙配置

```bash
# 开放HTTPS端口
sudo ufw allow 443/tcp

# 如果不需要HTTP访问，可以关闭80端口
# sudo ufw deny 80/tcp
```

### 3. 定期更新证书

```bash
# 手动测试证书续期
sudo certbot renew --dry-run

# 检查证书有效期
openssl x509 -in nginx/ssl/fullchain.pem -noout -dates
```

## 🔍 故障排除

### 1. 证书验证失败

```bash
# 检查证书文件是否存在
ls -la nginx/ssl/

# 验证证书内容
openssl x509 -in nginx/ssl/fullchain.pem -text -noout

# 检查私钥
openssl rsa -in nginx/ssl/privkey.pem -check
```

### 2. Nginx配置错误

```bash
# 测试Nginx配置
docker-compose exec nginx nginx -t

# 查看Nginx日志
docker-compose logs nginx
```

### 3. 证书过期

```bash
# 强制续期证书
sudo certbot renew --force-renewal

# 重启Nginx
docker-compose restart nginx
```

### 4. 域名解析问题

```bash
# 检查域名解析
nslookup your_domain.com

# 检查端口连通性
telnet your_domain.com 443
```

## 📋 SSL检查清单

- [ ] 域名已正确解析到服务器IP
- [ ] 防火墙已开放443端口
- [ ] SSL证书文件已正确放置
- [ ] Nginx配置已启用HTTPS
- [ ] 环境变量已更新为HTTPS地址
- [ ] 服务已重新部署
- [ ] 证书自动续期已配置
- [ ] 安全头已正确配置

## 🌐 在线SSL检测工具

部署完成后，可以使用以下工具检测SSL配置：

- [SSL Labs SSL Test](https://www.ssllabs.com/ssltest/)
- [SSL Checker](https://www.sslchecker.com/sslchecker)
- [DigiCert SSL Installation Checker](https://www.digicert.com/help/)

---

配置完成后，您的图床系统将支持安全的HTTPS访问！ 🔒