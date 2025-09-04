@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

REM 图床转换系统 Windows 部署脚本
REM 支持开发环境和生产环境部署

set "RED=[91m"
set "GREEN=[92m"
set "YELLOW=[93m"
set "BLUE=[94m"
set "NC=[0m"

REM 默认参数
set "ENV=dev"
set "DETACH=false"
set "BUILD=false"
set "CLEAN=false"
set "STOP=false"
set "RESTART=false"
set "LOGS=false"
set "BACKUP=false"
set "RESTORE_FILE="

REM 日志函数
:log_info
echo %BLUE%[INFO]%NC% %~1
goto :eof

:log_success
echo %GREEN%[SUCCESS]%NC% %~1
goto :eof

:log_warning
echo %YELLOW%[WARNING]%NC% %~1
goto :eof

:log_error
echo %RED%[ERROR]%NC% %~1
goto :eof

REM 检查命令是否存在
:check_command
where %1 >nul 2>&1
if errorlevel 1 (
    call :log_error "%1 未安装，请先安装 %1"
    exit /b 1
)
goto :eof

REM 检查环境变量文件
:check_env_file
if not exist ".env" (
    call :log_warning ".env 文件不存在"
    if exist ".env.example" (
        call :log_info "复制 .env.example 到 .env"
        copy ".env.example" ".env" >nul
        call :log_warning "请编辑 .env 文件并填入正确的配置值"
        set /p "choice=是否现在编辑 .env 文件? (y/n): "
        if /i "!choice!"=="y" (
            notepad .env
        )
    ) else (
        call :log_error ".env.example 文件也不存在，请检查项目完整性"
        exit /b 1
    )
)
goto :eof

REM 显示帮助信息
:show_help
echo 图床转换系统 Windows 部署脚本
echo.
echo 用法: %~nx0 [选项]
echo.
echo 选项:
echo   -h, --help              显示帮助信息
echo   -e, --env ENV           指定环境 (dev^|prod) [默认: dev]
echo   -d, --detach            后台运行 (仅适用于生产环境)
echo   -b, --build             强制重新构建镜像
echo   -c, --clean             清理旧的容器和镜像
echo   -s, --stop              停止所有服务
echo   -r, --restart           重启所有服务
echo   -l, --logs              查看服务日志
echo   --backup                备份数据库
echo   --restore FILE          从备份文件恢复数据库
echo.
echo 示例:
echo   %~nx0                      # 开发环境部署
echo   %~nx0 -e prod -d           # 生产环境后台部署
echo   %~nx0 -b                   # 重新构建并部署
echo   %~nx0 -s                   # 停止所有服务
echo   %~nx0 -l                   # 查看日志
goto :eof

REM 清理函数
:cleanup
call :log_info "清理旧的容器和镜像..."
docker-compose down --remove-orphans
docker system prune -f
docker volume prune -f
call :log_success "清理完成"
goto :eof

REM 备份数据库
:backup_database
call :log_info "备份数据库..."
if not exist "backups" mkdir backups
for /f "tokens=1-4 delims=/ " %%i in ('date /t') do set "mydate=%%k%%j%%i"
for /f "tokens=1-2 delims=: " %%i in ('time /t') do set "mytime=%%i%%j"
set "mytime=!mytime: =0!"
set "BACKUP_FILE=backup_!mydate!_!mytime!.sql"
docker-compose exec mysql mysqldump -u root -p%MYSQL_ROOT_PASSWORD% image_host > "backups\!BACKUP_FILE!"
call :log_success "数据库备份完成: backups\!BACKUP_FILE!"
goto :eof

REM 恢复数据库
:restore_database
if "%~1"=="" (
    call :log_error "请指定备份文件路径"
    exit /b 1
)

if not exist "%~1" (
    call :log_error "备份文件不存在: %~1"
    exit /b 1
)

call :log_info "从备份文件恢复数据库: %~1"
docker-compose exec -T mysql mysql -u root -p%MYSQL_ROOT_PASSWORD% image_host < "%~1"
call :log_success "数据库恢复完成"
goto :eof

REM 检查服务健康状态
:check_health
call :log_info "检查服务健康状态..."

REM 等待服务启动
timeout /t 10 /nobreak >nul

REM 检查数据库
docker-compose exec mysql mysqladmin ping -h localhost --silent >nul 2>&1
if errorlevel 1 (
    call :log_error "MySQL 服务异常"
) else (
    call :log_success "MySQL 服务正常"
)

REM 检查 Redis
docker-compose exec redis redis-cli ping | findstr "PONG" >nul 2>&1
if errorlevel 1 (
    call :log_warning "Redis 服务异常或未启用"
) else (
    call :log_success "Redis 服务正常"
)

REM 检查后端 API
curl -f http://localhost:8080/api/health >nul 2>&1
if errorlevel 1 (
    call :log_error "后端 API 服务异常"
) else (
    call :log_success "后端 API 服务正常"
)

REM 检查前端
curl -f http://localhost:3000 >nul 2>&1
if errorlevel 1 (
    call :log_error "前端服务异常"
) else (
    call :log_success "前端服务正常"
)
goto :eof

REM 显示服务信息
:show_services
echo.
call :log_info "服务访问信息:"
echo   前端地址: http://localhost:3000
echo   后端 API: http://localhost:8080
echo   数据库: localhost:3306
echo   Redis: localhost:6379
echo.
call :log_info "常用命令:"
echo   查看日志: docker-compose logs -f [service_name]
echo   进入容器: docker-compose exec [service_name] sh
echo   停止服务: docker-compose down
echo   重启服务: docker-compose restart [service_name]
echo.
goto :eof

REM 主要部署函数
:deploy
set "deploy_env=%~1"
set "deploy_detach=%~2"
set "deploy_build=%~3"

if "%deploy_env%"=="" set "deploy_env=dev"
if "%deploy_detach%"=="" set "deploy_detach=false"
if "%deploy_build%"=="" set "deploy_build=false"

call :log_info "开始部署图床转换系统 (环境: !deploy_env!)"

REM 检查必要的命令
call :check_command docker
if errorlevel 1 exit /b 1

call :check_command docker-compose
if errorlevel 1 exit /b 1

REM 检查环境变量文件
call :check_env_file
if errorlevel 1 exit /b 1

REM 创建必要的目录
if not exist "logs" mkdir logs
if not exist "backups" mkdir backups
if not exist "uploads" mkdir uploads
if not exist "temp" mkdir temp

REM 构建参数
set "compose_args="
if "%deploy_build%"=="true" (
    set "compose_args=--build"
)

if "%deploy_detach%"=="true" (
    set "compose_args=!compose_args! -d"
)

REM 选择 docker-compose 文件
set "compose_file=docker-compose.yml"
if "%deploy_env%"=="prod" (
    if exist "docker-compose.prod.yml" (
        set "compose_file=docker-compose.prod.yml"
    )
)

call :log_info "使用配置文件: !compose_file!"

REM 启动服务
call :log_info "启动服务..."
docker-compose -f !compose_file! up !compose_args!

if "%deploy_detach%"=="true" (
    REM 检查服务健康状态
    call :check_health
    
    REM 显示服务信息
    call :show_services
    
    call :log_success "部署完成！系统正在后台运行"
) else (
    call :log_info "按 Ctrl+C 停止服务"
)
goto :eof

REM 解析命令行参数
:parse_args
if "%~1"=="" goto :end_parse

if "%~1"=="-h" goto :help
if "%~1"=="--help" goto :help
if "%~1"=="-e" goto :set_env
if "%~1"=="--env" goto :set_env
if "%~1"=="-d" goto :set_detach
if "%~1"=="--detach" goto :set_detach
if "%~1"=="-b" goto :set_build
if "%~1"=="--build" goto :set_build
if "%~1"=="-c" goto :set_clean
if "%~1"=="--clean" goto :set_clean
if "%~1"=="-s" goto :set_stop
if "%~1"=="--stop" goto :set_stop
if "%~1"=="-r" goto :set_restart
if "%~1"=="--restart" goto :set_restart
if "%~1"=="-l" goto :set_logs
if "%~1"=="--logs" goto :set_logs
if "%~1"=="--backup" goto :set_backup
if "%~1"=="--restore" goto :set_restore

call :log_error "未知选项: %~1"
call :show_help
exit /b 1

:help
call :show_help
exit /b 0

:set_env
set "ENV=%~2"
shift
shift
goto :parse_args

:set_detach
set "DETACH=true"
shift
goto :parse_args

:set_build
set "BUILD=true"
shift
goto :parse_args

:set_clean
set "CLEAN=true"
shift
goto :parse_args

:set_stop
set "STOP=true"
shift
goto :parse_args

:set_restart
set "RESTART=true"
shift
goto :parse_args

:set_logs
set "LOGS=true"
shift
goto :parse_args

:set_backup
set "BACKUP=true"
shift
goto :parse_args

:set_restore
set "RESTORE_FILE=%~2"
shift
shift
goto :parse_args

:end_parse

REM 解析命令行参数
call :parse_args %*

REM 执行相应的操作
if "%CLEAN%"=="true" (
    call :cleanup
) else if "%STOP%"=="true" (
    call :log_info "停止所有服务..."
    docker-compose down
    call :log_success "所有服务已停止"
) else if "%RESTART%"=="true" (
    call :log_info "重启所有服务..."
    docker-compose restart
    call :log_success "所有服务已重启"
) else if "%LOGS%"=="true" (
    docker-compose logs -f
) else if "%BACKUP%"=="true" (
    call :backup_database
) else if not "%RESTORE_FILE%"=="" (
    call :restore_database "%RESTORE_FILE%"
) else (
    call :deploy "%ENV%" "%DETACH%" "%BUILD%"
)

endlocal