# 快速入门指南

本指南将帮助您快速搭建并运行小说视频生成工作流。

## 环境准备

### 1. 安装基础依赖

```bash
# 确保安装了 Go 1.19+
go version

# 安装 FFmpeg (用于音频处理)
# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt update && sudo apt install ffmpeg

# Windows (使用 Chocolatey)
choco install ffmpeg
```

### 2. 准备服务依赖

#### 启动 Ollama
```bash
# 安装 Ollama (如果尚未安装)
# macOS
brew install ollama

# 启动 Ollama 服务
ollama serve

# 在新的终端窗口中拉取模型
ollama pull qwen3:4b
```

#### 启动 Stable Diffusion WebUI (DrawThings)
```bash
# 克隆仓库
git clone https://github.com/AUTOMATIC1111/stable-diffusion-webui.git
cd stable-diffusion-webui

# 启动服务
python launch.py --port 7861
```

#### 启动 IndexTTS2
```bash
# 克隆 IndexTTS2 仓库
git clone https://github.com/your-index-tts-repo/index-tts.git
cd index-tts

# 启动服务
python app.py --port 7860
```

## 项目设置

### 1. 克隆项目
```bash
git clone https://github.com/your-repo/novel-video-workflow.git
cd novel-video-workflow
```

### 2. 安装 Go 依赖
```bash
go mod tidy
```

### 3. 准备参考音频
```bash
# 创建音频目录
mkdir -p assets/ref_audio

# 放置您的参考音频文件 (如 ref.m4a)
# 音频文件应放在 ./assets/ref_audio/ref.m4a
```

## 准备输入数据

### 1. 创建输入目录结构
```bash
mkdir -p input/我的小说/chapter_01
```

### 2. 添加小说文本
```bash
# 创建小说章节文件
cat > input/我的小说/chapter_01/chapter_01.txt << EOF
第1章

这里是您的小说第一章内容...

EOF
```

## 运行工作流

### 1. 检查服务状态
```bash
# 运行自检程序
go run cmd/test_workflow/main.go
```

如果所有服务都正常，您应该看到类似以下的输出：
```
✅ Ollama 服务可用
✅ DrawThings 服务可用
✅ IndexTTS2 服务可用
✅ Aegisub 脚本可用
```

### 2. 执行完整工作流
```bash
go run cmd/test_workflow/main.go
```

## 输出检查

### 1. 检查生成的文件
```bash
ls -la output/我的小说/chapter_01/
```

您应该看到：
- `chapter_01.wav` - 生成的音频文件
- `chapter_01.srt` - 生成的字幕文件
- `images/` - 包含生成的图像

### 2. 检查处理进度
```bash
go run cmd/check_progress/main.go
```

## 常见问题

### 服务连接失败
- 确认所有服务都在正确的端口上运行
- 检查防火墙设置

### 音频生成失败
- 确认参考音频文件存在
- 检查 IndexTTS2 服务状态

### 图像生成失败
- 确认 Stable Diffusion WebUI 正在运行
- 检查模型是否正确加载

### 字幕生成失败
- 确认音频文件已生成
- 检查 Aegisub 脚本路径

## 故障排除

### 检查服务状态
```bash
# 检查 Ollama
curl http://localhost:11434/api/tags

# 检查 DrawThings
curl http://localhost:7861/sdapi/v1/sd-models

# 检查 IndexTTS2
curl http://localhost:7860
```

### 查看详细日志
```bash
# 运行并查看详细输出
go run cmd/test_workflow/main.go 2>&1 | tee workflow.log
```

## 扩展功能

### 处理多个章节
创建多个章节目录：
```bash
mkdir -p input/我的小说/chapter_02
echo "第2章

这里是第二章内容..." > input/我的小说/chapter_02/chapter_02.txt
```

### 自定义配置
根据需要修改 [config.yaml](file:///Users/mac/code/ai/novel-video-workflow/config.yaml) 中的参数设置。

## 下一步

- 阅读 [MCP_SERVICES_GUIDE.md](file:///Users/mac/code/ai/novel-video-workflow/MCP_SERVICES_GUIDE.md) 了解详细的服务配置
- 查看 [USAGE.md](file:///Users/mac/code/ai/novel-video-workflow/USAGE.md) 了解高级用法
- 探索 [PROCESSING_WORKFLOW.md](file:///Users/mac/code/ai/novel-video-workflow/PROCESSING_WORKFLOW.md) 了解处理流程