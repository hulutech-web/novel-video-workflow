# Novel Video Workflow

这是一个自动化的小说视频生成工作流程，能够将小说文本转换为带有音频和字幕的视频内容。

## 功能特性

- 自动识别小说文本中的章节标记（如"第x章"）
- 创建对应的目录结构（如chapter_01, chapter_02等）
- 将章节内容保存到对应的文件中
- 使用AI生成高质量的语音音频
- 为音频生成时间轴匹配的字幕

## 目录结构

```
novel-video-workflow/
├── input/
│   └── 小说名称/
│       ├── chapter_01/
│       │   └── chapter_01.txt
│       ├── chapter_02/
│       │   └── chapter_02.txt
│       └── ...
├── output/
│   └── 小说名称/
│       ├── chapter_01/
│       │   ├── audio.wav
│       │   └── subtitles.srt
│       └── ...
└── assets/
    └── ref_audio/
        └── ref.m4a
```

## 快速开始

### 1. 安装依赖

确保已安装以下依赖：
- Go 1.19+
- Python 3.8+

### 2. 启动服务

在使用音频生成功能前，需要先启动依赖的服务：

```bash
# 在一个终端中启动IndexTTS2服务
cd /Users/mac/code/ai/tts/index-tts && python app.py
```

### 3. 准备小说文本

将小说文本放入 `input/小说名称/` 目录下，例如 `input/幽灵客栈/novel.txt`。

### 4. 拆分章节

使用项目中的工具自动拆分章节：

```bash
# 使用我们实现的文件管理工具
# 系统会自动识别"第x章"标记并创建对应的目录结构
```

### 5. 生成音频

运行音频生成程序：

```bash
go run cmd/process_chapter/main.go
```

### 6. 生成字幕

使用Aegisub工具生成SRT字幕文件。

## 核心功能

### 章节拆分

系统能够自动识别小说文本中的章节标记（如"第x章"），x可以是数字或汉字，获取这个数字，创建文件夹chapter_0x文件夹，然后阅读文本，将相应部分的内容放入chapter_0x.txt中，注意"第x章"内容也一并放入其中。

### 音频生成

使用Indextts2技术将文本转换为高质量语音，支持情感控制和音色克隆。

### 字幕生成

根据音频和原始文本生成时间轴匹配的SRT字幕文件。

## 使用MCP工具

项目集成了MCP（Model Context Protocol）工具，支持以下功能：

- `process_chapter`: 处理单个小说章节
- `generate_indextts2_audio`: 使用IndexTTS2生成音频
- `generate_subtitles_from_indextts2`: 为音频生成字幕

## 技术架构

- Go语言实现核心处理逻辑
- IndexTTS2用于语音合成
- Aegisub用于字幕生成
- MCP协议用于AI代理交互
