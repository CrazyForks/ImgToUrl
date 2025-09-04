# GitHub 部署指南

## 项目已准备就绪

✅ Git 仓库已初始化  
✅ .gitignore 文件已创建  
✅ 初始提交已完成  
✅ 敏感文件已排除（.env 文件不会被上传）

## 下一步：上传到 GitHub

### 1. 在 GitHub 上创建新仓库

1. 访问 [GitHub](https://github.com)
2. 点击右上角的 "+" 按钮，选择 "New repository"
3. 填写仓库信息：
   - **Repository name**: `image-hosting-system`（或您喜欢的名称）
   - **Description**: `图床系统 - Go后端 + Vue3前端`
   - **Visibility**: 选择 Public 或 Private
   - ⚠️ **不要**勾选 "Add a README file"、"Add .gitignore" 或 "Choose a license"（因为我们已经有了这些文件）
4. 点击 "Create repository"

### 2. 连接本地仓库到 GitHub

创建仓库后，GitHub 会显示一个页面，复制其中的远程仓库 URL（类似 `https://github.com/username/repository-name.git`）

然后在项目根目录执行以下命令：

```bash
# 添加远程仓库（替换为您的实际 GitHub 仓库 URL）
git remote add origin https://github.com/YOUR_USERNAME/YOUR_REPOSITORY.git

# 推送代码到 GitHub
git branch -M main
git push -u origin main
```

### 3. 验证上传

上传完成后，刷新 GitHub 仓库页面，您应该能看到所有项目文件。

## 重要提醒

- ✅ `.env` 文件已被 `.gitignore` 排除，不会上传敏感信息
- ✅ `uploads/` 目录已被排除，不会上传用户文件
- ✅ `node_modules/` 等依赖目录已被排除

## 项目结构说明

```
图床转换/
├── backend/          # Go 后端服务
├── frontend/         # Vue3 前端应用
├── database/         # 数据库初始化脚本
├── docker-compose.yml # Docker 部署配置
├── README.md         # 项目说明
├── R2_CONFIG_GUIDE.md # R2 配置指南
└── .gitignore        # Git 忽略文件
```

## 后续步骤

1. 在 GitHub 仓库中添加详细的 README.md
2. 设置 GitHub Actions 进行 CI/CD（可选）
3. 配置 GitHub Pages 部署前端（如需要）
4. 添加 Issues 和 Pull Request 模板（可选）

---

**注意**: 如果您需要部署到生产环境，请确保：
- 配置真实的 Cloudflare R2 凭证
- 设置安全的数据库连接
- 配置 HTTPS 和域名
- 设置适当的环境变量