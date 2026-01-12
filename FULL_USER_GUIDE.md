# 小说视频工作流完整用户指南

## 1. 硬件与软件要求

### 1.1 硬件配置
- **设备**: Mac Studio M2 Max 基础款
- **处理器**: Apple M2 Max
- **内存**: 建议至少16GB
- **存储**: 建议至少512GB SSD

### 1.2 必需软件
- **FFmpeg**: 音视频处理工具
  ```bash
  brew install ffmpeg
  ```
- **Go**: 项目开发语言 (1.19+)
- **Python**: 用于TTS服务
- **Node.js**: 用于前端工具 (如需要)

## 2. MCP服务配置详解

### 2.1 Ollama (端口: 11434)
- **功能**: 本地大语言模型服务
- **主要模型**: 
  - `llama3:8b` - 用于内容分析和提示词优化
- **用途**:
  - 章节内容理解
  - 场景描述生成
  - 提示词优化
  - 文本风格分析

### 2.2 IndexTTS2 (端口: 7860)
- **功能**: 文本转语音服务
- **用途**:
  - 将小说文本转换为语音
  - 支持音色克隆
  - 生成高质量音频
- **API端点**: `http://localhost:7860`

### 2.3 DrawThings (端口: 7861)
- **功能**: AI图像生成服务
- **用途**:
  - 根据文本内容生成图像
  - 场景可视化
  - 风格化图像生成
- **API端点**: `http://localhost:7861`

### 2.4 Aegisub (本地脚本)
- **功能**: 字幕生成与处理
- **用途**:
  - 生成同步字幕文件
  - 时间轴对齐
  - 字幕样式处理

## 3. 服务启动顺序

### 3.1 启动MCP服务
1. **启动Ollama**:
   ```bash
   ollama serve
   ```

2. **启动IndexTTS2**:
   ```bash
   cd /path/to/indexTTS2
   uv run webui.py
   ```

3. **启动DrawThings (Stable Diffusion)**:
   ```bash
   cd /path/to/stable-diffusion-webui
   ./webui.sh
   ```

4. **确认Aegisub已安装**:
   - macOS: 确保 `/Applications/Aegisub.app` 存在

### 3.2 验证服务状态
运行以下命令验证所有服务是否正常:
```bash
curl http://localhost:11434/api/tags
curl http://localhost:7860
curl http://localhost:7861
```

## 4. 项目结构与工作流

### 4.1 主要目录结构
```
novel-video-workflow/
├── cmd/                    # 命令行工具
│   ├── test_workflow/      # 主工作流测试
│   ├── split_chapters/     # 章节拆分工具
│   └── full_workflow/      # 完整工作流
├── pkg/                    # 项目包
│   ├── tools/
│   │   ├── file/          # 文件管理
│   │   ├── indextts2/     # TTS服务
│   │   ├── drawthings/    # 图像生成
│   │   └── aegisub/       # 字幕生成
├── input/                  # 输入目录
│   └── 小说名/
│       ├── 小说名.txt      # 原始小说文件
│       └── chapter_01/    # 章节子目录
├── output/                 # 输出目录
├── assets/                 # 资源文件
│   └── ref_audio/         # 参考音频
└── config.yaml            # 配置文件
```

### 4.2 工作流顺序
1. **输入处理**: 读取小说文件
2. **章节拆分**: 自动按"第X章"拆分
3. **内容分析**: Ollama分析章节内容
4. **语音生成**: IndexTTS2生成音频
5. **图像生成**: DrawThings生成图像
6. **字幕生成**: Aegisub生成字幕
7. **视频合成**: 组合成最终视频

## 5. MCP服务测试流程

### 5.1 单项服务测试

#### 5.1.1 Ollama测试
```bash
# 检查模型列表
curl -X POST http://localhost:11434/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3:8b",
    "prompt": "你好，测试Ollama服务",
    "stream": false
  }'
```

#### 5.1.2 IndexTTS2测试
```bash
# 检查服务状态
curl http://localhost:7860
```

#### 5.1.3 DrawThings测试
```bash
# 检查API可用性
curl http://localhost:7861
```

### 5.2 项目内测试

#### 5.2.1 章节拆分测试
```bash
cd /Users/mac/code/ai/novel-video-workflow
go run cmd/split_chapters/main.go
```

#### 5.2.2 完整工作流测试
```bash
go run cmd/test_workflow/main.go
```

#### 5.2.3 详细测试输出
测试过程会显示:
- 服务可用性检查
- 章节拆分进度
- 音频生成状态
- 图像生成状态
- 字幕生成状态

## 6. 配置文件说明

### 6.1 config.yaml 关键配置

```yaml
# Ollama配置
ollama:
  api_url: "http://localhost:11434"
  model: "llama3:8b"
  timeout_seconds: 120

# IndexTTS2配置
indextts2:
  api_url: "http://localhost:7860"
  timeout_seconds: 300

# DrawThings配置
drawthings:
  api_url: "http://localhost:7861"
  width: 512
  height: 896  # 适配手机竖屏

# 视频输出配置
video:
  resolution:
    width: 1080
    height: 1920  # 竖屏分辨率
```

## 7. 使用示例

### 7.1 准备输入文件
1. 在 `input/` 目录下创建小说目录
2. 将小说文件放入目录 (如 `input/幽灵客栈/幽灵客栈.txt`)
3. 确保文件包含"第X章"格式的章节标记

### 7.2 运行完整流程
```bash
cd /Users/mac/code/ai/novel-video-workflow
go run cmd/test_workflow/main.go
```

### 7.3 输出文件结构
```
output/
└── 小说名/
    └── chapter_01/
        ├── chapter_01.wav    # 音频文件
        ├── chapter_01.srt    # 字幕文件
        ├── chapter_01.mp4    # 视频片段
        └── images/          # 图像文件
            ├── scene_01.png
            ├── scene_02.png
            └── ...
```

## 8. 故障排除

### 8.1 常见问题
- **服务未启动**: 检查对应端口服务是否运行
- **音频生成失败**: 确认参考音频文件存在
- **图像生成失败**: 检查DrawThings服务状态
- **章节拆分失败**: 确认文本格式正确

### 8.2 日志查看
- 服务日志通常在对应服务的控制台输出
- 项目日志使用zap库记录

## 9. 性能优化建议

### 9.1 硬件优化
- 使用M2 Max的GPU加速AI推理
- 确保足够的内存处理大型模型

### 9.2 服务配置优化
- 调整并发数以平衡性能和资源
- 根据需要调整图像生成参数

### 9.3 工作流优化
- 合理设置超时时间
- 优化重试机制