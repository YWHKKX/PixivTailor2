#!/bin/bash

# PixivTailor 构建脚本

set -e

echo "🚀 开始构建 PixivTailor..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ Go环境未安装，请先安装Go 1.24+"
    exit 1
fi

# 检查Go版本
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.24"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "❌ Go版本过低，需要 $REQUIRED_VERSION+，当前版本: $GO_VERSION"
    exit 1
fi

echo "✅ Go环境检查通过: $GO_VERSION"

# 进入源码目录
cd src

# 下载依赖
echo "📦 下载依赖..."
go mod tidy

# 构建项目
echo "🔨 构建项目..."
go build -o ../pixiv-tailor main.go

# 检查构建结果
if [ -f "../pixiv-tailor" ]; then
    echo "✅ 构建成功！"
    echo "📁 可执行文件: ./pixiv-tailor"
    
    # 显示帮助信息
    echo ""
    echo "📖 使用帮助:"
    echo "  ./pixiv-tailor --help"
    echo ""
    echo "🎯 快速开始:"
    echo "  # 爬取Pixiv图片"
    echo "  ./pixiv-tailor crawl --type tag --query \"エルザ・グランヒルテ\" --limit 100"
    echo ""
    echo "  # 生成AI图像"
    echo "  ./pixiv-tailor generate --model \"chosenMix_bakedVae.safetensors\" --prompt \"1girl, beautiful\""
    echo ""
    echo "  # 训练LoRA模型"
    echo "  ./pixiv-tailor train --name \"my_model\" --data-dir \"data/images/train\""
    echo ""
    echo "  # 生成图像标签"
    echo "  ./pixiv-tailor tag --input-dir \"data/images\" --output-dir \"data/tags\""
    echo ""
    echo "  # 分类标签"
    echo "  ./pixiv-tailor classify --input \"data/tags\" --output \"global_configs/global_tags.json\""
    echo ""
    echo "🔧 配置说明:"
    echo "  - 编辑 global_configs/config.json 配置文件"
    echo "  - 设置Pixiv Cookie和API密钥"
    echo "  - 配置Stable Diffusion WebUI和Kohya-ss服务地址"
else
    echo "❌ 构建失败！"
    exit 1
fi
