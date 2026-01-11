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

## 小说文本处理

项目支持自动处理小说文本并按章节拆分存储。

### 输入目录结构

1. 将小说文本文件放在 `input/` 目录下，以小说名称创建子目录，例如 `input/幽灵客栈/`。
2. 系统会自动识别文本中的章节标记（如“第1章”、“第2章”等）并按章节拆分，创建 `chapter_01`、`chapter_02` 等子目录，每个目录中包含对应的章节文本文件。
3. 您也可以手动创建这些目录结构，只要确保章节目录命名格式为 `chapter_01`、`chapter_02` 等连续数字格式。

### 输出目录结构

处理后的输出文件会保存在 `output/` 目录下对应的子目录中，与输入目录结构相对应，便于管理和追踪处理结果。

### 使用方法

使用 [pkg/tools/file/file.go](file:///Users/mac/code/ai/novel-video-workflow/pkg/tools/file/file.go) 中的 [FileManager](file:///Users/mac/code/ai/novel-video-workflow/pkg/tools/file/file.go#L9-L11) 结构体提供的方法来处理小说文本：

- `CreateNovelInputStructure(novelName, novelText)` - 创建小说输入目录结构并拆分章节
- `CreateNovelOutputStructure(novelName)` - 创建对应的小说输出目录结构
- `GetNovelChaptersFromInput(novelName)` - 获取输入目录中的所有章节文件

## 使用 Ollama 本地服务

项目集成了 Ollama 作为本地 AI 服务，可以通过 Ollama 工具处理器调用各种功能。

### 启动 Ollama 服务

1. 确保已安装 Ollama：
   ```bash
   brew install ollama
   ```

2. 启动 Ollama 服务：
   ```bash
   ollama serve
   ```

3. 拉取所需模型（例如 llama3.1）：
   ```bash
   ollama pull llama3.1
   ```

### 使用 Ollama 工具处理器

Ollama 工具处理器作为 MCP 服务的代理，允许本地 Ollama 调用项目内的各种工具：

```bash
go run pkg/utils/ollama_tool_processor.go '{"name":"novel_video_workflow_generate_audio","arguments":{"text":"你好，这是一段测试文本","reference_audio":"./assets/ref_audio/ref.m4a","output_file":"./output/test.wav"}}'
```

### 可用的工具

- `novel_video_workflow_generate_audio`: 生成音频（映射到 `generate_indextts2_audio`）
- `process_chapter`: 处理单个章节
- `generate_subtitles_from_indextts2`: 生成字幕

### 配置说明

Ollama 工具处理器会自动为缺少参数的工具调用提供默认值：
- `reference_audio`: 默认为 `./assets/ref_audio/ref.m4a`
- `output_file`: 默认为 `./output/ollama_output_[时间戳].wav`

### 在 Ollama 对话框中使用工具

要在 Ollama 的对话框中使用这些工具，您需要理解这是一个自定义的 MCP 服务，仅适用于本项目环境。这些工具不是通用 AI 模型的内置功能，而是本地开发的工具接口。

1. **生成音频**：
   在本地 MCP 服务运行状态下，您可以通过 Ollama 工具处理器调用：
   ```
   调用 novel_video_workflow_generate_audio 工具将以下文本转换为音频："这里是要转换的文本内容"。使用默认参考音频，输出到 ./output/filename.wav
   ```

2. **处理章节**：
   ```
   调用 process_chapter 工具处理以下章节内容："这里是章节内容"
   ```

3. **生成字幕**：
   ```
   调用 generate_subtitles_from_indextts2 工具，基于音频文件 ./output/audio.wav 和对应文本生成字幕
   ```

**当前架构说明：**
- 目前 Ollama 无法像 Continue 等工具一样直接与 MCP 对话
- 我们的 ollama_tool_processor.go 充当代理角色，将 Ollama 的工具调用请求转发到本地 MCP 服务器
- 这种方式需要手动运行代理程序，不如原生 MCP 集成无缝
- 未来可考虑使用支持 MCP 协议的 IDE 或工具（如 Continue、Cursor 等）获得更好的集成体验

**注意事项：**
- 这些工具是本项目的自定义 MCP 接口，需要本地 MCP 服务支持
- 确保在发送提示词前，MCP 服务已启动 (`go run main.go mcp`)
- 工具调用需要适当的文件路径权限
- 输出文件将保存在项目 output 目录中

## 注意事项

- 在macOS上，Aegisub字幕生成器会自动使用Python备用方案，避免GUI启动导致的卡顿
- 确保参考音频文件放置在 `assets/ref_audio/` 目录下
- 输出文件将保存在 `output/` 目录中