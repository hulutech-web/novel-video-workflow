# Novel Video Workflow (精简版)

## 项目结构

清理后的精简项目包含以下核心文件：

### 核心代码
- [main.go](file:///Users/mac/code/ai/novel-video-workflow/main.go) - 主程序入口
- [config.yaml](file:///Users/mac/code/ai/novel-video-workflow/config.yaml) - 配置文件
- [internal/mcp/server.go](file:///Users/mac/code/ai/novel-video-workflow/internal/mcp/server.go) - MCP服务器
- [internal/mcp/handler.go](file:///Users/mac/code/ai/novel-video-workflow/internal/mcp/handler.go) - MCP请求处理器
- [internal/tools/tts.go](file:///Users/mac/code/ai/novel-video-workflow/internal/tools/tts.go) - TTS处理器
- [internal/tools/indextts2/client.go](file:///Users/mac/code/ai/novel-video-workflow/internal/tools/indextts2/client.go) - IndexTTS2客户端
- [internal/tools/tts_test.go](file:///Users/mac/code/ai/novel-video-workflow/internal/tools/tts_test.go) - TTS测试

### 脚本
- [scripts/start_indextts2_mcp_new.sh](file:///Users/mac/code/ai/novel-video-workflow/scripts/start_indextts2_mcp_new.sh) - 启动MCP服务器
- [scripts/quick_audio_gen.sh](file:///Users/mac/code/ai/novel-video-workflow/scripts/quick_audio_gen.sh) - 快速音频生成脚本
- [scripts/ollama_tool_processor.go](file:///Users/mac/code/ai/novel-video-workflow/scripts/ollama_tool_processor.go) - Ollama工具处理器

### 资源文件
- [ref.m4a](file:///Users/mac/code/ai/novel-video-workflow/ref.m4a) - 参考音频文件
- [音色.m4a](file:///Users/mac/code/ai/novel-video-workflow/音色.m4a) - 参考音频文件
- [output/](file:///Users/mac/code/ai/novel-video-workflow/internal/tools/subtitle.go#L30-L30) - 生成的音频文件输出目录

## 使用方法

### 1. 启动服务
```bash
# 启动IndexTTS2服务
cd /Users/mac/code/ai/tts/index-tts
python app.py

# 启动MCP服务器
cd /Users/mac/code/ai/novel-video-workflow
./scripts/start_indextts2_mcp_new.sh
```

### 2. 生成音频
```bash
cd /Users/mac/code/ai/novel-video-workflow
./scripts/quick_audio_gen.sh "您想要转换为语音的文本"
```

## 清理说明

已删除的文件包括：
- 临时和测试文件
- 开发过程文档
- 多余的脚本和配置文件
- 保留了所有核心功能所需文件

项目现在更简洁，但保持了全部功能。