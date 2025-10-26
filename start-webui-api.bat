@echo off
chcp 65001
title 启动 Stable Diffusion WebUI API

echo.
echo ========================================
echo      启动 Stable Diffusion WebUI API
echo ========================================
echo.

echo [1/3] 强制终止所有WebUI进程...
echo 正在终止Python进程...
taskkill /f /im python.exe
echo 正在终止WebUI相关进程...
taskkill /f /im webui.bat
taskkill /f /im webui-user.bat
echo 等待进程完全终止...
timeout /t 5
echo 检查端口状态...
netstat -ano | findstr ":7860"
if %errorlevel% equ 0 (
    echo 端口7860仍被占用，强制释放...
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":7860"') do (
        taskkill /f /pid %%a
    )
    timeout /t 3
)
echo ✅ 进程清理完成

echo [2/3] 切换到WebUI目录...
cd /d "D:\PythonProject\stable-diffusion-webui"
if not exist "webui.bat" (
    echo ❌ 错误: 找不到webui.bat文件
    echo 请检查路径: D:\PythonProject\stable-diffusion-webui
    pause
    exit /b 1
)

echo [3/3] 启动WebUI (API模式)...
echo.
echo 重要提示:
echo 1. 请等待看到 "Running on local URL: http://127.0.0.1:7860"
echo 2. 然后按任意键继续
echo 3. 保持此窗口打开
echo.

echo 执行命令: webui.bat --api --listen --port 7860 --skip-python-version-check
echo.

webui.bat --api --listen --port 7860 --skip-python-version-check

echo.
echo WebUI已停止
pause