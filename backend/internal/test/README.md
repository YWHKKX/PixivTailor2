# PixivTailor 测试套件

## 📋 概述

本测试套件为PixivTailor项目提供了全面的功能验证，包括单元测试、集成测试、性能测试和冒烟测试。

## 🧪 测试文件结构

```
src/tests/
├── integration_test.go    # 集成测试
├── config_test.go        # 配置系统测试
├── logger_test.go        # 日志系统测试
├── models_test.go        # 数据模型测试
├── quick_test.go         # 快速启动测试
├── smoke_test.go         # 冒烟测试
├── test_config.json      # 测试配置文件
├── run_tests.sh          # Linux/Mac测试脚本
└── run_tests.bat         # Windows测试脚本
```

## 🚀 快速开始

### 运行所有测试

**Linux/Mac:**
```bash
cd src/tests
chmod +x run_tests.sh
./run_tests.sh
```

**Windows:**
```cmd
cd src\tests
run_tests.bat
```

### 运行特定测试

```bash
# 运行冒烟测试
go test -v ./tests -run TestSmokeTests

# 运行配置系统测试
go test -v ./tests -run TestConfigSystem

# 运行日志系统测试
go test -v ./tests -run TestLoggerSystem

# 运行集成测试
go test -v ./tests -run TestIntegration
```

## 📊 测试类型

### 1. 冒烟测试 (Smoke Tests)
- **文件**: `smoke_test.go`
- **目的**: 快速验证基本功能是否正常
- **包含**:
  - 配置系统基本初始化
  - 日志系统基本功能
  - 错误处理基本功能
  - 模块配置基本验证

### 2. 快速启动测试 (Quick Start Tests)
- **文件**: `quick_test.go`
- **目的**: 验证快速启动流程
- **包含**:
  - 基本初始化流程
  - 配置验证
  - 错误处理
  - 模块集成

### 3. 配置系统测试 (Config Tests)
- **文件**: `config_test.go`
- **目的**: 验证配置管理功能
- **包含**:
  - 配置管理器测试
  - 模块配置测试
  - 配置工厂测试
  - 错误处理测试
  - 性能测试

### 4. 日志系统测试 (Logger Tests)
- **文件**: `logger_test.go`
- **目的**: 验证日志功能
- **包含**:
  - 日志初始化
  - 日志级别设置
  - 操作日志
  - 进度日志
  - 结构化日志

### 5. 数据模型测试 (Models Tests)
- **文件**: `models_test.go`
- **目的**: 验证数据模型
- **包含**:
  - PixivImage模型
  - GenerateRequest模型
  - TrainRequest模型
  - TagRequest模型
  - ClassifyRequest模型
  - 常量定义

### 6. 集成测试 (Integration Tests)
- **文件**: `integration_test.go`
- **目的**: 验证模块间集成
- **包含**:
  - 配置系统集成
  - 日志系统集成
  - AI模块集成
  - 爬虫模块集成
  - 完整工作流程

## 🔧 测试配置

测试使用专门的配置文件 `test_config.json`，包含：

- **调试模式**: 启用详细日志和调试信息
- **性能优化**: 降低资源使用，适合测试环境
- **快速响应**: 减少超时时间和重试次数
- **测试数据**: 使用测试专用的模型和配置

## 📈 性能测试

### 基准测试
```bash
# 运行所有基准测试
go test -v ./tests -bench=.

# 运行特定基准测试
go test -v ./tests -bench=BenchmarkConfigAccess
go test -v ./tests -bench=BenchmarkLogger
go test -v ./tests -bench=BenchmarkErrorCreation
```

### 性能指标
- **配置访问**: 每次访问 < 1ms
- **日志输出**: 每次输出 < 0.1ms
- **错误创建**: 每次创建 < 0.01ms
- **内存使用**: 测试期间 < 100MB

## 🔄 并发测试

测试套件包含并发测试，验证：
- 多协程同时访问配置
- 多协程同时输出日志
- 配置系统的线程安全性
- 日志系统的并发安全性

## ❌ 错误场景测试

测试各种错误情况：
- 无效配置文件
- 缺失配置文件
- 配置键不存在
- 模块配置错误
- 网络错误
- 文件系统错误

## 📊 测试覆盖率

运行测试后会自动生成覆盖率报告：

```bash
# 生成覆盖率报告
go test -v ./tests -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

目标覆盖率：
- **整体覆盖率**: > 80%
- **核心模块覆盖率**: > 90%
- **配置系统覆盖率**: > 95%
- **日志系统覆盖率**: > 90%

## 🐛 调试测试

### 启用详细输出
```bash
go test -v ./tests -run TestSmokeTests
```

### 运行单个测试
```bash
go test -v ./tests -run TestConfigSystem
```

### 调试特定测试
```bash
go test -v ./tests -run TestConfigSystem -args -test.v
```

## 📝 测试最佳实践

### 1. 测试命名
- 使用描述性的测试名称
- 包含测试类型和功能描述
- 使用驼峰命名法

### 2. 测试结构
- 使用 `t.Run()` 组织子测试
- 每个测试函数只测试一个功能
- 使用表格驱动测试处理多个测试用例

### 3. 测试数据
- 使用临时目录存储测试文件
- 清理测试产生的临时文件
- 使用测试专用的配置和数据

### 4. 错误处理
- 测试正常情况和错误情况
- 验证错误类型和消息
- 测试错误恢复机制

## 🔍 故障排除

### 常见问题

1. **测试失败**: 检查Go环境和依赖
2. **配置错误**: 验证测试配置文件
3. **权限问题**: 确保有写入临时目录的权限
4. **网络问题**: 某些测试可能需要网络连接

### 调试步骤

1. 运行单个测试函数
2. 检查测试日志输出
3. 验证测试环境配置
4. 检查依赖项是否正确安装

## 📚 相关文档

- [配置系统文档](../docs/unified-config-system.md)
- [日志模块文档](../docs/logger-module.md)
- [数据模型文档](../docs/models-module.md)
- [错误处理文档](../docs/errors-module.md)

## 🤝 贡献

添加新测试时请遵循：
1. 使用现有的测试结构
2. 添加适当的注释和文档
3. 确保测试的可重复性
4. 更新相关文档
