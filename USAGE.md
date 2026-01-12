# Novel Video Workflow 使用说明

## 概述

这是一个自动化的小说视频生成工作流程，能够将小说文本转换为带有音频、字幕和图像的视频内容。整个流程包括四个主要步骤：

1. **file mcp** - 将文字分出章节
2. **indexTTS2** - 生成音频
3. **aegisub** - 生成字幕
4. **drawthings** - 生成图像

## 系统要求

- Go 1.19+
- Python 3.8+
- IndexTTS2 服务 (运行在 http://localhost:7860)
- DrawThings 服务 (运行在 http://localhost:7861)
- Aegisub (用于字幕生成)
- Ollama (用于AI推理)

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
│           ├── chapter_01.wav    # 生成的音频
│           ├── chapter_01.srt    # 生成的字幕
│           └── images/           # 生成的图像
│               ├── scene_01.png
│               └── ...
└── assets/
    └── ref_audio/           # 参考音频
        └── ref.m4a
```

## 四步工作流程详解

### 步骤 1: file mcp - 拆分章节

#### 功能描述
- 识别小说文本中的章节标记（如"第x章"），x可以是数字也可以是汉字
- 获取章节号并创建对应的文件夹（如`chapter_0x`），使用两位数格式
- 将相应部分的内容放入`chapter_0x.txt`中
- 保留"第x章"内容本身

#### 章节编号处理
- **输入格式**: 支持"第7章"、"第七章"等格式
- **输出格式**: 统一使用两位数格式（如 `chapter_01`, `chapter_07`）
- **范围**: 支持最多99章的处理
- **映射**: 确保输入输出章节编号完全一致

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

### 步骤 4: drawthings - 生成图像

#### 功能描述
- 使用Ollama分析章节内容
- 生成适合的图像提示词
- 使用DrawThings API生成图像序列

#### 使用的工具
- 工具名称: `generate_images_from_chapter_with_ai_prompt`
- 参数:
  - `chapter_text` (章节文本)
  - `output_dir` (输出目录)
  - `width` (图像宽度)
  - `height` (图像高度)
  - `is_suspense` (是否悬疑风格)

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
- `generate_image_from_text` - 使用DrawThings API进行文生图
- `generate_image_from_image` - 使用DrawThings API进行图生图
- `generate_images_from_chapter` - 使用DrawThings API将章节文本转换为图像序列
- `generate_images_from_chapter_with_ai_prompt` - 使用AI生成提示词并将章节文本转换为图像序列

## 完整工作流程示例

以下是如何处理一个完整的章节：

1. **准备输入文件**
   ```bash
   # 将小说文本放入 input/小说名/chapter_07/chapter_07.txt
   # 确保章节内容以"第7章"开头
   ```

2. **启动服务**
   ```bash
   # 启动Ollama服务
   ollama serve
   
   # 启动Stable Diffusion WebUI (DrawThings)
   cd /path/to/stable-diffusion-webui && python launch.py --port 7861
   
   # 启动IndexTTS2服务
   cd /Users/mac/code/ai/tts/index-tts && python app.py --port 7860
   
   # 启动MCP服务器
   bash pkg/tools/aegisub/start_indextts2_mcp_new.sh
   ```

3. **执行四步流程**
   - 调用 `file_split_novel_into_chapters` 拆分章节
   - 调用 `generate_indextts2_audio` 生成音频
   - 调用 `generate_subtitles_from_indextts2` 生成字幕
   - 调用 `generate_images_from_chapter_with_ai_prompt` 生成图像

## 输出文件

处理完成后，将在 `output/` 目录下生成：

- **音频文件**: `chapter_07.wav` (约120MB)
- **字幕文件**: `chapter_07.srt` (约4KB)
- **图像文件**: `images/` 目录下的多个 `scene_XX.png` 文件

## 配置文件

系统使用 `config.yaml` 进行配置：

```yaml
# TTS配置
tts:
  indextts2:
    api_url: "http://localhost:7860"
    timeout_seconds: 600  # 增加超时时间以处理长文本
    max_retries: 3

# 图像生成配置
image:
  drawthings:
    api_url: "http://localhost:7861"
    model: "z_image_turbo_1.0_q6p.ckpt"
    default_size:
      width: 512
      height: 896

# Ollama配置
ollama:
  api_url: "http://localhost:11434"
  model: "qwen3:4b"
```

## 故障排除

1. **音频生成失败**
   - 检查IndexTTS2服务是否运行
   - 检查参考音频文件是否存在
   - 确认音频文件路径正确

2. **字幕生成失败**
   - 检查音频文件是否生成成功
   - 检查Aegisub是否正确安装
   - 确认音频和文本文件路径正确

3. **图像生成失败**
   - 检查DrawThings服务是否运行
   - 确认Ollama服务可用
   - 验证模型是否正确加载

4. **章节拆分失败**
   - 确保章节标记格式正确（如"第x章"）
   - 检查文本编码是否为UTF-8
   - 验证章节编号格式是否一致

5. **服务连接失败**
   - 确认所有服务在正确的端口运行
   - 检查防火墙设置
   - 验证网络连接

## 参考音频

系统需要参考音频文件用于声音克隆：
- 路径: `./assets/ref_audio/ref.m4a` 或 `./ref.m4a`
- 格式: 支持常见的音频格式（WAV, MP3, M4A等）
- 时长: 建议10-30秒的清晰语音
- 音质: 清晰、无噪音的语音样本

## 性能说明

- **章节拆分**: 几乎瞬时完成
- **音频生成**: 根据文本长度，可能需要几分钟到几十分钟
- **字幕生成**: 通常在几秒到几分钟内完成
- **图像生成**: 取决于图像数量和复杂度

## 项目特点

- **自动章节识别**: 支持数字和汉字章节号
- **两位数格式输出**: 所有章节编号使用两位数格式（如 chapter_01, chapter_07）
- **高质量音频**: 使用先进的TTS技术
- **精确字幕**: 基于字数占比的时间轴匹配
- **AI图像生成**: 使用Ollama优化的提示词生成图像
- **MCP协议**: 支持AI代理自动化调用
- **可扩展性**: 模块化设计，易于扩展新功能
- **输入输出一致性**: 确保章节编号在输入和输出目录中完全一致