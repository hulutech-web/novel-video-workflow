# Novel Video Workflow

## 项目概述

本项目提供了一个完整的本地小说视频生成工作流，包括音频生成、图像生成和视频合成功能。特别地，`novel_video_workflow_generate_audio` 功能允许您使用IndexTTS2生成高质量的语音合成。

## 项目结构

- [main.go](file:///Users/mac/code/ai/novel-video-workflow/main.go) - 主程序入口
- [config.yaml](file:///Users/mac/code/ai/novel-video-workflow/config.yaml) - 配置文件
- [internal/](file:///Users/mac/code/ai/novel-video-workflow/internal/mcp/server.go) - 核心功能模块
  - [mcp/](file:///Users/mac/code/ai/novel-video-workflow/internal/mcp/server.go) - MCP服务器实现
  - [tools/](file:///Users/mac/code/ai/novel-video-workflow/internal/tools/tts.go) - 各种工具实现
    - [indextts2/client.go](file:///Users/mac/code/ai/novel-video-workflow/internal/tools/indextts2/client.go) - IndexTTS2客户端
    - [tts.go](file:///Users/mac/code/ai/novel-video-workflow/internal/tools/tts.go) - TTS处理器
    - [tts_test.go](file:///Users/mac/code/ai/novel-video-workflow/internal/tools/tts_test.go) - TTS测试
- [scripts/](file:///Users/mac/code/ai/novel-video-workflow/scripts/start_indextts2_mcp_new.sh) - 实用脚本
  - [start_indextts2_mcp_new.sh](file:///Users/mac/code/ai/novel-video-workflow/scripts/start_indextts2_mcp_new.sh) - 启动MCP服务器
  - [quick_audio_gen.sh](file:///Users/mac/code/ai/novel-video-workflow/scripts/quick_audio_gen.sh) - 快速音频生成
  - [ollama_tool_processor.go](file:///Users/mac/code/ai/novel-video-workflow/scripts/ollama_tool_processor.go) - Ollama工具处理器
- [ref.m4a](file:///Users/mac/code/ai/novel-video-workflow/ref.m4a) / [音色.m4a](file:///Users/mac/code/ai/novel-video-workflow/音色.m4a) - 参考音频文件
- [output/](file:///Users/mac/code/ai/novel-video-workflow/internal/tools/subtitle.go#L30-L30) - 生成的音频文件输出目录

## 系统要求

- Go 1.19+
- Python 3.8+
- 至少8GB RAM（推荐16GB+用于IndexTTS2）
- IndexTTS2服务已安装并运行

## 安装和使用

### 1. 启动IndexTTS2服务
```bash
cd /Users/mac/code/ai/tts/index-tts
python app.py
```

### 2. 启动MCP服务器
```bash
cd /Users/mac/code/ai/novel-video-workflow
./scripts/start_indextts2_mcp_new.sh
```

### 3. 生成音频（命令行）
```bash
cd /Users/mac/code/ai/novel-video-workflow
./scripts/quick_audio_gen.sh "您想要转换为语音的文本"
```

例如：
```bash
./scripts/quick_audio_gen.sh "欢迎使用小说视频生成工具，这是一个完全本地的解决方案。"
```

## 主要功能

- **novel_video_workflow_generate_audio**: 使用IndexTTS2生成高质量音频文件，支持声音克隆和情感控制
- **完全本地运行**: 所有处理都在您的设备上完成
- **无费用**: 不需要付费订阅
- **无Token限制**: 不受API调用次数限制
- **隐私保护**: 数据不会离开您的设备
- **离线可用**: 不需要网络连接

## 输出文件

生成的音频文件将保存在 [output/](file:///Users/mac/code/ai/novel-video-workflow/internal/tools/subtitle.go#L30-L30) 目录中。

## 故障排除

### 音频生成失败

1. 检查IndexTTS2服务是否运行
2. 确认参考音频文件存在
3. 检查是否有足够的内存资源

### 输出文件未生成

1. 检查output目录权限
2. 确认项目根目录有写入权限