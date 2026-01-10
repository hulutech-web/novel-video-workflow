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
brew install --cask aegisub
brew install ffmpeg
brew install drawthings

# 5. 创建配置文件
mkdir -p ~/.novel-video
cp config.yaml ~/.novel-video/

# 6. 设置Claude Desktop MCP
CLAUDE_CONFIG="$HOME/Library/Application Support/Claude/claude_desktop_config.json"
if [ -f "$CLAUDE_CONFIG" ]; then
    echo "配置Claude Desktop MCP..."
    # 备份原配置
    cp "$CLAUDE_CONFIG" "${CLAUDE_CONFIG}.bak"

    # 添加MCP配置
    python3 -c "
import json
import os

config_path = '$CLAUDE_CONFIG'
with open(config_path, 'r') as f:
    config = json.load(f)

project_path = os.path.abspath('.')
mcp_config = {
    'mcpServers': {
        'novel-video': {
            'command': 'go',
            'args': ['run', os.path.join(project_path, 'main.go')]
        }
    }
}

config.update(mcp_config)
with open(config_path, 'w') as f:
    json.dump(config, f, indent=2)
"
fi

echo "安装完成！"
echo "使用方法："
echo "1. 启动Claude Desktop"
echo "2. 在聊天中可以使用："
echo "   - process_chapter: 处理单个章节"
echo "   - batch_process: 批量处理小说"
echo "   - generate_audio: 生成音频"