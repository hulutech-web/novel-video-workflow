#!/bin/bash

echo "安装小说视频工作流..."

# 1. 安装Go（如果未安装）
if ! command -v go &> /dev/null; then
    echo "安装Go..."
    brew install go
fi

# 2. 安装依赖
echo "安装Go依赖..."
go mod tidy

# 3. 安装Python依赖（用于indexTTS）
echo "安装Python依赖..."
pip3 install torch torchaudio
pip3 install git+https://github.com/IndexTTS/IndexTTS.git

# 4. 安装其他工具
echo "安装其他工具..."
brew install ffmpeg

# 5. 创建输出目录
mkdir -p ./output
mkdir -p ./assets

# 6. 配置Ollama（如果已安装）
if command -v ollama &> /dev/null; then
    echo "检测到Ollama已安装"
    echo "Ollama版本: $(ollama -v)"
    # 检查是否需要启动Ollama服务
    if ! pgrep -f "ollama serve" > /dev/null; then
        echo "请确保Ollama服务正在运行: ollama serve"
    fi
else
    echo "Ollama未安装，请访问 https://ollama.com/ 下载安装"
    echo "或者运行: brew install ollama"
fi

echo "安装完成！"
echo "使用方法："
echo "1. 启动Ollama服务: ollama serve"
echo "2. 使用ollama_tool_processor作为本地MCP服务"
echo "3. 在Ollama中可以调用以下工具:"
echo "   - generate_indextts2_audio: 生成音频"
echo "   - process_chapter: 处理单个章节"
echo "   - generate_subtitles_from_indextts2: 生成字幕"