# Novel Video Workflow

一个用于将小说内容转换为视频的自动化工作流系统，集成了文本转语音、字幕生成等功能。

## 项目结构

```
novel-video-workflow/
├── assets/                 # 资源文件
│   └── ref_audio/          # 参考音频文件
├── docs/                   # 文档
├── examples/               # 示例文件
│   ├── demo_mcp_workflow.go
│   ├── test_mcp_protocol.go
│   └── test_mcp_workflow.go
├── output/                 # 输出目录
├── pkg/                    # 项目包
│   ├── mcp/                # MCP协议相关
│   │   ├── handler.go      # MCP处理器
│   │   └── server.go       # MCP服务器
│   ├── types/              # 类型定义
│   ├── utils/              # 实用工具
│   │   └── ollama_tool_processor.go # Ollama工具处理器
│   └── tools/              # 工具集合
│       ├── aegisub/        # Aegisub字幕生成
│       │   ├── aegisub_generator.go     # 字幕生成器
│       │   ├── aegisub_generator.lua    # Lua脚本
│       │   ├── aegisub_generator_test.go # 测试文件
│       │   ├── aegisub_integration.go   # 集成接口
│       │   ├── aegisub_example.go       # 示例代码
│       │   ├── aegisub_subtitle_gen.sh  # Shell脚本
│       │   ├── quick_audio_gen.sh       # 快速音频生成脚本
│       │   └── start_indextts2_mcp_new.sh # 启动脚本
│       ├── file/           # 文件操作工具
│       │   └── file.go
│       ├── image/          # 图像处理工具
│       │   └── image.go
│       ├── indextts2/      # IndexTTS2集成
│       │   └── client.go
│       ├── tts/            # TTS工具
│       │   ├── tts.go
│       │   └── tts_test.go
│       └── video/          # 视频处理工具
│           └── video.go
├── pkg/workflow/           # 工作流处理
│   ├── processor.go        # 工作流处理器
├── test_data/              # 测试数据
│   └── test_text.txt
├── config.yaml             # 配置文件
├── main.go                 # 主程序入口
└── setup.sh                # 初始化脚本
```

## 功能模块

### 1. Aegisub字幕生成器 (pkg/tools/aegisub/)
- 自动生成SRT字幕文件
- 支持音频时长分析
- 按文本字数占比分配时间段
- 提供Lua脚本和Shell脚本实现
- 在macOS上自动使用Python备用方案

### 2. IndexTTS2集成 (pkg/tools/indextts2/)
- 与IndexTTS2服务集成
- 支持高质量文本转语音
- 音频生成与处理

### 3. MCP协议支持 (pkg/mcp/)
- Model Context Protocol (MCP) 服务
- 支持多种工具集成
- 可扩展的工具注册机制

### 4. Ollama本地集成 (pkg/utils/ollama_tool_processor.go)
- 本地Ollama工具处理器
- 作为MCP工具的代理
- 支持本地AI功能调用

## 安装与使用

1. 确保安装了必要的依赖:
   - Go 1.19+
   - FFmpeg
   - Python 3.x

2. 运行主程序:
   ```bash
   go run main.go
   ```

3. 运行示例:
   ```bash
   go run examples/demo_mcp_workflow.go
   ```

## MCP服务

项目提供了多种MCP服务:
- `process_chapter`: 章节处理
- `generate_audio`: 音频生成
- `generate_indextts2_audio`: IndexTTS2音频生成
- `generate_subtitles_from_indextts2`: 字幕生成

## 配置

修改 `config.yaml` 来配置各项参数，包括API密钥、服务地址等。

## 注意事项

- 在macOS上，Aegisub字幕生成器会自动使用Python备用方案，避免GUI启动导致的卡顿
- 确保参考音频文件放置在 `assets/ref_audio/` 目录下
- 输出文件将保存在 `output/` 目录中