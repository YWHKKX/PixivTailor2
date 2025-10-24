@echo off
REM 简单的测试运行脚本
echo 🧪 运行 PixivTailor 测试套件

cd src

echo 📋 运行冒烟测试...
go test -v ./tests -run TestSmokeTests

echo 🔧 运行基本功能测试...
go test -v ./tests -run TestBasicFunctionality

echo ⚡ 运行快速启动测试...
go test -v ./tests -run TestQuickStart

echo ✅ 测试完成！
pause
