#!/bin/bash

# 小说视频工作流 - MCP服务启动脚本

echo "🚀 启动小说视频工作流所需的服务..."

# 检查并启动Ollama服务
echo "🔍 检查Ollama服务..."
if ! pgrep -f "ollama serve" > /dev/null; then
    echo "📡 启动Ollama服务 (端口: 11434)..."
    ollama serve > /tmp/ollama.log 2>&1 &
    sleep 3
else
    echo "✅ Ollama服务已在运行"
fi

# 检查Ollama模型
echo "🔍 检查Ollama模型..."
if ollama list | grep -q "llama3"; then
    echo "✅ llama3模型已存在"
else
    echo "📥 拉取llama3模型..."
    ollama pull llama3:8b
fi

# 检查并启动DrawThings (Stable Diffusion)
echo "🔍 检查DrawThings服务 (端口: 7861)..."
DRAWTHINGS_PID=$(lsof -ti:7861)
if [ -z "$DRAWTHINGS_PID" ]; then
    echo "🎨 启动DrawThings服务..."
    # 请根据您的Stable Diffusion WebUI路径修改以下命令
    # cd /path/to/stable-diffusion-webui && ./webui.sh --port 7861 > /tmp/drawthings.log 2>&1 &
    echo "⚠️  请手动启动DrawThings (Stable Diffusion WebUI) 端口7861"
else
    echo "✅ DrawThings服务已在运行 (PID: $DRAWTHINGS_PID)"
fi

# 检查并启动IndexTTS2
echo "🔍 检查IndexTTS2服务 (端口: 7860)..."
INDXTTS2_PID=$(lsof -ti:7860)
if [ -z "$INDXTTS2_PID" ]; then
    echo "🗣️  启动IndexTTS2服务..."
    # 请根据您的IndexTTS2路径修改以下命令
    # cd /path/to/indexTTS2 && python app.py --port 7860 > /tmp/indextts2.log 2>&1 &
    echo "⚠️  请手动启动IndexTTS2服务 端口7860"
else
    echo "✅ IndexTTS2服务已在运行 (PID: $INDXTTS2_PID)"
fi

# 检查Aegisub安装
echo "🔍 检查Aegisub..."
if [ -d "/Applications/Aegisub.app" ]; then
    echo "✅ Aegisub已安装"
else
    echo "⚠️  Aegisub未找到，请安装Aegisub应用"
fi

echo ""
echo "📋 服务状态概览:"
echo "- Ollama: $(if curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then echo '✅ 运行中'; else echo '❌ 未运行'; fi)"
echo "- DrawThings: $(if curl -s http://localhost:7861 > /dev/null 2>&1; then echo '✅ 运行中'; else echo '❌ 未运行'; fi)"
echo "- IndexTTS2: $(if curl -s http://localhost:7860 > /dev/null 2>&1; then echo '✅ 运行中'; else echo '❌ 未运行'; fi)"

echo ""
echo "💡 使用说明:"
echo "1. 确保以上所有服务都处于运行状态"
echo "2. 将小说文件放入 input/ 目录"
echo "3. 运行: go run cmd/test_workflow/main.go"
echo ""

# 等待用户按键继续
read -p "按任意键继续查看服务详细状态..." -n1 -s
echo ""

echo ""
echo "🔧 详细服务状态:"
echo "Ollama模型列表:"
ollama list | head -10

echo ""
echo "端口占用情况:"
lsof -i :7860,:7861,:11434 | grep LISTEN

echo ""
echo "🎉 服务启动脚本执行完成！"