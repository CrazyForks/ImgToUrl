#!/bin/bash

# 图床转换系统部署脚本
# 支持开发环境和生产环境部署

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "$1 未安装，请先安装 $1"
        exit 1
    fi
}

# 检查环境变量文件
check_env_file() {
    if [ ! -f ".env" ]; then
        log_warning ".env 文件不存在"
        if [ -f ".env.example" ]; then
            log_info "复制 .env.example 到 .env"
            cp .env.example .env
            log_warning "请编辑 .env 文件并填入正确的配置值"
            read -p "是否现在编辑 .env 文件? (y/n): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                ${EDITOR:-nano} .env
            fi
        else
            log_error ".env.example 文件也不存在，请检查项目完整性"
            exit 1
        fi
    fi
}

# 显示帮助信息
show_help() {
    echo "图床转换系统部署脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示帮助信息"
    echo "  -e, --env ENV           指定环境 (dev|prod) [默认: dev]"
    echo "  -d, --detach            后台运行 (仅适用于生产环境)"
    echo "  -b, --build             强制重新构建镜像"
    echo "  -c, --clean             清理旧的容器和镜像"
    echo "  -s, --stop              停止所有服务"
    echo "  -r, --restart           重启所有服务"
    echo "  -l, --logs              查看服务日志"
    echo "  --backup                备份数据库"
    echo "  --restore FILE          从备份文件恢复数据库"
    echo ""
    echo "示例:"
    echo "  $0                      # 开发环境部署"
    echo "  $0 -e prod -d           # 生产环境后台部署"
    echo "  $0 -b                   # 重新构建并部署"
    echo "  $0 -s                   # 停止所有服务"
    echo "  $0 -l                   # 查看日志"
}

# 清理函数
cleanup() {
    log_info "清理旧的容器和镜像..."
    docker-compose down --remove-orphans
    docker system prune -f
    docker volume prune -f
    log_success "清理完成"
}

# 备份数据库
backup_database() {
    log_info "备份数据库..."
    BACKUP_FILE="backup_$(date +%Y%m%d_%H%M%S).sql"
    docker-compose exec mysql mysqldump -u root -p\$MYSQL_ROOT_PASSWORD image_host > "backups/$BACKUP_FILE"
    log_success "数据库备份完成: backups/$BACKUP_FILE"
}

# 恢复数据库
restore_database() {
    if [ -z "$1" ]; then
        log_error "请指定备份文件路径"
        exit 1
    fi
    
    if [ ! -f "$1" ]; then
        log_error "备份文件不存在: $1"
        exit 1
    fi
    
    log_info "从备份文件恢复数据库: $1"
    docker-compose exec -T mysql mysql -u root -p\$MYSQL_ROOT_PASSWORD image_host < "$1"
    log_success "数据库恢复完成"
}

# 检查服务健康状态
check_health() {
    log_info "检查服务健康状态..."
    
    # 等待服务启动
    sleep 10
    
    # 检查数据库
    if docker-compose exec mysql mysqladmin ping -h localhost --silent; then
        log_success "MySQL 服务正常"
    else
        log_error "MySQL 服务异常"
    fi
    
    # 检查 Redis
    if docker-compose exec redis redis-cli ping | grep -q PONG; then
        log_success "Redis 服务正常"
    else
        log_warning "Redis 服务异常或未启用"
    fi
    
    # 检查后端 API
    if curl -f http://localhost:8080/api/health &> /dev/null; then
        log_success "后端 API 服务正常"
    else
        log_error "后端 API 服务异常"
    fi
    
    # 检查前端
    if curl -f http://localhost:3000 &> /dev/null; then
        log_success "前端服务正常"
    else
        log_error "前端服务异常"
    fi
}

# 显示服务信息
show_services() {
    echo ""
    log_info "服务访问信息:"
    echo "  前端地址: http://localhost:3000"
    echo "  后端 API: http://localhost:8080"
    echo "  数据库: localhost:3306"
    echo "  Redis: localhost:6379"
    echo ""
    log_info "常用命令:"
    echo "  查看日志: docker-compose logs -f [service_name]"
    echo "  进入容器: docker-compose exec [service_name] sh"
    echo "  停止服务: docker-compose down"
    echo "  重启服务: docker-compose restart [service_name]"
    echo ""
}

# 主要部署函数
deploy() {
    local env=${1:-dev}
    local detach=${2:-false}
    local build=${3:-false}
    
    log_info "开始部署图床转换系统 (环境: $env)"
    
    # 检查必要的命令
    check_command docker
    check_command docker-compose
    
    # 检查环境变量文件
    check_env_file
    
    # 创建必要的目录
    mkdir -p logs backups uploads temp
    
    # 设置权限
    chmod 755 logs backups uploads temp
    
    # 构建参数
    local compose_args=""
    if [ "$build" = true ]; then
        compose_args="--build"
    fi
    
    if [ "$detach" = true ]; then
        compose_args="$compose_args -d"
    fi
    
    # 选择 docker-compose 文件
    local compose_file="docker-compose.yml"
    if [ "$env" = "prod" ]; then
        if [ -f "docker-compose.prod.yml" ]; then
            compose_file="docker-compose.prod.yml"
        fi
    fi
    
    log_info "使用配置文件: $compose_file"
    
    # 启动服务
    log_info "启动服务..."
    docker-compose -f $compose_file up $compose_args
    
    if [ "$detach" = true ]; then
        # 检查服务健康状态
        check_health
        
        # 显示服务信息
        show_services
        
        log_success "部署完成！系统正在后台运行"
    else
        log_info "按 Ctrl+C 停止服务"
    fi
}

# 解析命令行参数
ENV="dev"
DETACH=false
BUILD=false
CLEAN=false
STOP=false
RESTART=false
LOGS=false
BACKUP=false
RESTORE_FILE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -e|--env)
            ENV="$2"
            shift 2
            ;;
        -d|--detach)
            DETACH=true
            shift
            ;;
        -b|--build)
            BUILD=true
            shift
            ;;
        -c|--clean)
            CLEAN=true
            shift
            ;;
        -s|--stop)
            STOP=true
            shift
            ;;
        -r|--restart)
            RESTART=true
            shift
            ;;
        -l|--logs)
            LOGS=true
            shift
            ;;
        --backup)
            BACKUP=true
            shift
            ;;
        --restore)
            RESTORE_FILE="$2"
            shift 2
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 执行相应的操作
if [ "$CLEAN" = true ]; then
    cleanup
elif [ "$STOP" = true ]; then
    log_info "停止所有服务..."
    docker-compose down
    log_success "所有服务已停止"
elif [ "$RESTART" = true ]; then
    log_info "重启所有服务..."
    docker-compose restart
    log_success "所有服务已重启"
elif [ "$LOGS" = true ]; then
    docker-compose logs -f
elif [ "$BACKUP" = true ]; then
    backup_database
elif [ -n "$RESTORE_FILE" ]; then
    restore_database "$RESTORE_FILE"
else
    deploy "$ENV" "$DETACH" "$BUILD"
fi