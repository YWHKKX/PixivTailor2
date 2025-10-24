# PixivTailorEx

一个基于 Go 语言开发的现代化 Pixiv 爬虫与 AI 图像处理平台，集成了 Web 界面、实时监控、任务管理和智能缓存系统。**项目核心功能已完成开发**，提供了完整的 Pixiv 内容爬取、文件管理、实时日志和任务调度功能。

## ✨ 核心特性 (已完成)

### 🚀 已实现功能
- **🕷️ 智能爬取**: 支持按标签、用户、插画ID爬取 Pixiv 插画 ✅
- **🌐 现代化Web界面**: React + TypeScript + Ant Design 响应式界面 ✅
- **📈 实时监控**: WebSocket 实时通信，任务进度和日志实时更新 ✅
- **📁 文件管理**: 任务隔离的文件存储，可视化文件树浏览 ✅
- **⚙️ 智能配置**: 灵活的 Cookie 和代理配置，支持默认配置 ✅
- **🔄 任务管理**: 完整的任务生命周期管理，支持重试和取消 ✅
- **📊 实时日志**: 分级日志系统，实时日志流显示 ✅
- **🛡️ 错误处理**: 完善的错误重试和状态管理机制 ✅

### 🚧 规划功能
- **🎨 AI 生成**: 集成 Stable Diffusion WebUI，支持 LoRA 模型和姿态控制
- **🤖 模型训练**: 集成 Kohya-ss，支持 LoRA 模型训练
- **🏷️ 智能标签**: 使用 WD14Tagger 自动生成图像标签
- **📊 标签分类**: 使用 OpenAI API 对标签进行智能分类

## 🚀 快速开始

### 环境要求 ✅

- **Windows 10/11** (64位)
- **Go 1.21+** - 后端开发环境
- **Node.js 18+** - 前端开发环境
- **现代浏览器** - Chrome/Firefox/Edge (支持WebSocket)

### 一键启动 ✅

```bash
# 1. 克隆项目
git clone <repository-url>
cd PixivTailorEx

# 2. 一键启动 (推荐)
scripts/start-dev.bat

# 3. 访问系统
# Web界面: http://localhost:3002
# API接口: http://localhost:50052/api/
```

### 手动启动 ✅

```bash
# 终端1: 启动后端
cd backend
go run tailor.go server

# 终端2: 启动前端
cd frontend
npm install
npm run dev
```

## 📁 项目结构 (当前状态)

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
│   │   └── pixiv_example.json # 配置示例
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

## 🛠️ 开发指南

### 本地开发 ✅

```bash
# 启动开发模式
scripts/start-dev.bat

# 或者手动启动
# 后端
cd backend
go run tailor.go server

# 前端
cd frontend
npm run dev
```

### 生产部署 ✅

```bash
# 编译后端
cd backend
go build -o tailor.exe tailor.go

# 构建前端
cd frontend
npm run build

# 启动生产服务
scripts/start-prod.bat
```

## 📖 使用说明

### 1. 系统访问 ✅

- **Web界面**: http://localhost:3002
- **后端API**: http://localhost:50052/api/
- **WebSocket**: ws://localhost:50052/ws
- **健康检查**: http://localhost:50052/health

### 2. 爬虫功能 ✅

访问 http://localhost:3002 使用爬虫功能：

- **标签爬取**: 根据标签搜索和下载Pixiv插画
- **用户爬取**: 下载指定用户的所有作品
- **插画爬取**: 下载特定插画ID的图片
- **实时监控**: 查看任务进度和实时日志
- **文件管理**: 在文件树中浏览下载的图片

### 3. 配置管理 ✅

- **Cookie配置**: 支持默认Cookie和自定义Cookie
- **代理设置**: 支持HTTP代理配置
- **任务参数**: 设置下载数量、延迟等参数
- **文件存储**: 自动按任务ID组织文件结构

### 4. 任务管理 ✅

- **任务创建**: 通过Web界面创建爬取任务
- **实时监控**: WebSocket实时更新任务状态
- **任务控制**: 支持启动、停止、重试任务
- **日志查看**: 实时查看任务执行日志

## 🔧 API 文档

### HTTP API ✅

- **端口**: 50052
- **基础路径**: /api/
- **协议**: HTTP/HTTPS
- **格式**: JSON

### WebSocket API ✅

- **端口**: 50052
- **路径**: /ws
- **协议**: WebSocket
- **功能**: 实时通信、任务状态更新、日志流

### 主要接口

- **POST /api/crawl/create**: 创建爬取任务
- **POST /api/task/start**: 启动任务
- **POST /api/task/stop**: 停止任务
- **POST /api/filetree**: 获取文件树
- **GET /api/images/{path}**: 获取图片文件
- **GET /health**: 健康检查

## 📊 监控和日志

### 实时监控 ✅

- **任务状态**: 实时任务状态更新
- **进度跟踪**: 实时进度百分比显示
- **日志流**: 实时日志信息显示
- **错误报告**: 详细错误信息和重试状态

### 日志系统 ✅

- **分级日志**: info, warn, error 等级别
- **实时显示**: WebSocket 实时日志流
- **任务日志**: 每个任务的独立日志
- **全局日志**: 系统全局日志信息

## 🐛 故障排除

### 常见问题

1. **端口被占用** ✅
   ```bash
   # 检查端口占用
   netstat -ano | findstr :50052
   netstat -ano | findstr :3002
   
   # 停止占用进程
   taskkill /f /pid <进程ID>
   ```

2. **依赖安装失败** ✅
   ```bash
   # 清理缓存后重试
   npm cache clean --force
   go clean -modcache
   ```

3. **Cookie配置问题** ✅
   ```bash
   # 检查配置文件
   backend/configs/crawler_config.json
   
   # 确保Cookie格式正确
   # 格式: PHPSESSID=xxx; __utma=xxx; ...
   ```

4. **代理连接失败** ✅
   ```bash
   # 检查代理服务器是否运行
   # 默认代理地址: http://127.0.0.1:7890
   # 确保代理服务器正常运行
   ```

### 获取帮助

- **项目文档**: `docs/` 目录
- **配置示例**: `backend/configs/pixiv_example.json`
- **启动脚本**: `scripts/` 目录

## 🎯 项目状态总结

### ✅ 已完成功能 (核心功能)

1. **现代化Web界面** - React + TypeScript + Ant Design
2. **Pixiv爬虫系统** - 支持标签、用户、插画ID爬取
3. **任务管理系统** - 完整的任务生命周期管理
4. **文件管理系统** - 任务隔离的文件存储和浏览
5. **实时监控系统** - WebSocket实时通信和状态更新
6. **配置管理系统** - 灵活的Cookie和代理配置
7. **错误处理系统** - 完善的错误重试和状态管理

### 🚧 待实现功能 (扩展功能)

1. **AI图像生成** - Stable Diffusion WebUI集成
2. **模型训练** - Kohya-ss训练框架集成
3. **智能标签** - WD14Tagger自动标签生成
4. **标签分类** - OpenAI API智能分类

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [React](https://react.dev/) - 前端框架
- [Ant Design](https://ant.design/) - UI组件库
- [Go](https://golang.org/) - 后端语言
- [Pixiv](https://www.pixiv.net/) - 内容平台

---

**提示**: 如果遇到问题，请先运行 `scripts/status.bat` 检查服务状态，然后查看相关日志文件。
