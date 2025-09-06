# 多阶段：构建前端 + 打包到 Nginx
FROM node:20-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install --no-audit --no-fund
COPY frontend/ .
RUN npm run build

FROM nginx:alpine
# 拷贝 Nginx 配置
COPY nginx/nginx.conf /etc/nginx/nginx.conf
COPY nginx/conf.d /etc/nginx/conf.d
# 拷贝前端静态资源
COPY --from=frontend-builder /app/dist /usr/share/nginx/html
EXPOSE 80