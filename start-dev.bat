@echo off
chcp 65001 >nul
title PixivTailor Dev Mode (Fixed)

echo.
echo ========================================
echo    PixivTailor Development Mode (Fixed)
echo ========================================
echo.

echo [1/6] Stopping existing services...
call stop.bat
if %errorlevel% neq 0 (
    echo WARNING: Some services may not have stopped properly
)

echo.
echo [2/6] Environment Check...
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Go not found
    pause
    exit /b 1
)

node --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Node.js not found
    pause
    exit /b 1
)
echo Environment: OK

echo.

echo [3/6] Directory Setup...
if not exist "data" mkdir data
if not exist "logs" mkdir logs
if not exist "frontend\dist" mkdir frontend\dist
echo Directories: OK

echo.

echo [4/6] Dependencies...
cd backend
go mod tidy
if %errorlevel% neq 0 (
    echo ERROR: Backend dependencies failed
    pause
    exit /b 1
)
cd ..

cd frontend
if not exist "node_modules" (
    npm install
    if %errorlevel% neq 0 (
        echo ERROR: Frontend dependencies failed
        pause
        exit /b 1
    )
)
cd ..

echo Dependencies: OK

echo.

echo [5/6] Port Detection...

:: 检测前端可用端口
set FRONTEND_PORT=3000
:check_frontend_port
netstat -an | findstr :%FRONTEND_PORT% >nul
if %errorlevel% equ 0 (
    set /a FRONTEND_PORT+=1
    if %FRONTEND_PORT% LSS 3010 (
        goto :check_frontend_port
    ) else (
        echo ERROR: No available frontend port found
        pause
        exit /b 1
    )
)
echo Frontend port: %FRONTEND_PORT%

:: 检查后端固定端口
echo Checking backend fixed ports...
netstat -an | findstr :50051 >nul
if %errorlevel% equ 0 (
    echo ⚠️  Port 50051 occupied, killing process...
    for /f "tokens=5" %%a in ('netstat -aon ^| findstr :50051') do (
        taskkill /f /pid %%a >nul 2>&1
    )
)

netstat -an | findstr :50052 >nul
if %errorlevel% equ 0 (
    echo ⚠️  Port 50052 occupied, killing process...
    for /f "tokens=5" %%a in ('netstat -aon ^| findstr :50052') do (
        taskkill /f /pid %%a >nul 2>&1
    )
)

echo Backend will use fixed ports: 50051 (gRPC), 50052 (HTTP)

echo.

echo [6/6] Starting Dev Services...

goto :start_same_window


:start_same_window
echo Starting backend server with fixed ports...
start /b cmd /c "cd /d %~dp0backend && go run tailor.go --verbose server"

echo Waiting for backend to start...
timeout /t 5 >nul

:: 验证后端是否成功启动 (检测固定端口)
echo Checking backend status...
set BACKEND_FOUND=0

netstat -an | findstr :50051 >nul
if %errorlevel% equ 0 (
    echo gRPC server successfully started on port 50051
    set BACKEND_FOUND=1
)

netstat -an | findstr :50052 >nul
if %errorlevel% equ 0 (
    echo HTTP server successfully started on port 50052
    set BACKEND_FOUND=1
)

if %BACKEND_FOUND% equ 0 (
    echo WARNING: Backend may not have started properly
)

echo Starting frontend dev server (port %FRONTEND_PORT%) in background...
start /b cmd /c "cd /d %~dp0frontend && npm run dev -- --port %FRONTEND_PORT%"

echo Waiting for frontend...
timeout /t 8 >nul

:: 验证前端是否成功启动
echo Checking frontend status...
netstat -an | findstr :%FRONTEND_PORT% >nul
if %errorlevel% neq 0 (
    echo WARNING: Frontend may not have started properly on port %FRONTEND_PORT%
) else (
    echo Frontend successfully started on port %FRONTEND_PORT%
)

echo Opening browser...
start http://localhost:%FRONTEND_PORT%
goto :end


:end

echo.
echo ========================================
echo    Dev Mode Started!
echo ========================================
echo.
echo Frontend: http://localhost:%FRONTEND_PORT%
echo gRPC:     localhost:50051
echo HTTP:     http://localhost:50052
echo.
echo TIP: 
echo   - Frontend supports hot reload
echo   - Backend supports hot reload with air
echo   - Press Ctrl+C to stop services
echo   - Close windows to stop services
echo.
pause >nul
