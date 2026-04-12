<p align="center">
 <img src="https://github.com/hulutech-web/novel-video-workflow/blob/master/logo.png?raw=true" width="300" />
</p>
<p align="center">
特别说明 <br/>
剪映客户端版本 3.4.1 其他版本可自行尝试<br/>
下载链接： <br/><a href="https://www.ilanzou.com/s/szdnSewG">https://www.ilanzou.com/s/szdnSewG</a>  <br/>
Aegisub客户端下载   <br/>
<a href="https://www.ilanzou.com/s/uiDnSeXN">https://www.ilanzou.com/s/uiDnSeXN</a>
 <br/>
</p>

# 小说视频工作流 (Novel Video Workflow)

一个基于AI技术的小说转视频自动化生成系统，集成了多种AI工具（TTS、图像生成等），能够将小说文本转换为带有音频、字幕和图像的视频内容，并生成可用于剪映的一键出片项目结构。
> 2026.02.11 新增图像MV生成，通过歌词自动生成mv图片，  
> 详情  MCP工具-->generate_image_from_lyric_ai_prompt--> 输入歌词，点击生成即可   
> 🎉在线体验：https://yadou.net


## 联系方式
> - 邮箱：<EMAIL>yuanhaozhuzhu@hotmail.com
> - qq群： 1053223644
<div align="center">
<img src="assets/qrcode_1772414579837.jpg" width="200" alt="芽豆小说剧" title="芽豆小说剧"> 
</div>
## 🌟 功能特性

- ✨ **智能章节分割** - 自动将小说文本按章节拆分
- 🗣️ **AI驱动文本转语音** - 支持声音克隆的高质量语音合成
- 💬 **自动生成字幕** - 基于音频内容的精准时间轴字幕
- 🎨 **AI图像生成** - 基于章节内容的智能图像生成
- ⚙️ **自动化工作流** - 端到端的自动化处理流程
- 🔌 **MCP服务集成** - 与Ollama Desktop等AI代理平台集成
- 🌐 **Web控制台界面** - 直观易用的Web操作界面
- 🎬 **剪映项目导出** - 生成可直接导入剪映的项目结构

## 🖥️ Web控制台

![web_pic.png](web_pic.png)

## 🏗️ MCP服务架构图

```mermaid
graph TB
    subgraph "📦 用户输入层"
        A[📖 小说文本]
        B[🎵 参考音频]
    end
    
    subgraph "🤖 MCP服务层"
        subgraph "🧠 Ollama (11434)"
            O[🔍 内容分析与提示词优化]
        end
        
        subgraph "💬 IndexTTS2 (7860)"
            T[🗣️ 文本转语音]
        end
        
        subgraph "🖼️ DrawThings (7861)"
            D[🎨 AI图像生成]
        end
        
        subgraph "📝 Aegisub"
            S[💬 字幕生成]
        end
    end
    
    subgraph "⚙️ 处理层"
        P1[✂️ 章节拆分]
        P2[🔄 工作流编排]
        P3[📁 文件管理]
    end
    
    subgraph "📤 输出层"
        OUT1[🔊 音频]
        OUT2[🖼️ 图像]
        OUT3[📝 字幕]
        OUT4[🎥 剪映项目]
    end

    A --> P1
    B --> T
    P1 --> O
    P1 --> T
    P1 --> D
    O --> D
    T --> OUT1
    D --> OUT2
    T --> S
    S --> OUT3
    OUT1 --> P2
    OUT2 --> P2
    OUT3 --> P2
    P2 --> OUT4
    
    %% 颜色定义
    classDef inputClass fill:#e3f2fd,stroke:#1976d2,stroke-width:2px,color:#000
    classDef mcpClass fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px,color:#000
    classDef serviceClass fill:#e8f5e8,stroke:#388e3c,stroke-width:2px,color:#000
    classDef componentClass fill:#fff3e0,stroke:#f57c00,stroke-width:2px,color:#000
    classDef outputClass fill:#ffebee,stroke:#d32f2f,stroke-width:2px,color:#000
    classDef olamaClass fill:#e1f5fe,stroke:#0288d1,stroke-width:2px,color:#000
    classDef indexttsClass fill:#e0f7fa,stroke:#0097a7,stroke-width:2px,color:#000
    classDef drawthingsClass fill:#e8f5f0,stroke:#43a047,stroke-width:2px,color:#000
    classDef aegisubClass fill:#f1f8e9,stroke:#7cb342,stroke-width:2px,color:#000

    %% 应用颜色类
    class A,B inputClass
    class O olamaClass
    class T indexttsClass
    class D drawthingsClass
    class S aegisubClass
    class P1,P2,P3 componentClass
    class OUT1,OUT2,OUT3,OUT4 outputClass
```

## 🚀 快速开始

### 系统要求（项目测试,后期扩展到更多平台）

- **操作系统**: macOS
- **Go**: 1.25+ (推荐)
- **内存**: 16GB以上 (推荐32GB)
- **GPU**: Apple Silicon (Metal支持)
- **存储**: 100GB以上可用空间

### 依赖服务

在运行系统前，请确保以下服务已安装并运行：

1. **Ollama** (用于AI推理)
   ```bash
   # 安装Ollama
   curl -fsSL https://ollama.ai/install.sh | sh
   # 启动服务
   ollama serve
   # 下载模型
   ollama pull qwen3:4b
   ```

2. **Drawthings** (用于图像生成)  
苹果商店下载，开启http访问，7861端口


3. **IndexTTS2** (用于TTS语音合成)
   ```bash
   # 按照IndexTTS2项目说明安装并启动服务
   # 确保服务在 http://localhost:7860 运行
   ```

### 启动步骤

1. **准备输入文件**
   ```bash
   # 将小说文件放入input目录
   mkdir -p input/小说名称
   cp 你的小说.txt input/小说名称/小说名称.txt
   ```

2. **准备参考音频** (可选但推荐)
   ```bash
   # 将参考音频文件放入assets目录
   mkdir -p assets/ref_audio
   cp 你的参考音频.m4a assets/ref_audio/ref.m4a
   ```

3. **启动系统**
   ```bash
   # 方法1: 同时启动MCP和Web服务 (推荐，默认)
   go run main.go

   # 方法2: 仅启动MCP服务
   go run main.go mcp

   # 方法3: 仅启动Web服务
   go run main.go web

   # 方法4: 批量处理模式
   go run main.go batch
   ```

4. **访问Web界面**
   - 打开浏览器访问: http://localhost:8080
   - 上传小说文件并开始处理

## 🛠️ 使用方法

### 1. Web界面操作

1. 访问 `http://localhost:8080`
2. 上传小说文件夹至input目录
3. 选择需要处理的工具（章节分割、音频生成、图像生成等）
4. 点击"处理上传的文件夹"执行完整工作流
5. 查看output目录中的生成结果

### 2. MCP服务调用

系统支持通过MCP协议调用各种工具，适用于AI代理集成：

```bash
# 启动MCP服务
MCP_STDIO_MODE=true go run main.go

# 或使用桥接器
go run cmd/ollama_mcp_bridge/main.go -mode server
```

### 3. 命令行批量处理

```bash
go run cmd/full_workflow/main.go
```

### 4. 一键生成剪映草稿，修改后直接发布  
在output目录下，选择chapter_0x章节，点击一键发布，打开剪映，便可以看到草稿文件，文件名与章节名一致  

## 📁 目录结构

### 输入目录结构
```
input/
└── 小说名称/
    └── 小说名称.txt  # 或已拆分的 chapter_01 等目录
```

### 输出目录结构
```
output/
└── 小说名称/
    └── chapter_01/
        ├── chapter_01.wav      # 音频文件
        ├── chapter_01.srt      # 字幕文件
        ├── chapter_01.json     # 剪映项目文件
        └── images/             # 图像目录
            ├── scene_01.png
            ├── scene_02.png
            └── ...
    └── chapter_02/
        ├── chapter_02.wav
        ├── chapter_02.srt
        ├── chapter_02.json
        └── images/
            ├── scene_01.png
            ├── scene_02.png
            └── ...
```

## 🔧 主要工具列表

系统提供以下MCP工具供调用：

| 工具名称 | 功能描述            |
|---------|-----------------|
| `generate_indextts2_audio` | 使用IndexTTS2生成音频 |
| `generate_subtitles_from_indextts2` | 生成字幕文件          |
| `file_split_novel_into_chapters` | 分割小说章节          |
| `generate_image_from_text` | 根据文本生成图像        |
| `generate_image_from_image` | 图像风格转换          |
| `generate_images_from_chapter` | 章节转图像           |
| `generate_images_from_chapter_with_ai_prompt` | AI智能提示词图像生成     |
| `generate_image_from_lyric_ai_prompt` | 歌词生成MV          |

## ⚙️ 配置说明

系统通过 `config.yaml` 文件进行配置，主要配置项包括：

- **服务端点**: Ollama, Drawthings, IndexTTS2等服务地址
- **路径配置**: 输入输出目录、资源文件路径
- **图像设置**: 生成图像的尺寸、质量、样式等
- **音频设置**: 音频格式、采样率等
- **工作流设置**: 并发任务数、临时目录等

## 🧩 MCP服务集成

本项目实现了MCP（Model Context Protocol）协议，支持以下集成方式：

### 1. Ollama Desktop集成
- 通过MCP协议与Ollama Desktop无缝集成
- 提供丰富的工具集合供AI代理调用

### 2. 工具处理器
- 使用 [ollama_tool_processor.go](pkg/utils/ollama_tool_processor.go) 作为代理
- 将外部工具调用转发到本地MCP服务

### 3. MCP桥接器
- 通过 [cmd/ollama_mcp_bridge/main.go](cmd/ollama_mcp_bridge/main.go) 提供额外集成选项
- 支持多种运行模式

## 📋 依赖项

- **Go**: 1.25+
- **Ollama**: 用于AI推理
- **Drawthings**: 用于图像、视频生成
- **IndexTTS2**: 用于高质量语音合成
- **Aegisub**: 用于字幕生成
- **FFmpeg**: 用于音频处理

## 🧪 章节编号处理

- 支持阿拉伯数字和中文数字识别（如"第七章"或"第7章"）
- 输出使用两位数格式（如 `chapter_01`, `chapter_08`）
- 最多支持99章处理
- 自动检测重复内容并跳过处理

## 🔍 服务自检

程序启动时会自动检查所有必需服务的可用性：
- Ollama 服务 - 必需
- DrawThings 服务 - 必需  
- IndexTTS2 服务 - 必需
- Aegisub 脚本 - 必需

如果任一关键服务不可用，程序将停止执行并显示错误信息。

## 📁 输出文件

- **音频文件**: `chapter_01.wav` (高质量音频)
- **字幕文件**: `chapter_01.srt` (SRT格式)
- **图像文件**: `scene_01.png`, `scene_02.png`... (AI生成图像)
- **剪映项目**: `chapter_01.json` (可直接导入剪映的项目文件，或作为剪映配置文件的参考)

## 📚 详细文档

更多信息请参考以下文档：

- [SYSTEM_ARCHITECTURE.md](SYSTEM_ARCHITECTURE.md) - 系统架构详细说明
- [USER_GUIDE.md](USER_GUIDE.md) - 完整用户操作手册
- [pkg/tools/drawthings/README.md](pkg/tools/drawthings/README.md) - 图像生成模块说明


## 💻 剪映截图   一键到剪映，自动生成到剪映的草稿目录，无需人工导入 

<div align="center">

<img src="%E6%88%AA%E5%B1%8F2026-01-16%2002.27.50.png" width="400" alt="剪映草稿目录生成 - 截图1" title="剪映草稿目录生成 - 截图1"> <img src="%E6%88%AA%E5%B1%8F2026-01-16%2002.29.02.png" width="400" alt="剪映草稿目录生成 - 截图2" title="剪映草稿目录生成 - 截图2">

</div>

## 🎬 效果一览 

### 视频演示
<div align="center">
  
<video width="80%" controls poster="logo.png">
  <source src="幽灵客栈_chapter_08.mov" type="video/quicktime">
  您的浏览器不支持视频标签。
</video>

<p><em>AI自动生成的视频内容 - 展示了从文本到图像再到视频的完整转换流程</em></p>
</div>

### 音频与字幕示例
- 🎵 [chapter_08.wav](output/青花瓷/chapter_08/chapter_08.wav) - AI生成的配音
- 📄 [chapter_08.srt](output/青花瓷/chapter_08/chapter_08.srt) - 自动生成的字幕文件


### AI生成图像示例 (宫格展示)

<div align="center">

<img src="output/青花瓷/chapter_08/scene_01.png" width="200" alt="场景 01" title="AI生成图像 - 场景 01"> <img src="output/青花瓷/chapter_08/scene_02.png" width="200" alt="场景 02" title="AI生成图像 - 场景 02">  
<img src="output/青花瓷/chapter_08/scene_03.png" width="200" alt="场景 03" title="AI生成图像 - 场景 03"> <img src="output/青花瓷/chapter_08/scene_04.png" width="200" alt="场景 04" title="AI生成图像 - 场景 04">  
<img src="output/青花瓷/chapter_08/scene_05.png" width="200" alt="场景 05" title="AI生成图像 - 场景 05"> <img src="output/青花瓷/chapter_08/scene_06.png" width="200" alt="场景 06" title="AI生成图像 - 场景 06">  
<img src="output/青花瓷/chapter_08/scene_07.png" width="200" alt="场景 07" title="AI生成图像 - 场景 07"> <img src="output/青花瓷/chapter_08/scene_08.png" width="200" alt="场景 08" title="AI生成图像 - 场景 08">  

</div>


## 🤝 贡献

欢迎提交Issue和Pull Request来帮助改进项目！

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情