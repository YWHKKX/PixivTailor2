@echo off
chcp 65001 >nul
title PixivTailor Status

echo.
echo ========================================
echo    PixivTailor Service Status
echo ========================================
echo.

echo [1/3] Port Status
echo.

netstat -an | findstr :50051 >nul
if %errorlevel% equ 0 (
    echo gRPC:      RUNNING (port 50051)
    for /f "tokens=5" %%a in ('netstat -aon ^| findstr :50051') do (
        echo   PID: %%a
    )
) else (
    echo gRPC:      NOT RUNNING (port 50051)
)

netstat -an | findstr :50052 >nul
if %errorlevel% equ 0 (
    echo HTTP:      RUNNING (port 50052)
    for /f "tokens=5" %%a in ('netstat -aon ^| findstr :50052') do (
        echo   PID: %%a
    )
) else (
    echo HTTP:      NOT RUNNING (port 50052)
)

netstat -an | findstr :3000 >nul
if %errorlevel% equ 0 (
    echo Frontend:  RUNNING (port 3000)
    for /f "tokens=5" %%a in ('netstat -aon ^| findstr :3000') do (
        echo   PID: %%a
    )
) else (
    echo Frontend:  NOT RUNNING (port 3000)
)

echo.

echo [2/3] Process Status
echo.

tasklist | findstr tailor.exe >nul
if %errorlevel% equ 0 (
    echo tailor.exe: RUNNING
) else (
    echo tailor.exe: NOT RUNNING
)

tasklist | findstr go.exe >nul
if %errorlevel% equ 0 (
    echo Go processes: RUNNING
) else (
    echo Go processes: NOT RUNNING
)

tasklist | findstr node.exe >nul
if %errorlevel% equ 0 (
    echo Node.js processes: RUNNING
) else (
    echo Node.js processes: NOT RUNNING
)

echo.

echo [3/3] Service URLs
echo.
echo Frontend: http://localhost:3000
echo gRPC:     localhost:50051
echo HTTP:     http://localhost:50052
echo.

echo Press any key to close...
pause >nul
