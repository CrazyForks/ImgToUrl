# Cloudflare R2 配置指南

## 问题诊断

根据错误日志显示：`TLS handshake failure`，这表明当前 `.env` 文件中的 R2 配置使用的是测试数据，不是真实有效的 Cloudflare R2 凭证。

## 解决方案

### 1. 获取真实的 Cloudflare R2 凭证

#### 步骤 1：登录 Cloudflare 控制台
1. 访问 [Cloudflare Dashboard](https://dash.cloudflare.com/)
2. 登录您的 Cloudflare 账户

#### 步骤 2：创建 R2 存储桶
1. 在左侧菜单中选择 "R2 Object Storage"
2. 点击 "创建存储桶" 或 "Create bucket"
3. 输入存储桶名称（例如：`my-image-host`）
4. 选择位置（建议选择离您用户最近的区域）
5. 点击 "创建存储桶"

#### 步骤 3：获取 API 凭证
1. 在 R2 页面中，点击 "管理 R2 API 令牌" 或 "Manage R2 API tokens"
2. 点击 "创建 API 令牌" 或 "Create API token"
3. 选择 "自定义令牌" 或 "Custom token"
4. 配置权限：
   - **权限**：`Object Storage:Edit`
   - **资源**：选择您刚创建的存储桶
5. 点击 "继续以显示摘要" 然后 "创建令牌"
6. **重要**：复制并保存显示的 Access Key ID 和 Secret Access Key

#### 步骤 4：获取账户 ID 和端点
1. 在 Cloudflare 控制台右侧边栏找到 "账户 ID" 或 "Account ID"
2. 复制账户 ID
3. R2 端点格式：`https://[账户ID].r2.cloudflarestorage.com`

#### 步骤 5：配置公共访问（可选）
1. 在存储桶设置中，找到 "公共访问" 或 "Public access"
2. 可以选择：
   - 使用 R2.dev 子域名（免费）：`https://pub-[随机字符串].r2.dev`
   - 配置自定义域名（需要域名在 Cloudflare）

### 2. 更新配置文件

编辑 `backend/.env` 文件，替换以下配置：

```env
# Cloudflare R2 配置
R2_ACCOUNT_ID=你的真实账户ID
R2_ACCESS_KEY_ID=你的真实Access_Key_ID
R2_SECRET_ACCESS_KEY=你的真实Secret_Access_Key
R2_BUCKET_NAME=你的存储桶名称
R2_ENDPOINT=https://你的账户ID.r2.cloudflarestorage.com
R2_REGION=auto
R2_PUBLIC_URL=https://你的公共访问URL
```

### 3. 重启服务

配置更新后，需要重启后端服务：

```bash
# 停止当前服务（Ctrl+C）
# 然后重新启动
cd backend
go run main.go
```

## 配置示例

假设您的配置如下：
- 账户 ID：`abc123def456ghi789`
- Access Key ID：`1234567890abcdef`
- Secret Access Key：`abcdef1234567890abcdef1234567890abcdef12`
- 存储桶名称：`my-image-host`
- 公共 URL：`https://pub-xyz789.r2.dev`

那么您的 `.env` 配置应该是：

```env
R2_ACCOUNT_ID=abc123def456ghi789
R2_ACCESS_KEY_ID=1234567890abcdef
R2_SECRET_ACCESS_KEY=abcdef1234567890abcdef1234567890abcdef12
R2_BUCKET_NAME=my-image-host
R2_ENDPOINT=https://abc123def456ghi789.r2.cloudflarestorage.com
R2_REGION=auto
R2_PUBLIC_URL=https://pub-xyz789.r2.dev
```

## 测试配置

配置完成后，重启服务并尝试上传图片。如果配置正确，图片应该能够成功上传到 R2 存储桶。

## 常见问题

### Q: 如何验证配置是否正确？
A: 重启后端服务后，尝试上传一张图片。如果成功，说明配置正确。

### Q: 仍然出现 TLS 错误怎么办？
A: 请检查：
1. 账户 ID 是否正确
2. API 凭证是否有效
3. 存储桶名称是否正确
4. 网络连接是否正常

### Q: 如何设置自定义域名？
A: 在 Cloudflare 控制台中，进入您的域名设置，添加 CNAME 记录指向 R2 存储桶。

## 安全提醒

- **不要**将真实的 API 凭证提交到公共代码仓库
- **不要**在生产环境中使用过于宽泛的权限
- 定期轮换 API 密钥
- 考虑使用环境变量或密钥管理服务