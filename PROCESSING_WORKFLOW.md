# 小说章节处理工作流程

本项目提供了一个自动化的工作流程，用于处理小说文本并生成对应的音频和字幕。

## 工作流程概述

1. **文本拆分** - 将完整的小说文本按章节拆分并创建目录结构
2. **音频生成** - 使用Indextts2将每个章节文本转换为音频
3. **字幕生成** - 为生成的音频创建SRT字幕文件

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

## 服务依赖

在运行任何音频生成任务之前，必须确保以下服务正在运行：

1. **IndexTTS2服务** - 运行在 http://localhost:7860
   ```bash
   cd /Users/mac/code/ai/tts/index-tts && python app.py
   ```

## 使用方法

### 1. 准备输入文件

将你的小说文本文件放入 `input/小说名称/` 目录下，例如 `input/幽灵客栈/novel.txt`。

### 2. 文本预处理

使用以下命令拆分章节：

```bash
# 假设你有一个工具可以自动拆分章节
# 或者手动创建章节目录结构
```

### 3. 处理单个章节的音频生成

要处理特定章节（例如 chapter_07），你可以使用以下 Go 程序：

```go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"novel-video-workflow/pkg/tools/indextts2"
)

func main() {
	// 定义路径
	inputChapterPath := "./input/幽灵客栈/chapter_07/chapter_07.txt"
	referenceAudioPath := "./assets/ref_audio/ref.m4a"
	outputAudioPath := "./output/幽灵客栈/chapter_07/audio.wav"

	// 读取章节内容
	content, err := ioutil.ReadFile(inputChapterPath)
	if err != nil {
		log.Fatalf("无法读取章节文件: %v", err)
	}

	chapterText := string(content)
	fmt.Printf("成功读取章节内容，长度: %d 字符\n", len(chapterText))

	// 确保输出目录存在
	outputDir := filepath.Dir(outputAudioPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("无法创建输出目录: %v", err)
	}

	// 初始化日志
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("无法初始化日志: %v", err)
	}
	defer logger.Sync()

	// 使用Indextts2客户端直接生成音频
	client := indextts2.NewIndexTTS2Client(logger, "http://localhost:7860")

	fmt.Println("正在调用Indextts2 API生成音频...")
	err = client.GenerateTTSWithAudio(referenceAudioPath, chapterText, outputAudioPath)
	if err != nil {
		log.Fatalf("调用Indextts2 API失败: %v", err)
	}

	fmt.Printf("音频文件已成功生成: %s\n", outputAudioPath)
}
```

运行此程序：

```bash
cd cmd/process_chapter
go run main.go
```

### 4. 使用MCP工具

如果你想通过MCP协议调用工具，可以使用以下方式：

```go
// 示例：如何使用MCP客户端调用generate_indextts2_audio工具
package main

import (
	"context"
	"fmt"
	"io/ioutil"

	mcp "github.com/mark3labs/mcp-go/client"
)

func main() {
	// 读取章节内容
	content, err := ioutil.ReadFile("./input/幽灵客栈/chapter_07/chapter_07.txt")
	if err != nil {
		panic(err)
	}
	chapterText := string(content)

	// 创建MCP客户端连接到服务器
	client, err := mcp.NewClient(
		mcp.WithHost("localhost"),
		mcp.WithPort(3000), // 默认端口
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx := context.Background()

	// 调用generate_indextts2_audio工具
	result, err := client.CallTool(ctx, "generate_indextts2_audio", map[string]interface{}{
		"text":            chapterText,
		"reference_audio": "./assets/ref_audio/ref.m4a",
		"output_file":     "./output/幽灵客栈/chapter_07/audio.wav",
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("工具调用成功！结果: %+v\n", result)
}
```

## 可用的MCP工具

1. `process_chapter` - 处理单个章节
2. `generate_indextts2_audio` - 使用Indextts2生成音频
3. `generate_subtitles_from_indextts2` - 为Indextts2音频生成字幕

## 注意事项

- 确保Indextts2服务正在运行（默认在 http://localhost:7860）
- 文本较长的章节可能需要更多时间来处理
- 确保参考音频文件存在
- 输出目录会自动创建
- 如果需要将长文本按段落分隔处理，文本中的 `\n\n\n` 会被识别为段落分隔符，但整个章节仍会生成一个完整的音频文件