// Package capcut 定义剪映项目生成器
// 从output目录读取其他MCP工具生成的音频、图片和字幕文件，生成剪映项目
package capcut

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"novel-video-workflow/pkg/capcut/internal/material"
	"novel-video-workflow/pkg/capcut/internal/script"
	"novel-video-workflow/pkg/capcut/internal/segment"
	"novel-video-workflow/pkg/capcut/internal/srt"
	"novel-video-workflow/pkg/capcut/internal/track"
	"novel-video-workflow/pkg/capcut/internal/types"

	"github.com/google/uuid"
)

// CapcutGenerator 剪映项目生成器
type CapcutGenerator struct {
	Logger interface{} // 可以传入zap.Logger或其他日志记录器
}

// NewCapcutGenerator 创建新的剪映项目生成器
func NewCapcutGenerator(logger interface{}) *CapcutGenerator {
	return &CapcutGenerator{
		Logger: logger,
	}
}

// 清理路径中的特殊字符
func cleanPath(path string) string {
	// 创建一个新的字符串构建器
	var cleaned strings.Builder
	
	// 遍历每个字符，过滤掉控制字符，但保留中文等Unicode字符
	for _, r := range path {
		// 只过滤掉真正的控制字符（0-31），保留可打印ASCII、Unicode字符（如中文）
		if r >= 32 || r == 10 || r == 13 || r == 9 { // 32以上包括可打印ASCII和Unicode字符（如中文）
			cleaned.WriteRune(r)
		}
	}
	return cleaned.String()
}

// getAudioDuration 获取音频文件的实际时长（微秒）
func getAudioDuration(audioFilePath string) (int64, error) {
	// 检查文件是否存在
	if _, err := os.Stat(audioFilePath); os.IsNotExist(err) {
		return 0, fmt.Errorf("音频文件不存在: %s", audioFilePath)
	}

	// 使用 ffprobe 获取音频时长
	cmd := exec.Command("ffprobe", "-v", "quiet", "-show_entries", "format=duration", "-of", "csv=p=0", audioFilePath)
	output, err := cmd.Output()
	if err != nil {
		// 尝试检查ffprobe是否可用
		if _, err := exec.LookPath("ffprobe"); err != nil {
			return 0, fmt.Errorf("系统中未找到ffprobe命令，请确保已安装FFmpeg: %v", err)
		}
		return 0, fmt.Errorf("ffprobe命令执行失败: %v", err)
	}

	// 解析输出的时长（秒）
	durationStr := strings.TrimSpace(string(output))
	durationSec, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("无法解析音频时长: %v", err)
	}

	// 检查解析到的时长是否有效
	if durationSec <= 0 {
		return 0, fmt.Errorf("解析到无效的音频时长: %f", durationSec)
	}

	// 转换为微秒
	return int64(durationSec * 1000000), nil
}

// findJianyingDraftFolder 查找剪映草稿文件夹
func findJianyingDraftFolder() (string, error) {
	// 尝常见路径
	possiblePaths := []string{
		filepath.Join(os.Getenv("HOME"), "Movies", "JianyingPro", "User Data", "Projects", "com.lveditor.draft"),
		filepath.Join(os.Getenv("HOME"), "Movies", "CapCut", "User Data", "Projects", "com.lveditor.draft"),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("未找到剪映草稿文件夹")
}

// GenerateProject 生成剪映项目
// 输入目录应该是output/小说名称/chapter_XX，包含音频、图片和字幕文件
func (cg *CapcutGenerator) GenerateProject(inputDir string) error {
	// 获取输入目录的绝对路径
	inputDir, err := filepath.Abs(inputDir)
	if err != nil {
		return fmt.Errorf("获取输入目录绝对路径失败: %v", err)
	}

	// 清理输入目录路径中的特殊字符
	inputDir = cleanPath(inputDir)

	// 检查必要的文件
	audioFile := ""
	imageFiles := []string{}
	srtFile := ""

	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("读取输入目录失败: %v", err)
	}

	for _, file := range files {
		filename := strings.ToLower(file.Name())
		if strings.HasSuffix(filename, ".wav") || strings.HasSuffix(filename, ".mp3") {
			audioFile = cleanPath(filepath.Join(inputDir, file.Name())) // 清理音频文件路径
		} else if strings.HasSuffix(filename, ".png") || strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".jpeg") {
			imageFiles = append(imageFiles, cleanPath(filepath.Join(inputDir, file.Name()))) // 清理图片文件路径
		} else if strings.HasSuffix(filename, ".srt") {
			srtFile = cleanPath(filepath.Join(inputDir, file.Name())) // 清理字幕文件路径
		}
	}

	if audioFile == "" {
		return fmt.Errorf("未找到音频文件")
	}

	if len(imageFiles) == 0 {
		return fmt.Errorf("未找到图片文件")
	}

	// 获取音频文件实际时长
	audioDuration, err := getAudioDuration(audioFile)
	if err != nil {
		return fmt.Errorf("获取音频时长失败: %v", err)
	}

	// 计算音频时长（秒）
	audioDurationSec := float64(audioDuration) / 1000000.0

	// 计算台词总字数
	totalSubtitleChars := 0
	if srtFile != "" {
		srtEntries, err := srt.ParseSrtFile(srtFile)
		if err != nil {
			fmt.Printf("解析字幕文件失败: %v\n", err)
		} else {
			for _, entry := range srtEntries {
				totalSubtitleChars += len([]rune(entry.Text))
			}
		}
	}

	// 输出日志信息
	fmt.Printf("🎵 音频时长: %.2f秒 (%d微秒)\n", audioDurationSec, audioDuration)
	fmt.Printf("🖼️  图片资源: %d张\n", len(imageFiles))
	fmt.Printf("📄 视频资源: 1个 (音频文件: %s)\n", filepath.Base(audioFile))
	fmt.Printf("💬 台词字数: %d个字符\n", totalSubtitleChars)

	// 创建草稿文件 (1080x1920 手机竖屏视频)
	sf, err := script.NewScriptFile(1080, 1920, 30) // 宽度、高度、帧率
	if err != nil {
		return fmt.Errorf("创建草稿文件失败: %v", err)
	}

	// 设置草稿的基本信息
	sf.Duration = audioDuration

	// 计算每个图片的显示时间（平均分配音频总时长）
	numScenes := len(imageFiles)
	sceneDuration := audioDuration / int64(numScenes)

	// 添加图片素材到草稿
	for i, imageFile := range imageFiles {
		relPath := imageFile // 使用原始路径，NewVideoMaterial会自动转换为绝对路径
		imageName := filepath.Base(imageFile)
		videoMaterial, err := material.NewVideoMaterial(
			material.MaterialTypePhoto, // 静态图片
			&relPath,                   // 文件路径 (NewVideoMaterial会自动转换为绝对路径)
			nil,                        // 替换路径 (不需要，使用原始路径)
			&imageName,                 // 素材名称
			nil,                        // 远程URL
			nil,                        // 裁剪设置
			nil,                        // 时长
			nil,                        // 宽度
			nil,                        // 高度
		)
		if err != nil {
			fmt.Printf("创建视频素材失败: %v\n", err)
			continue
		}
		sf.AddMaterial(videoMaterial)

		// 添加到视频轨道
		videoTrack, err := sf.GetTrack("video", nil)
		if err != nil {
			videoTrackName := stringPtr(fmt.Sprintf("视频轨道_%d", i))
			sf.AddTrack(track.TrackTypeVideo, videoTrackName)
			videoTrack, _ = sf.GetTrack("video", videoTrackName)
		}

		startTime := int64(i) * sceneDuration
		endTime := startTime + sceneDuration

		// 确保最后一张图片精确结束于音频末尾
		if i == numScenes-1 {
			endTime = audioDuration
		}

		sourceTimeRange := types.NewTimerange(startTime, endTime-startTime)
		targetTimeRange := types.NewTimerange(startTime, endTime-startTime)

		videoSegment := segment.NewVideoSegment(
			videoMaterial.MaterialID, // materialID
			sourceTimeRange,          // sourceTimerange
			targetTimeRange,          // targetTimerange
			1.0,                      // speed
			1.0,                      // volume
			nil,                      // clipSettings
		)

		videoTrack.AddSegment(videoSegment)
	}

	// 添加音频素材到草稿
	audioFileName := filepath.Base(audioFile)
	audioMaterial, err := material.NewAudioMaterial(
		&audioFile,                             // 文件路径 (NewAudioMaterial会自动转换为绝对路径)
		nil,                                    // 替换路径 (不需要，使用原始路径)
		&audioFileName,                         // 素材名称
		nil,                                    // 远程URL
		float64Ptr(float64(audioDuration)/1e6), // 时长（秒）
	)
	if err != nil {
		return fmt.Errorf("创建音频素材失败: %v", err)
	} else {
		sf.AddMaterial(audioMaterial)

		// 添加到音频轨道
		audioTrackName := stringPtr("音频轨道")
		sf.AddTrack(track.TrackTypeAudio, audioTrackName)

		// 获取刚刚添加的音频轨道
		audioTrack, err := sf.GetTrack("audio", audioTrackName)
		if err != nil {
			return fmt.Errorf("获取音频轨道失败: %v", err)
		}

		audioSegment := segment.NewAudioSegment(
			audioMaterial.MaterialID,             // materialID
			types.NewTimerange(0, audioDuration), // targetTimerange - 整个音频时长
			nil,                                  // sourceTimerange
			1.0,                                  // speed
			1.0,                                  // volume
		)

		err = audioTrack.AddSegment(audioSegment)
		if err != nil {
			return fmt.Errorf("向音频轨道添加片段失败: %v", err)
		}
	}

	// 如果有SRT字幕文件，则添加字幕
	if srtFile != "" {
		srtEntries, err := srt.ParseSrtFile(srtFile)
		if err != nil {
			return fmt.Errorf("解析字幕文件失败: %v", err)
		} else {
			// 重新计算字幕时间戳，使其与音频总时长相匹配
			// 首先获取原始字幕总时长
			var originalSubtitleDuration int64
			if len(srtEntries) > 0 {
				lastEntry := srtEntries[len(srtEntries)-1]
				originalSubtitleDuration = lastEntry.End
			}

			// 添加文本轨道和字幕
			textTrackName := stringPtr("字幕轨道")
			sf.AddTrack(track.TrackTypeText, textTrackName)

			// 获取文本轨道并添加字幕片段
			textTrack, err := sf.GetTrack("text", textTrackName)
			if err != nil {
				return fmt.Errorf("获取文本轨道失败: %v", err)
			} else {
				for _, entry := range srtEntries {
					// 根据音频时长与原始字幕时长的比例调整字幕时间
					var adjustedStart, adjustedEnd int64
					if originalSubtitleDuration > 0 {
						// 按比例调整时间戳
						ratio := float64(audioDuration) / float64(originalSubtitleDuration)
						adjustedStart = int64(float64(entry.Start) * ratio)
						adjustedEnd = int64(float64(entry.End) * ratio)
						
						// 确保最后一个字幕精确结束于音频末尾
						if entry.End == originalSubtitleDuration && entry.End > 0 {
							adjustedEnd = audioDuration
						}
					} else {
						// 如果无法计算比例，直接使用原始时间
						adjustedStart = entry.Start
						adjustedEnd = entry.End
					}

					// 创建文本样式
					textStyle := segment.NewTextStyle()
					textStyle.Size = 24.0
					textStyle.Color = [3]float64{1.0, 1.0, 1.0} // 白色
					textStyle.Bold = true
					textStyle.Align = 1 // 居中对齐

					// 创建ClipSettings来设置字幕位置，使其显示在画面下方
					clipSettings := segment.NewClipSettingsWithParams(
						1.0,   // alpha
						0.0,   // rotation
						1.0,   // scaleX
						1.0,   // scaleY
						0.0,   // transformX
						-0.8,  // transformY - 负值使字幕靠下显示
						false, // flipH
						false, // flipV
					)

					// 创建文本素材并添加到素材库
					textMaterial := map[string]interface{}{
						"add_type":                     2,
						"alignment":                    1,
						"background_alpha":             1.0,
						"background_color":             "",
						"background_height":            1.0,
						"background_horizontal_offset": 0.0,
						"background_round_radius":      0.0,
						"background_vertical_offset":   0.0,
						"background_width":             1.0,
						"bold_width":                   0.0,
						"border_color":                 "",
						"border_width":                 0.08,
						"check_flag":                   7,
						"content":                      fmt.Sprintf("<font id=\"%s\" path=\"/Applications/VideoFusion-macOS.app/Contents/Resources/Font/SystemFont/zh-hans.ttf\"><color=(1.000000, 1.000000, 1.000000, 1.000000)><size=5.000000>%s</size></color></font>", uuid.New().String(), strings.ReplaceAll(entry.Text, "\n", "\u0001")),
						"font_category_id":             "",
						"font_category_name":           "",
						"font_id":                      "",
						"font_name":                    "",
						"font_path":                    "/Applications/VideoFusion-macOS.app/Contents/Resources/Font/SystemFont/zh-hans.ttf",
						"font_resource_id":             "",
						"font_size":                    5.0,
						"font_title":                   "none",
						"font_url":                     "",
						"fonts":                        []interface{}{},
						"global_alpha":                 1.0,
						"has_shadow":                   false,
						"id":                           uuid.New().String(), // 生成唯一ID
						"initial_scale":                1.0,
						"is_rich_text":                 false,
						"italic_degree":                0,
						"ktv_color":                    "",
						"layer_weight":                 1,
						"letter_spacing":               0.0,
						"line_spacing":                 0.02,
						"recognize_type":               0,
						"shadow_alpha":                 0.8,
						"shadow_angle":                 -45.0,
						"shadow_color":                 "",
						"shadow_distance":              8.0,
						"shadow_point":                 map[string]interface{}{"x": 1.0182337649086284, "y": -1.0182337649086284},
						"shadow_smoothing":             1.0,
						"shape_clip_x":                 false,
						"shape_clip_y":                 false,
						"style_name":                   "",
						"sub_type":                     0,
						"text_alpha":                   1.0,
						"text_color":                   "#FFFFFF",
						"text_size":                    30,
						"text_to_audio_ids":            []interface{}{},
						"type":                         "subtitle",
						"typesetting":                  0,
						"underline":                    false,
						"underline_offset":             0.22,
						"underline_width":              0.05,
						"use_effect_default_color":     true,
					}
					// 将文本素材添加到素材库
					sf.Materials.Texts = append(sf.Materials.Texts, textMaterial)

					// 创建文本片段，使用刚添加的文本素材ID
					textSegment := segment.NewTextSegment(
						entry.Text, // text
						types.NewTimerange(adjustedStart, adjustedEnd-adjustedStart), // targetTimerange - 调整后的时间
						"",           // font (空字符串使用默认字体)
						textStyle,    // style
						clipSettings, // clipSettings - 添加位置设置
					)
					// 设置正确的MaterialID（使用刚添加的文本素材ID）
					textSegment.MaterialID = textMaterial["id"].(string)

					textTrack.AddSegment(textSegment)
				}
			}
		}
	}

	// 生成项目ID
	projectID := uuid.New().String()

	// 将草稿内容写入临时文件
	outputPath := filepath.Join("output", projectID+".json")
	err = os.MkdirAll("output", 0755)
	if err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	err = sf.Dump(outputPath)
	if err != nil {
		return fmt.Errorf("保存草稿文件失败: %v", err)
	}

	fmt.Printf("剪映草稿文件已生成: %s\n", outputPath)

	// 查找剪映草稿文件夹并复制项目
	jianyingDir, err := findJianyingDraftFolder()
	if err != nil {
		return fmt.Errorf("查找剪映草稿文件夹失败: %v", err)
	}

	// 创建新项目文件夹
	newProjectDir := filepath.Join(jianyingDir, projectID)
	err = os.MkdirAll(newProjectDir, 0755)
	if err != nil {
		return fmt.Errorf("创建项目文件夹失败: %v", err)
	}

	// 复制必要的项目文件到剪映项目目录
	err = copyProjectFiles(outputPath, newProjectDir, inputDir)
	if err != nil {
			return fmt.Errorf("复制项目文件失败: %v", err)
	}

	fmt.Printf("项目已复制到剪映目录: %s\n", newProjectDir)
	fmt.Println("请在剪映中打开该项目进行最终调整和导出")

	return nil
}

// GenerateProjectWithOutputDir 生成剪映项目，支持指定输出目录
func (cg *CapcutGenerator) GenerateProjectWithOutputDir(inputDir, outputDir string) error {
	// 获取输入目录的绝对路径
	inputDir, err := filepath.Abs(inputDir)
	if err != nil {
		return fmt.Errorf("获取输入目录绝对路径失败: %v", err)
	}

	// 检查必要的文件
	audioFile := ""
	imageFiles := []string{}
	srtFile := ""

	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("读取输入目录失败: %v", err)
	}

	for _, file := range files {
		filename := strings.ToLower(file.Name())
		if strings.HasSuffix(filename, ".wav") || strings.HasSuffix(filename, ".mp3") {
			audioFile = cleanPath(filepath.Join(inputDir, file.Name())) // 清理音频文件路径
		} else if strings.HasSuffix(filename, ".png") || strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".jpeg") {
			imageFiles = append(imageFiles, cleanPath(filepath.Join(inputDir, file.Name()))) // 清理图片文件路径
		} else if strings.HasSuffix(filename, ".srt") {
			srtFile = cleanPath(filepath.Join(inputDir, file.Name())) // 清理字幕文件路径
		}
	}

	if audioFile == "" {
		return fmt.Errorf("未找到音频文件")
	}

	if len(imageFiles) == 0 {
		return fmt.Errorf("未找到图片文件")
	}

	// 获取音频文件实际时长
	audioDuration, err := getAudioDuration(audioFile)
	if err != nil {
		return fmt.Errorf("获取音频时长失败: %v", err)
	}

	// 计算音频时长（秒）
	audioDurationSec := float64(audioDuration) / 1000000.0

	// 计算台词总字数
	totalSubtitleChars := 0
	if srtFile != "" {
		srtEntries, err := srt.ParseSrtFile(srtFile)
		if err != nil {
			fmt.Printf("解析字幕文件失败: %v\n", err)
		} else {
			for _, entry := range srtEntries {
				totalSubtitleChars += len([]rune(entry.Text))
			}
		}
	}

	// 输出日志信息
	fmt.Printf("🎵 音频时长: %.2f秒 (%d微秒)\n", audioDurationSec, audioDuration)
	fmt.Printf("🖼️  图片资源: %d张\n", len(imageFiles))
	fmt.Printf("📄 视频资源: 1个 (音频文件: %s)\n", filepath.Base(audioFile))
	fmt.Printf("💬 台词字数: %d个字符\n", totalSubtitleChars)

	// 创建草稿文件 (1080x1920 竖屏视频)
	sf, err := script.NewScriptFile(1080, 1920, 30) // 宽度、高度、帧率
	if err != nil {
		return fmt.Errorf("创建草稿文件失败: %v", err)
	}

	// 设置草稿的基本信息
	sf.Duration = audioDuration

	// 计算每个图片的显示时间（平均分配音频总时长）
	numScenes := len(imageFiles)
	sceneDuration := audioDuration / int64(numScenes)

	// 添加图片素材到草稿
	for i, imageFile := range imageFiles {
		relPath := imageFile // 使用原始路径，NewVideoMaterial会自动转换为绝对路径
		imageName := filepath.Base(imageFile)
		videoMaterial, err := material.NewVideoMaterial(
			material.MaterialTypePhoto, // 静态图片
			&relPath,                   // 文件路径 (NewVideoMaterial会自动转换为绝对路径)
			nil,                        // 替换路径 (不需要，使用原始路径)
			&imageName,                 // 素材名称
			nil,                        // 远程URL
			nil,                        // 裁剪设置
			nil,                        // 时长
			nil,                        // 宽度
			nil,                        // 高度
		)
		if err != nil {
			fmt.Printf("创建视频素材失败: %v\n", err)
			continue
		}
		sf.AddMaterial(videoMaterial)

		// 添加到视频轨道
		videoTrack, err := sf.GetTrack("video", nil)
		if err != nil {
			videoTrackName := stringPtr(fmt.Sprintf("视频轨道_%d", i))
			sf.AddTrack(track.TrackTypeVideo, videoTrackName)
			videoTrack, _ = sf.GetTrack("video", videoTrackName)
		}

		startTime := int64(i) * sceneDuration
		endTime := startTime + sceneDuration

		// 确保最后一张图片精确结束于音频末尾
		if i == numScenes-1 {
			endTime = audioDuration
		}

		sourceTimeRange := types.NewTimerange(startTime, endTime-startTime)
		targetTimeRange := types.NewTimerange(startTime, endTime-startTime)

		videoSegment := segment.NewVideoSegment(
			videoMaterial.MaterialID, // materialID
			sourceTimeRange,          // sourceTimerange
			targetTimeRange,          // targetTimerange
			1.0,                      // speed
			1.0,                      // volume
			nil,                      // clipSettings
		)

		videoTrack.AddSegment(videoSegment)
	}

	// 添加音频素材到草稿
	audioFileName := filepath.Base(audioFile)
	audioMaterial, err := material.NewAudioMaterial(
		&audioFile,                             // 文件路径 (NewAudioMaterial会自动转换为绝对路径)
		nil,                                    // 替换路径 (不需要，使用原始路径)
		&audioFileName,                         // 素材名称
		nil,                                    // 远程URL
		float64Ptr(float64(audioDuration)/1e6), // 时长（秒）
	)
	if err != nil {
		return fmt.Errorf("创建音频素材失败: %v", err)
	} else {
		sf.AddMaterial(audioMaterial)

		// 添加到音频轨道
		audioTrackName := stringPtr("音频轨道")
		sf.AddTrack(track.TrackTypeAudio, audioTrackName)

		// 获取刚刚添加的音频轨道
		audioTrack, err := sf.GetTrack("audio", audioTrackName)
		if err != nil {
			return fmt.Errorf("获取音频轨道失败: %v", err)
		}

		audioSegment := segment.NewAudioSegment(
			audioMaterial.MaterialID,             // materialID
			types.NewTimerange(0, audioDuration), // targetTimerange - 整个音频时长
			nil,                                  // sourceTimerange
			1.0,                                  // speed
			1.0,                                  // volume
		)

		err = audioTrack.AddSegment(audioSegment)
		if err != nil {
			return fmt.Errorf("向音频轨道添加片段失败: %v", err)
		}
	}

	// 如果有SRT字幕文件，则添加字幕
	if srtFile != "" {
		srtEntries, err := srt.ParseSrtFile(srtFile)
		if err != nil {
			return fmt.Errorf("解析字幕文件失败: %v", err)
		} else {
			// 重新计算字幕时间戳，使其与音频总时长相匹配
			// 首先获取原始字幕总时长
			var originalSubtitleDuration int64
			if len(srtEntries) > 0 {
				lastEntry := srtEntries[len(srtEntries)-1]
				originalSubtitleDuration = lastEntry.End
			}

			// 添加文本轨道和字幕
			textTrackName := stringPtr("字幕轨道")
			sf.AddTrack(track.TrackTypeText, textTrackName)

			// 获取文本轨道并添加字幕片段
			textTrack, err := sf.GetTrack("text", textTrackName)
			if err != nil {
				return fmt.Errorf("获取文本轨道失败: %v", err)
			} else {
				for _, entry := range srtEntries {
					// 根据音频时长与原始字幕时长的比例调整字幕时间
					var adjustedStart, adjustedEnd int64
					if originalSubtitleDuration > 0 {
						// 按比例调整时间戳
						ratio := float64(audioDuration) / float64(originalSubtitleDuration)
						adjustedStart = int64(float64(entry.Start) * ratio)
						adjustedEnd = int64(float64(entry.End) * ratio)
						
						// 确保最后一个字幕精确结束于音频末尾
						if entry.End == originalSubtitleDuration && entry.End > 0 {
							adjustedEnd = audioDuration
						}
					} else {
						// 如果无法计算比例，直接使用原始时间
						adjustedStart = entry.Start
						adjustedEnd = entry.End
					}

					// 创建文本样式
					textStyle := segment.NewTextStyle()
					textStyle.Size = 24.0
					textStyle.Color = [3]float64{1.0, 1.0, 1.0} // 白色
					textStyle.Bold = true
					textStyle.Align = 1 // 居中对齐

					// 创建ClipSettings来设置字幕位置，使其显示在画面下方
					clipSettings := segment.NewClipSettingsWithParams(
						1.0,   // alpha
						0.0,   // rotation
						1.0,   // scaleX
						1.0,   // scaleY
						0.0,   // transformX
						-0.8,  // transformY - 负值使字幕靠下显示
						false, // flipH
						false, // flipV
					)

					// 创建文本素材并添加到素材库
					textMaterial := map[string]interface{}{
						"add_type":                     2,
						"alignment":                    1,
						"background_alpha":             1.0,
						"background_color":             "",
						"background_height":            1.0,
						"background_horizontal_offset": 0.0,
						"background_round_radius":      0.0,
						"background_vertical_offset":   0.0,
						"background_width":             1.0,
						"bold_width":                   0.0,
						"border_color":                 "",
						"border_width":                 0.08,
						"check_flag":                   7,
						"content":                      fmt.Sprintf("<font id=\"%s\" path=\"/Applications/VideoFusion-macOS.app/Contents/Resources/Font/SystemFont/zh-hans.ttf\"><color=(1.000000, 1.000000, 1.000000, 1.000000)><size=5.000000>%s</size></color></font>", uuid.New().String(), strings.ReplaceAll(entry.Text, "\n", "\u0001")),
						"font_category_id":             "",
						"font_category_name":           "",
						"font_id":                      "",
						"font_name":                    "",
						"font_path":                    "/Applications/VideoFusion-macOS.app/Contents/Resources/Font/SystemFont/zh-hans.ttf",
						"font_resource_id":             "",
						"font_size":                    5.0,
						"font_title":                   "none",
						"font_url":                     "",
						"fonts":                        []interface{}{},
						"global_alpha":                 1.0,
						"has_shadow":                   false,
						"id":                           uuid.New().String(), // 生成唯一ID
						"initial_scale":                1.0,
						"is_rich_text":                 false,
						"italic_degree":                0,
						"ktv_color":                    "",
						"layer_weight":                 1,
						"letter_spacing":               0.0,
						"line_spacing":                 0.02,
						"recognize_type":               0,
						"shadow_alpha":                 0.8,
						"shadow_angle":                 -45.0,
						"shadow_color":                 "",
						"shadow_distance":              8.0,
						"shadow_point":                 map[string]interface{}{"x": 1.0182337649086284, "y": -1.0182337649086284},
						"shadow_smoothing":             1.0,
						"shape_clip_x":                 false,
						"shape_clip_y":                 false,
						"style_name":                   "",
						"sub_type":                     0,
						"text_alpha":                   1.0,
						"text_color":                   "#FFFFFF",
						"text_size":                    30,
						"text_to_audio_ids":            []interface{}{},
						"type":                         "subtitle",
						"typesetting":                  0,
						"underline":                    false,
						"underline_offset":             0.22,
						"underline_width":              0.05,
						"use_effect_default_color":     true,
					}
					// 将文本素材添加到素材库
					sf.Materials.Texts = append(sf.Materials.Texts, textMaterial)

					// 创建文本片段，使用刚添加的文本素材ID
					textSegment := segment.NewTextSegment(
						entry.Text, // text
						types.NewTimerange(adjustedStart, adjustedEnd-adjustedStart), // targetTimerange - 调整后的时间
						"",           // font (空字符串使用默认字体)
						textStyle,    // style
						clipSettings, // clipSettings - 添加位置设置
					)
					// 设置正确的MaterialID（使用刚添加的文本素材ID）
					textSegment.MaterialID = textMaterial["id"].(string)

					textTrack.AddSegment(textSegment)
				}
			}
		}
	}

	// 生成项目ID
	projectID := uuid.New().String()

	// 将草稿内容写入指定输出目录
	outputPath := filepath.Join(outputDir, projectID+".json")
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	err = sf.Dump(outputPath)
	if err != nil {
		return fmt.Errorf("保存草稿文件失败: %v", err)
	}

	fmt.Printf("剪映草稿文件已生成: %s\n", outputPath)

	return nil
}

// GenerateAndImportProject 生成剪映项目并导入到剪映
func (cg *CapcutGenerator) GenerateAndImportProject(inputDir, projectName string) error {
	// 获取输入目录的绝对路径
	inputDir, err := filepath.Abs(inputDir)
	if err != nil {
		return fmt.Errorf("获取输入目录绝对路径失败: %v", err)
	}

	// 检查必要的文件
	audioFile := ""
	imageFiles := []string{}
	srtFile := ""

	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("读取输入目录失败: %v", err)
	}

	for _, file := range files {
		filename := strings.ToLower(file.Name())
		if strings.HasSuffix(filename, ".wav") || strings.HasSuffix(filename, ".mp3") {
			audioFile = cleanPath(filepath.Join(inputDir, file.Name())) // 清理音频文件路径
		} else if strings.HasSuffix(filename, ".png") || strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".jpeg") {
			imageFiles = append(imageFiles, cleanPath(filepath.Join(inputDir, file.Name()))) // 清理图片文件路径
		} else if strings.HasSuffix(filename, ".srt") {
			srtFile = cleanPath(filepath.Join(inputDir, file.Name())) // 清理字幕文件路径
		}
	}

	if audioFile == "" {
		return fmt.Errorf("未找到音频文件")
	}

	if len(imageFiles) == 0 {
		return fmt.Errorf("未找到图片文件")
	}

	// 获取音频文件实际时长
	audioDuration, err := getAudioDuration(audioFile)
	if err != nil {
		return fmt.Errorf("获取音频时长失败: %v", err)
	}

	// 创建草稿文件 (1080x1920 竖屏视频)
	sf, err := script.NewScriptFile(1080, 1920, 30) // 宽度、高度、帧率
	if err != nil {
		return fmt.Errorf("创建草稿文件失败: %v", err)
	}

	// 设置草稿的基本信息
	sf.Duration = audioDuration

	// 计算每个图片的显示时间（平均分配音频总时长）
	numScenes := len(imageFiles)
	sceneDuration := audioDuration / int64(numScenes)

	// 添加图片素材到草稿
	for i, imageFile := range imageFiles {
		relPath := imageFile // 使用原始路径，NewVideoMaterial会自动转换为绝对路径
		imageName := filepath.Base(imageFile)
		videoMaterial, err := material.NewVideoMaterial(
			material.MaterialTypePhoto, // 静态图片
			&relPath,                   // 文件路径 (NewVideoMaterial会自动转换为绝对路径)
			nil,                        // 替换路径 (不需要，使用原始路径)
			&imageName,                 // 素材名称
			nil,                        // 远程URL
			nil,                        // 裁剪设置
			nil,                        // 时长
			nil,                        // 宽度
			nil,                        // 高度
		)
		if err != nil {
			fmt.Printf("创建视频素材失败: %v\n", err)
			continue
		}
		sf.AddMaterial(videoMaterial)

		// 添加到视频轨道
		videoTrack, err := sf.GetTrack("video", nil)
		if err != nil {
			videoTrackName := stringPtr(fmt.Sprintf("视频轨道_%d", i))
			sf.AddTrack(track.TrackTypeVideo, videoTrackName)
			videoTrack, _ = sf.GetTrack("video", videoTrackName)
		}

		startTime := int64(i) * sceneDuration
		endTime := startTime + sceneDuration

		// 确保最后一张图片精确结束于音频末尾
		if i == numScenes-1 {
			endTime = audioDuration
		}

		sourceTimeRange := types.NewTimerange(startTime, endTime-startTime)
		targetTimeRange := types.NewTimerange(startTime, endTime-startTime)

		videoSegment := segment.NewVideoSegment(
			videoMaterial.MaterialID, // materialID
			sourceTimeRange,          // sourceTimerange
			targetTimeRange,          // targetTimerange
			1.0,                      // speed
			1.0,                      // volume
			nil,                      // clipSettings
		)

		videoTrack.AddSegment(videoSegment)
	}

	// 添加音频素材到草稿
	audioFileName := filepath.Base(audioFile)
	audioMaterial, err := material.NewAudioMaterial(
		&audioFile,                             // 文件路径 (NewAudioMaterial会自动转换为绝对路径)
		nil,                                    // 替换路径 (不需要，使用原始路径)
		&audioFileName,                         // 素材名称
		nil,                                    // 远程URL
		float64Ptr(float64(audioDuration)/1e6), // 时长（秒）
	)
	if err != nil {
		return fmt.Errorf("创建音频素材失败: %v", err)
	} else {
		sf.AddMaterial(audioMaterial)

		// 添加到音频轨道
		audioTrackName := stringPtr("音频轨道")
		sf.AddTrack(track.TrackTypeAudio, audioTrackName)

		// 获取刚刚添加的音频轨道
		audioTrack, err := sf.GetTrack("audio", audioTrackName)
		if err != nil {
			return fmt.Errorf("获取音频轨道失败: %v", err)
		}

		audioSegment := segment.NewAudioSegment(
			audioMaterial.MaterialID,             // materialID
			types.NewTimerange(0, audioDuration), // targetTimerange - 整个音频时长
			nil,                                  // sourceTimerange
			1.0,                                  // speed
			1.0,                                  // volume
		)

		err = audioTrack.AddSegment(audioSegment)
		if err != nil {
			return fmt.Errorf("向音频轨道添加片段失败: %v", err)
		}
	}

	// 如果有SRT字幕文件，则添加字幕
	if srtFile != "" {
		srtEntries, err := srt.ParseSrtFile(srtFile)
		if err != nil {
			return fmt.Errorf("解析字幕文件失败: %v", err)
		} else {
			// 重新计算字幕时间戳，使其与音频总时长相匹配
			// 首先获取原始字幕总时长
			var originalSubtitleDuration int64
			if len(srtEntries) > 0 {
				lastEntry := srtEntries[len(srtEntries)-1]
				originalSubtitleDuration = lastEntry.End
			}

			// 添加文本轨道和字幕
			textTrackName := stringPtr("字幕轨道")
			sf.AddTrack(track.TrackTypeText, textTrackName)

			// 获取文本轨道并添加字幕片段
			textTrack, err := sf.GetTrack("text", textTrackName)
			if err != nil {
				return fmt.Errorf("获取文本轨道失败: %v", err)
			} else {
				for _, entry := range srtEntries {
					// 根据音频时长与原始字幕时长的比例调整字幕时间
					var adjustedStart, adjustedEnd int64
					if originalSubtitleDuration > 0 {
						// 按比例调整时间戳
						ratio := float64(audioDuration) / float64(originalSubtitleDuration)
						adjustedStart = int64(float64(entry.Start) * ratio)
						adjustedEnd = int64(float64(entry.End) * ratio)
						
						// 确保最后一个字幕精确结束于音频末尾
						if entry.End == originalSubtitleDuration && entry.End > 0 {
							adjustedEnd = audioDuration
						}
					} else {
						// 如果无法计算比例，直接使用原始时间
						adjustedStart = entry.Start
						adjustedEnd = entry.End
					}

					// 创建文本样式
					textStyle := segment.NewTextStyle()
					textStyle.Size = 24.0
					textStyle.Color = [3]float64{1.0, 1.0, 1.0} // 白色
					textStyle.Bold = true
					textStyle.Align = 1 // 居中对齐

					// 创建ClipSettings来设置字幕位置，使其显示在画面下方
					clipSettings := segment.NewClipSettingsWithParams(
						1.0,   // alpha
						0.0,   // rotation
						1.0,   // scaleX
						1.0,   // scaleY
						0.0,   // transformX
						-0.8,  // transformY - 负值使字幕靠下显示
						false, // flipH
						false, // flipV
					)

					// 创建文本素材并添加到素材库
					textMaterial := map[string]interface{}{
						"add_type":                     2,
						"alignment":                    1,
						"background_alpha":             1.0,
						"background_color":             "",
						"background_height":            1.0,
						"background_horizontal_offset": 0.0,
						"background_round_radius":      0.0,
						"background_vertical_offset":   0.0,
						"background_width":             1.0,
						"bold_width":                   0.0,
						"border_color":                 "",
						"border_width":                 0.08,
						"check_flag":                   7,
						"content":                      fmt.Sprintf("<font id=\"%s\" path=\"/Applications/VideoFusion-macOS.app/Contents/Resources/Font/SystemFont/zh-hans.ttf\"><color=(1.000000, 1.000000, 1.000000, 1.000000)><size=5.000000>%s</size></color></font>", uuid.New().String(), strings.ReplaceAll(entry.Text, "\n", "\u0001")),
						"font_category_id":             "",
						"font_category_name":           "",
						"font_id":                      "",
						"font_name":                    "",
						"font_path":                    "/Applications/VideoFusion-macOS.app/Contents/Resources/Font/SystemFont/zh-hans.ttf",
						"font_resource_id":             "",
						"font_size":                    5.0,
						"font_title":                   "none",
						"font_url":                     "",
						"fonts":                        []interface{}{},
						"global_alpha":                 1.0,
						"has_shadow":                   false,
						"id":                           uuid.New().String(), // 生成唯一ID
						"initial_scale":                1.0,
						"is_rich_text":                 false,
						"italic_degree":                0,
						"ktv_color":                    "",
						"layer_weight":                 1,
						"letter_spacing":               0.0,
						"line_spacing":                 0.02,
						"recognize_type":               0,
						"shadow_alpha":                 0.8,
						"shadow_angle":                 -45.0,
						"shadow_color":                 "",
						"shadow_distance":              8.0,
						"shadow_point":                 map[string]interface{}{"x": 1.0182337649086284, "y": -1.0182337649086284},
						"shadow_smoothing":             1.0,
						"shape_clip_x":                 false,
						"shape_clip_y":                 false,
						"style_name":                   "",
						"sub_type":                     0,
						"text_alpha":                   1.0,
						"text_color":                   "#FFFFFF",
						"text_size":                    30,
						"text_to_audio_ids":            []interface{}{},
						"type":                         "subtitle",
						"typesetting":                  0,
						"underline":                    false,
						"underline_offset":             0.22,
						"underline_width":              0.05,
						"use_effect_default_color":     true,
					}
					// 将文本素材添加到素材库
					sf.Materials.Texts = append(sf.Materials.Texts, textMaterial)

					// 创建文本片段，使用刚添加的文本素材ID
					textSegment := segment.NewTextSegment(
						entry.Text, // text
						types.NewTimerange(adjustedStart, adjustedEnd-adjustedStart), // targetTimerange - 调整后的时间
						"",           // font (空字符串使用默认字体)
						textStyle,    // style
						clipSettings, // clipSettings - 添加位置设置
					)
					// 设置正确的MaterialID（使用刚添加的文本素材ID）
					textSegment.MaterialID = textMaterial["id"].(string)

					textTrack.AddSegment(textSegment)
				}
			}
		}
	}

	// 生成项目ID - 使用传入的项目名
	projectID := projectName

	// 将草稿内容写入临时文件
	outputPath := filepath.Join("output", projectID+".json")
	err = os.MkdirAll("output", 0755)
	if err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	err = sf.Dump(outputPath)
	if err != nil {
		return fmt.Errorf("保存草稿文件失败: %v", err)
	}

	fmt.Printf("剪映草稿文件已生成: %s\n", outputPath)

	// 查找剪映草稿文件夹并复制项目
	jianyingDir, err := findJianyingDraftFolder()
	if err != nil {
		return fmt.Errorf("查找剪映草稿文件夹失败: %v", err)
	}

	// 创建新项目文件夹
	newProjectDir := filepath.Join(jianyingDir, projectID)
	err = os.MkdirAll(newProjectDir, 0755)
	if err != nil {
		return fmt.Errorf("创建项目文件夹失败: %v", err)
	}

	// 复制必要的项目文件到剪映项目目录
	err = copyProjectFiles(outputPath, newProjectDir, inputDir)
	if err != nil {
		return fmt.Errorf("复制项目文件失败: %v", err)
	}

	fmt.Printf("项目已复制到剪映目录: %s\n", newProjectDir)
	fmt.Println("请在剪映中打开该项目进行最终调整和导出")

	return nil
}

// copyProjectFiles 复制项目文件到剪映目录
func copyProjectFiles(sourceDraftPath, targetProjectDir, inputDir string) error {
	// 读取源草稿文件
	content, err := ioutil.ReadFile(sourceDraftPath)
	if err != nil {
		return err
	}

	// 将内容写入目标目录的 draft_info.json
	draftInfoPath := filepath.Join(targetProjectDir, "draft_info.json")
	err = ioutil.WriteFile(draftInfoPath, content, 0644)
	if err != nil {
		return err
	}

	// 复制原始媒体文件到项目目录，并收集文件信息用于更新路径
	mediaFiles := make(map[string]string) // 原始路径 -> 目标路径映射
	files, _ := ioutil.ReadDir(inputDir)
	for _, file := range files {
		ext := filepath.Ext(strings.ToLower(file.Name()))
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".wav" || ext == ".mp3" || ext == ".srt" {
			srcPath := filepath.Join(inputDir, file.Name())
			dstPath := filepath.Join(targetProjectDir, file.Name())

			// 复制文件
			err = copyFile(srcPath, dstPath)
			if err != nil {
				fmt.Printf("复制媒体文件失败 %s: %v\n", srcPath, err)
				// 继续处理其他文件
			} else {
				mediaFiles[cleanPath(srcPath)] = cleanPath(dstPath) // 清理路径
			}
		}
	}

	// 读取刚刚写入的draft_info.json，更新其中的素材路径
	updatedContent, err := ioutil.ReadFile(draftInfoPath)
	if err != nil {
		return err
	}

	var draftData map[string]interface{}
	err = json.Unmarshal(updatedContent, &draftData)
	if err != nil {
		return err
	}

	// 更新素材路径
	updateMediaPaths(draftData, mediaFiles)

	// 写回更新后的draft_info.json
	updatedJSON, err := json.MarshalIndent(draftData, "", "    ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(draftInfoPath, updatedJSON, 0644)
	if err != nil {
		return err
	}

	// 创建 draft_agency_config.json
	agencyConfig := createAgencyConfig(targetProjectDir) // 使用目标项目目录而不是输入目录
	agencyConfigPath := filepath.Join(targetProjectDir, "draft_agency_config.json")
	agencyConfigContent, err := json.Marshal(agencyConfig)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(agencyConfigPath, agencyConfigContent, 0644)
	if err != nil {
		return err
	}

	// 创建 draft_virtual_store.json
	virtualStore := createVirtualStore(draftInfoPath)
	virtualStorePath := filepath.Join(targetProjectDir, "draft_virtual_store.json")
	virtualStoreContent, err := json.Marshal(virtualStore)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(virtualStorePath, virtualStoreContent, 0644)
	if err != nil {
		return err
	}

	// 创建 draft_meta_info.json
	metaInfo := createMetaInfo()
	metaInfoPath := filepath.Join(targetProjectDir, "draft_meta_info.json")
	metaInfoContent, err := json.Marshal(metaInfo)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(metaInfoPath, metaInfoContent, 0644)
	if err != nil {
		return err
	}

	// 创建 template.tmp 文件
	templatePath := filepath.Join(targetProjectDir, "template.tmp")
	templateContent := "{}"
	err = ioutil.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		return err
	}

	return nil
}

// updateMediaPaths 更新素材路径
func updateMediaPaths(data map[string]interface{}, mediaFiles map[string]string) {
	if materials, ok := data["materials"].(map[string]interface{}); ok {
		// 更新视频素材路径
		if videos, ok := materials["videos"].([]interface{}); ok {
			for _, video := range videos {
				if videoObj, ok := video.(map[string]interface{}); ok {
					if path, ok := videoObj["path"].(string); ok {
						if newPath, exists := mediaFiles[path]; exists {
							videoObj["path"] = newPath
						}
					}
				}
			}
		}

		// 更新音频素材路径
		if audios, ok := materials["audios"].([]interface{}); ok {
			for _, audio := range audios {
				if audioObj, ok := audio.(map[string]interface{}); ok {
					if path, ok := audioObj["path"].(string); ok {
						if newPath, exists := mediaFiles[path]; exists {
							audioObj["path"] = newPath
						}
					}
				}
			}
		}
	}
}

// AgencyConfig 剪映代理配置
type AgencyConfig struct {
	Materials       []map[string]interface{} `json:"marterials"`
	UseConverter    bool                     `json:"use_converter"`
	VideoResolution int                      `json:"video_resolution"`
}

// createAgencyConfig 创建代理配置
func createAgencyConfig(inputDir string) *AgencyConfig {
	config := &AgencyConfig{
		Materials:       []map[string]interface{}{},
		UseConverter:    false,
		VideoResolution: 720,
	}

	// 获取输入目录中的所有媒体文件
	files, _ := ioutil.ReadDir(inputDir)
	for _, file := range files {
		filename := strings.ToLower(cleanPath(file.Name())) // 清理文件名中的特殊字符
		ext := filepath.Ext(filename)
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".wav" || ext == ".mp3" {
			absPath := cleanPath(filepath.Join(inputDir, file.Name())) // 使用cleanPath清理路径
			material := map[string]interface{}{
				"source_path":   absPath,
				"use_converter": true,
			}
			config.Materials = append(config.Materials, material)
		}
	}

	return config
}

// VirtualStore 虚拟存储配置
type VirtualStore struct {
	DraftMaterials []string      `json:"draft_materials"`
	VirtualStore   []interface{} `json:"draft_virtual_store"`
}

// createVirtualStore 创建虚拟存储配置
func createVirtualStore(draftInfoPath string) *VirtualStore {
	// 读取draft_info.json获取素材ID
	content, err := ioutil.ReadFile(draftInfoPath)
	if err != nil {
		// 如果无法读取，返回空配置
		return &VirtualStore{
			DraftMaterials: []string{},
			VirtualStore: []interface{}{
				map[string]interface{}{
					"type": 0,
					"value": []interface{}{
						map[string]interface{}{
							"creation_time": 0,
							"display_name":  "",
							"filter_type":   0,
							"id":            "",
							"import_time":   0,
							"sort_sub_type": 0,
							"sort_type":     0,
						},
					},
				},
				map[string]interface{}{
					"type":  1,
					"value": []interface{}{},
				},
			},
		}
	}

	var data map[string]interface{}
	err = json.Unmarshal(content, &data)
	if err != nil {
		return &VirtualStore{
			DraftMaterials: []string{},
			VirtualStore: []interface{}{
				map[string]interface{}{
					"type": 0,
					"value": []interface{}{
						map[string]interface{}{
							"creation_time": 0,
							"display_name":  "",
							"filter_type":   0,
							"id":            "",
							"import_time":   0,
							"sort_sub_type": 0,
							"sort_type":     0,
						},
					},
				},
				map[string]interface{}{
					"type":  1,
					"value": []interface{}{},
				},
			},
		}
	}

	materialsInterface, ok := data["materials"]
	if !ok {
		return &VirtualStore{
			DraftMaterials: []string{},
			VirtualStore: []interface{}{
				map[string]interface{}{
					"type": 0,
					"value": []interface{}{
						map[string]interface{}{
							"creation_time": 0,
							"display_name":  "",
							"filter_type":   0,
							"id":            "",
							"import_time":   0,
							"sort_sub_type": 0,
							"sort_type":     0,
						},
					},
				},
				map[string]interface{}{
					"type":  1,
					"value": []interface{}{},
				},
			},
		}
	}

	materialsMap, ok := materialsInterface.(map[string]interface{})
	if !ok {
		return &VirtualStore{
			DraftMaterials: []string{},
			VirtualStore: []interface{}{
				map[string]interface{}{
					"type": 0,
					"value": []interface{}{
						map[string]interface{}{
							"creation_time": 0,
							"display_name":  "",
							"filter_type":   0,
							"id":            "",
							"import_time":   0,
							"sort_sub_type": 0,
							"sort_type":     0,
						},
					},
				},
				map[string]interface{}{
					"type":  1,
					"value": []interface{}{},
				},
			},
		}
	}

	// 提取各种素材的ID
	var draftMaterials []string

	// 添加视频素材ID
	if videosInterface, ok := materialsMap["videos"]; ok {
		if videos, ok := videosInterface.([]interface{}); ok {
			for _, videoInterface := range videos {
				if video, ok := videoInterface.(map[string]interface{}); ok {
					if id, ok := video["id"]; ok {
						if idStr, ok := id.(string); ok {
							draftMaterials = append(draftMaterials, idStr)
						}
					}
				}
			}
		}
	}

	// 添加音频素材ID
	if audiosInterface, ok := materialsMap["audios"]; ok {
		if audios, ok := audiosInterface.([]interface{}); ok {
			for _, audioInterface := range audios {
				if audio, ok := audioInterface.(map[string]interface{}); ok {
					if id, ok := audio["id"]; ok {
						if idStr, ok := id.(string); ok {
							draftMaterials = append(draftMaterials, idStr)
						}
					}
				}
			}
		}
	}

	// 添加文本素材ID
	if textsInterface, ok := materialsMap["texts"]; ok {
		if texts, ok := textsInterface.([]interface{}); ok {
			for _, textInterface := range texts {
				if text, ok := textInterface.(map[string]interface{}); ok {
					if id, ok := text["id"]; ok {
						if idStr, ok := id.(string); ok {
							draftMaterials = append(draftMaterials, idStr)
						}
					}
				}
			}
		}
	}

	// 构建虚拟存储的值数组
	var valueItems []interface{}

	// 添加基础项
	valueItems = append(valueItems, map[string]interface{}{
		"creation_time": 0,
		"display_name":  "",
		"filter_type":   0,
		"id":            "",
		"import_time":   0,
		"sort_sub_type": 0,
		"sort_type":     0,
	})

	// 为每个素材添加虚拟存储项
	for _, id := range draftMaterials {
		item := map[string]interface{}{
			"creation_time": 0,
			"display_name":  "素材",
			"filter_type":   0,
			"id":            id,
			"import_time":   0,
			"sort_sub_type": 0,
			"sort_type":     0,
		}
		valueItems = append(valueItems, item)
	}

	return &VirtualStore{
		DraftMaterials: draftMaterials,
		VirtualStore: []interface{}{
			map[string]interface{}{
				"type":  0,
				"value": valueItems,
			},
			map[string]interface{}{
				"type":  1,
				"value": []interface{}{},
			},
		},
	}
}

// MetaInfo 元信息
type MetaInfo struct {
	Version string `json:"version"`
	Name    string `json:"name"`
}

// createMetaInfo 创建元信息
func createMetaInfo() *MetaInfo {
	return &MetaInfo{
		Version: "1.0",
		Name:    "Generated Project",
	}
}

// 辅助函数：字符串指针
func stringPtr(s string) *string {
	return &s
}

// 辅助函数：浮点数指针
func float64Ptr(f float64) *float64 {
	return &f
}

// copyFile 复制文件的辅助函数
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}