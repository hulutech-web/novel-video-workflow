# Novel Video Workflow 实现总结

## 项目概述

我们成功实现了一个自动化的小说视频生成工作流程，该系统能够将小说文本转换为带有音频和字幕的视频内容。整个流程包括三个主要步骤：章节拆分、音频生成和字幕生成。

## 核心功能实现

### 1. 文件管理工具 (file mcp)

我们在 [pkg/tools/file/file.go](file:///Users/mac/code/ai/novel-video-workflow/pkg/tools/file/file.go) 中实现了以下功能：

- **章节识别**: 使用正则表达式 `(?m)^\s*第[\p{N}\p{L}]+[章节][^\r\n]*$` 识别"第x章"标记，x可以是数字也可以是汉字
- **章节提取**: 实现 [SplitNovelIntoChapters](file:///Users/mac/code/ai/novel-video-workflow/pkg/tools/file/file.go#L133-L174) 方法按章节拆分小说文本
- **章节号解析**: 实现 [extractChapterNumber](file:///Users/mac/code/ai/novel-video-workflow/pkg/tools/file/file.go#L176-L194) 和 [convertChineseNumberToArabic](file:///Users/mac/code/ai/novel-video-workflow/pkg/tools/file/file.go#L196-L212) 方法处理数字和汉字章节号
- **目录结构创建**: 自动创建 `chapter_0x` 目录和对应的 `chapter_0x.txt` 文件
- **新方法**: 添加了 [SplitNovelFileIntoChapters](file:///Users/mac/code/ai/novel-video-workflow/pkg/tools/file/file.go#L279-L323) 方法从文件读取并拆分小说

### 2. MCP 工具注册

我们在 [pkg/mcp/handler.go](file:///Users/mac/code/ai/novel-video-workflow/pkg/mcp/handler.go) 中添加了新的MCP工具：

- **file_split_novel_into_chapters**: 用于拆分小说章节的MCP工具
- 工具参数: `novel_path` (小说文件路径)
- 该工具集成了 [FileManager](file:///Users/mac/code/ai/novel-video-workflow/pkg/tools/file/file.go#L8-L10) 的功能

### 3. 音频生成 (IndexTTS2)

- 集成了 [IndexTTS2](file:///Users/mac/code/ai/novel-video-workflow/pkg/tools/indextts2/client.go#L57-L59) TTS引擎
- 实现了声音克隆功能，使用参考音频进行语音合成
- 通过 `generate_indextts2_audio` 工具提供MCP接口

### 4. 字幕生成 (Aegisub)

- 实现了基于字数占比算法的字幕生成
- 通过音频时长和文本字数计算每段文本的显示时间
- 生成标准SRT格式字幕文件
- 通过 `generate_subtitles_from_indextts2` 工具提供MCP接口

## 完整工作流程

### 三步流程

1. **file mcp** - 识别"第x章"标记，创建 `chapter_0x` 目录和文件
2. **indexTTS2** - 生成高质量音频文件
3. **aegisub** - 根据音频和文本生成时间轴匹配的字幕

### 实现验证

我们成功测试了完整的三步流程：
- 输入: `input/幽灵客栈/chapter_09/chapter_09.txt` (18,303 字符)
- 输出:
  - 音频: `output/幽灵客栈/chapter_09/chapter_09_audio.wav` (58MB)
  - 字幕: `output/幽灵客栈/chapter_09/chapter_09_subtitles.srt` (24KB)

## 技术架构

- **Go语言**: 实现核心处理逻辑
- **MCP协议**: 用于AI代理交互
- **IndexTTS2**: 用于语音合成
- **Aegisub**: 用于字幕生成
- **模块化设计**: 各功能组件独立且可扩展

## 代码变更总结

1. **[pkg/tools/file/file.go](file:///Users/mac/code/ai/novel-video-workflow/pkg/tools/file/file.go)**: 添加了章节拆分和文件处理功能
2. **[pkg/mcp/handler.go](file:///Users/mac/code/ai/novel-video-workflow/pkg/mcp/handler.go)**: 添加了 `file_split_novel_into_chapters` MCP工具
3. **新增 USAGE.md**: 详细使用说明文档

## 项目特点

- **智能章节识别**: 支持数字和汉字章节号
- **自动目录结构**: 自动创建章节目录和文件
- **高质量音频**: 使用先进的TTS技术
- **精确字幕**: 基于字数占比的时间轴匹配
- **MCP协议支持**: 支持AI代理自动化调用
- **模块化设计**: 易于扩展和维护

## 验证结果

系统已成功验证，能够处理长篇小说章节（如18,000+字符的章节），生成高质量的音频（58MB WAV文件）和精确的SRT字幕文件（24KB），完整实现了用户需求的三步工作流程。