# Novel Video Workflow 使用说明

## 概述

这是一个自动化的小说视频生成工作流程，能够将小说文本转换为带有音频和字幕的视频内容。整个流程包括三个主要步骤：

1. **file mcp** - 将文字分出章节
2. **indexTTS2** - 生成音频
3. **aegisub** - 生成字幕

## 系统要求

- Go 1.19+
- Python 3.8+
- IndexTTS2 服务 (运行在 http://localhost:7860)
- Aegisub (用于字幕生成)

## 安装依赖

```bash
# 运行安装脚本
bash setup.sh
```

## 项目结构

```
novel-video-workflow/
├── input/                    # 输入目录
│   └── 小说名称/
│       ├── chapter_01/
│       │   └── chapter_01.txt
│       └── ...
├── output/                   # 输出目录
│   └── 小说名称/
│       └── chapter_01/
│           ├── audio.wav    # 生成的音频
│           └── subtitles.srt # 生成的字幕
└── assets/
    └── ref_audio/           # 参考音频
        └── ref.m4a
```

## 三步工作流程详解

### 步骤 1: file mcp - 拆分章节

#### 功能描述
- 识别小说文本中的章节标记（如"第x章"），x可以是数字也可以是汉字
- 获取章节号并创建对应的文件夹（如`chapter_0x`）
- 将相应部分的内容放入`chapter_0x.txt`中
- 保留"第x章"内容本身

#### 使用的工具
- 工具名称: `file_split_novel_into_chapters`
- 参数: `novel_path` (小说文件路径)

#### 示例
```go
// 使用FileManager进行章节拆分
fm := file.NewFileManager()
chapters, err := fm.SplitNovelFileIntoChapters("./input/novel.txt")
```

### 步骤 2: indexTTS2 - 生成音频

#### 功能描述
- 使用Indextts2 TTS引擎生成高质量音频
- 支持声音克隆，使用参考音频进行语音合成
- 生成完整的章节音频文件

#### 使用的工具
- 工具名称: `generate_indextts2_audio`
- 参数: 
  - `text` (要转换的文本)
  - `reference_audio` (参考音频路径)
  - `output_file` (输出音频路径)

#### 启动服务
```bash
# 启动IndexTTS2服务
cd /Users/mac/code/ai/tts/index-tts && python app.py
```

### 步骤 3: aegisub - 生成字幕

#### 功能描述
- 根据音频和原文生成时间轴匹配的字幕
- 使用Aegisub工具生成SRT格式字幕文件
- 通过字数占比算法匹配音频时长

#### 使用的工具
- 工具名称: `generate_subtitles_from_indextts2`
- 参数:
  - `audio_file` (音频文件路径)
  - `text_content` (文本内容)
  - `output_file` (输出SRT路径)

## 启动MCP服务器

```bash
bash pkg/tools/aegisub/start_indextts2_mcp_new.sh
```

服务器将提供以下MCP工具：
- `process_chapter` - 处理单个章节
- `generate_audio` - 生成音频文件
- `generate_indextts2_audio` - 使用IndexTTS2生成音频
- `generate_subtitles_from_indextts2` - 生成字幕
- `file_split_novel_into_chapters` - 拆分小说章节

## 完整工作流程示例

以下是如何处理一个完整的章节：

1. **准备输入文件**
   ```bash
   # 将小说文本放入 input/小说名/chapter_09/chapter_09.txt
   ```

2. **启动服务**
   ```bash
   # 启动IndexTTS2服务
   cd /Users/mac/code/ai/tts/index-tts && python app.py
   
   # 启动MCP服务器
   bash pkg/tools/aegisub/start_indextts2_mcp_new.sh
   ```

3. **执行三步流程**
   - 调用 `file_split_novel_into_chapters` 拆分章节
   - 调用 `generate_indextts2_audio` 生成音频
   - 调用 `generate_subtitles_from_indextts2` 生成字幕

## 输出文件

处理完成后，将在 `output/` 目录下生成：

- **音频文件**: `chapter_09_audio.wav` (约58MB)
- **字幕文件**: `chapter_09_subtitles.srt` (约24KB)

## 配置文件

系统使用 `config.yaml` 进行配置：

```yaml
# TTS配置
tts:
  indextts2:
    api_url: "http://localhost:7860"
    timeout_seconds: 300
    max_retries: 3
```

## 故障排除

1. **音频生成失败**
   - 检查IndexTTS2服务是否运行
   - 检查参考音频文件是否存在

2. **字幕生成失败**
   - 检查音频文件是否生成成功
   - 检查Aegisub是否正确安装

3. **章节拆分失败**
   - 确保章节标记格式正确（如"第x章"）
   - 检查文本编码是否为UTF-8

## 参考音频

系统需要参考音频文件用于声音克隆：
- 路径: `./assets/ref_audio/ref.m4a` 或 `./ref.m4a`
- 格式: 支持常见的音频格式（WAV, MP3, M4A等）
- 时长: 建议10-30秒的清晰语音

## 性能说明

- **章节拆分**: 几乎瞬时完成
- **音频生成**: 根据文本长度，可能需要几分钟到几十分钟
- **字幕生成**: 通常在几秒到几分钟内完成

## 项目特点

- **自动章节识别**: 支持数字和汉字章节号
- **高质量音频**: 使用先进的TTS技术
- **精确字幕**: 基于字数占比的时间轴匹配
- **MCP协议**: 支持AI代理自动化调用
- **可扩展性**: 模块化设计，易于扩展新功能