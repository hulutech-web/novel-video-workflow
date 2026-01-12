# MCP 服务集成与自动化工作流指南

## 概述

本项目实现了基于MCP（Model Context Protocol）协议的AI工具链集成，支持自动化处理小说文本生成音频、字幕和图像的完整工作流。

## 系统架构

### 核心组件

1. **IndexTTS2** - 高质量语音合成服务
2. **Aegisub** - 字幕生成与时间轴匹配
3. **DrawThings** - 图像生成服务（基于Stable Diffusion）
4. **Ollama** - 大语言模型服务

### 工作流处理器

- [pkg/mcp/handler.go](file:///Users/mac/code/ai/novel-video-workflow/pkg/mcp/handler.go) - MCP服务注册与处理器
- [cmd/test_workflow/main.go](file:///Users/mac/code/ai/novel-video-workflow/cmd/test_workflow/main.go) - 自动化工作流入口

## 服务配置要求

### 1. IndexTTS2 服务
- **端口**: `http://localhost:7860`
- **功能**: 语音合成，支持声音克隆
- **依赖**: Python环境，IndexTTS2服务
- **参考音频路径**: `./assets/ref_audio/ref.m4a`

### 2. Aegisub 服务
- **功能**: 为音频生成SRT字幕文件
- **依赖**: Aegisub应用或Aegisub Lua API
- **脚本路径**: [pkg/tools/aegisub/aegisub_generator.lua](file:///Users/mac/code/ai/novel-video-workflow/pkg/tools/aegisub/aegisub_generator.lua)

### 3. DrawThings 服务
- **端口**: `http://localhost:7861`
- **功能**: 基于文本生成图像
- **依赖**: Stable Diffusion WebUI
- **模型**: `z_image_turbo_1.0_q6p.ckpt`

### 4. Ollama 服务
- **端口**: `http://localhost:11434`
- **功能**: 大语言模型推理
- **模型**: `qwen3:4b` 或其他兼容模型
- **用途**: 生成图像提示词、章节分析

## 服务注册与MCP工具

### 注册的MCP工具

1. **process_chapter**
   - 处理单个章节的完整工作流
   - 参数: `chapter_text`, `chapter_number`

2. **generate_indextts2_audio**
   - 生成音频文件
   - 参数: `text`, `reference_audio`, `output_file`

3. **generate_subtitles_from_indextts2**
   - 生成字幕文件
   - 参数: `audio_file`, `text_content`, `output_file`

4. **file_split_novel_into_chapters**
   - 拆分小说章节
   - 参数: `novel_path`

5. **generate_image_from_text**
   - 文生图
   - 参数: `text`, `output_file`, `width`, `height`, `is_suspense`

6. **generate_images_from_chapter**
   - 从章节生成图像序列
   - 参数: `chapter_text`, `output_dir`, `width`, `height`, `is_suspense`

7. **generate_images_from_chapter_with_ai_prompt**
   - 使用AI生成提示词并生成图像
   - 参数: `chapter_text`, `output_dir`, `width`, `height`, `is_suspense`

## 目录结构

```
novel-video-workflow/
├── input/                      # 输入目录
│   └── 小说名称/
│       ├── chapter_01/         # 章节目录
│       │   └── chapter_01.txt  # 章节文本
│       └── ...
├── output/                     # 输出目录
│   └── 小说名称/
│       └── chapter_01/         # 输出章节目录
│           ├── audio.wav       # 音频文件
│           ├── subtitles.srt   # 字幕文件
│           └── images/         # 图像目录
├── assets/
│   └── ref_audio/              # 参考音频
│       └── ref.m4a
├── pkg/
│   ├── mcp/                    # MCP服务
│   ├── tools/
│   │   ├── indextts2/          # IndexTTS2客户端
│   │   ├── aegisub/            # Aegisub集成
│   │   └── drawthings/         # DrawThings客户端
└── cmd/
    └── test_workflow/          # 测试工作流
```

## 运行要求

### 环境依赖

1. **Go 1.19+**
2. **Python 3.8+**
3. **Node.js (用于MCP)**
4. **FFmpeg (音频处理)**

### 服务启动顺序

1. **启动Ollama服务**
   ```bash
   ollama serve
   ```

2. **启动DrawThings服务 (Stable Diffusion WebUI)**
   ```bash
   cd /path/to/stable-diffusion-webui
   python launch.py --port 7861
   ```

3. **启动IndexTTS2服务**
   ```bash
   cd /path/to/index-tts
   python app.py --port 7860
   ```

4. **启动MCP服务器**
   ```bash
   bash pkg/tools/aegisub/start_indextts2_mcp_new.sh
   ```

### 服务自检

程序启动时会自动检查所有服务的可用性：
- Ollama 服务 - 必需
- DrawThings 服务 - 必需
- IndexTTS2 服务 - 必需
- Aegisub 脚本 - 必需

## 章节编号处理

### 格式规范
- 所有章节编号使用两位数格式 (如 `chapter_01`, `chapter_07`)
- 支持中文数字识别 (如"第七章" → 7)
- 支持最多99章的处理

### 文件命名规则
- 输入: `chapter_07.txt`
- 音频输出: `chapter_07.wav`
- 字幕输出: `chapter_07.srt`
- 图像输出: `scene_01.png`, `scene_02.png`...

## 工作流执行流程

### 完整处理流程

1. **章节解析**
   - 从文本中识别"第x章"标记
   - 拆分章节并创建目录结构

2. **音频生成**
   - 使用IndexTTS2进行语音合成
   - 基于参考音频进行声音克隆

3. **字幕生成**
   - 使用Aegisub为音频生成SRT字幕
   - 基于字数占比算法匹配时间轴

4. **图像生成**
   - 使用Ollama分析章节内容
   - 生成分镜提示词
   - 使用DrawThings生成图像序列

### 错误处理与降级

- 服务不可用时提供清晰错误信息
- 关键服务不可用时停止整个工作流
- 支持多种音频格式输入

## 测试与验证

### 运行测试工作流

```bash
go run cmd/test_workflow/main.go
```

### 检查处理进度

```bash
go run cmd/check_progress/main.go
```

### 验证输出文件

- 音频文件大小和时长
- 字幕文件时间轴准确性
- 图像文件数量和质量

## 故障排除

### 常见问题

1. **服务连接失败**
   - 检查端口是否正确
   - 确认服务是否已启动

2. **音频生成失败**
   - 检查参考音频文件是否存在
   - 确认IndexTTS2服务状态

3. **字幕生成失败**
   - 验证音频文件是否生成成功
   - 检查Aegisub安装和配置

4. **图像生成失败**
   - 确认Ollama模型可用性
   - 检查DrawThings服务状态

### 调试方法

- 查看服务日志输出
- 验证网络连接
- 检查文件权限和路径

## 性能说明

- **章节拆分**: 几乎瞬时完成
- **音频生成**: 根据文本长度，可能需要几分钟到几十分钟
- **字幕生成**: 通常在几秒到几分钟内完成
- **图像生成**: 取决于图像数量和复杂度

## 扩展与定制

### 添加新工具

1. 在[pkg/mcp/handler.go](file:///Users/mac/code/ai/novel-video-workflow/pkg/mcp/handler.go)中注册新MCP工具
2. 实现对应的处理函数
3. 添加参数验证和错误处理

### 自定义工作流

- 修改[cmd/test_workflow/main.go](file:///Users/mac/code/ai/novel-video-workflow/cmd/test_workflow/main.go)中的处理逻辑
- 调整服务依赖检查
- 定制输出格式和路径

## 安全注意事项

- 确保服务运行在安全的网络环境中
- 验证输入文件的安全性
- 限制文件上传大小和类型
- 定期更新依赖组件

## 维护与更新

- 定期检查服务版本兼容性
- 更新模型和工具到最新版本
- 监控系统性能和资源使用
- 备份重要配置和数据