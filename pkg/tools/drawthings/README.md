# DrawThings API 集成

本模块集成了 DrawThings API（Stable Diffusion API），支持文生图和图生图功能，特别适用于悬疑小说视频生成。

## 服务准备

在使用 DrawThings API 之前，请确保已正确安装并运行 Stable Diffusion WebUI（AUTOMATIC1111）：

1. 下载并安装 [AUTOMATIC1111/stable-diffusion-webui](https://github.com/AUTOMATIC1111/stable-diffusion-webui)
2. 启动服务时启用 API 模式：
   ```bash
   # Linux/Mac
   ./webui.sh --api --nowebui
   
   # Windows
   webui-user.bat --api --nowebui
   ```
3. 确认服务在 http://localhost:7861 上运行

## 验证服务可用性

运行以下命令验证服务是否可用：

```bash
# 测试连接
curl http://localhost:7861
```

## 功能

- `generate_image_from_text` - 文生图工具
- `generate_image_from_image` - 图生图工具  
- `generate_images_from_chapter` - 章节文生图工具（将章节文本按段落分割并生成对应的图像序列）
- `generate_images_from_chapter_with_ai_prompt` - 使用AI生成提示词的章节文生图工具（使用Ollama生成优化的提示词）

## API 参数

### 文生图 (txt2img) 参数

- `prompt`: 提示词
- `negative_prompt`: 负面提示词
- `width`: 图像宽度 (默认: 512)
- `height`: 图像高度 (默认: 896)
- `steps`: 生成步数 (默认: 35)
- `seed`: 随机种子 (默认: -1)
- `sampler_name`: 采样器 (默认: "DPM++ 2M Trailing")
- `guidance_scale`: 提示词引导尺度 (默认: 0)
- `batch_size`: 批次大小 (默认: 1)
- `enable_hr`: 启用高清修复 (默认: false)
- `tiled_diffusion`: 启用平铺扩散 (默认: true)
- `tiled_decoding`: 启用平铺解码 (默认: true)

### 图生图 (img2img) 参数

- `init_images`: 参考图像 Base64 编码字符串数组
- `strength`: 变换强度 (默认: 0.7)
- `prompt`: 提示词
- `negative_prompt`: 负面提示词
- `width`: 图像宽度 (默认: 1024)
- `height`: 图像高度 (默认: 1792)

## Ollama 集成

新功能通过Ollama集成，使用AI模型（如 llama3.1）来生成更精确的图像提示词：

1. 将章节文本按段落分割
2. 对每个段落使用Ollama生成优化的图像提示词
3. 使用生成的提示词调用DrawThings API生成图像
4. 将生成的图像保存到指定的章节输出目录

## 悬疑风格

当 `is_suspense` 参数设置为 `true` 时，系统会自动添加以下悬疑风格描述：
- 周围环境模糊成黑影
- 空气凝滞
- 浅景深
- 胶片颗粒感
- 低饱和度
- 极致悬疑氛围
- 阴沉窒息感
- 夏季，环境阴霾
- 其他部分模糊不可见

## 使用示例

### 在 Ollama 中使用

```
请使用 generate_image_from_text 工具将以下文本转换为图像：
{
  "text": "夜晚的古老庄园，闪电划过天空，神秘的黑影在窗前闪过",
  "output_file": "./output/scene1.png",
  "width": 1024,
  "height": 1792,
  "is_suspense": true
}
```

### 章节图像生成（使用AI提示词）

```
请使用 generate_images_from_chapter_with_ai_prompt 工具将章节文本转换为图像序列：
{
  "chapter_text": "第一章的内容文本...",
  "output_dir": "./output/chapter_01_images/",
  "width": 1024,
  "height": 1792,
  "is_suspense": true
}
```

## 依赖

- DrawThings API (通过 http://localhost:7861 访问)
- Ollama (通过 http://localhost:11434 访问，默认模型 qwen3:4b)
- Stable Diffusion WebUI