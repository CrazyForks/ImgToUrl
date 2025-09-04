# 域名配置指南

本指南将帮助您为图床系统配置域名和DNS解析。

## 🌐 域名准备

### 1. 购买域名

推荐的域名注册商：
- **国外**: Namecheap, GoDaddy, Cloudflare
- **国内**: 阿里云, 腾讯云, 华为云

### 2. 域名选择建议

- 选择简短易记的域名
- 建议使用 `.com`, `.net`, `.org` 等通用顶级域名
- 可以考虑包含 `img`, `pic`, `image` 等关键词

**示例域名**:
- `img.yourdomain.com`
- `pic.yourdomain.com`
- `images.yourdomain.com`
- `cdn.yourdomain.com`

## 🔧 DNS配置

### 1. 获取VPS IP地址

```bash
# 在VPS上执行，获取公网IP
curl ifconfig.me
# 或
curl ipinfo.io/ip
```

### 2. 配置DNS记录

在域名注册商或DNS服务商的控制面板中添加以下记录：

#### A记录（必需）
```
类型: A
主机记录: @
记录值: 您的VPS公网IP
TTL: 600
```

#### CNAME记录（可选）
```
类型: CNAME
主机记录: www
记录值: yourdomain.com
TTL: 600
```

#### 子域名配置（推荐）
如果使用子域名（如 img.yourdomain.com）：
```
类型: A
主机记录: img
记录值: 您的VPS公网IP
TTL: 600
```

### 3. 验证DNS解析

```bash
# 检查域名解析
nslookup yourdomain.com

# 检查特定DNS服务器
nslookup yourdomain.com 8.8.8.8

# 使用dig命令（更详细）
dig yourdomain.com
```

## 🚀 Cloudflare配置（推荐）

Cloudflare提供免费的CDN和DNS服务，可以提高网站访问速度和安全性。

### 1. 注册Cloudflare账户

访问 [Cloudflare官网](https://www.cloudflare.com/) 注册账户。

### 2. 添加站点

1. 在Cloudflare控制台点击「添加站点」
2. 输入您的域名
3. 选择免费计划
4. Cloudflare会自动扫描现有DNS记录

### 3. 更新域名服务器

将域名的DNS服务器更改为Cloudflare提供的服务器：
```
ns1.cloudflare.com
ns2.cloudflare.com
```

### 4. 配置DNS记录

在Cloudflare DNS管理中添加：

```
类型: A
名称: @
内容: 您的VPS IP
代理状态: 已代理（橙色云朵）

类型: A
名称: www
内容: 您的VPS IP
代理状态: 已代理（橙色云朵）
```

### 5. SSL/TLS配置

在Cloudflare SSL/TLS设置中：
1. 选择「完全（严格）」模式
2. 启用「始终使用HTTPS」
3. 启用「HSTS」

### 6. 性能优化

在「速度」选项卡中：
1. 启用「自动缩小」
2. 启用「Brotli压缩」
3. 配置「缓存规则」

## 📝 更新应用配置

### 1. 修改环境变量

编辑 `.env` 文件，更新API地址：

```bash
# 使用您的实际域名替换
VITE_API_BASE_URL=https://yourdomain.com/api

# 如果使用子域名
# VITE_API_BASE_URL=https://img.yourdomain.com/api
```

### 2. 更新Nginx配置

编辑 `nginx/conf.d/default.conf`，更新server_name：

```nginx
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;  # 更新为您的域名
    # ...
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;  # 更新为您的域名
    # ...
}
```

### 3. 重新部署

```bash
# 重新构建前端（因为API地址变更）
docker-compose build frontend

# 重启服务
docker-compose restart
```

## 🔒 SSL证书配置

### 使用Let's Encrypt

```bash
# 获取SSL证书
sudo certbot certonly --standalone -d yourdomain.com -d www.yourdomain.com

# 复制证书到项目目录
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx/ssl/
sudo chown $USER:$USER nginx/ssl/*
```

详细SSL配置请参考 `SSL_SETUP_GUIDE.md`。

## 🌍 多域名配置

如果您有多个域名指向同一个服务：

### 1. DNS配置

为每个域名配置A记录指向同一个IP。

### 2. Nginx配置

```nginx
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com img.yourdomain.com;
    # ...
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com www.yourdomain.com img.yourdomain.com;
    # ...
}
```

### 3. SSL证书

```bash
# 为多个域名获取证书
sudo certbot certonly --standalone \
  -d yourdomain.com \
  -d www.yourdomain.com \
  -d img.yourdomain.com
```

## 📊 域名监控

### 1. DNS监控工具

- [DNSChecker](https://dnschecker.org/)
- [WhatsMyDNS](https://www.whatsmydns.net/)
- [DNS Propagation Checker](https://www.dnspropagation.net/)

### 2. 网站监控

- [UptimeRobot](https://uptimerobot.com/)
- [Pingdom](https://www.pingdom.com/)
- [StatusCake](https://www.statuscake.com/)

### 3. 性能测试

- [GTmetrix](https://gtmetrix.com/)
- [PageSpeed Insights](https://pagespeed.web.dev/)
- [WebPageTest](https://www.webpagetest.org/)

## 🚨 故障排除

### 1. DNS解析问题

```bash
# 清除本地DNS缓存
# Windows
ipconfig /flushdns

# macOS
sudo dscacheutil -flushcache

# Linux
sudo systemctl restart systemd-resolved
```

### 2. 域名未生效

- DNS记录修改后需要等待传播（通常5-30分钟）
- 检查TTL设置，较低的TTL值传播更快
- 使用不同的DNS服务器测试

### 3. SSL证书问题

```bash
# 检查证书状态
openssl s_client -connect yourdomain.com:443 -servername yourdomain.com

# 检查证书有效期
echo | openssl s_client -connect yourdomain.com:443 2>/dev/null | openssl x509 -noout -dates
```

### 4. Cloudflare相关问题

- 确保代理状态正确（橙色云朵表示已代理）
- 检查SSL/TLS模式设置
- 查看Cloudflare的「分析」页面了解流量情况

## 📋 域名配置清单

- [ ] 域名已购买并可管理
- [ ] DNS A记录已配置
- [ ] DNS解析已生效
- [ ] 环境变量已更新
- [ ] Nginx配置已更新
- [ ] SSL证书已配置
- [ ] 服务已重新部署
- [ ] 网站可通过域名正常访问
- [ ] HTTPS重定向正常工作
- [ ] 域名监控已设置

## 🎯 最佳实践

1. **使用CDN**: 推荐使用Cloudflare等CDN服务
2. **启用HTTPS**: 始终使用SSL证书
3. **设置监控**: 配置域名和网站监控
4. **备份配置**: 定期备份DNS配置
5. **文档记录**: 记录所有配置信息

---

配置完成后，您的图床系统将可以通过自定义域名访问！ 🌐