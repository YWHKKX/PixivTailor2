@echo off
chcp 65001 >nul
title PixivTailor 生产模式

echo.
echo ========================================
echo    PixivTailor 生产模式启动
echo ========================================
echo.

:: 检查 Docker
echo [1/5] 检查 Docker 环境...
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ 错误: 未找到 Docker 环境，请先安装 Docker
    pause
    exit /b 1
)
echo ✅ Docker 环境检查通过

:: 检查 Docker Compose
docker-compose --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ 错误: 未找到 Docker Compose，请先安装 Docker Compose
    pause
    exit /b 1
)
echo ✅ Docker Compose 环境检查通过

:: 停止现有服务
echo [2/5] 停止现有服务...
docker-compose down
echo ✅ 现有服务已停止

:: 构建镜像
echo [3/5] 构建 Docker 镜像...
docker-compose build --no-cache
if %errorlevel% neq 0 (
    echo ❌ Docker 镜像构建失败
    pause
    exit /b 1
)
echo ✅ Docker 镜像构建完成

:: 启动服务
echo [4/5] 启动生产服务...
docker-compose up -d
if %errorlevel% neq 0 (
    echo ❌ 服务启动失败
    pause
    exit /b 1
)
echo ✅ 服务启动完成

:: 等待服务就绪
echo [5/5] 等待服务就绪...
timeout /t 10 >nul

:: 检查服务状态
echo 检查服务状态...
docker-compose ps

:: 打开浏览器
echo 打开前端页面...
start http://localhost:3000

echo.
echo ========================================
echo    生产模式启动完成！
echo ========================================
echo.
echo 🌐 前端地址: http://localhost:3000
echo 🔧 后端地址: localhost:50051
echo 📊 监控地址: http://localhost:3001
echo 📈 指标地址: http://localhost:9090
echo.
echo 💡 管理命令:
echo    - 查看日志: docker-compose logs -f
echo    - 停止服务: docker-compose down
echo    - 重启服务: docker-compose restart
echo.
echo 按任意键关闭此窗口...
pause >nul
