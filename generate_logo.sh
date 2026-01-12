#!/bin/bash

# 小说视频工作流项目Logo生成脚本

echo "🎨 为小说视频工作流项目生成Logo..."

# 检查是否设置了DrawThings API
if [ -z "$DRAWTHINGS_API_URL" ]; then
    DRAWTHINGS_API_URL="http://localhost:7861"
fi

echo "🔗 使用DrawThings API: $DRAWTHINGS_API_URL"

# Logo生成提示词
PROMPT="A minimalist logo design for \"Novel Video Workflow\" project, featuring an open book transforming into video waves or sound waves, with modern clean lines, using blue and orange gradient colors, professional typography, vector art style, centered composition, white background, high resolution, elegant and tech-savvy appearance, 4k quality"

echo "📝 提示词: $PROMPT"
echo ""
echo "💡 要生成Logo，请使用以下方式之一："
echo ""
echo "1. 在DrawThings/Stable Diffusion WebUI中使用上述提示词"
echo "2. 使用curl命令（示例）："
echo "   curl -X POST $DRAWTHINGS_API_URL/sdapi/v1/txt2img \\"
echo "   -H \"Content-Type: application/json\" \\"
echo "   -d '{"
echo "     \"prompt\": \"$PROMPT\","
echo "     \"negative_prompt\": \"low quality, blurry, distorted, extra limbs, bad anatomy\","
echo "     \"steps\": 30,"
echo "     \"width\": 512,"
echo "     \"height\": 512,"
echo "     \"cfg_scale\": 7.0,"
echo "     \"sampler_name\": \"DPM++ 2M Karras\""
echo "   }'"
echo ""
echo "3. 或者在Ollama中优化提示词后使用："
echo "   ollama run llama3:8b"
echo ""
echo "🏆 建议生成多个版本，选择最适合项目风格的Logo"