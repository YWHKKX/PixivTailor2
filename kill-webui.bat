@echo off
chcp 65001 >nul
echo.
echo ========================================
echo      强制关闭 WebUI 进程
echo ========================================
echo.

echo [1/4] 正在关闭所有 Python 进程...
taskkill /F /IM python.exe 2>nul
if %errorlevel% equ 0 (
    echo ✅ Python 进程已关闭
) else (
    echo ℹ️  没有发现运行中的 Python 进程
)

echo.
echo [2/4] 正在检查端口 7860...
netstat -ano | findstr :7860 >nul
if %errorlevel% equ 0 (
    echo ⚠️  端口 7860 仍被占用，正在强制释放...
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr :7860') do (
        taskkill /F /PID %%a 2>nul
    )
    echo ✅ 端口 7860 已释放
) else (
    echo ✅ 端口 7860 未被占用
)

echo.
echo [3/4] 正在关闭所有 WebUI 相关批处理...
taskkill /F /IM webui.bat 2>nul
taskkill /F /IM webui-user.bat 2>nul
taskkill /F /IM start-webui-api.bat 2>nul

echo.
echo [4/4] 正在关闭所有 cmd 进程（可能包含 WebUI）...
for /f "tokens=2" %%a in ('tasklist /fi "imagename eq cmd.exe" /fo csv ^| findstr /v "PID"') do (
    taskkill /F /PID %%a 2>nul
)

echo.
echo ========================================
echo      WebUI 进程清理完成
echo ========================================
echo.
echo 现在可以安全地重新启动主程序了
echo.
