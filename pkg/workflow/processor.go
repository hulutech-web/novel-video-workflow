package workflow

import (
	"context"
	"fmt"
	"novel-video-workflow/pkg/tools/file"
	image "novel-video-workflow/pkg/tools/image"
	video "novel-video-workflow/pkg/tools/video"
	aegisub "novel-video-workflow/pkg/tools/aegisub"
	drawthings "novel-video-workflow/pkg/tools/drawthings"
	"path/filepath"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"novel-video-workflow/pkg/tools/indextts2"
)

type ChapterParams struct {
	Text           string
	Number         int
	ReferenceAudio string
	OutputDir      string
	MaxImages      int // 添加最大图片数参数
}

type ChapterResult struct {
	ChapterDir   string
	TextFile     string
	AudioFile    string
	SubtitleFile string
	ImageFiles   []string
	Status       string
	Message      string
	VideoProject string
	EditListFile string
}

type Processor struct {
	fileTool       *file.FileManager
	ttsTool        *indextts2.IndexTTS2Client
	aegisubTool    *aegisub.AegisubIntegration
	imageTool      *image.ImageGenerator
	drawThingsTool *drawthings.ChapterImageGenerator
	logger         *zap.Logger
}

func NewProcessor(logger *zap.Logger) (*Processor, error) {
	// 初始化各个工具
	fileTool := file.NewFileManager()
	ttsTool := indextts2.NewIndexTTS2Client(logger, "http://localhost:7860")
	aegisubTool := aegisub.NewAegisubIntegration()
	imageTool := image.NewImageGenerator(logger)
	drawThingsTool := drawthings.NewChapterImageGenerator(logger)

	return &Processor{
		fileTool:       fileTool,
		ttsTool:        ttsTool,
		aegisubTool:    aegisubTool,
		imageTool:      imageTool,
		drawThingsTool: drawThingsTool,
		logger:         logger,
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

	audioFile := filepath.Join(dirInfo.AudioDir, fmt.Sprintf("chapter_%02d.wav", params.Number))
	
	// 直接使用IndexTTS2客户端生成音频
	var audioErr error
	if params.ReferenceAudio != "" {
		audioErr = p.ttsTool.GenerateTTSWithAudio(params.ReferenceAudio, params.Text, audioFile)
		if audioErr != nil {
			p.logger.Warn("音频生成失败，继续处理",
				zap.Error(audioErr),
			)
			audioFile = "" // 清空音频文件路径
		}
	} else {
		p.logger.Info("未提供参考音频，跳过音频生成")
		audioFile = ""
	}

	// 3. 生成字幕
	p.logger.Info("开始生成字幕...")

	subtitleFile := filepath.Join(dirInfo.SubtitleDir, fmt.Sprintf("chapter_%02d.srt", params.Number))

	var subtitleErr error

	// 根据配置选择字幕生成方式
	generatorType := viper.GetString("subtitle.generator")

	switch generatorType {
	case "aegisub":
		if audioFile != "" {
			// 使用AegisubIntegration生成字幕
			subtitleErr = p.aegisubTool.ProcessIndextts2OutputWithCustomName(
				audioFile,
				params.Text,
				subtitleFile,
			)
		}
	case "static":
		// 如果选择静态字幕生成，我们暂时跳过（因为没有实现）
		p.logger.Info("静态字幕生成暂未实现")
	default: // auto
		if audioFile != "" {
			// 使用AegisubIntegration生成字幕
			subtitleErr = p.aegisubTool.ProcessIndextts2OutputWithCustomName(
				audioFile,
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
	} else {
		p.logger.Info("字幕生成成功", zap.String("文件", subtitleFile))
	}

	// 4. 使用DrawThings API生成场景图片 - 通过AI分析场景生成提示词
	p.logger.Info("开始使用DrawThings生成场景图片...",
		zap.Int("章节", params.Number),
		zap.String("输出目录", dirInfo.ImageDir),
		zap.Int("最大图片数", params.MaxImages),
	)

	// 使用AI生成提示词并生成图像
	imageDir := dirInfo.ImageDir
	sceneImages, err := p.drawThingsTool.GenerateImagesFromChapter(params.Text, imageDir, 1024, 1792, true)
	if err != nil {
		p.logger.Warn("使用DrawThings生成场景图片失败", zap.Error(err))
		
		// 回退到原有图片生成方法
		_, err2 := p.imageTool.GenerateSceneImages(
			params.Text,
			dirInfo.ImageDir,
			params.MaxImages,
		)
		if err2 != nil {
			p.logger.Warn("回退图片生成也失败", zap.Error(err2))
		} else {
			p.logger.Info("使用回退方法生成图片成功")
		}
	}

	var imageFiles []string
	for _, img := range sceneImages {
		imageFiles = append(imageFiles, img.ImageFile)
	}

	// 5. 创建视频项目
	videoTool := video.NewVideoProcessor(p.logger)
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
	ttsResult *indextts2.TTSResult, subtitleFile string, images []string) map[string]interface{} {

	return map[string]interface{}{
		"chapter": chapterNum,
		"assets": map[string]interface{}{
			"audio":    ttsResult.AudioPath,
			"subtitle": subtitleFile,
			"images":   images,
		},
		"timeline": []map[string]interface{}{
			{
				"time": "00:00",
				"type": "audio_start",
				"file": ttsResult.AudioPath,
			},
		},
	}
}

func (p *Processor) GetProgress() any {
	return nil
}