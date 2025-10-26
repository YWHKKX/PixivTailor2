# 配置模块文档

配置模块是PixivTailor的统一配置管理模块，负责加载、验证和管理应用程序的所有配置参数，支持多种配置源、动态配置更新和模块化配置管理。该模块已完全实现并集成到Web界面中。

## 📁 模块结构

- **backend/internal/config/**: 配置管理模块目录
- **config.go**: 配置管理接口实现
- **modules.go**: 模块配置定义
- **ai_config.go**: AI模块配置定义
- **backend/internal/http/generation_config_handler.go**: 配置文件管理API

## 🔧 核心组件

### 1. 配置管理器 ✅

**功能描述**: 统一管理所有配置，提供配置的加载、保存、验证和监听功能

**主要功能**:
- **Load**: 加载配置文件
- **Save**: 保存配置更改
- **Reload**: 重新加载配置
- **Validate**: 验证配置有效性
- **Get**: 获取配置值
- **Set**: 设置配置值
- **GetModuleConfig**: 获取模块配置
- **SetModuleConfig**: 设置模块配置
- **parseModuleConfig**: 解析模块配置
- **createDefaultConfigUnlocked**: 创建默认配置

### 2. 全局配置结构 ✅

**功能描述**: 定义应用程序的完整配置结构

**主要配置项**:
- **系统配置**: 基础系统设置
- **模块配置**: 各功能模块配置
- **用户配置**: 用户个性化设置
- **元数据**: 配置版本和元信息

### 3. 模块配置工厂 ✅

**功能描述**: 创建和管理不同类型的模块配置

**支持的模块**:
- **AI模块**: AI生成和训练相关配置
- **爬虫模块**: Pixiv爬虫相关配置
- **日志模块**: 日志系统配置

## 🚀 主要功能

### 1. 配置加载 ✅

**功能描述**: 从多种配置源加载配置信息

**支持格式**:
- **JSON格式**: 主要的配置文件格式
- **环境变量**: 支持环境变量覆盖
- **命令行参数**: 支持命令行参数覆盖
- **默认配置**: 提供默认配置值

**加载流程**:
- 加载默认配置
- 加载主配置文件
- 应用环境变量覆盖
- 应用命令行参数覆盖
- 验证配置有效性

### 2. 配置验证 ✅

**功能描述**: 验证配置参数的有效性和完整性

**验证项目**:
- **参数类型**: 检查参数数据类型
- **参数范围**: 验证参数取值范围
- **依赖关系**: 检查配置项之间的依赖关系
- **必需参数**: 验证必需参数是否存在

### 3. 动态配置 ✅

**功能描述**: 支持配置的动态更新和热重载

**动态特性**:
- **热重载**: 无需重启应用即可应用配置更改
- **配置监听**: 监听配置文件变化
- **实时更新**: 实时更新配置参数
- **回滚机制**: 支持配置回滚

### 4. 模块化配置 ✅

**功能描述**: 支持模块化的配置管理

**模块支持**:
- **爬虫模块**: Pixiv爬虫相关配置
- **AI模块**: AI生成和训练相关配置
- **日志模块**: 日志系统配置
- **工具模块**: 工具函数相关配置

## ⚙️ 配置参数

### 系统配置

**基础路径配置**:
- **base_dir**: 基础目录路径
- **config_dir**: 配置目录路径
- **data_dir**: 数据目录路径
- **log_dir**: 日志目录路径
- **temp_dir**: 临时目录路径
- **cache_dir**: 缓存目录路径

**安全配置**:
- **encryption_key**: 加密密钥
- **allowed_hosts**: 允许的主机列表
- **max_file_size**: 最大文件大小
- **max_memory**: 最大内存使用量
- **rate_limit**: 速率限制

**性能配置**:
- **max_workers**: 最大工作线程数
- **queue_size**: 队列大小
- **cache_size**: 缓存大小
- **batch_size**: 批处理大小
- **timeout**: 超时时间
- **retry_count**: 重试次数

### 爬虫模块配置

**Pixiv配置**:
- **cookie**: Pixiv登录Cookie
- **user_agent**: 用户代理字符串
- **accept**: 接受的内容类型
- **accept_language**: 接受的语言类型
- **referer**: 请求来源页面
- **timeout**: 请求超时时间
- **retry_count**: 重试次数
- **max_concurrent**: 最大并发数
- **default_delay**: 默认延迟时间
- **max_delay**: 最大延迟时间

**代理配置**:
- **enabled**: 是否启用代理
- **url**: 代理服务器地址

**API配置**:
- **search_tag**: 标签搜索API地址
- **user_works**: 用户作品API地址
- **illust_pages**: 插画页面API地址
- **artworks**: 作品详情API地址

**下载配置**:
- **save_dir**: 保存目录
- **filename_template**: 文件名模板
- **overwrite**: 是否覆盖已存在文件
- **create_subdirs**: 是否创建子目录
- **max_file_size**: 最大文件大小
- **allowed_types**: 允许的文件类型
- **skip_existing**: 是否跳过已存在文件
- **verify_checksum**: 是否验证校验和
- **progress_display**: 是否显示下载进度
- **auto_retry**: 是否自动重试

### AI模块配置

**Stable Diffusion配置**:
- **api_url**: WebUI API地址
- **timeout**: 请求超时时间
- **retry_count**: 重试次数
- **default_model**: 默认模型
- **default_sampler**: 默认采样器
- **default_steps**: 默认采样步数
- **default_cfg**: 默认CFG Scale
- **default_width**: 默认图像宽度
- **default_height**: 默认图像高度
- **max_batch_size**: 最大批次大小

**Kohya-ss配置**:
- **api_url**: 训练API地址
- **timeout**: 请求超时时间
- **retry_count**: 重试次数
- **default_epochs**: 默认训练轮数
- **default_batch_size**: 默认批次大小
- **default_lr**: 默认学习率

**OpenAI配置**:
- **api_keys**: API密钥列表
- **model**: 使用的模型
- **timeout**: 请求超时时间
- **retry_count**: 重试次数
- **max_tokens**: 最大令牌数
- **temperature**: 温度参数
- **rate_limit**: 速率限制

**WD14Tagger配置**:
- **model_path**: 模型文件路径
- **threshold**: 标签阈值
- **batch_size**: 批次大小
- **max_tags**: 最大标签数
- **skip_tags**: 跳过的标签列表
- **extend_tags**: 扩展的标签列表
- **tag_order**: 标签顺序
- **save_type**: 保存格式

### 日志模块配置

**日志级别配置**:
- **default**: 默认日志级别
- **modules**: 各模块的日志级别
- **dynamic**: 是否支持动态调整

**输出配置**:
- **console**: 是否输出到控制台
- **file**: 是否输出到文件
- **network**: 是否输出到网络
- **format**: 输出格式

**文件配置**:
- **path**: 日志文件路径
- **max_size**: 最大文件大小
- **max_backups**: 最大备份数量
- **max_age**: 最大保存天数
- **compress**: 是否压缩日志文件

**格式配置**:
- **type**: 格式类型
- **timestamp**: 是否包含时间戳
- **level**: 是否包含日志级别
- **module**: 是否包含模块名
- **caller**: 是否包含调用者信息
- **stack**: 是否包含堆栈信息
- **colors**: 是否使用颜色
- **full_time**: 是否使用完整时间格式

### 工具模块配置

**文件配置**:
- **max_file_size**: 最大文件大小
- **allowed_types**: 允许的文件类型
- **create_backup**: 是否创建备份
- **backup_suffix**: 备份文件后缀
- **temp_dir**: 临时目录
- **cleanup_temp**: 是否清理临时文件
- **cleanup_interval**: 清理间隔

**字符串配置**:
- **max_length**: 最大长度
- **sanitize_chars**: 需要清理的字符
- **replace_char**: 替换字符
- **trim_spaces**: 是否去除空格
- **normalize_unicode**: 是否标准化Unicode
- **case_sensitive**: 是否区分大小写

**网络配置**:
- **timeout**: 超时时间
- **retry_count**: 重试次数
- **user_agent**: 用户代理
- **follow_redirects**: 是否跟随重定向
- **max_redirects**: 最大重定向次数
- **proxy_url**: 代理地址
- **insecure_skip_verify**: 是否跳过SSL验证

**缓存配置**:
- **enabled**: 是否启用缓存
- **max_size**: 最大缓存大小
- **ttl**: 生存时间
- **cleanup_interval**: 清理间隔
- **storage**: 存储类型
- **path**: 缓存路径

## 🔄 工作流程

### 1. 配置加载流程
- 初始化默认配置
- 加载主配置文件
- 应用环境变量覆盖
- 应用命令行参数覆盖
- 验证配置有效性
- 初始化各模块配置

### 2. 配置更新流程
- 检测配置变化
- 验证新配置有效性
- 应用配置更改
- 通知相关模块
- 记录配置变更日志

### 3. 配置验证流程
- 检查配置格式
- 验证参数类型
- 检查参数范围
- 验证依赖关系
- 生成验证报告

## 🛡️ 错误处理

### 常见错误类型
- **配置文件不存在**: 配置文件路径错误
- **配置格式错误**: JSON格式不正确
- **参数验证失败**: 参数值不符合要求
- **依赖关系错误**: 配置项依赖关系不正确
- **权限不足**: 无法读取或写入配置文件

### 错误处理策略
- **默认值回退**: 使用默认配置值
- **配置验证**: 严格的配置验证机制
- **错误日志**: 详细的错误日志记录
- **用户提示**: 友好的错误提示信息
- **自动修复**: 尝试自动修复配置问题

## 📈 性能优化

### 1. 配置缓存
- **内存缓存**: 将配置缓存在内存中
- **文件缓存**: 缓存解析后的配置
- **增量更新**: 只更新变化的配置项

### 2. 配置预加载
- **预加载机制**: 预先加载常用配置
- **懒加载**: 按需加载配置项
- **批量加载**: 批量加载相关配置

### 3. 配置优化
- **配置压缩**: 压缩配置文件大小
- **配置合并**: 合并重复的配置项
- **配置清理**: 清理无用的配置项

## 🌐 API接口

### HTTP API端点

**配置文件管理**:
- `GET /api/configs` - 获取配置列表
- `POST /api/configs/create` - 创建配置
- `GET /api/configs/categories` - 获取配置分类
- `GET /api/configs/default` - 获取默认配置
- `POST /api/configs/default` - 设置默认配置
- `POST /api/configs/import` - 导入配置
- `POST /api/configs/export` - 导出配置
- `POST /api/configs/import-file` - 从文件导入配置
- `POST /api/configs/load-from-files` - 从文件加载配置
- `GET /api/configs/name/{name}` - 按名称获取配置
- `POST /api/configs/{id}/use` - 使用配置
- `GET /api/configs/{id}` - 获取配置详情
- `PUT /api/configs/{id}` - 更新配置
- `DELETE /api/configs/{id}` - 删除配置

**文件系统配置管理**:
- `POST /api/configs/file/create` - 创建配置文件
- `PUT /api/configs/file/{id}/update` - 更新配置文件
- `DELETE /api/configs/file/{id}/delete` - 删除配置文件

**配置管理**:
- `POST /api/config/get` - 获取配置
- `POST /api/config/update` - 更新配置

### 配置数据结构

**GenerationConfig配置结构**:
```json
{
  "id": "config_id",
  "name": "配置名称",
  "category": "分类",
  "description": "描述",
  "model": "模型文件",
  "loras": [
    {
      "path": "lora文件路径",
      "weight": 0.8
    }
  ],
  "steps": 20,
  "cfg_scale": 7,
  "width": 512,
  "height": 512,
  "sampler": "采样器",
  "enable_hr": false,
  "hr_scale": 2,
  "hr_steps": 0,
  "hr_upscaler": "upscaler",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

**API响应结构**:
```json
{
  "status": {
    "code": 200,
    "message": "success",
    "details": "操作成功"
  },
  "data": {
    // 具体数据内容
  }
}
```