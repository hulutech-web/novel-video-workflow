#!/bin/bash

# 启动 Novel Video Workflow 的 IndexTTS2 MCP 服务器
# 这个脚本将启动一个MCP服务器，使AI代理可以调用IndexTTS2功能

set -e

echo "🚀 启动 Novel Video Workflow IndexTTS2 MCP 服务器..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装Go语言环境"
    exit 1
fi

# 检查是否在项目根目录
if [ ! -f "go.mod" ] || [ ! -f "main.go" ]; then
    echo "❌ 请在项目根目录执行此脚本"
    exit 1
fi

# 检查IndexTTS2服务是否运行
TTS_URL="http://localhost:7860"
echo "🔍 检查 IndexTTS2 服务: $TTS_URL"

if ! curl -s --connect-timeout 5 $TTS_URL >/dev/null 2>&1; then
    echo "⚠️  IndexTTS2 服务未运行或不可访问"
    echo "💡 提示: 如果您还没有启动IndexTTS2服务，请在另一个终端运行:"
    echo "   cd <YOUR_INDEX_TTS_PATH> && python app.py"
else
    echo "✅ IndexTTS2 服务可访问"
fi

echo "🔧 编译并启动 Novel Video Workflow MCP 服务器..."

# 设置环境变量
export NOVEL_VIDEO_WORKFLOW_MCP=true

# 运行MCP服务器
exec go run main.go mcp