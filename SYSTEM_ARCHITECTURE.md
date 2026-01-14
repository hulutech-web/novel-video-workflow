# 小说视频工作流系统架构说明

## 项目概述
这是一个基于AI的小说转视频自动化生成系统，集成了多种AI工具（TTS、图像生成等），能够将小说文本转换为带有音频和图像的视频内容。

## 系统架构

### 核心组件
1. **MCP服务器** (`pkg/mcp`) - 提供标准化工具接口
2. **Web服务器** (`cmd/web_server`) - 提供用户界面
3. **工作流处理器** (`pkg/workflow`) - 协调各组件工作
4. **工具适配器** (`pkg/tools`) - 封装具体AI服务调用

### 目录结构
```
novel-video-workflow/
├── cmd/                      # 命令行工具入口
│   ├── full_workflow/        # 完整工作流执行器
│   ├── ollama_mcp_bridge/    # Ollama MCP桥接器
│   └── web_server/           # Web服务器
├── pkg/                      # 核心包
│   ├── mcp/                  # MCP协议实现
│   ├── tools/                # 各类AI工具适配器
│   │   ├── drawthings/       # 图像生成工具
│   │   ├── indextts2/        # TTS语音合成
│   │   └── aegisub/          # 字幕处理
│   ├── workflow/             # 工作流处理器
│   └── utils/                # 工具函数
├── assets/                   # 静态资源
│   └── ref_audio/            # 参考音频
├── input/                    # 输入小说目录
├── output/                   # 输出目录
├── config.yaml               # 系统配置
└── main.go                   # 主程序入口
```

## 详细文件说明

### 主程序入口
- **`main.go`** - 系统主入口，支持多种运行模式：
  - `mcp`: 启动MCP服务器供AI代理调用
  - `web`: 启动Web界面供用户操作
  - `batch`: 批量处理模式
  - 默认: 同时启动MCP和Web服务

### MCP服务器组件 (`pkg/mcp/`)
- **`server.go`** - MCP协议服务器核心实现
- **`handler.go`** - 工具调用处理器，包含所有可用工具的实现
- **`adapter.go`** - MCP协议适配器

### 工具组件 (`pkg/tools/`)

#### 图像生成 (`pkg/tools/drawthings/`)
- **`client.go`** - Stable Diffusion WebUI API客户端
- **`ollama_client.go`** - Ollama API客户端，用于智能提示词生成
- **`chapter_image_generator.go`** - 章节图像生成器，将文本转换为图像

#### TTS语音合成 (`pkg/tools/indextts2/`)
- **`client.go`** - IndexTTS2 API客户端，用于语音合成

#### 字幕处理 (`pkg/tools/aegisub/`)
- **`aegisub_generator.go`** - Aegisub字幕生成器
- **`quick_audio_gen.sh`** - 快速音频生成脚本

### 工作流处理器 (`pkg/workflow/`)
- **`processor.go`** - 核心工作流处理器，协调各组件执行

### Web服务器 (`cmd/web_server/`)
- **`web_server.go`** - Web界面实现，包含：
  - 实时日志展示
  - 工具调用界面
  - 文件上传功能
  - WebSocket实时通信

### 配置文件
- **`config.yaml`** - 系统配置，包含API端点、模型设置等

### 资源文件
- **`assets/ref_audio/`** - 参考音频文件，用于TTS克隆声音
- **`input/`** - 小说输入目录
- **`output/`** - 生成结果输出目录

## 系统功能

### 主要工具
1. **`generate_indextts2_audio`** - 生成TTS音频
2. **`generate_subtitles_from_indextts2`** - 生成字幕
3. **`file_split_novel_into_chapters`** - 分割小说章节
4. **`generate_image_from_text`** - 文本转图像
5. **`generate_image_from_image`** - 图像转图像
6. **`generate_images_from_chapter`** - 章节转图像
7. **`generate_images_from_chapter_with_ai_prompt`** - AI智能提示词图像生成

### 运行模式

#### 1. MCP服务器模式
```bash
go run main.go mcp
```
启动MCP服务器，供AI代理调用工具。

#### 2. Web界面模式
```bash
go run main.go web
```
启动Web界面，地址：http://localhost:8080

#### 3. 批量处理模式
```bash
go run main.go batch
```
执行完整批处理工作流。

#### 4. 默认模式（推荐）
```bash
go run main.go
```
同时启动MCP和Web服务。

## 使用方法

### 环境准备
1. 安装Go 1.25+
2. 安装Ollama并下载模型
3. 安装Stable Diffusion WebUI
4. 安装IndexTTS2服务

### 快速开始
1. **准备参考音频**
   - 将参考音频文件放入 `assets/ref_audio/` 目录
   - 推荐使用高质量人声音频

2. **配置系统**
   - 修改 `config.yaml` 中的API端点和参数

3. **启动系统**
   ```bash
   go run main.go
   ```

4. **使用Web界面**
   - 访问 http://localhost:8080
   - 上传小说文件夹或使用单个工具

5. **使用MCP服务**
   - AI代理可通过MCP协议调用各种工具

### 工作流程
1. **输入** - 提供小说文本
2. **预处理** - 分割章节、提取关键信息
3. **音频生成** - 使用TTS生成语音
4. **图像生成** - 使用AI生成匹配图像
5. **后期处理** - 生成字幕、合成视频
6. **输出** - 生成最终视频内容

## 部署说明

### 本地部署
1. 克隆仓库
2. 安装依赖：`go mod tidy`
3. 启动服务：`go run main.go`

### 服务配置
- Web服务器：http://localhost:8080
- MCP服务器：stdio协议（通过MCP客户端调用）
- Ollama服务：http://localhost:11434
- SD WebUI：http://localhost:7861
- IndexTTS2：http://localhost:7860

## 开发说明

### 添加新工具
1. 在 `pkg/mcp/handler.go` 中添加工具处理函数
2. 注册工具到MCP服务器
3. 如需要，更新Web界面以支持新工具

### 测试方法
- 单元测试：`go test ./...`
- 集成测试：使用Web界面进行端到端测试
- 工具测试：直接调用MCP工具

### 日志系统
- 使用zap日志库
- 支持结构化日志
- Web界面实时显示日志

## 注意事项
1. 确保所有依赖服务正常运行
2. 配置文件路径使用相对路径以提高移植性
3. 音频和图像生成需要较长时间，请耐心等待
4. 大文件处理时注意内存使用情况