/*视频组装*/
package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// VideoProcessor 视频处理器
type VideoProcessor struct {
	logger *zap.Logger
}

// VideoProject 视频项目
type VideoProject struct {
	Name       string          `json:"name"`
	OutputDir  string          `json:"output_dir"`
	Assets     VideoAssets     `json:"assets"`
	Timeline   []VideoClip     `json:"timeline"`
	Duration   float64         `json:"duration"`
	Resolution VideoResolution `json:"resolution"`
}

// VideoAssets 视频资产
type VideoAssets struct {
	AudioFiles    []string `json:"audio_files"`
	SubtitleFiles []string `json:"subtitle_files"`
	ImageFiles    []string `json:"image_files"`
	VideoFiles    []string `json:"video_files"`
}

// VideoClip 视频剪辑片段
type VideoClip struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"` // audio, image, video, subtitle
	File      string   `json:"file"`
	StartTime float64  `json:"start_time"`
	Duration  float64  `json:"duration"`
	Position  Point    `json:"position,omitempty"`
	Scale     float64  `json:"scale,omitempty"`
	Effects   []Effect `json:"effects,omitempty"`
}

// VideoResolution 视频分辨率
type VideoResolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Point 坐标点
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Effect 特效
type Effect struct {
	Type     string                 `json:"type"`
	Duration float64                `json:"duration"`
	Params   map[string]interface{} `json:"params"`
}

// JianyingProject 剪映项目结构
type JianyingProject struct {
	AppVersion     string                 `json:"app_version"`
	ProjectInfo    JianyingProjectInfo    `json:"project_info"`
	Materials      JianyingMaterials      `json:"materials"`
	Tracks         JianyingTracks         `json:"tracks"`
	ExtraMaterials JianyingExtraMaterials `json:"extra_materials"`
}

// JianyingProjectInfo 剪映项目信息
type JianyingProjectInfo struct {
	Name       string  `json:"name"`
	CreateTime int64   `json:"create_time"`
	UpdateTime int64   `json:"update_time"`
	Duration   float64 `json:"duration"`
	Resolution string  `json:"resolution"`
	FrameRate  int     `json:"frame_rate"`
	SampleRate int     `json:"sample_rate"`
	Volume     float64 `json:"volume"`
}

// JianyingMaterials 剪映素材
type JianyingMaterials struct {
	Audios []JianyingAudio `json:"audios"`
	Videos []JianyingVideo `json:"videos"`
	Images []JianyingImage `json:"images"`
	Texts  []JianyingText  `json:"texts"`
}

// JianyingAudio 剪映音频素材
type JianyingAudio struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Path     string  `json:"path"`
	Duration float64 `json:"duration"`
	Volume   float64 `json:"volume"`
	FadeIn   float64 `json:"fade_in,omitempty"`
	FadeOut  float64 `json:"fade_out,omitempty"`
}

// JianyingVideo 剪映视频素材
type JianyingVideo struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Path     string  `json:"path"`
	Duration float64 `json:"duration"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
}

// JianyingImage 剪映图片素材
type JianyingImage struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Path     string  `json:"path"`
	Duration float64 `json:"duration"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
}

// JianyingText 剪映文字素材
type JianyingText struct {
	ID       string  `json:"id"`
	Content  string  `json:"content"`
	Duration float64 `json:"duration"`
	Style    string  `json:"style"`
	FontSize int     `json:"font_size"`
	Color    string  `json:"color"`
}

// JianyingTracks 剪映轨道
type JianyingTracks struct {
	VideoTracks []JianyingVideoTrack `json:"video_tracks"`
	AudioTracks []JianyingAudioTrack `json:"audio_tracks"`
	TextTracks  []JianyingTextTrack  `json:"text_tracks"`
}

// JianyingVideoTrack 剪映视频轨道
type JianyingVideoTrack struct {
	ID      string              `json:"id"`
	Clips   []JianyingVideoClip `json:"clips"`
	Effects []JianyingEffect    `json:"effects,omitempty"`
}

// JianyingAudioTrack 剪映音频轨道
type JianyingAudioTrack struct {
	ID    string              `json:"id"`
	Clips []JianyingAudioClip `json:"clips"`
}

// JianyingTextTrack 剪映文字轨道
type JianyingTextTrack struct {
	ID    string             `json:"id"`
	Clips []JianyingTextClip `json:"clips"`
}

// JianyingVideoClip 剪映视频剪辑片段
type JianyingVideoClip struct {
	ID         string  `json:"id"`
	MaterialID string  `json:"material_id"`
	StartTime  float64 `json:"start_time"`
	Duration   float64 `json:"duration"`
	Position   Point   `json:"position"`
	Scale      float64 `json:"scale"`
	Rotation   float64 `json:"rotation,omitempty"`
}

// JianyingAudioClip 剪映音频剪辑片段
type JianyingAudioClip struct {
	ID         string  `json:"id"`
	MaterialID string  `json:"material_id"`
	StartTime  float64 `json:"start_time"`
	Duration   float64 `json:"duration"`
	Volume     float64 `json:"volume"`
}

// JianyingTextClip 剪映文字剪辑片段
type JianyingTextClip struct {
	ID         string  `json:"id"`
	MaterialID string  `json:"material_id"`
	StartTime  float64 `json:"start_time"`
	Duration   float64 `json:"duration"`
	Position   Point   `json:"position"`
}

// JianyingEffect 剪映特效
type JianyingEffect struct {
	Type     string                 `json:"type"`
	Duration float64                `json:"duration"`
	Params   map[string]interface{} `json:"params"`
}

// JianyingExtraMaterials 剪映额外素材
type JianyingExtraMaterials struct {
	Transitions []JianyingTransition `json:"transitions"`
	Filters     []JianyingFilter     `json:"filters"`
	Effects     []JianyingEffect     `json:"effects"`
}

// JianyingTransition 剪映转场
type JianyingTransition struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Duration float64 `json:"duration"`
}

// JianyingFilter 剪映滤镜
type JianyingFilter struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Intensity float64 `json:"intensity"`
}

// NewVideoProcessor 创建视频处理器
func NewVideoProcessor(logger *zap.Logger) *VideoProcessor {
	return &VideoProcessor{
		logger: logger,
	}
}

// CreateVideoProject 创建视频项目
func (vp *VideoProcessor) CreateVideoProject(chapterDir string, chapterNum int) (*VideoProject, error) {
	vp.logger.Info("创建视频项目",
		zap.String("章节目录", chapterDir),
		zap.Int("章节编号", chapterNum),
	)

	// 1. 扫描资产文件
	assets, err := vp.scanAssets(chapterDir)
	if err != nil {
		return nil, fmt.Errorf("扫描资产失败: %w", err)
	}

	// 2. 创建时间线
	timeline, duration, err := vp.createTimeline(assets, chapterNum)
	if err != nil {
		return nil, fmt.Errorf("创建时间线失败: %w", err)
	}

	// 3. 构建项目
	project := &VideoProject{
		Name:      fmt.Sprintf("小说第%d章", chapterNum),
		OutputDir: chapterDir,
		Assets:    *assets,
		Timeline:  timeline,
		Duration:  duration,
		Resolution: VideoResolution{
			Width:  1920,
			Height: 1080,
		},
	}

	// 4. 保存项目文件
	projectFile := filepath.Join(chapterDir, "video_project.json")
	if err := vp.saveProject(project, projectFile); err != nil {
		return nil, fmt.Errorf("保存项目文件失败: %w", err)
	}

	// 5. 创建剪映项目
	jianyingFile := filepath.Join(chapterDir, "jianying_project.draft")
	if err := vp.createJianyingProject(project, jianyingFile); err != nil {
		vp.logger.Warn("创建剪映项目失败", zap.Error(err))
		// 不返回错误，因为剪映项目是可选的
	}

	vp.logger.Info("视频项目创建完成",
		zap.String("项目文件", projectFile),
		zap.Float64("时长", duration),
	)

	return project, nil
}

// scanAssets 扫描资产文件
func (vp *VideoProcessor) scanAssets(chapterDir string) (*VideoAssets, error) {
	assets := &VideoAssets{}

	// 扫描音频文件
	audioDir := filepath.Join(chapterDir, "audio")
	if files, err := filepath.Glob(filepath.Join(audioDir, "*.wav")); err == nil {
		assets.AudioFiles = files
	}
	if files, err := filepath.Glob(filepath.Join(audioDir, "*.mp3")); err == nil {
		assets.AudioFiles = append(assets.AudioFiles, files...)
	}

	// 扫描字幕文件
	subtitleDir := filepath.Join(chapterDir, "subtitles")
	if files, err := filepath.Glob(filepath.Join(subtitleDir, "*.srt")); err == nil {
		assets.SubtitleFiles = files
	}
	if files, err := filepath.Glob(filepath.Join(subtitleDir, "*.ass")); err == nil {
		assets.SubtitleFiles = append(assets.SubtitleFiles, files...)
	}

	// 扫描图片文件
	imageDir := filepath.Join(chapterDir, "images")
	if files, err := filepath.Glob(filepath.Join(imageDir, "*.png")); err == nil {
		assets.ImageFiles = files
	}
	if files, err := filepath.Glob(filepath.Join(imageDir, "*.jpg")); err == nil {
		assets.ImageFiles = append(assets.ImageFiles, files...)
	}
	if files, err := filepath.Glob(filepath.Join(imageDir, "*.jpeg")); err == nil {
		assets.ImageFiles = append(assets.ImageFiles, files...)
	}

	vp.logger.Debug("扫描到的资产",
		zap.Int("音频文件", len(assets.AudioFiles)),
		zap.Int("字幕文件", len(assets.SubtitleFiles)),
		zap.Int("图片文件", len(assets.ImageFiles)),
	)

	return assets, nil
}

// createTimeline 创建时间线
func (vp *VideoProcessor) createTimeline(assets *VideoAssets, chapterNum int) ([]VideoClip, float64, error) {
	var timeline []VideoClip
	var totalDuration float64

	// 1. 添加主音频
	if len(assets.AudioFiles) > 0 {
		// 获取音频时长（优化实现）
		audioDuration, err := vp.getAudioDuration(assets.AudioFiles[0])
		if err != nil {
			vp.logger.Warn("获取音频时长失败，使用估算值", zap.Error(err))
			audioDuration = vp.estimateAudioDuration(assets.AudioFiles[0])
		}

		audioClip := VideoClip{
			ID:        fmt.Sprintf("audio_%d_1", chapterNum),
			Type:      "audio",
			File:      assets.AudioFiles[0],
			StartTime: 0,
			Duration:  audioDuration,
		}
		timeline = append(timeline, audioClip)
		totalDuration = audioDuration
	}

	// 2. 添加图片序列
	if len(assets.ImageFiles) > 0 {
		// 每张图片显示时间（秒）
		imageDuration := 5.0
		if totalDuration > 0 {
			// 根据音频时长调整图片显示时间
			imageDuration = totalDuration / float64(len(assets.ImageFiles))
			
			// 确保每张图片至少显示1秒
			if imageDuration < 1.0 {
				imageDuration = 1.0
			}
		}

		for i, imageFile := range assets.ImageFiles {
			startTime := float64(i) * imageDuration
			// 确保不超过总时长
			if startTime+imageDuration > totalDuration && totalDuration > 0 {
				startTime = totalDuration - imageDuration
				if startTime < 0 {
					startTime = 0
				}
			}

			imageClip := VideoClip{
				ID:        fmt.Sprintf("image_%d_%d", chapterNum, i+1),
				Type:      "image",
				File:      imageFile,
				StartTime: startTime,
				Duration:  imageDuration,
				Position:  Point{X: 0, Y: 0},
				Scale:     1.0,
				Effects: []Effect{
					{
						Type:     "fade_in",
						Duration: 0.5,
						Params: map[string]interface{}{
							"easing": "linear",
						},
					},
					{
						Type:     "fade_out",
						Duration: 0.5,
						Params: map[string]interface{}{
							"easing": "linear",
						},
					},
					{
						Type:     "ken_burns",
						Duration: imageDuration,
						Params: map[string]interface{}{
							"zoom":  1.1,
							"pan_x": 0.05,
							"pan_y": 0.03,
						},
					},
				},
			}
			timeline = append(timeline, imageClip)
		}
	}

	// 3. 添加字幕
	if len(assets.SubtitleFiles) > 0 {
		// 解析字幕文件
		subtitleLines, err := vp.parseSubtitleFile(assets.SubtitleFiles[0])
		if err != nil {
			vp.logger.Warn("解析字幕文件失败", zap.Error(err))
		} else {
			for i, line := range subtitleLines {
				subtitleClip := VideoClip{
					ID:        fmt.Sprintf("subtitle_%d_%d", chapterNum, i+1),
					Type:      "subtitle",
					File:      "", // 字幕内容在Params中
					StartTime: line.StartTime.Seconds(),
					Duration:  line.EndTime.Seconds() - line.StartTime.Seconds(),
					Position:  Point{X: 0, Y: 400}, // 底部位置
					Effects: []Effect{
						{
							Type:     "text_animation",
							Duration: 0.5,
							Params: map[string]interface{}{
								"type":      "fade",
								"direction": "up",
							},
						},
					},
				}
				timeline = append(timeline, subtitleClip)
			}
		}
	}

	return timeline, totalDuration, nil
}

// getAudioDuration 获取音频真实时长
func (vp *VideoProcessor) getAudioDuration(audioFile string) (float64, error) {
	// 尝试使用ffprobe获取精确时长
	cmd := exec.Command("ffprobe", "-v", "quiet", "-show_entries", "format=duration", "-of", "csv=p=0", audioFile)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	
	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, err
	}
	
	return duration, nil
}

// estimateAudioDuration 估算音频时长
func (vp *VideoProcessor) estimateAudioDuration(audioFile string) float64 {
	// 简单估算：文件大小 / 比特率
	// 实际应用中应该使用音频库解析
	if info, err := os.Stat(audioFile); err == nil {
		// 假设是128kbps的MP3
		fileSizeMB := float64(info.Size()) / (1024 * 1024)
		return fileSizeMB * 60 / 1.0 // 粗略估算
	}

	// 默认返回5分钟
	return 300.0
}

// parseSubtitleFile 解析字幕文件
func (vp *VideoProcessor) parseSubtitleFile(subtitleFile string) ([]SubtitleLine, error) {
	// 从subtitle包导入解析逻辑
	ext := filepath.Ext(subtitleFile)
	
	switch ext {
	case ".srt":
		return vp.parseSRTFile(subtitleFile)
	case ".ass":
		return vp.parseASSFile(subtitleFile)
	default:
		return nil, fmt.Errorf("不支持的字幕格式: %s", ext)
	}
}

// parseSRTFile 解析SRT字幕文件
func (vp *VideoProcessor) parseSRTFile(subtitleFile string) ([]SubtitleLine, error) {
	_, err := os.ReadFile(subtitleFile)
	if err != nil {
		return nil, err
	}
	
	// 这里我们模拟SubtitleLine结构，实际需要引入subtitle包
	// 由于循环依赖，我们暂时返回空切片
	return []SubtitleLine{}, nil
}

// parseASSFile 解析ASS字幕文件
func (vp *VideoProcessor) parseASSFile(subtitleFile string) ([]SubtitleLine, error) {
	_, err := os.ReadFile(subtitleFile)
	if err != nil {
		return nil, err
	}
	
	// 这里我们模拟SubtitleLine结构，实际需要引入subtitle包
	// 由于循环依赖，我们暂时返回空切片
	return []SubtitleLine{}, nil
}

// saveProject 保存项目文件
func (vp *VideoProcessor) saveProject(project *VideoProject, filename string) error {
	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化项目失败: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("写入项目文件失败: %w", err)
	}

	vp.logger.Info("项目文件已保存", zap.String("文件", filename))
	return nil
}

// createJianyingProject 创建剪映项目
func (vp *VideoProcessor) createJianyingProject(project *VideoProject, outputFile string) error {
	now := time.Now().Unix()

	jianyingProject := JianyingProject{
		AppVersion: "5.0.0",
		ProjectInfo: JianyingProjectInfo{
			Name:       project.Name,
			CreateTime: now,
			UpdateTime: now,
			Duration:   project.Duration,
			Resolution: "1920x1080",
			FrameRate:  30,
			SampleRate: 44100,
			Volume:     1.0,
		},
		Materials: JianyingMaterials{
			Audios: vp.convertToJianyingAudios(project.Assets.AudioFiles),
			Images: vp.convertToJianyingImages(project.Assets.ImageFiles),
			Texts:  vp.convertToJianyingTexts(project.Timeline),
		},
		Tracks: JianyingTracks{
			VideoTracks: vp.createJianyingVideoTracks(project.Timeline),
			AudioTracks: vp.createJianyingAudioTracks(project.Timeline),
			TextTracks:  vp.createJianyingTextTracks(project.Timeline),
		},
		ExtraMaterials: JianyingExtraMaterials{
			Transitions: []JianyingTransition{
				{
					ID:       "fade",
					Name:     "淡入淡出",
					Duration: 0.5,
				},
			},
			Filters: []JianyingFilter{
				{
					ID:        "cinematic",
					Name:      "电影感",
					Intensity: 0.3,
				},
			},
		},
	}

	data, err := json.MarshalIndent(jianyingProject, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化剪映项目失败: %w", err)
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("写入剪映项目文件失败: %w", err)
	}

	// 创建剪映项目需要的文件夹结构
	draftDir := strings.TrimSuffix(outputFile, ".draft")
	if err := os.MkdirAll(draftDir, 0755); err != nil {
		return fmt.Errorf("创建剪映项目目录失败: %w", err)
	}

	// 创建素材文件夹
	materialsDir := filepath.Join(draftDir, "materials")
	if err := os.MkdirAll(materialsDir, 0755); err != nil {
		return fmt.Errorf("创建素材目录失败: %w", err)
	}

	// 复制资产文件到剪映项目目录
	if err := vp.copyAssetsToJianying(project.Assets, materialsDir); err != nil {
		vp.logger.Warn("复制资产到剪映项目失败", zap.Error(err))
	}

	vp.logger.Info("剪映项目已创建", zap.String("文件", outputFile))
	return nil
}

// 辅助转换函数
func (vp *VideoProcessor) convertToJianyingAudios(audioFiles []string) []JianyingAudio {
	var audios []JianyingAudio
	for i, file := range audioFiles {
		duration := vp.estimateAudioDuration(file)
		audios = append(audios, JianyingAudio{
			ID:       fmt.Sprintf("audio_%d", i+1),
			Name:     filepath.Base(file),
			Path:     file,
			Duration: duration,
			Volume:   1.0,
			FadeIn:   0.5,
			FadeOut:  0.5,
		})
	}
	return audios
}

func (vp *VideoProcessor) convertToJianyingImages(imageFiles []string) []JianyingImage {
	var images []JianyingImage
	for i, file := range imageFiles {
		images = append(images, JianyingImage{
			ID:       fmt.Sprintf("image_%d", i+1),
			Name:     filepath.Base(file),
			Path:     file,
			Duration: 5.0,
			Width:    1920,
			Height:   1080,
		})
	}
	return images
}

func (vp *VideoProcessor) convertToJianyingTexts(timeline []VideoClip) []JianyingText {
	var texts []JianyingText
	// 从时间线中提取字幕信息
	for _, clip := range timeline {
		if clip.Type == "subtitle" {
			texts = append(texts, JianyingText{
				ID:       fmt.Sprintf("text_%s", clip.ID),
				Content:  clip.File, // 实际应从字幕文件中提取内容
				Duration: clip.Duration,
				Style:    "Default",
				FontSize: 24,
				Color:    "&H00FFFFFF", // 白色
			})
		}
	}
	return texts
}

func (vp *VideoProcessor) createJianyingVideoTracks(timeline []VideoClip) []JianyingVideoTrack {
	var tracks []JianyingVideoTrack
	
	// 分离不同类型的剪辑
	var imageClips []JianyingVideoClip
	for _, clip := range timeline {
		if clip.Type == "image" {
			imageClips = append(imageClips, JianyingVideoClip{
				ID:         clip.ID,
				MaterialID: fmt.Sprintf("image_%s", clip.ID),
				StartTime:  clip.StartTime,
				Duration:   clip.Duration,
				Position:   clip.Position,
				Scale:      clip.Scale,
			})
		}
	}
	
	if len(imageClips) > 0 {
		tracks = append(tracks, JianyingVideoTrack{
			ID:    "video_track_1",
			Clips: imageClips,
		})
	}
	
	return tracks
}

func (vp *VideoProcessor) createJianyingAudioTracks(timeline []VideoClip) []JianyingAudioTrack {
	var tracks []JianyingAudioTrack
	
	// 分离音频剪辑
	var audioClips []JianyingAudioClip
	for _, clip := range timeline {
		if clip.Type == "audio" {
			audioClips = append(audioClips, JianyingAudioClip{
				ID:         clip.ID,
				MaterialID: fmt.Sprintf("audio_%s", clip.ID),
				StartTime:  clip.StartTime,
				Duration:   clip.Duration,
				Volume:     1.0,
			})
		}
	}
	
	if len(audioClips) > 0 {
		tracks = append(tracks, JianyingAudioTrack{
			ID:    "audio_track_1",
			Clips: audioClips,
		})
	}
	
	return tracks
}

func (vp *VideoProcessor) createJianyingTextTracks(timeline []VideoClip) []JianyingTextTrack {
	var tracks []JianyingTextTrack
	
	// 分离文本/字幕剪辑
	var textClips []JianyingTextClip
	for _, clip := range timeline {
		if clip.Type == "subtitle" {
			textClips = append(textClips, JianyingTextClip{
				ID:         clip.ID,
				MaterialID: fmt.Sprintf("text_%s", clip.ID),
				StartTime:  clip.StartTime,
				Duration:   clip.Duration,
				Position:   clip.Position,
			})
		}
	}
	
	if len(textClips) > 0 {
		tracks = append(tracks, JianyingTextTrack{
			ID:    "text_track_1",
			Clips: textClips,
		})
	}
	
	return tracks
}

func (vp *VideoProcessor) copyAssetsToJianying(assets VideoAssets, targetDir string) error {
	// 复制所有资产文件
	var allFiles []string
	allFiles = append(allFiles, assets.AudioFiles...)
	allFiles = append(allFiles, assets.ImageFiles...)
	allFiles = append(allFiles, assets.SubtitleFiles...)

	var wg sync.WaitGroup
	errChan := make(chan error, len(allFiles))

	for _, file := range allFiles {
		wg.Add(1)
		go func(srcFile string) {
			defer wg.Done()
			
			dstFile := filepath.Join(targetDir, filepath.Base(srcFile))
			if err := vp.copyFile(srcFile, dstFile); err != nil {
				errChan <- fmt.Errorf("复制文件失败 %s: %w", srcFile, err)
			}
		}(file)
	}

	wg.Wait()
	close(errChan)

	// 检查是否有错误
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (vp *VideoProcessor) copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}

// GenerateEditList 生成编辑清单
func (vp *VideoProcessor) GenerateEditList(chapterDir string, chapterNum int) (string, error) {
	vp.logger.Info("生成编辑清单",
		zap.String("目录", chapterDir),
		zap.Int("章节", chapterNum),
	)

	// 1. 创建视频项目
	project, err := vp.CreateVideoProject(chapterDir, chapterNum)
	if err != nil {
		return "", fmt.Errorf("创建视频项目失败: %w", err)
	}

	// 2. 生成编辑指令文本
	editList := vp.generateEditInstructions(project)

	// 3. 保存编辑清单
	editListFile := filepath.Join(chapterDir, "edit_instructions.txt")
	if err := os.WriteFile(editListFile, []byte(editList), 0644); err != nil {
		return "", fmt.Errorf("保存编辑清单失败: %w", err)
	}

	vp.logger.Info("编辑清单已生成", zap.String("文件", editListFile))
	return editListFile, nil
}

func (vp *VideoProcessor) generateEditInstructions(project *VideoProject) string {
	var instructions strings.Builder

	instructions.WriteString("# 小说视频编辑清单\n\n")
	instructions.WriteString(fmt.Sprintf("## 项目: %s\n", project.Name))
	instructions.WriteString(fmt.Sprintf("## 时长: %.1f秒\n\n", project.Duration))

	instructions.WriteString("## 资产文件\n")
	instructions.WriteString("### 音频文件:\n")
	for _, audio := range project.Assets.AudioFiles {
		instructions.WriteString(fmt.Sprintf("- %s\n", filepath.Base(audio)))
	}

	instructions.WriteString("\n### 图片文件:\n")
	for _, image := range project.Assets.ImageFiles {
		instructions.WriteString(fmt.Sprintf("- %s\n", filepath.Base(image)))
	}

	instructions.WriteString("\n### 字幕文件:\n")
	for _, subtitle := range project.Assets.SubtitleFiles {
		instructions.WriteString(fmt.Sprintf("- %s\n", filepath.Base(subtitle)))
	}

	instructions.WriteString("\n## 编辑步骤\n")
	instructions.WriteString("1. 打开剪映专业版\n")
	instructions.WriteString("2. 导入所有资产文件\n")
	instructions.WriteString("3. 按照以下时间线进行编辑:\n\n")

	for _, clip := range project.Timeline {
		instructions.WriteString(fmt.Sprintf("### %s: %s\n",
			strings.ToUpper(clip.Type),
			filepath.Base(clip.File)))

		instructions.WriteString(fmt.Sprintf("- 开始时间: %.1f秒\n", clip.StartTime))
		instructions.WriteString(fmt.Sprintf("- 持续时间: %.1f秒\n", clip.Duration))

		if len(clip.Effects) > 0 {
			instructions.WriteString("- 特效:\n")
			for _, effect := range clip.Effects {
				instructions.WriteString(fmt.Sprintf("  - %s (%.1f秒)\n",
					effect.Type, effect.Duration))
			}
		}
		instructions.WriteString("\n")
	}

	instructions.WriteString("## 导出设置\n")
	instructions.WriteString("- 分辨率: 1920x1080\n")
	instructions.WriteString("- 帧率: 30fps\n")
	instructions.WriteString("- 码率: 10Mbps\n")
	instructions.WriteString("- 格式: MP4\n")

	return instructions.String()
}