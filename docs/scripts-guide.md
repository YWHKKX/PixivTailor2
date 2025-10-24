# PixivTailor 脚本使用指南

本文档详细介绍了PixivTailor项目中的所有启动和管理脚本，帮助用户快速理解和使用各种启动方式。

## 📋 脚本概览

| 脚本名称           | 功能             | 适用场景         | 端口       |
| ------------------ | ---------------- | ---------------- | ---------- |
| `start.bat`        | 统一启动器       | 开发/生产环境    | 3000/50051 |
| `start-dev.bat`    | 开发模式启动     | 本地开发调试     | 3000/50051 |
| `start-prod.bat`   | 生产模式启动     | 生产环境部署     | 3000/50051 |
| `start-server.bat` | 仅启动后端服务器 | 仅运行后端服务   | 50051      |
| `status.bat`       | 服务状态检查     | 监控服务运行状态 | -          |
| `stop.bat`         | 停止所有服务     | 清理运行的服务   | -          |
| `build.sh`         | 构建脚本         | Linux/Mac构建    | -          |

## 🚀 启动脚本详解

### 1. start.bat - 统一启动器

**功能描述**: 最全面的启动脚本，支持Docker和本地两种模式

**特点**:
- 🔄 自动检测Docker环境
- 🛠️ 完整的依赖检查（Go、Node.js、Docker）
- 📦 自动安装依赖和生成gRPC代码
- 🌐 支持Docker Compose和本地开发模式
- 🔧 自动打开浏览器

**启动流程**:
1. 环境检查（Go、Node.js、Docker）
2. 目录结构创建
3. gRPC代码生成
4. 依赖安装（后端+前端）
5. 服务启动（Docker优先，失败则本地模式）

**使用场景**:
- 首次部署
- 不确定使用哪种模式
- 需要完整的服务栈

### 2. start-dev.bat - 开发模式启动

**功能描述**: 专门用于本地开发的轻量级启动脚本

**特点**:
- 🏃‍♂️ 快速启动（跳过Docker检查）
- 🔧 仅支持本地开发模式
- 📦 自动安装依赖
- 🌐 自动打开浏览器

**启动流程**:
1. 环境检查（Go、Node.js）
2. 目录结构创建
3. 依赖安装
4. 启动后端和前端服务

**使用场景**:
- 日常开发调试
- 快速启动开发环境
- 不需要Docker环境

**⚠️ 注意事项**:
- 脚本中使用了正确的命令 `tailor.exe server`
- 端口配置为3000，但实际可能运行在其他端口

### 3. start-prod.bat - 生产模式启动

**功能描述**: 专门用于生产环境的Docker部署脚本

**特点**:
- 🐳 强制使用Docker模式
- 🔄 自动构建和重启服务
- 📊 提供监控和管理命令
- 🛡️ 生产级别的错误处理

**启动流程**:
1. Docker环境检查
2. 停止现有服务
3. 构建Docker镜像
4. 启动生产服务
5. 服务状态检查

**使用场景**:
- 生产环境部署
- 正式发布版本
- 需要容器化部署

### 4. start-server.bat - 仅启动后端服务器

**功能描述**: 只启动gRPC后端服务器，不启动前端

**特点**:
- 🎯 专注于后端服务
- 🔧 自动编译CLI工具
- 📦 自动生成gRPC代码
- 🛠️ 适合API开发和测试

**启动流程**:
1. Go环境检查
2. 编译CLI工具
3. 生成gRPC代码
4. 启动后端服务器

**使用场景**:
- API开发和测试
- 后端服务调试
- 不需要前端界面

## 🔧 管理脚本详解

### 5. status.bat - 服务状态检查

**功能描述**: 检查所有服务的运行状态

**检查内容**:
- 端口占用情况（50051、3000）
- 进程状态（Go、Node.js）
- 服务URL信息

**使用场景**:
- 故障排查
- 服务监控
- 状态确认

### 6. stop.bat - 停止所有服务

**功能描述**: 停止所有运行的服务和进程

**停止内容**:
- Docker服务（docker-compose down）
- 后端进程（端口50051）
- 前端进程（端口3000）
- Node.js进程

**使用场景**:
- 清理运行环境
- 重启服务前清理
- 故障恢复

### 7. build.sh - Linux/Mac构建脚本

**功能描述**: 用于Linux和Mac系统的构建脚本

**特点**:
- 🐧 跨平台支持
- 🔧 Go版本检查
- 📦 依赖管理
- 📖 使用帮助

## 📊 脚本对比分析

### start.bat vs start-dev.bat

| 特性         | start.bat | start-dev.bat |
| ------------ | --------- | ------------- |
| Docker支持   | ✅ 是      | ❌ 否          |
| 环境检查     | 完整      | 基础          |
| gRPC代码生成 | ✅ 是      | ❌ 否          |
| 启动速度     | 较慢      | 较快          |
| 适用场景     | 部署/开发 | 开发调试      |

### 端口配置问题

**发现的问题**:
- 所有脚本都假设前端运行在3000端口
- 实际运行时可能因为端口占用而使用其他端口
- 需要动态检测实际运行端口

**建议修复**:
```batch
:: 检查可用端口
for /L %%i in (3000,1,3010) do (
    netstat -an | findstr :%%i >nul
    if errorlevel 1 (
        set FRONTEND_PORT=%%i
        goto :port_found
    )
)
```

## 🛠️ 脚本修复建议

### 1. start-dev.bat 修复

**问题**: 使用了错误的启动命令
```batch
# 错误的命令
start "PixivTailor Backend" cmd /k "tailor.exe server --port :50051"

# 正确的命令
start "PixivTailor Backend" cmd /k "tailor.exe --verbose server --port :50051"
```

### 2. 端口动态检测

**建议**: 添加端口检测逻辑
```batch
:: 检测可用端口
:find_port
for /L %%i in (3000,1,3010) do (
    netstat -an | findstr :%%i >nul
    if errorlevel 1 (
        set FRONTEND_PORT=%%i
        goto :port_found
    )
)
echo 错误: 未找到可用端口
pause
exit /b 1

:port_found
echo 使用端口: %FRONTEND_PORT%
```

### 3. 错误处理改进

**建议**: 添加更详细的错误处理
```batch
:: 检查服务启动状态
timeout /t 5 >nul
netstat -an | findstr :50051 >nul
if errorlevel 1 (
    echo 错误: 后端服务启动失败
    pause
    exit /b 1
)
```

## 🎯 推荐使用方式

### 开发环境
```bash
# 推荐使用 start-dev.bat（修复后）
.\start-dev.bat
```

### 生产环境
```bash
# 推荐使用 start-prod.bat
.\start-prod.bat
```

### 仅后端开发
```bash
# 推荐使用 start-server.bat
.\start-server.bat
```

### 服务管理
```bash
# 检查状态
.\status.bat

# 停止服务
.\stop.bat
```

## 📝 总结

PixivTailor项目提供了完整的脚本生态系统，支持从开发到生产的各种场景。主要脚本包括：

- **start.bat**: 最全面的启动器，适合首次部署
- **start-dev.bat**: 开发模式，适合日常开发
- **start-prod.bat**: 生产模式，适合正式部署
- **start-server.bat**: 仅后端，适合API开发
- **status.bat**: 状态检查，适合故障排查
- **stop.bat**: 服务停止，适合环境清理

建议根据具体使用场景选择合适的脚本，并注意修复已知的端口和命令问题。
