-- 脚本名称：Aegisub 原生 API 字数占比 SRT 生成器
-- 脚本描述：调用 Aegisub 内置 API 加载音频，按字数占比分配时长，生成 SRT 文件
-- 支持：1. 界面运行（硬编码配置） 2. MCP 无界面调用（命令行传参）
script_name = "Aegisub 原生 API_SRT 生成器（字数占比）"
script_description = "调用 Aegisub API 加载音频，按字数占比生成 SRT，支持 MCP 自动化"
script_author = "Custom Script"
script_version = "51.0"

-- 【可选】界面运行配置（手动运行时修改，MCP 调用时会被命令行参数覆盖）
local default_config = {
    text_file = "/Users/mac/Documents/ai/chapter6/novel06.txt",
    audio_file = "/Users/mac/Documents/ai/chapter6/spk_1767952937.wav",
    output_srt = "/Users/mac/Documents/ai/chapter6/novel_word_ratio_final.srt",
}

-- SRT 时间格式转换（保留原逻辑，优化变量名）
function ms_to_srt(ms)
    ms = math.max(ms, 0)
    local h = math.floor(ms / 3600000)
    local m = math.floor((ms % 3600000) / 60000)
    local s = math.floor((ms % 60000) / 1000)
    local ms_remain = ms % 1000
    return string.format("%02d:%02d:%02d,%03d", h, m, s, ms_remain)
end

-- 读取文本+统计总字数/每段字数（优化：UTF-8 中文字数统计，支持中文）
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

-- 核心：按字数占比计算每段时长（保留你的原始逻辑）
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

-- 生成 SRT 文件（保留原逻辑，优化容错）
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

-- 主逻辑（兼容界面运行 + MCP 无界面运行）
function main()
    -- 步骤 1：获取配置（优先命令行参数，适配 MCP；无参数则使用默认配置，适配界面运行）
    local audio_file, text_file, output_srt
    if arg[1] and arg[2] and arg[3] then
        -- MCP 无界面调用：从命令行获取参数（arg[1]=音频路径，arg[2]=文本路径，arg[3]=SRT 输出路径）
        audio_file = arg[1]
        text_file = arg[2]
        output_srt = arg[3]
    else
        -- 界面手动运行：使用默认硬编码配置
        audio_file = default_config.audio_file
        text_file = default_config.text_file
        output_srt = default_config.output_srt
    end

    -- 步骤 2：调用 Aegisub API 获取音频时长
    local audio_sec, audio_msg = get_audio_duration(audio_file)
    aegisub.debug.out(audio_msg)
    if audio_sec <= 0 then
        aegisub.cancel()
        return
    end

    -- 步骤 3：读取文本并统计字数
    local paras, para_words, total_words, text_msg = read_text(text_file)
    aegisub.debug.out(text_msg)
    if total_words <= 0 then
        aegisub.cancel()
        return
    end

    -- 步骤 4：按字数占比计算时间轴
    local timeline = calc_timeline(paras, para_words, total_words, audio_sec)

    -- 步骤 5：生成 SRT 文件
    local srt_msg = gen_srt(timeline, output_srt)
    aegisub.debug.out(srt_msg)

    -- 步骤 6：输出验证信息
    local avg_speed = total_words / audio_sec
    aegisub.debug.out(" 按字数占比计算完成！")
    aegisub.debug.out(" 总字数："..total_words.." | 音频时长："..string.format("%.2f", audio_sec).."秒")
    aegisub.debug.out("️  平均语速："..string.format("%.2f", avg_speed).."字/秒")
    aegisub.debug.out(" 前5段时长（按占比分配）：")
    for i=1,math.min(5,#timeline) do
        local dur = timeline[i]["end"] - timeline[i].start
        aegisub.debug.out(string.format("   第%d段：%d字  %.2f秒", i, timeline[i].wc, dur))
    end
end

-- 注册 Aegisub 宏（支持界面运行），同时兼容无界面 MCP 调用
aegisub.register_macro(script_name, script_description, main)