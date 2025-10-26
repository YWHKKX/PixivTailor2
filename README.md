# PixivTailor2

一个基于 Go 语言开发的现代化 Pixiv 爬虫与 AI 图像处理平台，集成了 Web 界面、实时监控、任务管理和智能缓存系统。**项目核心功能已完成开发**，提供了完整的 Pixiv 内容爬取、AI图像生成、文件管理、实时日志和任务调度功能。

## ✨ 核心特性 (已完成)

### 🚀 已实现功能
- **🕷️ 智能爬取**: 支持按标签、用户、插画ID爬取 Pixiv 插画 ✅
- **🎨 AI 生成**: 集成 Stable Diffusion WebUI，支持 LoRA 模型和配置管理 ✅
- **🌐 现代化Web界面**: React + TypeScript + Ant Design 响应式界面 ✅
- **📈 实时监控**: WebSocket 实时通信，任务进度和日志实时更新 ✅
- **📁 文件管理**: 任务隔离的文件存储，可视化文件树浏览 ✅
- **⚙️ 智能配置**: 灵活的 Cookie 和代理配置，支持默认配置 ✅
- **🔄 任务管理**: 完整的任务生命周期管理，支持重试和取消 ✅
- **📊 实时日志**: 分级日志系统，实时日志流显示 ✅
- **🛡️ 错误处理**: 完善的错误重试和状态管理机制 ✅
- **🎯 AI配置管理**: 文件系统配置管理，支持LoRA和参数模板 ✅
- **🖼️ 图片查看器**: 内置图片查看器，支持键盘导航和批量操作 ✅

### 🚧 规划功能
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
PixivTailor2/
├── backend/                    # Go后端服务 ✅
│   ├── internal/              # 内部包
│   │   ├── http/              # HTTP服务器和API路由 ✅
│   │   │   ├── ai_handler.go  # AI处理器 ✅
│   │   │   ├── generation_config_handler.go # 生成配置处理器 ✅
│   │   │   └── server.go      # HTTP服务器 ✅
│   │   ├── service/           # 业务服务层 ✅
│   │   │   ├── generation_config_service.go # 生成配置服务 ✅
│   │   │   └── task_service.go # 任务服务 ✅
│   │   ├── repository/        # 数据访问层 ✅
│   │   │   ├── generation_config.go # 生成配置存储 ✅
│   │   │   └── storage.go     # 数据存储 ✅
│   │   ├── crawler/           # 爬虫核心功能 ✅
│   │   └── config/            # 配置管理 ✅
│   ├── pkg/                   # 公共包 ✅
│   │   ├── paths/             # 路径管理服务 ✅
│   │   ├── models/            # 数据模型 ✅
│   │   │   └── generation_config.go # 生成配置模型 ✅
│   │   └── api/               # API客户端 ✅
│   ├── global_configs/        # 全局配置文件 ✅
│   │   ├── crawler_config.json # 爬虫配置 ✅
│   │   ├── pixiv_example.json # 配置示例 ✅
│   │   └── anime-style-config.json # AI生成配置 ✅
│   ├── data/                  # 数据目录 ✅
│   │   ├── images/            # 图片存储 (按任务ID组织) ✅
│   │   ├── cache/             # 缓存目录 (按任务ID组织) ✅
│   │   └── database/          # SQLite数据库 ✅
│   ├── tailor.go              # 主程序入口 ✅
│   └── tailor.exe             # 编译后的可执行文件 ✅
├── frontend/                  # React前端应用 ✅
│   ├── src/                   # 源代码 ✅
│   │   ├── pages/             # 页面组件 ✅
│   │   │   ├── AIGeneratorPage.tsx # AI生成器页面 ✅
│   │   │   ├── ConfigManagerPage.tsx # 配置管理页面 ✅
│   │   │   └── CrawlerPage.tsx # 爬虫页面 ✅
│   │   ├── services/          # API服务 ✅
│   │   │   ├── aiService.ts   # AI服务客户端 ✅
│   │   │   └── websocket.ts   # WebSocket服务 ✅
│   │   ├── components/        # 通用组件 ✅
│   │   └── utils/             # 工具函数 ✅
│   ├── public/                # 静态资源 ✅
│   ├── dist/                  # 构建产物 ✅
│   └── package.json           # 依赖管理 ✅
├── docs/                      # 项目文档 ✅
│   ├── README.md              # 主文档 ✅
│   ├── ai-module.md           # AI模块文档 ✅
│   ├── cli-usage.md           # CLI使用说明 ✅
│   ├── deployment-guide.md    # 部署指南 ✅
│   └── crawler-module.md      # 爬虫模块文档 ✅
├── scripts/                   # 启动脚本 ✅
│   ├── start-dev.bat          # 开发环境启动 ✅
│   ├── start-webui-api.bat    # WebUI API启动 ✅
│   └── kill-webui.bat         # WebUI停止 ✅
└── README.md                  # 项目说明 ✅
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

访问 http://localhost:3000 使用爬虫功能：

- **标签爬取**: 根据标签搜索和下载Pixiv插画
- **用户爬取**: 下载指定用户的所有作品
- **插画爬取**: 下载特定插画ID的图片
- **实时监控**: 查看任务进度和实时日志
- **文件管理**: 在文件树中浏览下载的图片

### 3. AI图像生成 ✅

访问 http://localhost:3000 使用AI生成功能：

- **配置管理**: 管理AI生成配置和LoRA模型
- **参数调整**: 实时调整生成参数(提示词、步数、CFG等)
- **批量生成**: 支持批次数量和循环发包
- **任务监控**: 实时监控生成任务状态和进度
- **图片查看**: 内置图片查看器，支持键盘导航
- **批量操作**: 支持批量删除任务

### 4. 配置管理 ✅

- **Cookie配置**: 支持默认Cookie和自定义Cookie
- **代理设置**: 支持HTTP代理配置
- **任务参数**: 设置下载数量、延迟等参数
- **文件存储**: 自动按任务ID组织文件结构
- **AI配置**: 文件系统配置管理，支持LoRA和参数模板

### 5. 任务管理 ✅

- **任务创建**: 通过Web界面创建爬取和AI生成任务
- **实时监控**: WebSocket实时更新任务状态
- **任务控制**: 支持启动、停止、重试任务
- **日志查看**: 实时查看任务执行日志
- **批量操作**: 支持批量删除和清理任务

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

**爬虫相关**:
- **POST /api/crawl/create**: 创建爬取任务
- **POST /api/task/start**: 启动任务
- **POST /api/task/stop**: 停止任务
- **POST /api/filetree**: 获取文件树
- **GET /api/images/{path}**: 获取图片文件

**AI生成相关**:
- **POST /api/generate-with-config**: 使用配置生成AI图像
- **GET /api/configs**: 获取配置列表
- **GET /api/configs/{id}**: 获取单个配置
- **POST /api/configs**: 创建配置
- **PUT /api/configs/{id}**: 更新配置
- **DELETE /api/configs/{id}**: 删除配置
- **GET /api/tasks/{taskId}/images/{index}**: 获取生成的图片

**系统相关**:
- **GET /health**: 健康检查
- **GET /api/webui/status**: WebUI状态检查
- **POST /api/webui/start**: 启动WebUI
- **POST /api/webui/stop**: 停止WebUI

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
   backend/global_configs/crawler_config.json
   
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
- **配置示例**: `backend/global_configs/pixiv_example.json`
- **启动脚本**: `scripts/` 目录

## 🎯 项目状态总结

### ✅ 已完成功能 (核心功能)

1. **现代化Web界面** - React + TypeScript + Ant Design ✅
2. **Pixiv爬虫系统** - 支持标签、用户、插画ID爬取 ✅
3. **AI图像生成系统** - 集成Stable Diffusion WebUI，支持LoRA模型 ✅
4. **任务管理系统** - 完整的任务生命周期管理 ✅
5. **文件管理系统** - 任务隔离的文件存储和浏览 ✅
6. **实时监控系统** - WebSocket实时通信和状态更新 ✅
7. **配置管理系统** - 灵活的Cookie、代理和AI配置 ✅
8. **错误处理系统** - 完善的错误重试和状态管理 ✅
9. **图片查看器** - 内置图片查看器，支持键盘导航 ✅
10. **批量操作** - 支持批量删除和清理任务 ✅

### 🚧 待实现功能 (扩展功能)

1. **模型训练** - Kohya-ss训练框架集成
2. **智能标签** - WD14Tagger自动标签生成
3. **标签分类** - OpenAI API智能分类

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
