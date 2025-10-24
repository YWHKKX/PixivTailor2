# PixivTailor - 现代化Pixiv爬虫与AI图像处理平台

PixivTailor 是一个基于Go语言开发的现代化Pixiv爬虫与AI图像处理平台，集成了Web界面、实时监控、任务管理和智能缓存系统。该项目已完成核心功能开发，提供了完整的Pixiv内容爬取、文件管理、实时日志和任务调度功能。

## 🚀 核心功能 (已完成)

### 1. 现代化Web界面 ✅
- **React + TypeScript**: 现代化前端技术栈
- **Ant Design**: 美观的UI组件库
- **实时更新**: WebSocket实时通信
- **响应式设计**: 支持多种屏幕尺寸
- **文件树视图**: 直观的文件管理系统

### 2. Pixiv爬虫系统 ✅
- **多类型爬取**: 支持标签、用户ID、插画ID爬取
- **智能配置**: 灵活的Cookie和代理配置
- **任务管理**: 完整的任务生命周期管理
- **实时监控**: 实时进度和日志显示
- **错误处理**: 完善的错误重试和状态管理

### 3. 文件管理系统 ✅
- **任务隔离**: 按任务ID组织文件结构
- **智能缓存**: 任务特定的缓存机制
- **图片预览**: 内置图片预览功能
- **文件树**: 可视化的文件浏览界面
- **路径管理**: 统一的路径管理服务

### 4. 实时监控系统 ✅
- **WebSocket通信**: 实时双向通信
- **任务状态**: 实时任务状态更新
- **进度跟踪**: 详细的进度信息显示
- **日志系统**: 分级日志和实时日志流
- **错误报告**: 详细的错误信息和重试机制

### 5. 配置管理系统 ✅
- **多环境配置**: 支持开发和生产环境
- **动态配置**: 运行时配置更新
- **Cookie管理**: 智能Cookie配置和默认值
- **代理支持**: 灵活的代理配置选项
- **路径配置**: 统一的路径管理

## 🏗️ 项目架构 (当前状态)

```
PixivTailorEx/
├── backend/                    # Go后端服务 ✅
│   ├── internal/              # 内部包
│   │   ├── http/              # HTTP服务器和API路由
│   │   ├── service/           # 业务服务层
│   │   ├── repository/        # 数据访问层
│   │   ├── crawler/           # 爬虫核心功能
│   │   └── middleware/        # 中间件
│   ├── pkg/                   # 公共包
│   │   ├── paths/             # 路径管理服务
│   │   ├── models/            # 数据模型
│   │   └── api/               # API客户端
│   ├── configs/               # 配置文件
│   │   ├── crawler_config.json # 爬虫配置
│   │   └── default.json       # 默认配置
│   ├── data/                  # 数据目录
│   │   ├── images/            # 图片存储 (按任务ID组织)
│   │   ├── cache/             # 缓存目录 (按任务ID组织)
│   │   └── database/          # SQLite数据库
│   ├── tailor.go              # 主程序入口
│   └── tailor.exe             # 编译后的可执行文件
├── frontend/                  # React前端应用 ✅
│   ├── src/                   # 源代码
│   │   ├── pages/             # 页面组件
│   │   ├── services/          # API服务
│   │   ├── components/        # 通用组件
│   │   └── utils/             # 工具函数
│   ├── public/                # 静态资源
│   ├── dist/                  # 构建产物
│   └── package.json           # 依赖管理
├── docs/                      # 项目文档 ✅
│   ├── README.md              # 主文档
│   ├── cli-usage.md           # CLI使用说明
│   ├── deployment-guide.md    # 部署指南
│   └── crawler-module.md      # 爬虫模块文档
├── scripts/                   # 启动脚本 ✅
│   ├── start-dev.bat          # 开发环境启动
│   ├── start-prod.bat         # 生产环境启动
│   └── status.bat             # 状态检查
└── README.md                  # 项目说明
```

## 🛠️ 技术栈 (当前实现)

### 后端技术栈 ✅
- **语言**: Go 1.21+
- **Web框架**: Gorilla Mux (HTTP路由)
- **数据库**: SQLite (轻量级数据库)
- **日志系统**: 自定义日志系统
- **并发处理**: Goroutines + Channels
- **HTTP客户端**: 自定义HTTP客户端 (支持代理)
- **WebSocket**: Gorilla WebSocket
- **配置管理**: JSON配置文件

### 前端技术栈 ✅
- **框架**: React 18 + TypeScript
- **UI库**: Ant Design 5.x
- **构建工具**: Vite
- **状态管理**: React Hooks
- **HTTP客户端**: Fetch API
- **WebSocket**: 原生WebSocket API
- **样式**: CSS-in-JS + Ant Design主题

### 外部服务集成 ✅
- **Pixiv API**: 爬取Pixiv内容
- **代理支持**: HTTP代理配置
- **文件系统**: 本地文件存储和管理

## 📋 环境要求 (当前版本)

### 必需环境 ✅
- **Go 1.21+** - 后端开发环境
- **Node.js 18+** - 前端开发环境
- **Windows 10/11** - 操作系统支持
- **现代浏览器** - Chrome/Firefox/Edge (支持WebSocket)

### 可选环境
- **代理服务器** - 用于访问Pixiv (如Clash for Windows)
- **Pixiv账号** - 用于获取Cookie进行身份验证

## 🚀 快速开始 (当前版本)

### 1. 环境准备 ✅
- 安装Go 1.21+运行环境
- 安装Node.js 18+运行环境
- 克隆项目到本地

### 2. 一键启动 ✅
- 使用提供的启动脚本快速启动所有服务
- 支持开发环境和生产环境启动
- 自动启动后端服务和前端应用

### 3. 访问系统 ✅
- **Web界面**: 通过浏览器访问管理界面
- **后端API**: 提供REST API接口服务
- **WebSocket**: 实时通信和状态更新

### 4. 基本使用 ✅
1. **配置Cookie**: 在爬虫页面配置Pixiv身份验证
2. **创建任务**: 选择爬取类型（标签/用户/插画）
3. **监控进度**: 实时查看任务进度和日志
4. **查看结果**: 在文件树中浏览下载的图片

## 📖 详细文档

- [CLI使用说明](cli-usage.md) - 命令行使用指南
- [部署指南](deployment-guide.md) - 完整部署教程
- [爬虫模块](crawler-module.md) - 爬虫功能详解
- [配置模块](config-module.md) - 配置参数详解
- [脚本指南](scripts-guide.md) - 启动脚本说明
- [AI模块](ai-module.md) - AI功能规划 (待实现)

## 🎯 项目状态

### ✅ 已完成功能
- **Web界面**: 完整的React前端应用
- **爬虫系统**: 支持多种Pixiv内容爬取
- **任务管理**: 完整的任务生命周期管理
- **文件管理**: 任务隔离的文件存储系统
- **实时监控**: WebSocket实时通信和状态更新
- **配置管理**: 灵活的配置和Cookie管理
- **代理支持**: HTTP代理配置和错误处理

### 🚧 待实现功能
- **AI图像生成**: Stable Diffusion WebUI集成
- **模型训练**: Kohya-ss训练框架集成
- **标签系统**: WD14Tagger自动标签生成
- **标签分类**: OpenAI API智能分类

## 🤝 贡献

欢迎提交Issue和Pull Request来改进项目。

## 📄 许可证

本项目采用MIT许可证。

## ⚠️ 注意事项

1. **请遵守Pixiv的使用条款和版权规定**
2. **合理使用爬虫功能，避免对Pixiv服务器造成压力**
3. **妥善保管Cookie信息，不要泄露给他人**
4. **使用代理时请注意相关法律法规**

## 🔗 相关链接

- [Pixiv官网](https://www.pixiv.net/)
- [React官方文档](https://react.dev/)
- [Ant Design组件库](https://ant.design/)
- [Go语言官网](https://golang.org/)
