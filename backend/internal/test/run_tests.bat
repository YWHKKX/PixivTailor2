@echo off
REM PixivTailor 测试运行脚本 (Windows)
REM 用于快速验证各个模块的功能是否正常

echo 🚀 开始运行 PixivTailor 测试套件
echo ==================================

REM 进入src目录
cd src

REM 检查Go环境
echo 📋 检查Go环境...
go version

REM 运行测试
echo.
echo 🧪 运行冒烟测试...
go test -v ./tests -run TestSmokeTests

echo.
echo 🔧 运行基本功能测试...
go test -v ./tests -run TestBasicFunctionality

echo.
echo ⚡ 运行快速启动测试...
go test -v ./tests -run TestQuickStart

echo.
echo 🔍 运行配置系统测试...
go test -v ./tests -run TestConfigSystem

echo.
echo 📝 运行日志系统测试...
go test -v ./tests -run TestLoggerSystem

echo.
echo 📊 运行数据模型测试...
go test -v ./tests -run TestDataModels

echo.
echo 🔗 运行集成测试...
go test -v ./tests -run TestIntegration

echo.
echo ❌ 运行错误场景测试...
go test -v ./tests -run TestErrorScenarios

echo.
echo ⚡ 运行性能测试...
go test -v ./tests -run TestPerformance

echo.
echo 🔄 运行并发测试...
go test -v ./tests -run TestConcurrency

echo.
echo 📈 运行基准测试...
go test -v ./tests -bench=.

echo.
echo ✅ 所有测试完成！
echo ==================================

REM 显示测试覆盖率
echo.
echo 📊 生成测试覆盖率报告...
go test -v ./tests -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

echo 📄 覆盖率报告已生成: coverage.html
echo 🎉 测试套件运行完成！

pause
