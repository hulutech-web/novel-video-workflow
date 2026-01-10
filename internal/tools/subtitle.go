package tools

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// SubtitleGenerator 字幕生成器
type SubtitleGenerator struct {
	logger *zap.Logger
}

// SubtitleConfig 字幕配置
type SubtitleConfig struct {
	Style          string  `json:"style"`
	FontName       string  `json:"font_name"`
	FontSize       int     `json:"font_size"`
	PrimaryColor   string  `json:"primary_color"`   // 主要颜色
	SecondaryColor string  `json:"secondary_color"` // 次要颜色
	OutlineColor   string  `json:"outline_color"`   // 轮廓颜色
	BackColor      string  `json:"back_color"`      // 背景颜色
	Bold           bool    `json:"bold"`            // 是否加粗
	Italic         bool    `json:"italic"`          // 是否斜体
	Underline      bool    `json:"underline"`       // 是否下划线
	Alignment      int     `json:"alignment"`       // 对齐方式
	MarginL        int     `json:"margin_l"`        // 左边距
	MarginR        int     `json:"margin_r"`        // 右边距
	MarginV        int     `json:"margin_v"`        // 垂直边距
	Outline        float64 `json:"outline"`         // 轮廓大小
	Shadow         float64 `json:"shadow"`          // 阴影大小
}

// SubtitleLine 单行字幕
type SubtitleLine struct {
	Index     int           `json:"index"`
	StartTime time.Duration `json:"start_time"`
	EndTime   time.Duration `json:"end_time"`
	Text      string        `json:"text"`
	Style     string        `json:"style"`
}

// SubtitleResult 字幕生成结果
type SubtitleResult struct {
	Success      bool    `json:"success"`
	SubtitleFile string  `json:"subtitle_file,omitempty"`
	Format       string  `json:"format,omitempty"`
	LineCount    int     `json:"line_count,omitempty"`
	Duration     float64 `json:"duration,omitempty"`
	Error        string  `json:"error,omitempty"`
}

// ASSStyle ASS字幕样式
type ASSStyle struct {
	Name            string
	FontName        string
	FontSize        int
	PrimaryColour   string
	SecondaryColour string
	OutlineColour   string
	BackColour      string
	Bold            bool
	Italic          bool
	Underline       bool
	BorderStyle     int
	Outline         float64
	Shadow          float64
	Alignment       int
	MarginL         int
	MarginR         int
	MarginV         int
	Encoding        int
}

// NewSubtitleGenerator 创建字幕生成器
func NewSubtitleGenerator(logger *zap.Logger) *SubtitleGenerator {
	return &SubtitleGenerator{
		logger: logger,
	}
}

// GenerateFromAudio 基于音频生成字幕（自动时间轴）
func (sg *SubtitleGenerator) GenerateFromAudio(text, audioFile, outputFile string) (*SubtitleResult, error) {
	sg.logger.Info("基于音频生成字幕",
		zap.String("音频文件", audioFile),
		zap.String("输出文件", outputFile),
	)

	// 1. 获取音频时长
	duration, err := sg.getAudioDuration(audioFile)
	if err != nil {
		sg.logger.Warn("获取音频时长失败，使用默认时长", zap.Error(err))
		duration = 300.0 // 默认5分钟
	}

	sg.logger.Debug("音频时长", zap.Float64("秒", duration))

	// 2. 分割文本
	paragraphs := sg.splitTextToParagraphs(text)
	if len(paragraphs) == 0 {
		return nil, fmt.Errorf("文本为空或无法分割")
	}

	sg.logger.Debug("文本段落数量", zap.Int("段落数", len(paragraphs)))

	// 3. 生成时间轴
	subtitleLines := sg.generateTimeline(paragraphs, duration)

	// 4. 根据文件扩展名决定输出格式
	ext := filepath.Ext(outputFile)
	switch strings.ToLower(ext) {
	case ".ass":
		return sg.generateASS(subtitleLines, outputFile)
	case ".srt":
		return sg.generateSRT(subtitleLines, outputFile)
	case ".vtt":
		return sg.generateVTT(subtitleLines, outputFile)
	default:
		// 默认使用SRT格式
		if outputFile == "" {
			outputFile = strings.TrimSuffix(audioFile, filepath.Ext(audioFile)) + ".srt"
		}
		return sg.generateSRT(subtitleLines, outputFile)
	}
}

// GenerateStatic 生成静态字幕（固定时间间隔）
func (sg *SubtitleGenerator) GenerateStatic(text, outputFile string) (*SubtitleResult, error) {
	sg.logger.Info("生成静态字幕", zap.String("输出文件", outputFile))

	// 分割文本
	paragraphs := sg.splitTextToParagraphs(text)
	if len(paragraphs) == 0 {
		return nil, fmt.Errorf("文本为空或无法分割")
	}

	// 默认每段显示5秒
	var subtitleLines []SubtitleLine
	for i, para := range paragraphs {
		startTime := time.Duration(i*5) * time.Second
		endTime := time.Duration((i+1)*5) * time.Second

		subtitleLines = append(subtitleLines, SubtitleLine{
			Index:     i + 1,
			StartTime: startTime,
			EndTime:   endTime,
			Text:      para,
			Style:     "Default",
		})
	}

	// 生成字幕文件
	ext := filepath.Ext(outputFile)
	switch strings.ToLower(ext) {
	case ".ass":
		return sg.generateASS(subtitleLines, outputFile)
	default:
		return sg.generateSRT(subtitleLines, outputFile)
	}
}

// GenerateWithAegisub 使用Aegisub生成字幕
func (sg *SubtitleGenerator) GenerateWithAegisub(text, audioFile, outputFile string) (*SubtitleResult, error) {
	sg.logger.Info("使用Aegisub生成字幕",
		zap.String("音频文件", audioFile),
		zap.String("输出文件", outputFile),
	)

	// 1. 检查Aegisub是否安装
	if !sg.checkAegisubInstalled() {
		return &SubtitleResult{
			Success: false,
			Error:   "Aegisub未安装，请先安装Aegisub或使用其他字幕生成方式",
		}, nil
	}

	// 2. 创建临时文本文件
	tempDir := os.TempDir()
	tempTextFile := filepath.Join(tempDir, "subtitle_text.txt")

	if err := os.WriteFile(tempTextFile, []byte(text), 0644); err != nil {
		return &SubtitleResult{
			Success: false,
			Error:   fmt.Sprintf("创建临时文本文件失败: %v", err),
		}, err
	}
	defer os.Remove(tempTextFile)

	// 3. 创建Aegisub自动化脚本
	luaScript := sg.createAegisubLuaScript()
	luaScriptFile := filepath.Join(tempDir, "auto_subtitle.lua")

	if err := os.WriteFile(luaScriptFile, []byte(luaScript), 0644); err != nil {
		return &SubtitleResult{
			Success: false,
			Error:   fmt.Sprintf("创建Lua脚本失败: %v", err),
		}, err
	}
	defer os.Remove(luaScriptFile)

	// 4. 调用Aegisub
	result, err := sg.callAegisub(audioFile, tempTextFile, outputFile, luaScriptFile)
	if err != nil {
		sg.logger.Warn("Aegisub调用失败，回退到自动生成", zap.Error(err))
		// 回退到自动生成
		return sg.GenerateFromAudio(text, audioFile, outputFile)
	}

	return result, nil
}

// getAudioDuration 获取音频时长
func (sg *SubtitleGenerator) getAudioDuration(audioFile string) (float64, error) {
	// 使用ffprobe获取音频时长
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioFile,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe执行失败: %v", err)
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("解析时长失败: %v", err)
	}

	return duration, nil
}

// splitTextToParagraphs 分割文本为段落
func (sg *SubtitleGenerator) splitTextToParagraphs(text string) []string {
	// 按空行分割
	paragraphs := strings.Split(text, "\n\n")

	var result []string
	for _, p := range paragraphs {
		trimmed := strings.TrimSpace(p)
		if len(trimmed) > 0 {
			// 如果段落太长，进一步分割
			if len(trimmed) > 200 {
				result = append(result, sg.splitLongParagraph(trimmed)...)
			} else {
				result = append(result, trimmed)
			}
		}
	}

	return result
}

// splitLongParagraph 分割长段落
func (sg *SubtitleGenerator) splitLongParagraph(paragraph string) []string {
	// 按句子分割
	sentences := strings.FieldsFunc(paragraph, func(r rune) bool {
		return r == '。' || r == '！' || r == '？' || r == '.' || r == '!' || r == '?'
	})

	var result []string
	var current strings.Builder

	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)
		if len(trimmed) == 0 {
			continue
		}

		// 添加标点（简化处理）
		trimmed += "。"

		if current.Len()+len(trimmed) <= 100 {
			if current.Len() > 0 {
				current.WriteString(" ")
			}
			current.WriteString(trimmed)
		} else {
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
			current.WriteString(trimmed)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// generateTimeline 生成时间轴
func (sg *SubtitleGenerator) generateTimeline(paragraphs []string, totalDuration float64) []SubtitleLine {
	var subtitleLines []SubtitleLine

	// 计算每段的基本时长
	baseDuration := totalDuration / float64(len(paragraphs))

	// 根据段落长度调整时长
	for i, paragraph := range paragraphs {
		// 基础时长 + 根据字数调整
		wordCount := len([]rune(paragraph))
		adjustment := float64(wordCount) * 0.5 // 每个字0.5秒

		// 限制调整范围
		if adjustment > baseDuration*0.5 {
			adjustment = baseDuration * 0.5
		}

		startTime := time.Duration(float64(i)*baseDuration*1000) * time.Millisecond
		duration := time.Duration((baseDuration+adjustment)*1000) * time.Millisecond
		endTime := startTime + duration

		subtitleLines = append(subtitleLines, SubtitleLine{
			Index:     i + 1,
			StartTime: startTime,
			EndTime:   endTime,
			Text:      paragraph,
			Style:     "Default",
		})
	}

	return subtitleLines
}

// generateSRT 生成SRT格式字幕
func (sg *SubtitleGenerator) generateSRT(subtitleLines []SubtitleLine, outputFile string) (*SubtitleResult, error) {
	var srtContent strings.Builder

	for _, line := range subtitleLines {
		// 序号
		srtContent.WriteString(fmt.Sprintf("%d\n", line.Index))

		// 时间码
		startStr := sg.formatSRTTime(line.StartTime)
		endStr := sg.formatSRTTime(line.EndTime)
		srtContent.WriteString(fmt.Sprintf("%s --> %s\n", startStr, endStr))

		// 文本
		srtContent.WriteString(line.Text)
		srtContent.WriteString("\n\n")
	}

	// 写入文件
	if err := os.WriteFile(outputFile, []byte(srtContent.String()), 0644); err != nil {
		return &SubtitleResult{
			Success: false,
			Error:   fmt.Sprintf("写入SRT文件失败: %v", err),
		}, err
	}

	return &SubtitleResult{
		Success:      true,
		SubtitleFile: outputFile,
		Format:       "srt",
		LineCount:    len(subtitleLines),
		Duration:     subtitleLines[len(subtitleLines)-1].EndTime.Seconds(),
	}, nil
}

// generateASS 生成ASS格式字幕
func (sg *SubtitleGenerator) generateASS(subtitleLines []SubtitleLine, outputFile string) (*SubtitleResult, error) {
	var assContent strings.Builder

	// ASS文件头部
	assContent.WriteString("[Script Info]\n")
	assContent.WriteString("; Script generated by Novel Video Workflow\n")
	assContent.WriteString("Title: Novel Subtitle\n")
	assContent.WriteString("ScriptType: v4.00+\n")
	assContent.WriteString("WrapStyle: 0\n")
	assContent.WriteString("ScaledBorderAndShadow: yes\n")
	assContent.WriteString("YCbCr Matrix: None\n")
	assContent.WriteString(fmt.Sprintf("PlayResX: %d\n", viper.GetInt("video.resolution.width")))
	assContent.WriteString(fmt.Sprintf("PlayResY: %d\n", viper.GetInt("video.resolution.height")))
	assContent.WriteString("\n")

	// 样式定义
	assContent.WriteString("[V4+ Styles]\n")
	assContent.WriteString("Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\n")

	// 默认样式
	defaultStyle := sg.getDefaultASSStyle()
	assContent.WriteString(fmt.Sprintf("Style: %s,%s,%d,%s,%s,%s,%s,%d,%d,%d,0,100,100,0,0,1,%.1f,%.1f,%d,%d,%d,%d,1\n",
		defaultStyle.Name,
		defaultStyle.FontName,
		defaultStyle.FontSize,
		defaultStyle.PrimaryColour,
		defaultStyle.SecondaryColour,
		defaultStyle.OutlineColour,
		defaultStyle.BackColour,
		boolToInt(defaultStyle.Bold),
		boolToInt(defaultStyle.Italic),
		boolToInt(defaultStyle.Underline),
		defaultStyle.Outline,
		defaultStyle.Shadow,
		defaultStyle.Alignment,
		defaultStyle.MarginL,
		defaultStyle.MarginR,
		defaultStyle.MarginV,
	))

	// 事件部分
	assContent.WriteString("\n[Events]\n")
	assContent.WriteString("Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n")

	// 字幕行
	for _, line := range subtitleLines {
		startStr := sg.formatASSTime(line.StartTime)
		endStr := sg.formatASSTime(line.EndTime)

		assContent.WriteString(fmt.Sprintf("Dialogue: 0,%s,%s,%s,,0,0,0,,%s\n",
			startStr,
			endStr,
			line.Style,
			line.Text,
		))
	}

	// 写入文件
	if err := os.WriteFile(outputFile, []byte(assContent.String()), 0644); err != nil {
		return &SubtitleResult{
			Success: false,
			Error:   fmt.Sprintf("写入ASS文件失败: %v", err),
		}, err
	}

	return &SubtitleResult{
		Success:      true,
		SubtitleFile: outputFile,
		Format:       "ass",
		LineCount:    len(subtitleLines),
		Duration:     subtitleLines[len(subtitleLines)-1].EndTime.Seconds(),
	}, nil
}

// generateVTT 生成WebVTT格式字幕
func (sg *SubtitleGenerator) generateVTT(subtitleLines []SubtitleLine, outputFile string) (*SubtitleResult, error) {
	var vttContent strings.Builder

	// VTT文件头部
	vttContent.WriteString("WEBVTT\n\n")

	// 字幕行
	for _, line := range subtitleLines {
		vttContent.WriteString(fmt.Sprintf("%d\n", line.Index))

		// 时间码
		startStr := sg.formatVTTTime(line.StartTime)
		endStr := sg.formatVTTTime(line.EndTime)
		vttContent.WriteString(fmt.Sprintf("%s --> %s\n", startStr, endStr))

		// 文本
		vttContent.WriteString(line.Text)
		vttContent.WriteString("\n\n")
	}

	// 写入文件
	if err := os.WriteFile(outputFile, []byte(vttContent.String()), 0644); err != nil {
		return &SubtitleResult{
			Success: false,
			Error:   fmt.Sprintf("写入VTT文件失败: %v", err),
		}, err
	}

	return &SubtitleResult{
		Success:      true,
		SubtitleFile: outputFile,
		Format:       "vtt",
		LineCount:    len(subtitleLines),
		Duration:     subtitleLines[len(subtitleLines)-1].EndTime.Seconds(),
	}, nil
}

// 时间格式化辅助函数
func (sg *SubtitleGenerator) formatSRTTime(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	milliseconds := int(d.Milliseconds()) % 1000

	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, seconds, milliseconds)
}

func (sg *SubtitleGenerator) formatASSTime(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	centiseconds := int(d.Milliseconds()/10) % 100

	return fmt.Sprintf("%d:%02d:%02d.%02d", hours, minutes, seconds, centiseconds)
}

func (sg *SubtitleGenerator) formatVTTTime(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	milliseconds := int(d.Milliseconds()) % 1000

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
}

// 辅助函数
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (sg *SubtitleGenerator) getDefaultASSStyle() ASSStyle {
	return ASSStyle{
		Name:            "Default",
		FontName:        viper.GetString("subtitle.font_name"),
		FontSize:        viper.GetInt("subtitle.font_size"),
		PrimaryColour:   viper.GetString("subtitle.primary_color"),
		SecondaryColour: viper.GetString("subtitle.secondary_color"),
		OutlineColour:   viper.GetString("subtitle.outline_color"),
		BackColour:      viper.GetString("subtitle.back_color"),
		Bold:            viper.GetBool("subtitle.bold"),
		Italic:          viper.GetBool("subtitle.italic"),
		Underline:       viper.GetBool("subtitle.underline"),
		BorderStyle:     1,
		Outline:         viper.GetFloat64("subtitle.outline"),
		Shadow:          viper.GetFloat64("subtitle.shadow"),
		Alignment:       viper.GetInt("subtitle.alignment"),
		MarginL:         viper.GetInt("subtitle.margin_l"),
		MarginR:         viper.GetInt("subtitle.margin_r"),
		MarginV:         viper.GetInt("subtitle.margin_v"),
		Encoding:        1,
	}
}

// checkAegisubInstalled 检查Aegisub是否安装
func (sg *SubtitleGenerator) checkAegisubInstalled() bool {
	// 方法1: 检查应用程序
	appPath := "/Applications/Aegisub.app"
	if _, err := os.Stat(appPath); err == nil {
		return true
	}

	// 方法2: 检查命令行
	cmd := exec.Command("which", "aegisub")
	if err := cmd.Run(); err == nil {
		return true
	}

	return false
}

// createAegisubLuaScript 创建Aegisub自动化脚本
func (sg *SubtitleGenerator) createAegisubLuaScript() string {
	return `-- Aegisub 自动化脚本
-- 自动生成字幕时间轴

function auto_generate_subtitles(audio_file, text_file, output_file)
    -- 加载音频
    local audio = loadaudio(audio_file)
    if not audio then
        error("无法加载音频文件: " .. audio_file)
    end
    
    -- 读取文本
    local f = io.open(text_file, "r")
    if not f then
        error("无法打开文本文件: " .. text_file)
    end
    local text = f:read("*all")
    f:close()
    
    -- 清除现有字幕
    aegisub.progress.task("清除现有字幕...")
    local subs = {}
    
    -- 分割文本为段落
    aegisub.progress.task("分割文本...")
    local paragraphs = {}
    for para in text:gmatch("([^\n]+)") do
        if #para > 0 then
            table.insert(paragraphs, para)
        end
    end
    
    -- 计算时间轴
    aegisub.progress.task("计算时间轴...")
    local total_duration = audio.duration
    local paragraph_count = #paragraphs
    
    if paragraph_count == 0 then
        error("文本为空")
    end
    
    -- 智能时间分配
    local current_time = 0
    for i, para in ipairs(paragraphs) do
        -- 根据段落长度计算显示时间
        local para_length = #para
        local base_time = 3.0  -- 基础显示时间
        
        -- 根据字数调整
        if para_length > 100 then
            base_time = base_time + (para_length - 100) * 0.05
        end
        
        -- 限制最大显示时间
        if base_time > 8.0 then
            base_time = 8.0
        end
        
        -- 计算开始和结束时间
        local start_time = current_time
        local end_time = start_time + base_time
        
        -- 如果超过音频时长，进行调整
        if end_time > total_duration then
            end_time = total_duration
            base_time = end_time - start_time
        end
        
        -- 创建字幕行
        local line = {
            layer = 0,
            start_time = start_time * 1000,  -- 转换为毫秒
            end_time = end_time * 1000,
            style = "Default",
            actor = "",
            margin_l = 0,
            margin_r = 0,
            margin_v = 0,
            effect = "",
            text = para
        }
        
        table.insert(subs, line)
        current_time = end_time + 0.1  -- 添加0.1秒间隔
        
        -- 如果已经超过音频时长，停止
        if current_time >= total_duration then
            break
        end
    end
    
    -- 保存字幕
    aegisub.progress.task("保存字幕文件...")
    saveass(output_file, subs)
    
    return #subs
end

-- 主程序
local audio_file = arg[1]
local text_file = arg[2]
local output_file = arg[3]

if audio_file and text_file and output_file then
    local success, result = pcall(function()
        return auto_generate_subtitles(audio_file, text_file, output_file)
    end)
    
    if success then
        print("生成字幕成功，共" .. result .. "行")
    else
        print("生成字幕失败: " .. result)
        os.exit(1)
    end
else
    print("用法: aegisub --automation script.lua <音频文件> <文本文件> <输出文件>")
    os.exit(1)
end`
}

// callAegisub 调用Aegisub执行自动化脚本
func (sg *SubtitleGenerator) callAegisub(audioFile, textFile, outputFile, luaScript string) (*SubtitleResult, error) {
	// 构建命令
	cmd := exec.Command("aegisub",
		"--automation", luaScript,
		audioFile,
		textFile,
		outputFile,
	)

	// 设置环境变量
	cmd.Env = os.Environ()

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &SubtitleResult{
			Success: false,
			Error:   fmt.Sprintf("Aegisub执行失败: %v\n输出: %s", err, output),
		}, err
	}

	sg.logger.Debug("Aegisub输出", zap.String("output", string(output)))

	// 检查输出文件
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		return &SubtitleResult{
			Success: false,
			Error:   "Aegisub未生成输出文件",
		}, fmt.Errorf("输出文件不存在")
	}

	// 计算行数
	lineCount, err := sg.countSubtitleLines(outputFile)
	if err != nil {
		sg.logger.Warn("统计字幕行数失败", zap.Error(err))
		lineCount = 0
	}

	return &SubtitleResult{
		Success:      true,
		SubtitleFile: outputFile,
		Format:       strings.ToLower(filepath.Ext(outputFile))[1:], // 去掉点
		LineCount:    lineCount,
	}, nil
}

// countSubtitleLines 统计字幕行数
func (sg *SubtitleGenerator) countSubtitleLines(subtitleFile string) (int, error) {
	ext := strings.ToLower(filepath.Ext(subtitleFile))

	switch ext {
	case ".srt":
		return sg.countSRTLines(subtitleFile)
	case ".ass":
		return sg.countASSLines(subtitleFile)
	default:
		return 0, fmt.Errorf("不支持的格式: %s", ext)
	}
}

// countSRTLines 统计SRT字幕行数
func (sg *SubtitleGenerator) countSRTLines(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "-->") {
			count++
		}
	}

	return count, scanner.Err()
}

// countASSLines 统计ASS字幕行数
func (sg *SubtitleGenerator) countASSLines(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	inEventsSection := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "[Events]") {
			inEventsSection = true
			continue
		}

		if inEventsSection && strings.HasPrefix(line, "Dialogue:") {
			count++
		}

		if inEventsSection && strings.HasPrefix(line, "[") && !strings.HasPrefix(line, "[Events]") {
			break
		}
	}

	return count, scanner.Err()
}

// ConvertFormat 转换字幕格式
func (sg *SubtitleGenerator) ConvertFormat(inputFile, outputFormat string) (string, error) {
	sg.logger.Info("转换字幕格式",
		zap.String("输入文件", inputFile),
		zap.String("目标格式", outputFormat),
	)

	// 读取输入文件
	content, err := os.ReadFile(inputFile)
	if err != nil {
		return "", fmt.Errorf("读取字幕文件失败: %w", err)
	}

	// 解析字幕
	var subtitleLines []SubtitleLine
	ext := strings.ToLower(filepath.Ext(inputFile))

	switch ext {
	case ".srt":
		subtitleLines, err = sg.parseSRT(string(content))
	case ".ass":
		subtitleLines, err = sg.parseASS(string(content))
	default:
		return "", fmt.Errorf("不支持的输入格式: %s", ext)
	}

	if err != nil {
		return "", fmt.Errorf("解析字幕失败: %w", err)
	}

	// 生成输出文件名
	outputFile := strings.TrimSuffix(inputFile, ext) + "." + outputFormat

	// 生成输出文件
	switch outputFormat {
	case "srt":
		_, err = sg.generateSRT(subtitleLines, outputFile)
	case "ass":
		_, err = sg.generateASS(subtitleLines, outputFile)
	case "vtt":
		_, err = sg.generateVTT(subtitleLines, outputFile)
	default:
		return "", fmt.Errorf("不支持的输出格式: %s", outputFormat)
	}

	if err != nil {
		return "", fmt.Errorf("生成字幕失败: %w", err)
	}

	return outputFile, nil
}

// parseSRT 解析SRT格式字幕
func (sg *SubtitleGenerator) parseSRT(content string) ([]SubtitleLine, error) {
	var lines []SubtitleLine
	scanner := bufio.NewScanner(strings.NewReader(content))

	var currentLine SubtitleLine
	var lineText strings.Builder
	state := 0 // 0: 等待序号, 1: 等待时间码, 2: 收集文本

	for scanner.Scan() {
		text := scanner.Text()

		switch state {
		case 0: // 等待序号
			if _, err := strconv.Atoi(text); err == nil {
				currentLine.Index, _ = strconv.Atoi(text)
				state = 1
			}
		case 1: // 等待时间码
			if strings.Contains(text, "-->") {
				parts := strings.Split(text, "-->")
				if len(parts) == 2 {
					startTime, err1 := sg.parseSRTTime(parts[0])
					endTime, err2 := sg.parseSRTTime(parts[1])

					if err1 == nil && err2 == nil {
						currentLine.StartTime = startTime
						currentLine.EndTime = endTime
						state = 2
					}
				}
			}
		case 2: // 收集文本
			if text == "" {
				// 空行表示当前字幕结束
				currentLine.Text = lineText.String()
				lines = append(lines, currentLine)

				// 重置状态
				currentLine = SubtitleLine{}
				lineText.Reset()
				state = 0
			} else {
				if lineText.Len() > 0 {
					lineText.WriteString("\n")
				}
				lineText.WriteString(text)
			}
		}
	}

	// 处理最后一行
	if lineText.Len() > 0 {
		currentLine.Text = lineText.String()
		lines = append(lines, currentLine)
	}

	return lines, scanner.Err()
}

// parseASS 解析ASS格式字幕
func (sg *SubtitleGenerator) parseASS(content string) ([]SubtitleLine, error) {
	var lines []SubtitleLine
	scanner := bufio.NewScanner(strings.NewReader(content))

	inEventsSection := false
	index := 1

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "[Events]") {
			inEventsSection = true
			continue
		}

		if inEventsSection && strings.HasPrefix(line, "Dialogue:") {
			// 解析Dialogue行
			parts := strings.SplitN(line, ",", 10)
			if len(parts) >= 10 {
				startTime, err1 := sg.parseASSTime(parts[1])
				endTime, err2 := sg.parseASSTime(parts[2])
				style := parts[3]
				text := parts[9]

				if err1 == nil && err2 == nil {
					lines = append(lines, SubtitleLine{
						Index:     index,
						StartTime: startTime,
						EndTime:   endTime,
						Style:     style,
						Text:      text,
					})
					index++
				}
			}
		}

		if inEventsSection && strings.HasPrefix(line, "[") && !strings.HasPrefix(line, "[Events]") {
			break
		}
	}

	return lines, scanner.Err()
}

// parseSRTTime 解析SRT时间格式
func (sg *SubtitleGenerator) parseSRTTime(timeStr string) (time.Duration, error) {
	timeStr = strings.TrimSpace(timeStr)

	// SRT格式: HH:MM:SS,mmm
	parts := strings.Split(timeStr, ",")
	if len(parts) != 2 {
		return 0, fmt.Errorf("无效的SRT时间格式: %s", timeStr)
	}

	timePart := parts[0]
	millisPart := parts[1]

	// 解析时分秒
	timeParts := strings.Split(timePart, ":")
	if len(timeParts) != 3 {
		return 0, fmt.Errorf("无效的时间部分: %s", timePart)
	}

	hours, err1 := strconv.Atoi(timeParts[0])
	minutes, err2 := strconv.Atoi(timeParts[1])
	seconds, err3 := strconv.Atoi(timeParts[2])
	milliseconds, err4 := strconv.Atoi(millisPart)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return 0, fmt.Errorf("解析时间数字失败: %s", timeStr)
	}

	totalMillis := (hours*3600+minutes*60+seconds)*1000 + milliseconds
	return time.Duration(totalMillis) * time.Millisecond, nil
}

// parseASSTime 解析ASS时间格式
func (sg *SubtitleGenerator) parseASSTime(timeStr string) (time.Duration, error) {
	timeStr = strings.TrimSpace(timeStr)

	// ASS格式: H:MM:SS.cc
	parts := strings.Split(timeStr, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("无效的ASS时间格式: %s", timeStr)
	}

	timePart := parts[0]
	centiseconds, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("解析厘秒失败: %s", parts[1])
	}

	// 解析时分秒
	timeParts := strings.Split(timePart, ":")
	if len(timeParts) != 3 {
		return 0, fmt.Errorf("无效的时间部分: %s", timePart)
	}

	hours, err1 := strconv.Atoi(timeParts[0])
	minutes, err2 := strconv.Atoi(timeParts[1])
	seconds, err3 := strconv.Atoi(timeParts[2])

	if err1 != nil || err2 != nil || err3 != nil {
		return 0, fmt.Errorf("解析时间数字失败: %s", timeStr)
	}

	totalMillis := (hours*3600+minutes*60+seconds)*1000 + centiseconds*10
	return time.Duration(totalMillis) * time.Millisecond, nil
}
