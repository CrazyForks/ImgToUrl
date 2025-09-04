#!/bin/bash

# ImgToUrl VPS 快速部署脚本
# 使用方法: chmod +x deploy-vps.sh && ./deploy-vps.sh

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否为root用户
check_root() {
    if [[ $EUID -eq 0 ]]; then
        print_warning "检测到您正在使用root用户运行此脚本"
        print_warning "建议创建普通用户来运行Docker服务"
        read -p "是否继续？(y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# 检测操作系统
detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$NAME
        VER=$VERSION_ID
    else
        print_error "无法检测操作系统"
        exit 1
    fi
    print_info "检测到操作系统: $OS $VER"
}

# 安装Docker
install_docker() {
    print_info "检查Docker安装状态..."
    
    if command -v docker &> /dev/null; then
        print_success "Docker已安装: $(docker --version)"
    else
        print_info "开始安装Docker..."
        
        if [[ "$OS" == *"Ubuntu"* ]] || [[ "$OS" == *"Debian"* ]]; then
            # Ubuntu/Debian
            sudo apt update
            sudo apt install -y apt-transport-https ca-certificates curl gnupg lsb-release
            
            curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
            
            echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
            
            sudo apt update
            sudo apt install -y docker-ce docker-ce-cli containerd.io
            
        elif [[ "$OS" == *"CentOS"* ]] || [[ "$OS" == *"Red Hat"* ]]; then
            # CentOS/RHEL
            sudo yum install -y yum-utils
            sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
            sudo yum install -y docker-ce docker-ce-cli containerd.io
        else
            print_error "不支持的操作系统: $OS"
            exit 1
        fi
        
        # 启动Docker服务
        sudo systemctl start docker
        sudo systemctl enable docker
        
        # 添加当前用户到docker组
        sudo usermod -aG docker $USER
        
        print_success "Docker安装完成"
    fi
}

# 安装Docker Compose
install_docker_compose() {
    print_info "检查Docker Compose安装状态..."
    
    if command -v docker-compose &> /dev/null; then
        print_success "Docker Compose已安装: $(docker-compose --version)"
    else
        print_info "开始安装Docker Compose..."
        
        # 获取最新版本
        COMPOSE_VERSION=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep 'tag_name' | cut -d'"' -f4)
        
        # 下载并安装
        sudo curl -L "https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        
        sudo chmod +x /usr/local/bin/docker-compose
        
        print_success "Docker Compose安装完成: $(docker-compose --version)"
    fi
}

# 配置防火墙
setup_firewall() {
    print_info "配置防火墙..."
    
    if command -v ufw &> /dev/null; then
        # Ubuntu/Debian - ufw
        sudo ufw allow 22/tcp
        sudo ufw allow 80/tcp
        sudo ufw allow 443/tcp
        sudo ufw --force enable
        print_success "UFW防火墙配置完成"
    elif command -v firewall-cmd &> /dev/null; then
        # CentOS/RHEL - firewalld
        sudo systemctl start firewalld
        sudo systemctl enable firewalld
        sudo firewall-cmd --permanent --add-service=ssh
        sudo firewall-cmd --permanent --add-service=http
        sudo firewall-cmd --permanent --add-service=https
        sudo firewall-cmd --reload
        print_success "Firewalld防火墙配置完成"
    else
        print_warning "未检测到防火墙管理工具，请手动配置防火墙"
    fi
}

# 创建项目目录和配置
setup_project() {
    print_info "设置项目配置..."
    
    # 创建必要的目录
    mkdir -p nginx/conf.d nginx/ssl nginx/logs
    mkdir -p backend/logs backend/uploads
    
    # 复制生产环境配置
    if [[ ! -f .env ]]; then
        if [[ -f .env.prod ]]; then
            cp .env.prod .env
            print_success "已复制生产环境配置模板到 .env"
            print_warning "请编辑 .env 文件，配置您的数据库密码和域名"
        else
            print_error "未找到 .env.prod 模板文件"
            exit 1
        fi
    else
        print_info ".env 文件已存在，跳过复制"
    fi
    
    # 创建基础Nginx配置
    if [[ ! -f nginx/conf.d/default.conf ]]; then
        cat > nginx/conf.d/default.conf << 'EOF'
server {
    listen 80;
    server_name _;
    
    client_max_body_size 100M;
    
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
        print_success "已创建基础Nginx配置"
    fi
}

# 部署应用
deploy_app() {
    print_info "开始部署应用..."
    
    # 构建并启动服务
    if [[ -f docker-compose.prod.yml ]]; then
        print_info "使用生产环境配置部署..."
        docker-compose -f docker-compose.prod.yml build
        docker-compose -f docker-compose.prod.yml up -d
    else
        print_info "使用默认配置部署..."
        docker-compose build
        docker-compose up -d mysql redis backend frontend
    fi
    
    print_success "应用部署完成！"
}

# 检查服务状态
check_services() {
    print_info "检查服务状态..."
    
    sleep 10
    
    if [[ -f docker-compose.prod.yml ]]; then
        docker-compose -f docker-compose.prod.yml ps
    else
        docker-compose ps
    fi
    
    # 检查健康状态
    print_info "等待服务启动..."
    sleep 30
    
    if curl -f http://localhost:8080/api/health &> /dev/null; then
        print_success "后端服务运行正常"
    else
        print_warning "后端服务可能还在启动中，请稍后检查"
    fi
    
    if curl -f http://localhost &> /dev/null; then
        print_success "前端服务运行正常"
    else
        print_warning "前端服务可能还在启动中，请稍后检查"
    fi
}

# 显示部署信息
show_info() {
    echo
    print_success "=== 部署完成 ==="
    echo
    print_info "服务访问地址:"
    echo "  前端界面: http://$(curl -s ifconfig.me)"
    echo "  API接口:  http://$(curl -s ifconfig.me):8080/api"
    echo "  健康检查: http://$(curl -s ifconfig.me):8080/api/health"
    echo
    print_info "常用管理命令:"
    echo "  查看服务状态: docker-compose ps"
    echo "  查看日志:     docker-compose logs -f"
    echo "  重启服务:     docker-compose restart"
    echo "  停止服务:     docker-compose down"
    echo
    print_warning "重要提醒:"
    echo "  1. 请修改 .env 文件中的数据库密码"
    echo "  2. 如需HTTPS，请配置SSL证书"
    echo "  3. 建议配置域名和反向代理"
    echo "  4. 定期备份数据库和上传的文件"
    echo
    print_info "详细文档请查看: VPS_DEPLOY_GUIDE.md"
}

# 主函数
main() {
    echo
    print_info "=== ImgToUrl VPS 部署脚本 ==="
    echo
    
    check_root
    detect_os
    install_docker
    install_docker_compose
    setup_firewall
    setup_project
    deploy_app
    check_services
    show_info
    
    print_success "部署脚本执行完成！"
}

# 执行主函数
main "$@"