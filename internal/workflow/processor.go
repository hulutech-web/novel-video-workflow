package workflow

import (
	"context"
	"fmt"
	"novel-video-workflow/internal/tools"
	"path/filepath"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type ChapterParams struct {
	Text           string
	Number         int
	ReferenceAudio string
	OutputDir      string
	MaxImages      int // 添加最大图片数参数
}

type ChapterResult struct {
	ChapterDir     string   `json:"chapter_dir"`
	TextFile       string   `json:"text_file"`
	AudioFile      string   `json:"audio_file"`
	SubtitleFile   string   `json:"subtitle_file"`
	ImageFiles     []string `json:"image_files"`
	Status         string   `json:"status"`
	Message        string   `json:"message"`
	VideoProject   string   `json:"video_project,omitempty"`   // 添加视频项目路径
	EditListFile   string   `json:"edit_list_file,omitempty"`  // 添加编辑清单路径
	ProcessingTime float64  `json:"processing_time,omitempty"` // 添加处理时间
}

type Processor struct {
	fileTool     *tools.FileManager
	ttsTool      *tools.TTSProcessor
	subtitleTool *tools.SubtitleGenerator
	imageTool    *tools.ImageGenerator
	logger       *zap.Logger
}

func NewProcessor(logger *zap.Logger) (*Processor, error) {
	// 初始化各个工具
	fileTool := tools.NewFileManager()
	ttsTool := tools.NewTTSProcessor(logger)
	subtitleTool := tools.NewSubtitleGenerator(logger)
	imageTool := tools.NewImageGenerator(logger)

	return &Processor{
		fileTool:     fileTool,
		ttsTool:      ttsTool,
		subtitleTool: subtitleTool,
		imageTool:    imageTool,
		logger:       logger,
	}, nil
}

// ProcessChapter 处理单个章节
func (p *Processor) ProcessChapter(ctx context.Context, params ChapterParams) (*ChapterResult, error) {
	if params.MaxImages == 0 {
		params.MaxImages = 5 // 默认生成5张图片
	}

	// 1. 创建章节文件结构
	p.logger.Info("开始处理章节",
		zap.Int("章节", params.Number),
		zap.String("输出目录", params.OutputDir),
	)

	dirInfo, err := p.fileTool.CreateChapterStructure(params.Number, params.Text, params.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("创建章节结构失败: %w", err)
	}

	// 2. 生成音频文件
	p.logger.Info("开始生成音频...",
		zap.Int("章节", params.Number),
	)

	audioFile := filepath.Join(dirInfo.AudioDir, fmt.Sprintf("chapter_%d.wav", params.Number))
	ttsResult, err := p.ttsTool.Generate(params.Text, audioFile, params.ReferenceAudio)
	if err != nil {
		p.logger.Warn("音频生成失败，继续处理",
			zap.Error(err),
		)
		audioFile = "" // 清空音频文件路径
	} else if !ttsResult.Success {
		p.logger.Warn("音频生成失败",
			zap.String("错误", ttsResult.Error),
		)
		audioFile = ""
	}

	// 3. 生成字幕
	p.logger.Info("开始生成字幕...")

	subtitleFile := filepath.Join(dirInfo.SubtitleDir, fmt.Sprintf("chapter_%d.srt", params.Number))

	var subtitleResult *tools.SubtitleResult
	var subtitleErr error

	// 根据配置选择字幕生成方式
	generatorType := viper.GetString("subtitle.generator")

	switch generatorType {
	case "aegisub":
		if ttsResult != nil && ttsResult.Success {
			subtitleResult, subtitleErr = p.subtitleTool.GenerateWithAegisub(
				params.Text,
				audioFile,
				subtitleFile,
			)
		}
	case "static":
		subtitleResult, subtitleErr = p.subtitleTool.GenerateStatic(
			params.Text,
			subtitleFile,
		)
	default: // auto
		if ttsResult != nil && ttsResult.Success {
			subtitleResult, subtitleErr = p.subtitleTool.GenerateFromAudio(
				params.Text,
				audioFile,
				subtitleFile,
			)
		} else {
			subtitleResult, subtitleErr = p.subtitleTool.GenerateStatic(
				params.Text,
				subtitleFile,
			)
		}
	}

	if subtitleErr != nil {
		p.logger.Warn("字幕生成失败",
			zap.Error(subtitleErr),
			zap.String("生成方式", generatorType),
		)
		subtitleFile = "" // 清空文件路径，表示生成失败
	} else if subtitleResult != nil && subtitleResult.Success {
		p.logger.Info("字幕生成成功",
			zap.String("文件", subtitleResult.SubtitleFile),
			zap.String("格式", subtitleResult.Format),
			zap.Int("行数", subtitleResult.LineCount),
		)

		// 如果需要，转换格式
		if subtitleResult.Format != "srt" && subtitleResult.Format != "ass" {
			// 转换为ASS格式，便于视频编辑
			if converted, err := p.subtitleTool.ConvertFormat(subtitleFile, "ass"); err == nil {
				subtitleFile = converted
				p.logger.Info("字幕已转换为ASS格式", zap.String("文件", subtitleFile))
			}
		}
	}

	// 4. 生成场景图片
	p.logger.Info("开始生成场景图片...",
		zap.Int("章节", params.Number),
		zap.String("输出目录", dirInfo.ImageDir),
		zap.Int("最大图片数", params.MaxImages),
	)

	sceneImages, err := p.imageTool.GenerateSceneImages(
		params.Text,
		dirInfo.ImageDir,
		params.MaxImages,
	)
	if err != nil {
		p.logger.Warn("场景图片生成失败", zap.Error(err))
	}

	var imageFiles []string
	for _, img := range sceneImages {
		imageFiles = append(imageFiles, img.ImageFile)
	}

	// 5. 创建视频项目
	videoTool := tools.NewVideoProcessor(p.logger)
	videoProject, err := videoTool.CreateVideoProject(dirInfo.ChapterDir, params.Number)
	if err != nil {
		p.logger.Warn("视频项目创建失败", zap.Error(err))
	}

	// 6. 生成编辑清单
	editListFile, err := videoTool.GenerateEditList(dirInfo.ChapterDir, params.Number)
	if err != nil {
		p.logger.Warn("编辑清单生成失败", zap.Error(err))
	}

	result := &ChapterResult{
		ChapterDir:   dirInfo.ChapterDir,
		TextFile:     dirInfo.TextFile,
		AudioFile:    audioFile,
		SubtitleFile: subtitleFile,
		ImageFiles:   imageFiles,
		Status:       "completed",
		Message:      "章节处理完成",
		VideoProject: videoProject.Name,
		EditListFile: editListFile,
	}

	return result, nil
}

func (p *Processor) generateEditList(chapterDir string, chapterNum int,
	ttsResult *tools.TTSResult, subtitleFile string, images []string) map[string]interface{} {

	return map[string]interface{}{
		"chapter": chapterNum,
		"assets": map[string]interface{}{
			"audio":    ttsResult.OutputFile,
			"subtitle": subtitleFile,
			"images":   images,
		},
		"timeline": []map[string]interface{}{
			{
				"time": "00:00",
				"type": "audio_start",
				"file": ttsResult.OutputFile,
			},
		},
	}
}

func (p *Processor) GetProgress() any {
	return nil
}
