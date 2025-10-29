# 任务管理器文档

任务管理器是PixivTailor的核心调度系统，负责统一管理所有任务的生命周期、状态跟踪、队列调度和并发控制。

## 📁 模块结构

### 核心文件
- **backend/internal/service/task_service.go**: 任务服务实现（核心调度器）
- **backend/internal/http/ai_handler.go**: AI生成任务处理器
- **backend/internal/http/crawler_handler.go**: 爬虫任务处理器
- **backend/internal/http/tagger_handler.go**: 标签任务处理器

## 🏗️ 架构设计

### 核心架构

```
┌─────────────────────────────────────────────────────────────┐
│                   TaskService                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  核心职责:                                              │  │
│  │  1. 任务生命周期管理                                     │  │
│  │  2. 任务状态跟踪 (pending→running→completed/failed)    │  │
│  │  3. 任务队列管理 (按类型分组)                           │  │
│  │  4. 执行器注册机制 (TaskExecutor)                       │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                      ▲
                      │
        ┌─────────────┼─────────────┐
        │             │             │
    AIHandler    CrawlerHandler   TaggerHandler
```

### 任务状态流转

```
pending (等待) → running (运行中) → completed (已完成)
                                      ↓
                                   failed (失败)
                                   cancelled (已取消)
```

## 🎯 核心功能

### 1. 任务类型管理

支持的任务类型：
- **generate**: AI图像生成任务
- **crawl**: Pixiv爬虫任务
- **tag**: 图像标签生成任务
- **train**: 模型训练任务（待实现）
- **classify**: 图像分类任务（待实现）

### 2. 并发控制策略

#### 不同类型任务可以并发执行
```
generate:G1  ──────────┐  同时运行
crawl:C1               ├───┐
tag:T1                 └───┤  (不同类型并发)
```

#### 同类型任务串行执行
```
generate:G1  ────────────────────┐
generate:G2             等待  ────┘ 启动
generate:G3             等待  ────┘ 启动
```

### 3. 任务队列机制

按任务类型分组管理等待队列：

```go
waitingTasksByType := map[string][]string{
    "generate": ["task1", "task2", "task3"],
    "crawl":    ["task4"],
    "tag":      ["task5", "task6"],
}
```

**调度逻辑**：
1. 检查同类型是否有运行中的任务
2. 如果有，加入对应类型的等待队列
3. 如果没有，立即启动
4. 任务完成后，从对应类型的等待队列中取出下一个任务启动

### 4. 执行器注册机制

针对复杂的任务（如AI生成），使用执行器模式：

```go
// AIHandler 实现 TaskExecutor 接口
type TaskExecutor interface {
    ExecuteGenerateTask(ctx context.Context, taskID string, config map[string]interface{})
}

// 在初始化时注册
taskService.RegisterExecutor("generate", aiHandler)
```

**优势**：
- 解耦任务调度和执行逻辑
- 允许不同任务类型使用不同的执行方式
- 便于扩展新任务类型

### 5. 状态一致性保证

**双重检查机制**：
1. 检查内存中的 `runningTasks` map
2. 检查数据库中的 `running` 状态任务
3. 如果发现状态不一致（数据库有但内存没有），自动清理

**示例**：
```go
// 检测到状态不一致
if actualRunningCount == 0 && dbRunningCount > 0 {
    // 强制清理数据库中的不一致状态
    s.UpdateTaskStatus(task.ID, "failed")
    s.UpdateTaskError(task.ID, "任务状态不一致，已自动清理")
}
```

## 📡 API 接口

### 任务管理 API

#### 1. 获取任务列表
```http
POST /api/tasks
Content-Type: application/json

{
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 0
  },
  "status": "",          // 筛选状态 (pending, running, completed, failed, cancelled)
  "type": ""             // 筛选类型 (generate, crawl, tag)
}

Response:
{
  "status": {
    "code": 0,
    "message": "success"
  },
  "data": {
    "tasks": [
      {
        "id": "abc123",
        "type": "generate",
        "status": "running",
        "progress": 50,
        "created_at": "2025-01-01T10:00:00Z",
        ...
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 50
    }
  }
}
```

#### 2. 获取单个任务状态
```http
POST /api/status
Content-Type: application/json

{
  "task_id": "abc123"
}

Response:
{
  "task_id": "abc123",
  "status": "running",
  "progress": 50,
  "message": "...",
  "error_message": null,
  "created_at": "2025-01-01T10:00:00Z",
  "completed_at": null
}
```

#### 3. 启动任务
```http
POST /api/task/start
Content-Type: application/json

{
  "task_id": "abc123"
}
```

#### 4. 停止任务
```http
POST /api/task/stop
Content-Type: application/json

{
  "task_id": "abc123"
}
```

#### 5. 取消任务
```http
POST /api/cancel
Content-Type: application/json

{
  "task_id": "abc123"
}
```

#### 6. 删除任务
```http
POST /api/delete
Content-Type: application/json

{
  "task_id": "abc123"
}
```

#### 7. 清理任务
```http
POST /api/task/cleanup
Content-Type: application/json

{
  "cleanup_type": "completed"  // 或 "failed", "all"
}
```

### 实时更新 - WebSocket

```javascript
// 连接到 WebSocket
const ws = new WebSocket('ws://localhost:50052/ws');

// 订阅任务更新
ws.send(JSON.stringify({
  type: 'subscribe_task',
  task_id: 'abc123'
}));

// 接收更新
ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  // 处理任务状态、进度、日志更新
};
```

## 🔄 任务执行流程

### AI 生成任务

```
1. 前端调用 API
   POST /api/generate-with-config
   {
     "config_id": "...",
     "override": {...}
   }

2. AIHandler 创建任务
   taskService.CreateTask("generate", config)

3. TaskService 调度
   - 检查是否有同类型运行中任务
   - 如果有，加入等待队列
   - 如果没有，立即启动

4. TaskService 调用执行器
   executeTask() → executeGenerateTask()
   → AIHandler.ExecuteGenerateTask()

5. AIHandler 执行实际业务逻辑
   - 调用 WebUI API
   - 下载和保存图片
   - 更新任务进度

6. 任务完成
   - 从 runningTasks 移除
   - 触发 processNextTask()
   - 启动等待队列中的下一个任务
```

### 爬虫任务

```
1. 前端调用 API
   POST /api/crawl/create
   {
     "crawl_type": "tag",
     "query": "...",
     "limit": 10
   }

2. CrawlerHandler 创建任务
   taskService.CreateTask("crawl", config)

3. TaskService 调度
   - 检查是否有同类型运行中任务
   - 启动或加入等待队列

4. TaskService 执行
   executeTask() → executeCrawlTask()
   - 创建爬虫实例
   - 执行爬取
   - 保存结果

5. 任务完成
   - 更新状态为 completed
   - 触发下一个任务
```

### 标签任务

```
1. 前端调用 API
   POST /api/tag/create
   {
     "input_dir": "...",
     "analyzer": "wd14tagger",
     ...
   }

2. TaggerHandler 创建任务
   taskService.CreateTask("tag", config)

3. TaskService 执行
   executeTask() → executeTagTask()
   - 调用 WD14Tagger
   - 生成标签文件
   - 保存结果

4. 任务完成
```

## 🎯 并发控制实现

### 核心数据结构

```go
type taskServiceImpl struct {
    // 运行中的任务
    runningTasks map[string]context.CancelFunc
    
    // 按类型分组的等待队列
    waitingTasksByType map[string][]string
    
    // 任务执行器映射
    executors map[string]TaskExecutor
    
    // 互斥锁
    taskMutex    sync.RWMutex
    queueMutex   sync.Mutex
}
```

### 调度算法

```go
// 检查是否有同类型运行中任务
func hasRunningTaskOfType(taskType string) bool {
    // 1. 检查 runningTasks map
    // 2. 检查数据库中的 running 状态
    return hasRunning
}

// 启动任务前检查
if hasRunningTaskOfType(task.Type) {
    // 加入等待队列
    waitingTasksByType[task.Type] = append(..., task.ID)
} else {
    // 立即启动
    startTaskExecution(task.ID)
}

// 任务完成后调度下一个
func processNextTask() {
    for taskType, waitingTasks := range waitingTasksByType {
        if !hasRunningTaskOfType(taskType) && len(waitingTasks) > 0 {
            // 启动该类型的下一个任务
            nextTaskID := waitingTasks[0]
            startTaskExecution(nextTaskID)
            return
        }
    }
}
```

## 📊 前端集成

### 使用任务管理 API

```typescript
import apiService from '@/services/api';

// 获取任务列表
const tasks = await apiService.getTasks();

// 获取单个任务
const task = await apiService.getTask(taskId);

// 启动任务
await apiService.startTask(taskId);

// 停止任务
await apiService.stopTask(taskId);

// 取消任务
await apiService.cancelTask(taskId);

// 删除任务
await apiService.deleteTask(taskId);
```

### 实时更新

```typescript
// 创建 WebSocket 连接
const ws = new WebSocket('ws://localhost:50052/ws');

// 监听消息
ws.addEventListener('message', (event) => {
  const data = JSON.parse(event.data);
  
  // 更新任务状态
  if (data.type === 'task_status') {
    updateTaskStatus(data.task_id, data.status, data.progress);
  }
  
  // 新增日志
  if (data.type === 'task_log') {
    addTaskLog(data.task_id, data.level, data.message);
  }
});
```

## 🎨 前端模块与后端 API 映射

| 前端模块     | 任务类型   | 后端 API                         | Handler        |
| ------------ | ---------- | -------------------------------- | -------------- |
| **AI生成器** | `generate` | `POST /api/generate-with-config` | AIHandler      |
| **爬虫管理** | `crawl`    | `POST /api/crawl/create`         | CrawlerHandler |
| **标签生成** | `tag`      | `POST /api/tag/create`           | TaggerHandler  |

所有模块共用相同的任务查看接口：
- `POST /api/tasks` - 获取任务列表
- `POST /api/status` - 获取任务状态

## 🔍 监控和调试

### 任务状态查询

```bash
# 查看所有任务
curl -X POST http://localhost:50052/api/tasks \
  -H "Content-Type: application/json" \
  -d '{"pagination": {"page": 1, "page_size": 100}}'

# 查看指定类型的任务
curl -X POST http://localhost:50052/api/tasks \
  -H "Content-Type: application/json" \
  -d '{"pagination": {"page": 1, "page_size": 100}, "type": "generate"}'

# 查看指定状态的任务
curl -X POST http://localhost:50052/api/tasks \
  -H "Content-Type: application/json" \
  -d '{"pagination": {"page": 1, "page_size": 100}, "status": "running"}'
```

### 日志分析

任务管理器会记录详细的调度日志：

```
time="2025-01-01 10:00:00" level=info msg="CreateTask 开始执行: taskType=generate"
time="2025-01-01 10:00:01" level=info msg="检测到同类型的运行中任务，任务 abc123 (类型: generate) 加入等待队列"
time="2025-01-01 10:00:02" level=info msg="executeGenerateTask: 开始执行AI生成任务 abc123"
time="2025-01-01 10:00:05" level=info msg="processNextTask: 启动 generate 类型任务 def456 (该类型还有 2 个等待任务)"
```

## 🚀 最佳实践

### 1. 任务创建
- 任务创建后会自动加入队列
- 如果没有同类型运行中任务，会立即启动
- 如果有，会进入对应类型的等待队列

### 2. 任务监控
- 使用 WebSocket 实时接收任务更新
- 定期轮询任务状态（如果需要）
- 监控任务进度和日志

### 3. 资源管理
- 任务完成后会自动从 `runningTasks` 中移除
- 等待队列按类型分组，不会互相阻塞
- 不同类型的任务可以并发执行

### 4. 错误处理
- 任务失败时会记录错误信息
- 可以重新启动失败的任务
- 支持任务取消和删除

## 🔧 扩展指南

### 添加新任务类型

1. 定义新的任务类型常量
2. 在 TaskService 中添加执行逻辑
3. 创建对应的 Handler（如果需要特殊逻辑）
4. 实现 TaskExecutor 接口（如果需要）

```go
// 1. 添加任务类型
switch task.Type {
case "generate":
    s.executeGenerateTask(ctx, task, config)
case "crawl":
    s.executeCrawlTask(ctx, task, config)
case "tag":
    s.executeTagTask(ctx, task, config)
case "my_new_type":  // 新任务类型
    s.executeMyNewTask(ctx, task, config)
}
```

### 实现执行器

```go
// 实现 TaskExecutor 接口
func (h *MyHandler) ExecuteMyNewTask(ctx context.Context, taskID string, config map[string]interface{}) {
    // 实现具体的任务逻辑
    // ...
}

// 注册执行器
taskService.RegisterExecutor("my_new_type", myHandler)
```

## 📝 总结

任务管理器提供了统一的任务调度框架，具有以下特点：

✅ **类型级别并发控制** - 不同类型任务可并发，同类型任务串行  
✅ **智能队列管理** - 按类型分组，避免互相阻塞  
✅ **执行器模式** - 灵活的任务执行机制  
✅ **状态一致性** - 双重检查，保证状态准确  
✅ **实时更新** - WebSocket 推送任务状态  
✅ **易于扩展** - 简单的接口实现新任务类型  

通过统一的任务管理接口，前端各模块可以方便地查看和管理各自的任务。

## 🔗 相关文档

- [AI模块文档](./ai-module.md) - AI生成任务实现
- [爬虫模块文档](./crawler-module.md) - 爬虫任务实现
- [标签模块文档](./tagger-module.md) - 标签任务实现
- [配置模块文档](./config-module.md) - 配置文件管理

