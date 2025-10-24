@echo off
chcp 65001 >nul
title PixivTailor Stop Services

echo.
echo ========================================
echo    PixivTailor Stop Services
echo ========================================
echo.

echo [1/2] Stopping Docker services...
docker-compose down
if %errorlevel% neq 0 (
    echo WARNING: Docker services not running
) else (
    echo Docker services: STOPPED
)

echo.

echo [2/2] Stopping local processes...

echo Stopping backend processes...
REM 停止gRPC服务器进程 (固定端口50051)
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :50051') do (
    taskkill /f /pid %%a >nul 2>&1
    if !errorlevel! equ 0 (
        echo   gRPC Backend process stopped (PID: %%a, Port: 50051)
    )
)

REM 停止HTTP服务器进程 (固定端口50052)
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :50052') do (
    taskkill /f /pid %%a >nul 2>&1
    if !errorlevel! equ 0 (
        echo   HTTP Backend process stopped (PID: %%a, Port: 50052)
    )
)

echo Stopping tailor.exe processes...
taskkill /f /im tailor.exe >nul 2>&1
if %errorlevel% equ 0 (
    echo   tailor.exe processes stopped
)

echo Stopping frontend processes...
for /L %%i in (3000,1,3010) do (
    for /f "tokens=5" %%a in ('netstat -aon ^| findstr :%%i') do (
        taskkill /f /pid %%a >nul 2>&1
        if !errorlevel! equ 0 (
            echo   Frontend process stopped (PID: %%a, Port: %%i)
        )
    )
)

echo Stopping Node.js processes...
taskkill /f /im node.exe >nul 2>&1
if %errorlevel% equ 0 (
    echo   Node.js processes stopped
)

echo.
echo ========================================
echo    All Services Stopped
echo ========================================
echo.