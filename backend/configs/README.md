# 配置文件说明

## 概述

PixivTailor 使用 JSON 格式的配置文件来管理所有模块的设置，支持动态配置加载和前端配置管理。

## 配置文件结构

### 主配置文件 (config.json)
包含所有模块的完整配置，包括：
- 系统配置
- AI模块配置
- 爬虫模块配置
- 日志配置
- 工具模块配置

### 专用配置文件

#### 爬虫配置文件 (crawler_config.json)
专门用于爬虫模块的配置，包含：
- Pixiv API 配置
- 代理设置
- 下载配置
- 请求参数配置

#### 示例配置文件 (pixiv_example.json)
Pixiv 配置的示例文件，用于参考和快速配置。

## 配置加载优先级

1. **专用配置文件** - 如果存在专用配置文件，优先使用
2. **主配置文件** - 从主配置文件中读取对应模块的配置
3. **默认配置** - 如果配置文件不存在或格式错误，使用内置默认配置

## 前端配置管理

配置文件设计为前端友好的格式，支持：
- 实时配置修改
- 配置验证
- 配置导入导出
- 配置模板管理

## 配置验证

所有配置在加载时都会进行验证，包括：
- 数据类型检查
- 参数范围验证
- 依赖关系检查
- 必需参数验证

## 使用示例

### 从配置文件创建爬虫
```go
// 使用专用配置文件
crawler, err := NewCrawlerFromFile("configs/crawler_config.json")

// 使用默认配置
crawler, err := NewCrawler()

// 使用自定义配置
config := map[string]interface{}{
    "pixiv": map[string]interface{}{
        "cookie": "your_cookie_here",
        // ... 其他配置
    },
}
crawler, err := NewCrawlerWithConfig(config)
```

### 动态配置更新
```go
// 获取当前配置
currentConfig := crawler.GetConfig()

// 更新配置
newConfig := map[string]interface{}{
    "pixiv": map[string]interface{}{
        "timeout": 60,
        "retry_count": 5,
    },
}
err := crawler.UpdateConfig(newConfig)
```

## 注意事项

1. 配置文件使用 UTF-8 编码
2. JSON 格式必须正确，否则会使用默认配置
3. 敏感信息（如 Cookie）请妥善保管
4. 修改配置后建议重启服务以确保生效
5. 定期备份配置文件
