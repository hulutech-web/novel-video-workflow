#!/bin/bash

# Aegisub字幕生成脚本
# 参数: $1 - 音频文件路径, $2 - 文本文件路径, $3 - 输出SRT文件路径

AUDIO_FILE="$1"
TEXT_FILE="$2"
OUTPUT_SRT="$3"

if [ -z "$AUDIO_FILE" ] || [ -z "$TEXT_FILE" ] || [ -z "$OUTPUT_SRT" ]; then
    echo "错误: 缺少必要参数"
    echo "用法: $0 <音频文件> <文本文件> <输出SRT文件>"
    exit 1
fi

# 检查输入文件是否存在
if [ ! -f "$AUDIO_FILE" ]; then
    echo "错误: 音频文件 '$AUDIO_FILE' 不存在"
    exit 1
fi

if [ ! -f "$TEXT_FILE" ]; then
    echo "错误: 文本文件 '$TEXT_FILE' 不存在"
    exit 1
fi

# 确保输出目录存在
OUTPUT_DIR=$(dirname "$OUTPUT_SRT")
mkdir -p "$OUTPUT_DIR"

# 检查操作系统
UNAME=$(uname)

# 检查Aegisub是否已安装
AEGISUB_AVAILABLE=false
if command -v aegisub &> /dev/null; then
    AEGISUB_CMD="aegisub"
    AEGISUB_AVAILABLE=true
elif [ -f "/Applications/Aegisub.app/Contents/MacOS/aegisub" ]; then
    AEGISUB_CMD="/Applications/Aegisub.app/Contents/MacOS/aegisub"
    AEGISUB_AVAILABLE=true
elif [ -f "/Applications/Aegisub.app/Contents/MacOS/Aegisub" ]; then
    AEGISUB_CMD="/Applications/Aegisub.app/Contents/MacOS/Aegisub"
    AEGISUB_AVAILABLE=true
elif [ -f "/usr/bin/aegisub" ]; then
    AEGISUB_CMD="/usr/bin/aegisub"
    AEGISUB_AVAILABLE=true
elif [ -f "/usr/local/bin/aegisub" ]; then
    AEGISUB_CMD="/usr/local/bin/aegisub"
    AEGISUB_AVAILABLE=true
fi

# 在macOS上，即使Aegisub可用，我们也优先使用Python备用方案，因为Aegisub自动化脚本可能会启动GUI
if [ "$UNAME" = "Darwin" ]; then
    AEGISUB_AVAILABLE=false
fi

echo "正在生成字幕..."
echo "音频文件: $AUDIO_FILE"
echo "文本文件: $TEXT_FILE"
echo "输出文件: $OUTPUT_SRT"

# 如果Aegisub可用，尝试使用它
if [ "$AEGISUB_AVAILABLE" = true ]; then
    echo "尝试使用Aegisub..."

    # 创建一个临时的Lua脚本，它将接收命令行参数
    TEMP_LUA_SCRIPT="/tmp/aegisub_gen_$$.lua"
    cat > "$TEMP_LUA_SCRIPT" << 'LUA_EOF'
-- Aegisub字幕生成器（命令行版本）
arg = {os.getenv("AUDIO_FILE"), os.getenv("TEXT_FILE"), os.getenv("OUTPUT_SRT")}

-- SRT 时间格式转换
function ms_to_srt(ms)
    ms = math.max(ms, 0)
    local h = math.floor(ms / 3600000)
    local m = math.floor((ms % 3600000) / 60000)
    local s = math.floor((ms % 60000) / 1000)
    local ms_remain = ms % 1000
    return string.format("%02d:%02d:%02d,%03d", h, m, s, ms_remain)
end

-- 读取文本+统计总字数/每段字数
function read_text(text_path)
    local file, err = io.open(text_path, "rb")
    if not file then
        return nil, nil, 0, "[ERROR] 无法打开文本文件：" .. (err or "未知错误")
    end
    local content = file:read("*a")
    file:close()

    -- 统一换行符+去首尾空行
    content = string.gsub(content, "\r\n", "\n")
    content = string.gsub(content, "\r", "\n")
    content = string.gsub(content, "^%s+", "")
    content = string.gsub(content, "%s+$", "")

    local paras = {}
    local para_words = {}
    local total_words = 0

    for line in string.gmatch(content, "[^\n]+") do
        local trimed = string.gsub(line, "^%s+", "")
        trimed = string.gsub(trimed, "%s+$", "")
        if trimed ~= "" and trimed ~= "544" then
            -- 优化：使用 utf8.len() 统计 UTF-8 字数（Aegisub Lua 内置支持，中文按 1 字计算）
            local wc = utf8.len(trimed) or 0
            table.insert(paras, trimed)
            table.insert(para_words, wc)
            total_words = total_words + wc
        end
    end

    if total_words == 0 then
        return nil, nil, 0, "[ERROR] 文本文件无有效内容！"
    end
    return paras, para_words, total_words, "[INFO] 文本读取成功，总字数：" .. total_words
end

-- 【核心】调用 Aegisub 原生 API 加载音频并获取总时长（无需 ffprobe）
function get_audio_duration(audio_path)
    -- 调用 Aegisub 内置 API open_audio() 加载音频（原生支持，无需外部工具）
    local audio_handle, err = aegisub.open_audio(audio_path)
    if not audio_handle then
        return 0, "[ERROR] Aegisub 加载音频失败：" .. (err or "路径错误或音频格式不支持")
    end

    -- 调用 Aegisub 音频句柄方法 len() 获取音频总时长（单位：毫秒）
    local audio_ms = audio_handle:len()
    -- 转换为秒，方便后续计算
    local audio_sec = audio_ms / 1000.0

    return audio_sec, "[INFO] Aegisub 音频加载成功，总时长：" .. string.format("%.2f", audio_sec) .. " 秒"
end

-- 核心：按字数占比计算每段时长
function calc_timeline(paras, para_words, total_words, audio_sec)
    local timeline = {}
    local current_start = 0.0

    for i = 1, #paras do
        local para_sec = (para_words[i] * audio_sec) / total_words
        local current_end = current_start + para_sec
        
        -- 最后一段强制对齐，消除浮点计算误差
        if i == #paras then
            current_end = audio_sec
        end

        table.insert(timeline, {
            start = current_start,
            ["end"] = current_end,
            text = paras[i],
            wc = para_words[i]
        })
        current_start = current_end
    end
    return timeline
end

-- 生成 SRT 文件
function gen_srt(timeline, output_srt)
    local f, err = io.open(output_srt, "w")
    if not f then
        return "[ERROR] 无法写入 SRT 文件：" .. (err or "权限不足或路径不存在")
    end
    local srt = ""
    for i, seg in ipairs(timeline) do
        local s_ms = math.floor(seg.start * 1000)
        local e_ms = math.floor(seg["end"] * 1000)
        srt = srt .. i .. "\n"
        srt = srt .. ms_to_srt(s_ms) .. " --> " .. ms_to_srt(e_ms) .. "\n"
        srt = srt .. seg.text .. "\n\n"
    end
    f:write(srt)
    f:close()
    return "[SUCCESS] SRT 文件已保存到：" .. output_srt
end

-- 主逻辑
function main()
    -- 步骤 1：获取配置（优先命令行参数，适配 MCP；无参数则使用默认配置，适配界面运行）
    local audio_file, text_file, output_srt
    if arg[1] and arg[2] and arg[3] then
        -- MCP 无界面调用：从命令行获取参数（arg[1]=音频路径，arg[2]=文本路径，arg[3]=SRT 输出路径）
        audio_file = arg[1]
        text_file = arg[2]
        output_srt = arg[3]
    else
        print("错误: 需要提供命令行参数")
        return
    end

    -- 步骤 2：调用 Aegisub API 获取音频时长
    local audio_sec, audio_msg = get_audio_duration(audio_file)
    if audio_sec <= 0 then
        print(audio_msg)
        return
    end

    -- 步骤 3：读取文本并统计字数
    local paras, para_words, total_words, text_msg = read_text(text_file)
    if total_words <= 0 then
        print(text_msg)
        return
    end

    -- 步骤 4：按字数占比计算时间轴
    local timeline = calc_timeline(paras, para_words, total_words, audio_sec)

    -- 步骤 5：生成 SRT 文件
    local srt_msg = gen_srt(timeline, output_srt)
    print(srt_msg)

    -- 步骤 6：输出验证信息
    local avg_speed = total_words / audio_sec
    print(" 按字数占比计算完成！")
    print(" 总字数："..total_words.." | 音频时长："..string.format("%.2f", audio_sec).."秒")
    print("️  平均语速："..string.format("%.2f", avg_speed).."字/秒")
    for i=1,math.min(5,#timeline) do
        local dur = timeline[i]["end"] - timeline[i].start
        print(string.format("   第%d段：%d字  %.2f秒", i, timeline[i].wc, dur))
    end
end

-- 调用主函数
main()
LUA_EOF

    # 设置环境变量供Lua脚本使用
    export AUDIO_FILE="$AUDIO_FILE"
    export TEXT_FILE="$TEXT_FILE"
    export OUTPUT_SRT="$OUTPUT_SRT"
    
    # 启动Aegisub执行自动化脚本
    "$AEGISUB_CMD" --automation-script="$TEMP_LUA_SCRIPT" --quit >/dev/null 2>&1 &
    SCRIPT_PID=$!
    
    # 等待脚本执行完成，最多等待30秒
    TIMEOUT=30
    COUNT=0
    while kill -0 $SCRIPT_PID 2>/dev/null; do
        if [ $COUNT -ge $TIMEOUT ]; then
            echo "Aegisub执行超时，切换到备用方案..."
            kill $SCRIPT_PID 2>/dev/null
            break
        fi
        sleep 1
        COUNT=$((COUNT+1))
    done
    
    # 检查是否成功执行
    if [ $COUNT -lt $TIMEOUT ]; then
        echo "Aegisub字幕生成成功"
        # 清理临时脚本
        rm -f "$TEMP_LUA_SCRIPT"
        exit 0
    else
        echo "Aegisub自动化执行失败，切换到备用方案..."
    fi
    
    # 确保清理临时脚本
    rm -f "$TEMP_LUA_SCRIPT"
fi

# 备用方案：使用Python脚本按比例生成SRT文件
echo "使用备用Python方案生成字幕..."

PYTHON_SCRIPT=$(cat << 'EOF'
import sys
import os
import json
from datetime import timedelta

def read_text(text_path):
    """读取文本并统计总字数/每段字数"""
    with open(text_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # 统一换行符+去首尾空行
    content = content.replace('\r\n', '\n').replace('\r', '\n')
    content = content.strip()
    
    paras = []
    para_words = []
    total_words = 0
    
    for line in content.split('\n'):
        trimmed = line.strip()
        if trimmed and trimmed != "544":  # 忽略特定内容
            # 计算UTF-8字符数
            wc = len(trimmed)
            paras.append(trimmed)
            para_words.append(wc)
            total_words += wc
    
    return paras, para_words, total_words

def get_audio_duration(audio_path):
    """获取音频时长（秒）- 使用ffprobe或估算"""
    try:
        # 尝试使用ffprobe获取准确时长
        import subprocess
        result = subprocess.run(['ffprobe', '-v', 'quiet', '-show_entries', 'format=duration', '-of', 'csv=p=0', audio_path], 
                                capture_output=True, text=True, timeout=10)
        if result.returncode == 0:
            duration = float(result.stdout.strip())
            return duration
    except:
        pass
    
    # 如果ffprobe不可用，则估算（假设每分钟400字的标准朗读速度）
    # 这只是一个近似值，实际使用中可能需要更精确的方法
    print(f"警告: 无法使用ffprobe获取音频时长，使用估算值", file=sys.stderr)
    return 60  # 假设音频长度为60秒

def ms_to_srt(ms):
    """毫秒转SRT时间格式"""
    ms = max(ms, 0)
    td = timedelta(milliseconds=ms)
    total_seconds = td.total_seconds()
    hours = int(total_seconds // 3600)
    minutes = int((total_seconds % 3600) // 60)
    seconds = int(total_seconds % 60)
    milliseconds = int((total_seconds - int(total_seconds)) * 1000)
    return f"{hours:02d}:{minutes:02d}:{seconds:02d},{milliseconds:03d}"

def calc_timeline(paras, para_words, total_words, audio_sec):
    """按字数占比计算时间轴"""
    timeline = []
    current_start = 0.0

    for i in range(len(paras)):
        para_sec = (para_words[i] * audio_sec) / total_words
        current_end = current_start + para_sec
        
        # 最后一段强制对齐，消除浮点计算误差
        if i == len(paras) - 1:
            current_end = audio_sec

        timeline.append({
            'start': current_start,
            'end': current_end,
            'text': paras[i],
            'wc': para_words[i]
        })
        current_start = current_end
    
    return timeline

def gen_srt(timeline, output_srt):
    """生成SRT文件"""
    with open(output_srt, 'w', encoding='utf-8') as f:
        for i, seg in enumerate(timeline, 1):
            s_ms = int(seg['start'] * 1000)
            e_ms = int(seg['end'] * 1000)
            
            f.write(f"{i}\n")
            f.write(f"{ms_to_srt(s_ms)} --> {ms_to_srt(e_ms)}\n")
            f.write(f"{seg['text']}\n\n")

def main():
    if len(sys.argv) != 4:
        print("用法: python script.py <音频文件> <文本文件> <输出SRT文件>")
        sys.exit(1)
    
    audio_file = sys.argv[1]
    text_file = sys.argv[2]
    output_srt = sys.argv[3]
    
    # 读取文本并统计字数
    paras, para_words, total_words = read_text(text_file)
    
    if total_words <= 0:
        print("[ERROR] 文本文件无有效内容！", file=sys.stderr)
        sys.exit(1)
    
    # 获取音频时长
    audio_sec = get_audio_duration(audio_file)
    
    # 按字数占比计算时间轴
    timeline = calc_timeline(paras, para_words, total_words, audio_sec)
    
    # 生成SRT文件
    gen_srt(timeline, output_srt)
    
    print(f"[SUCCESS] SRT 文件已保存到：{output_srt}")
    print(f" 总字数：{total_words} | 音频时长：{audio_sec:.2f}秒")
    avg_speed = total_words / audio_sec if audio_sec > 0 else 0
    print(f"️  平均语速：{avg_speed:.2f}字/秒")

if __name__ == "__main__":
    main()
EOF
)

# 创建临时Python脚本并执行
TEMP_PYTHON_SCRIPT="/tmp/aegisub_fallback_$$.py"
echo "$PYTHON_SCRIPT" > "$TEMP_PYTHON_SCRIPT"

if command -v python3 &> /dev/null; then
    python3 "$TEMP_PYTHON_SCRIPT" "$AUDIO_FILE" "$TEXT_FILE" "$OUTPUT_SRT"
    PYTHON_RESULT=$?
elif command -v python &> /dev/null; then
    python "$TEMP_PYTHON_SCRIPT" "$AUDIO_FILE" "$TEXT_FILE" "$OUTPUT_SRT"
    PYTHON_RESULT=$?
else
    echo "错误: 未找到Python解释器，无法执行备用方案"
    exit 1
fi

# 清理临时文件
rm -f "$TEMP_PYTHON_SCRIPT"

if [ $PYTHON_RESULT -eq 0 ]; then
    echo "备用方案执行成功"
    exit 0
else
    echo "备用方案执行失败"
    exit $PYTHON_RESULT
fi