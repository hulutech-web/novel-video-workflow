package workflow

import (
	aegisub "novel-video-workflow/pkg/tools/aegisub"
	drawthings "novel-video-workflow/pkg/tools/drawthings"
	"novel-video-workflow/pkg/tools/file"
	image "novel-video-workflow/pkg/tools/image"
	"novel-video-workflow/pkg/tools/indextts2"
	"novel-video-workflow/pkg/capcut"

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
	capcutTool     *capcut.CapcutGenerator
	logger         *zap.Logger
}

func NewProcessor(logger *zap.Logger) (*Processor, error) {
	// 初始化各个工具
	fileTool := file.NewFileManager()
	ttsTool := indextts2.NewIndexTTS2Client(logger, "http://localhost:7860")
	aegisubTool := aegisub.NewAegisubIntegration()
	imageTool := image.NewImageGenerator(logger)
	drawThingsTool := drawthings.NewChapterImageGenerator(logger)
	capcutTool := capcut.NewCapcutGenerator(logger)

	return &Processor{
		fileTool:       fileTool,
		ttsTool:        ttsTool,
		aegisubTool:    aegisubTool,
		imageTool:      imageTool,
		drawThingsTool: drawThingsTool,
		capcutTool:     capcutTool,
		logger:         logger,
	}, nil
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

func (p *Processor) GenerateCapcutProject(chapterDir string) error {
	// 使用 CapCut 生成器生成剪映项目
	return p.capcutTool.GenerateProject(chapterDir)
}

func (p *Processor) GetProgress() any {
	return nil
}