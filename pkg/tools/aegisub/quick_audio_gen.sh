#!/bin/bash
# 快速音频生成脚本
# 用法: ./quick_audio_gen.sh "要转换为语音的文本"

TEXT="$1"
if [ -z "$TEXT" ]; then
    echo "用法: $0 \"要转换为语音的文本\""
    exit 1
fi

FILENAME="./output/quick_audio_$(date +%s).wav"

echo "正在生成音频: $TEXT"
cd "$(dirname "$(realpath "$0")")/../.."
go run pkg/utils/ollama_tool_processor.go "{\"name\":\"novel_video_workflow_generate_audio\",\"arguments\":{\"text\":\"$TEXT\",\"reference_audio\":\"./ref.m4a\",\"output_file\":\"$FILENAME\"}}"

echo "音频已保存到: $FILENAME"