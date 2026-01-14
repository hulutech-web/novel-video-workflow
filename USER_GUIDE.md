# 小说视频工作流系统用户手册

## 系统要求

### 硬件要求
- CPU: 支持AVX指令集（Intel/AMD 64-bit）
- 内存: 16GB以上（推荐32GB）
- 存储: 10GB以上可用空间
- GPU: NVIDIA GPU (CUDA支持，推荐8GB显存以上) 或 Apple Silicon (Metal支持)

### 软件要求
- Go 1.25或更高版本
- Python 3.8或更高版本
- Git
- Ollama (用于AI模型)
- Stable Diffusion WebUI (用于图像生成)

## 安装指南

### 1. 克隆项目
```bash
git clone <your-repo-url>
cd novel-video-workflow
```

### 2. 安装Go依赖
```bash
go mod tidy
```

### 3. 安装和配置依赖服务

#### 3.1 安装Ollama
```bash
# macOS
brew install ollama

# Linux
curl -fsSL https://ollama.ai/install.sh | sh

# 启动Ollama
ollama serve
```

下载所需模型：
```bash
ollama pull qwen3:4b
ollama pull llama3.2
```

#### 3.2 安装Stable Diffusion WebUI
```bash
git clone https://github.com/AUTOMATIC1111/stable-diffusion-webui.git
cd stable-diffusion-webui
./webui.sh
```

#### 3.3 安装IndexTTS2 (可选)
如果您需要TTS功能，需要安装IndexTTS2服务。

### 4. 配置系统

#### 4.1 配置config.yaml
```yaml
# config.yaml 示例配置
services:
  ollama:
    url: "http://localhost:11434"
    model: "qwen3:4b"
  stable_diffusion:
    url: "http://localhost:7861"
  indextts2:
    url: "http://localhost:7860"

paths:
  input_dir: "./input"
  output_dir: "./output"
  assets_dir: "./assets"
  ref_audio: "./assets/ref_audio/ref.m4a"

settings:
  image_width: 512
  image_height: 896
  audio_format: "wav"
```

#### 4.2 准备参考音频
在 `assets/ref_audio/` 目录下放置参考音频文件（.m4a或.wav格式），用于TTS声音克隆。

## 系统启动

### 1. 启动完整服务
```bash
# 启动MCP和Web服务
go run main.go
```

### 2. 启动特定服务
```bash
# 仅启动MCP服务
go run main.go mcp

# 仅启动Web服务
go run main.go web

# 批量处理模式
go run main.go batch

# 查看帮助
go run main.go help
```

## 使用Web界面

### 1. 访问界面
打开浏览器访问：http://localhost:8080

### 2. 界面功能

#### 2.1 仪表板
- 显示当前系统状态
- 可执行所有MCP工具

#### 2.2 MCP工具
- 显示所有可用的MCP工具
- 可单独执行每个工具
- 特殊工具支持自定义参数：
  - `generate_indextts2_audio`: 可输入文本生成音频
  - `generate_images_from_chapter_with_ai_prompt`: 可输入章节文本生成图像

#### 2.3 上传并处理
- 支持拖拽上传文件夹
- 批量处理小说文件

#### 2.4 教程
- 使用指南和说明

### 3. 实时日志
- 底部控制台显示实时日志
- 不同颜色标识不同类型消息：
  - 绿色：信息消息
  - 亮绿：成功消息
  - 红色：错误消息

## 主要功能使用

### 1. 文本转音频
1. 在MCP工具页面找到 `generate_indextts2_audio`
2. 点击"执行工具"
3. 输入要转换的文本
4. 设置输出目录
5. 点击"生成音频"

### 2. 章节转图像 (AI提示词)
1. 在MCP工具页面找到 `generate_images_from_chapter_with_ai_prompt`
2. 点击"执行工具"
3. 输入章节文本内容
4. 设置输出目录、图像宽度和高度
5. 点击"生成图像"

### 3. 批量处理小说
1. 在"上传并处理"页面
2. 拖拽小说文件夹到上传区域
3. 点击"处理上传的文件夹"
4. 系统将自动执行完整工作流

## 工具功能详解

### 1. generate_indextts2_audio
- 功能：使用IndexTTS2生成音频
- 参数：
  - text: 输入文本
  - reference_audio: 参考音频路径
  - output_file: 输出文件路径

### 2. generate_subtitles_from_indextts2
- 功能：生成字幕文件
- 参数：
  - audio_file: 音频文件路径
  - text: 原始文本
  - output_file: 输出字幕文件

### 3. file_split_novel_into_chapters
- 功能：将小说分割为章节
- 参数：
  - novel_path: 小说文件路径
  - output_dir: 输出目录

### 4. generate_image_from_text
- 功能：根据文本生成图像
- 参数：
  - text: 描述文本
  - output_file: 输出图像路径
  - width, height: 图像尺寸

### 5. generate_image_from_image
- 功能：图像风格转换
- 参数：
  - init_image_path: 输入图像
  - prompt: 提示词
  - output_file: 输出图像

### 6. generate_images_from_chapter
- 功能：章节转图像
- 参数：
  - chapter_text: 章节文本
  - output_dir: 输出目录

### 7. generate_images_from_chapter_with_ai_prompt
- 功能：AI智能提示词生成图像
- 参数：
  - chapter_text: 章节文本
  - output_dir: 输出目录
  - width, height: 图像尺寸

## 配置说明

### config.yaml 详细配置

```yaml
# 服务配置
services:
  ollama:
    url: "http://localhost:11434"  # Ollama服务地址
    model: "qwen3:4b"             # 使用的模型
    timeout: 300                   # 请求超时时间（秒）
  
  stable_diffusion:
    url: "http://localhost:7861"   # SD WebUI地址
    timeout: 120                   # 请求超时时间
    sampler: "Euler a"             # 采样器
    steps: 20                      # 采样步数
  
  indextts2:
    url: "http://localhost:7860"   # IndexTTS2服务地址
    timeout: 180                   # 请求超时时间

# 路径配置
paths:
  input_dir: "./input"             # 输入目录
  output_dir: "./output"           # 输出目录
  assets_dir: "./assets"           # 资源目录
  ref_audio: "./assets/ref_audio/ref.m4a"  # 默认参考音频

# 图像设置
image_settings:
  default_width: 512               # 默认宽度
  default_height: 896              # 默认高度
  quality: 95                      # JPEG质量（1-100）
  format: "png"                    # 默认格式

# 音频设置
audio_settings:
  format: "wav"                    # 音频格式
  sample_rate: 44100               # 采样率
  channels: 1                      # 声道数

# 工作流设置
workflow:
  concurrent_tasks: 2              # 并发任务数
  temp_dir: "./temp"               # 临时目录
  cleanup_after: true              # 处理完成后清理临时文件
```

## 故障排除

### 1. 服务启动失败
- 检查依赖服务是否运行
- 检查配置文件路径是否正确
- 检查防火墙设置

### 2. API连接失败
- 检查API端点URL
- 检查网络连接
- 检查服务是否正常运行

### 3. 生成质量不佳
- 调整模型参数
- 更换参考音频
- 优化提示词

### 4. 性能问题
- 检查硬件资源
- 调整并发设置
- 优化图像分辨率

## 开发扩展

### 添加新工具
1. 在 `pkg/mcp/handler.go` 中添加处理函数
2. 使用 `Handle{ToolName}Direct` 格式命名
3. 在MCP服务器中注册工具
4. 如需要，更新Web界面支持

### 自定义工作流
1. 修改 `pkg/workflow/processor.go`
2. 调整工具执行顺序
3. 添加自定义逻辑

## 技术支持

如遇问题，请检查：
1. 系统日志输出
2. 依赖服务状态
3. 配置文件设置
4. 网络连接状况

更多技术支持请参考相关文档或联系开发者。