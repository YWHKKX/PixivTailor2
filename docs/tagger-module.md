# 图像标签模块文档

图像标签模块是PixivTailor的新增功能模块，负责为图像生成标签，包括自动标签识别、批量处理、标签管理等核心功能。

## 📁 模块结构

- **backend/internal/http/tagger_handler.go**: 标签处理器实现 ✅
- **backend/internal/http/server.go**: HTTP服务器集成 ✅
- **backend/internal/ai/ai.go**: AI服务集成 ✅
- **backend/pkg/models/models.go**: 标签数据模型 ✅
- **frontend/src/pages/TaggerPage.tsx**: 标签生成器页面 ✅
- **frontend/src/services/api.ts**: API服务客户端 ✅

## 🔧 核心组件

### 1. TaggerHandler - 标签处理器 ✅

**功能描述**: 负责处理图像标签生成的HTTP请求

**主要功能**:
- **CreateTagTask**: 创建标签生成任务 ✅
- **GetTaggedImages**: 获取已标签的图像 ✅
- **GetAvailableAnalyzers**: 获取可用的分析器列表 ✅
- **AnalyzeImage**: 分析单张图像 ✅
- **GetTagTaskStatus**: 获取标签任务状态 ✅
- **StopTagTask**: 停止标签任务 ✅

**支持功能**:
- 多种标签分析器支持（WD14Tagger、DeepBooru）✅
- 批量图像处理 ✅
- 标签置信度评分 ✅
- 多种保存格式（TXT、JSON、CSV）✅
- 任务管理和监控 ✅
- 实时日志流（WebSocket）✅
- 动态分析器切换 ✅
- 路径自动解析和验证 ✅
- 异步批量处理 ✅
- 实时状态更新 ✅
- 错误处理和响应 ✅

### 2. TaggerPage - 标签生成器前端页面 ✅

**功能描述**: 提供图像标签生成的Web界面

**主要功能**:
- **配置管理**: 设置输入目录、分析器、输出格式等 ✅
- **任务管理**: 创建、监控、停止标签生成任务 ✅
- **结果展示**: 查看生成的标签和置信度 ✅
- **文件树集成**: 与现有文件系统集成 ✅
- **实时更新**: WebSocket实时状态更新 ✅

**支持功能**:
- 响应式界面设计 ✅
- 实时进度显示 ✅
- 标签可视化展示 ✅
- 任务详情查看 ✅
- 错误处理 ✅
- 实时日志流显示（info/warning/error级别）✅
- 分析器下拉选择（WD14 Tagger / DeepBooru）✅
- 自动路径解析（images/前缀自动添加）✅

### 3. 数据模型 ✅

**TagRequest** - 标签请求模型:
```go
type TagRequest struct {
    InputDir   string   `json:"input_dir"`    // 输入目录
    OutputDir  string   `json:"output_dir"`    // 输出目录
    Analyzer   string   `json:"analyzer"`      // 分析器
    SkipTags   []string `json:"skip_tags"`     // 跳过标签
    ExtendTags []string `json:"extend_tags"`   // 扩展标签
    TagOrder   string   `json:"tag_order"`     // 标签排序
    SaveType   string   `json:"save_type"`     // 保存格式
    Limit      int      `json:"limit"`         // 处理数量限制
}
```

**TaggedImage** - 已标签图像模型:
```go
type TaggedImage struct {
    BaseModel
    ImagePath string            `json:"image_path"`  // 图像路径
    Tags      []Tag             `json:"tags"`        // 标签列表
    Analyzer  string            `json:"analyzer"`     // 使用的分析器
    Metadata  map[string]string `json:"metadata"`    // 元数据
}

type Tag struct {
    Name      string  `json:"name"`       // 标签名称
    Score     float64 `json:"score"`      // 置信度分数
    Category  string  `json:"category"`   // 标签分类
    IsGeneral bool    `json:"is_general"` // 是否通用标签
}
```

## 🚀 主要功能

### 1. 自动标签生成 ✅

**功能描述**: 使用AI模型自动为图像生成标签

**支持特性**:
- **多种分析器**: 支持WD14Tagger、DeepDanbooru等 ✅
- **批量处理**: 支持批量图像标签生成 ✅
- **智能识别**: 自动识别图像中的对象、场景、风格等 ✅
- **置信度评分**: 每个标签都有置信度分数 ✅
- **标签分类**: 标签按分类组织（角色、风格、质量等）✅

**使用方式**:
```bash
# 通过前端界面
1. 访问"图像标签"页面
2. 选择输入目录
3. 配置分析器和其他参数
4. 点击"开始生成标签"
5. 实时查看生成进度和结果

# 通过CLI（待实现）
tailor tag --input ./images --analyzer wd14tagger --output ./tags
```

### 2. 标签管理 ✅

**功能描述**: 管理和查看生成的标签

**支持特性**:
- **标签展示**: 可视化展示标签和置信度 ✅
- **标签过滤**: 按置信度、分类过滤标签 ✅
- **标签编辑**: 手动添加、删除、修改标签 ✅
- **批量操作**: 批量导出、删除标签 ✅
- **标签搜索**: 快速搜索特定标签 ✅

### 3. 多种保存格式 ✅

**功能描述**: 支持多种标签保存格式

**支持格式**:
- **TXT**: 简单的文本格式，每行一个标签 ✅
- **JSON**: 完整的结构化数据，包含所有元数据 ✅
- **CSV**: 表格格式，便于数据分析 ✅

**示例**:
```json
// JSON格式示例
{
  "image_path": "data/images/sample.jpg",
  "tags": [
    {
      "name": "1girl",
      "score": 0.95,
      "category": "character",
      "is_general": true
    },
    {
      "name": "anime",
      "score": 0.90,
      "category": "style",
      "is_general": true
    }
  ],
  "analyzer": "wd14tagger",
  "metadata": {
    "width": "512",
    "height": "512"
  }
}
```

```txt
# TXT格式示例
1girl, anime, cute, solo, looking at viewer
```

### 4. 任务管理 ✅

**功能描述**: 完整的任务管理和监控功能

**支持特性**:
- **任务创建**: 创建新的标签生成任务 ✅
- **进度监控**: 实时查看任务进度 ✅
- **任务停止**: 随时停止正在运行的任务 ✅
- **任务详情**: 查看任务详细信息和日志 ✅
- **批量管理**: 批量清理已完成或失败的任务 ✅

### 5. 分析器支持 ✅

**功能描述**: 支持多种标签分析器

**可用分析器**:
- **WD14 Tagger**: 基于ConvNeXt V2的深度学习标签识别（默认）✅
  - 支持CPU/GPU模式
  - 高准确率
  - 支持广泛标签类别
  - API端点: `/tagger/v1/interrogate`
  - 响应字段: `caption`
  
- **DeepBooru**: 基于深度学习的标签分析系统 ✅
  - 高性能标签检测
  - 丰富的标签库
  - API端点: `/sdapi/v1/interrogate`
  - 响应字段: `result`

**特点**:
- 高准确率
- 支持大量标签类别
- 快速处理速度
- 可定制化配置
- 支持动态切换分析器
- 自动调整模型列表（基于分析器选择）

## 📊 API接口

### 1. 创建标签任务

**端点**: `POST /api/tag/create`

**请求体**:
```json
{
  "input_dir": ["data/images/task_xxx", "data/images/task_yyy"],
  "output_dir": "data/tags",
  "analyzer": "wd14tagger",
  "model": "wd14-convnext-v2.onnx",
  "skip_tags": ["lowres", "normal quality"],
  "extend_tags": ["beautiful", "detailed"],
  "tag_order": "score",
  "save_type": "json",
  "limit": 100
}
```

**注意**:
- `input_dir` 支持字符串或字符串数组
- 如果路径不以 `data/`、`images/` 等开头，会自动添加 `images/` 前缀
- 使用 `pathManager.ResolvePath()` 进行路径解析
- `analyzer` 选项: `wd14tagger` 或 `deepbooru`

**响应**:
```json
{
  "status": {
    "code": 200,
    "message": "任务创建成功"
  },
  "data": {
    "id": "task_xxx",
    "type": "tag",
    "status": "pending",
    "progress": 0,
    "created_at": "2025-01-20T10:00:00Z"
  }
}
```

### 2. 获取已标签图像

**端点**: `GET /api/tag/images?task_id=xxx`

**响应**:
```json
{
  "status": {
    "code": 200,
    "message": "获取成功"
  },
  "data": [
    {
      "id": 1,
      "image_path": "data/images/sample.jpg",
      "tags": [
        {
          "name": "1girl",
          "score": 0.95,
          "category": "character",
          "is_general": true
        }
      ],
      "analyzer": "wd14tagger",
      "metadata": {
        "width": "512",
        "height": "512"
      }
    }
  ]
}
```

### 3. 获取可用分析器

**端点**: `GET /api/tag/analyzers`

**响应**:
```json
{
  "status": {
    "code": 200,
    "message": "获取成功"
  },
  "data": [
    "wd14tagger",
    "deepdanbooru"
  ]
}
```

### 4. 分析单张图像

**端点**: `POST /api/tag/analyze`

**请求体**:
```json
{
  "image_path": "data/images/sample.jpg",
  "analyzer": "wd14tagger"
}
```

**响应**:
```json
{
  "status": {
    "code": 200,
    "message": "分析成功"
  },
  "data": {
    "id": 1,
    "image_path": "data/images/sample.jpg",
    "tags": [...],
    "analyzer": "wd14tagger",
    "metadata": {...}
  }
}
```

### 5. 获取任务状态

**端点**: `GET /api/tag/status?task_id=xxx`

### 6. 停止任务

**端点**: `POST /api/tag/stop?task_id=xxx`

## 🎯 使用场景

### 场景1: 批量标签生成

**需求**: 为大量图像批量生成标签

**步骤**:
1. 将图像放入指定目录
2. 访问"图像标签"页面
3. 配置参数（输入目录、分析器、输出格式等）
4. 点击"开始生成标签"
5. 等待处理完成
6. 查看和导出生成的标签

### 场景2: 单张图像快速分析

**需求**: 快速分析单张图像并获取标签

**步骤**:
1. 在文件树中选择图像
2. 点击"分析"按钮
3. 选择分析器
4. 获取标签结果
5. 查看置信度和分类

### 场景3: 标签管理

**需求**: 管理和编辑已生成的标签

**步骤**:
1. 查看生成的标签列表
2. 按置信度或分类筛选
3. 编辑或删除不需要的标签
4. 添加自定义标签
5. 导出标签到指定格式

## 🔍 技术细节

### 1. 标签分析流程

```
1. 选择输入目录（支持多目录）
2. 解析并验证路径（自动添加 images/ 前缀）
3. 扫描图像文件（getImageFiles，先收集所有文件）
4. 应用全局限制（limit 参数）
5. 加载分析器（根据 analyzer 选择 WD14 Tagger 或 DeepBooru）
6. 动态构建 API 端点:
   - WD14 Tagger: /tagger/v1/interrogate
   - DeepBooru: /sdapi/v1/interrogate
7. 发送请求到 Stable Diffusion WebUI
8. 解析响应（caption 或 result 字段）
9. 对每张图像进行分析
10. 生成标签和置信度分数
11. 应用过滤规则（跳过标签）
12. 添加扩展标签
13. 按指定顺序排序
14. 保存到指定格式（TXT/JSON/CSV）
15. 实时发送日志（info/warning/error）
```

### 2. 任务管理机制

- 异步处理: 标签生成在后台异步进行
- 进度跟踪: 实时更新处理进度
- 错误处理: 自动重试失败的图像
- 资源控制: 控制并发数量和处理速度

### 3. 数据流

```
图像文件 → 分析器 → 标签识别 → 置信度评分 → 标签分类 → 过滤规则 → 扩展标签 → 排序 → 保存
```

## 🚧 未来规划

### 规划功能:

1. **更多分析器支持**
   - Danbooru
   - NovelAI
   自定义分析器

2. **高级标签管理**
   - 标签相似度检测
   - 标签推荐系统
   - 自动标签优化

3. **批量操作增强**
   - 标签去重
   - 标签合并
   - 标签统计分析

4. **集成增强**
   - 与爬虫模块集成
   - 与AI生成模块集成
   - 标签数据库集成

## 📝 注意事项

1. **性能考虑**: 批量处理大量图像时可能需要较长时间
2. **资源消耗**: 标签分析需要一定的计算资源
3. **模型依赖**: 确保安装和配置了相应的分析器
4. **存储空间**: 生成的标签文件会占用一定存储空间

## 🐛 故障排除

### 常见问题:

1. **标签生成失败**
   - 检查图像文件是否有效
   - 确认分析器是否正确安装
   - 查看任务日志了解详细错误

2. **进度卡住**
   - 检查任务状态
   - 尝试重启任务
   - 查看系统资源使用情况

3. **标签质量不佳**
   - 尝试不同的分析器
   - 调整置信度阈值
   - 使用标签过滤和扩展功能

4. **CUDA/ONNX Runtime 错误** ⚠️
   
   **错误信息**: `LocalEntryNotFoundError`, `需要cuDNN 9.和CUDA 12.`
   
   **原因分析**:
   - WD14 Tagger 使用 ONNX Runtime GPU 版本
   - 需要特定的 CUDA 和 cuDNN 版本（CUDA 12.x + cuDNN 9.x）
   - ONNX Runtime 无法加载 CUDA 执行提供程序
   
   **解决方案**:
   
   **方案1: 安装正确的 CUDA 和 cuDNN**
   ```bash
   # 1. 下载并安装 CUDA 12.x Toolkit
   # 访问: https://developer.nvidia.com/cuda-downloads
   
   # 2. 下载匹配的 cuDNN (需要 NVIDIA 账号)
   # 访问: https://developer.nvidia.com/cudnn
   # 选择 "for CUDA 12.x" 的 cuDNN 9.x 版本
   
   # 3. 检查安装
   nvcc --version
   nvidia-smi
   
   # 4. 在 Python 中检查 ONNX Runtime
   python -c "import onnxruntime as ort; print(ort.__version__); print(ort.get_available_providers())"
   ```
   
   **方案2: 使用 CPU 模式** (推荐，更简单)
   - 在 Stable Diffusion WebUI 的 WD14 Tagger 扩展设置中启用 CPU 模式
   - 或在 WebUI 启动参数中添加: `--tagger-device cpu`
   - 这需要修改 Stable Diffusion WebUI 的配置
   
   **方案3: 使用不同的标签器**
   - 如果 WebUI 支持，可以尝试其他标签器
   - 或者使用本地的 WD14 Tagger 而不依赖 WebUI
   
   **重要提示**:
   - 当前 PixivTailor 使用 `wd14-convnext-v2` 模型（默认）
   - 该模型**支持 CPU 模式**，即使没有 CUDA 也能正常工作
   - 如果仍然遇到 CUDA 错误，这通常是 WebUI 的警告，不影响功能
   - ConvNeXt V2 模型通常在 CPU 模式下比 VIT 模型更稳定

### 环境要求:

- **Stable Diffusion WebUI**: 已安装并运行
- **WD14 Tagger 扩展**: 已安装并启用
- **GPU 模式**:
  - CUDA 12.x
  - cuDNN 9.x (for CUDA 12.x)
  - 匹配的 GPU 驱动
- **CPU 模式**: 
  - ONNX Runtime CPU 版本
  - 足够的系统内存

## 🎯 任务管理集成

标签模块已集成到统一的任务管理系统中，实现智能队列调度和并发控制。

### 任务执行流程

```
1. TaggerHandler 接收请求
   ↓
2. 创建标签任务（taskService.CreateTask("tag", config)）
   ↓
3. TaskService 检查队列
   - 如果同类型有运行中任务 → 加入等待队列
   - 如果没有 → 立即启动
   ↓
4. executeTagTask 执行标签生成
   - 调用 WD14Tagger
   - 批量处理图像
   - 生成标签文件
   ↓
5. 任务完成 → 自动启动下一个等待中的任务
```

### 并发控制

- **同类型串行**: 多个 tag 任务依次执行
- **不同类型并发**: generate、crawl、tag 任务可以同时运行
- **智能队列**: 按类型分组的等待队列，互不干扰

详见[任务管理器文档](./task-manager.md)了解更多细节。

## 🔗 相关文档

- [任务管理器文档](./task-manager.md) - 统一的任务调度系统
- [AI模块文档](./ai-module.md) - AI生成任务
- [爬虫模块文档](./crawler-module.md) - Pixiv爬虫功能
- [配置模块文档](./config-module.md) - 配置文件管理