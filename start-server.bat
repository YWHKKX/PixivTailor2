@echo off
chcp 65001 >nul
title PixivTailor gRPC 服务器

echo.
echo ========================================
echo    PixivTailor gRPC 服务器启动
echo ========================================
echo.

:: 检查 Go 环境
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ 错误: 未找到 Go 环境
    pause
    exit /b 1
)

:: 进入后端目录
cd /d "%~dp0backend"

:: 检查 tailor.go 是否存在
if not exist "tailor.go" (
    echo ❌ 错误: 未找到 tailor.go 文件
    pause
    exit /b 1
)
echo ✅ tailor.go 文件存在

:: 生成 gRPC 代码
echo [2/3] 生成 gRPC 代码...
cd ..\proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pixiv_tailor.proto
if %errorlevel% neq 0 (
    echo ❌ gRPC 代码生成失败
    pause
    exit /b 1
)
echo ✅ gRPC 代码生成完成

:: 复制 proto 文件到 backend
copy pixiv_tailor.pb.go ..\backend\proto\ >nul
copy pixiv_tailor_grpc.pb.go ..\backend\proto\ >nul
echo ✅ Proto 文件已更新

:: 检查端口占用并自动处理
echo [3/4] 检查端口占用...
netstat -an | findstr :50051 >nul
if %errorlevel% equ 0 (
    echo ⚠️  发现端口 50051 被占用，正在自动处理...
    for /f "tokens=5" %%a in ('netstat -aon ^| findstr :50051') do (
        echo 终止占用进程 PID: %%a
        taskkill /f /pid %%a >nul 2>&1
        if !errorlevel! equ 0 (
            echo ✅ 成功释放端口 50051
        ) else (
            echo ❌ 无法释放端口 50051
            pause
            exit /b 1
        )
    )
) else (
    echo ✅ 端口 50051 可用
)

netstat -an | findstr :50052 >nul
if %errorlevel% equ 0 (
    echo ⚠️  发现端口 50052 被占用，正在自动处理...
    for /f "tokens=5" %%a in ('netstat -aon ^| findstr :50052') do (
        echo 终止占用进程 PID: %%a
        taskkill /f /pid %%a >nul 2>&1
        if !errorlevel! equ 0 (
            echo ✅ 成功释放端口 50052
        ) else (
            echo ❌ 无法释放端口 50052
            pause
            exit /b 1
        )
    )
) else (
    echo ✅ 端口 50052 可用
)

:: 启动服务
echo [4/4] 启动服务器...
cd ..\backend
echo 启动服务器 (gRPC: 50051, HTTP: 50052)...
echo 按 Ctrl+C 停止服务器
echo.
go run tailor.go --verbose server

pause
