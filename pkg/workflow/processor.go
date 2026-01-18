package workflow

import (
	aegisub "novel-video-workflow/pkg/tools/aegisub"
	drawthings "novel-video-workflow/pkg/tools/drawthings"
	"novel-video-workflow/pkg/tools/file"
	image "novel-video-workflow/pkg/tools/image"
	"novel-video-workflow/pkg/tools/indextts2"
	"novel-video-workflow/pkg/capcut"
	"novel-video-workflow/pkg/database"

	"fmt"
	"os"
	"path/filepath"
	"time"

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
	dbManager      *database.GormManager
}

func NewProcessor(logger *zap.Logger) (*Processor, error) {
	// 初始化各个工具
	fileTool := file.NewFileManager()
	ttsTool := indextts2.NewIndexTTS2Client(logger, "http://localhost:7860")
	aegisubTool := aegisub.NewAegisubIntegration()
	imageTool := image.NewImageGenerator(logger)
	drawThingsTool := drawthings.NewChapterImageGenerator(logger)
	capcutTool := capcut.NewCapcutGenerator(logger)

	// 初始化数据库管理器
	dbManager, err := database.NewGormManager()
	if err != nil {
		logger.Error("Failed to initialize database manager", zap.Error(err))
		return nil, err
	}

	return &Processor{
		fileTool:       fileTool,
		ttsTool:        ttsTool,
		aegisubTool:    aegisubTool,
		imageTool:      imageTool,
		drawThingsTool: drawThingsTool,
		capcutTool:     capcutTool,
		logger:         logger,
		dbManager:      dbManager,
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

// CreateProject 创建新项目
func (p *Processor) CreateProject(name, description, genre, atmosphere string) (*database.Project, error) {
	project, err := p.dbManager.CreateProject(name, description, genre, atmosphere)
	if err != nil {
		p.logger.Error("Failed to create project", zap.Error(err))
		return nil, err
	}
	
	p.logger.Info("Created new project", zap.String("name", name), zap.String("genre", genre))
	return project, nil
}

// ProcessChapterWithTracking 处理章节并跟踪进度
func (p *Processor) ProcessChapterWithTracking(novelName, chapterName, inputFilePath, outputDir string) (*ChapterResult, error) {
	// 首先检查是否已存在项目，如果不存在则创建
	project, err := p.dbManager.GetProjectByName(novelName)
	if err != nil {
		p.logger.Error("Failed to get project", zap.Error(err))
	}
	
	if project == nil {
		project, err = p.dbManager.CreateProject(novelName, fmt.Sprintf("Project for novel: %s", novelName), "", "")
		if err != nil {
			p.logger.Error("Failed to create project", zap.Error(err))
			return nil, err
		}
		p.logger.Info("Created new project", zap.String("novel", novelName))
	}

	// 检查是否已存在处理记录
	existingProcess, err := p.dbManager.GetChapterProcessByProjectAndName(uint(project.ID), chapterName)
	if err != nil {
		p.logger.Error("Failed to get existing chapter process", zap.Error(err))
	}

	var chapterProcess *database.ChapterProcess
	if existingProcess != nil {
		// 如果存在失败的记录，允许重试
		if existingProcess.Status == database.StatusFailed || existingProcess.Status == database.StatusSkipped {
			err = p.dbManager.RetryChapterProcess(uint(existingProcess.ID))
			if err != nil {
				p.logger.Error("Failed to reset chapter process for retry", zap.Error(err))
				return nil, err
			}
			chapterProcess = existingProcess
			p.logger.Info("Reset chapter process for retry", zap.String("novel", novelName), zap.String("chapter", chapterName))
		} else {
			// 如果已有成功的记录，直接返回
			p.logger.Info("Chapter process already exists and completed", zap.String("novel", novelName), zap.String("chapter", chapterName))
			result := &ChapterResult{
				Status:  string(existingProcess.Status),
				Message: "Chapter already processed",
			}
			return result, nil
		}
	} else {
		// 创建新的处理记录
		chapterProcess = &database.ChapterProcess{
			ProjectID:     uint(project.ID),
			ChapterName:   chapterName,
			InputFilePath: inputFilePath,
			OutputDir:     outputDir,
			Status:        database.StatusPending,
		}

		err = p.dbManager.CreateChapterProcess(chapterProcess)
		if err != nil {
			p.logger.Error("Failed to create chapter process record", zap.Error(err))
			return nil, err
		}
		p.logger.Info("Created new chapter process record", zap.String("novel", novelName), zap.String("chapter", chapterName))
	}

	// 更新状态为处理中
	chapterProcess.Status = database.StatusProcessing
	err = p.dbManager.UpdateChapterProcessStatus(uint(chapterProcess.ID), database.StatusProcessing, "")
	if err != nil {
		p.logger.Error("Failed to update chapter process status to processing", zap.Error(err))
		return nil, err
	}

	// 创建章节目录
	chapterDir := filepath.Join(outputDir, chapterName)
	if err := os.MkdirAll(chapterDir, 0755); err != nil {
		errMsg := fmt.Sprintf("Failed to create chapter directory: %v", err)
		p.logger.Error(errMsg)
		p.dbManager.UpdateChapterProcessStatus(uint(chapterProcess.ID), database.StatusFailed, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	result := &ChapterResult{
		ChapterDir: chapterDir,
		Status:     string(database.StatusProcessing),
	}

	// 处理步骤1: 音频生成
	audioStep, err := p.dbManager.GetProcessStepByChapterAndName(uint(chapterProcess.ID), "audio")
	if err != nil {
		p.logger.Error("Failed to get audio step", zap.Error(err))
	}
	
	if audioStep == nil || audioStep.Status != database.StatusCompleted {
		err = p.dbManager.UpdateChapterProcessStatus(uint(chapterProcess.ID), database.StatusProcessing, "Generating audio")
		if err != nil {
			p.logger.Error("Failed to update chapter process status for audio generation", zap.Error(err))
			return nil, err
		}

		// 创建或重置音频步骤
		audioStep = &database.ProcessStep{
			ChapterID: uint(chapterProcess.ID),
			StepName:  "audio",
			Status:    database.StatusPending,
		}

		if audioStep.ID == 0 {
			err = p.dbManager.CreateProcessStep(audioStep)
			if err != nil {
				p.logger.Error("Failed to create audio step", zap.Error(err))
				return nil, err
			}
		}

		audioStep.Status = database.StatusProcessing
		audioStep.StartTime = database.MyTime{Time: time.Now()}
		err = p.dbManager.UpdateProcessStep(audioStep)
		if err != nil {
			p.logger.Error("Failed to update audio step status to processing", zap.Error(err))
			return nil, err
		}

		// 执行音频生成
		audioFile := filepath.Join(chapterDir, chapterName+".wav")
		refAudioPath := filepath.Join("assets", "ref_audio", "ref.m4a")
		if _, statErr := os.Stat(refAudioPath); statErr == nil {
			// 读取章节文本
			chapterText, err := os.ReadFile(inputFilePath)
			if err != nil {
				errMsg := fmt.Sprintf("Failed to read chapter file: %v", err)
				audioStep.Status = database.StatusFailed
				audioStep.ErrorMsg = errMsg
				audioStep.EndTime = database.MyTime{Time: time.Now()}
				p.dbManager.UpdateProcessStep(audioStep)
				p.dbManager.UpdateChapterProcessStatus(uint(chapterProcess.ID), database.StatusFailed, errMsg)
				return nil, fmt.Errorf(errMsg)
			}

			err = p.ttsTool.GenerateTTSWithAudio(refAudioPath, string(chapterText), audioFile)
			if err != nil {
				errMsg := fmt.Sprintf("Failed to generate audio: %v", err)
				audioStep.Status = database.StatusFailed
				audioStep.ErrorMsg = errMsg
				audioStep.EndTime = database.MyTime{Time: time.Now()}
				p.dbManager.UpdateProcessStep(audioStep)
				p.dbManager.UpdateChapterProcessStatus(uint(chapterProcess.ID), database.StatusFailed, errMsg)
				return nil, fmt.Errorf(errMsg)
			}

			audioStep.Status = database.StatusCompleted
			audioStep.EndTime = database.MyTime{Time: time.Now()}
			audioStep.Duration = int64(time.Since(audioStep.StartTime.Time) / time.Second)
			err = p.dbManager.UpdateProcessStep(audioStep)
			if err != nil {
				p.logger.Error("Failed to update audio step status to completed", zap.Error(err))
			}

			result.AudioFile = audioFile
			chapterProcess.AudioGenerated = true
		} else {
			p.logger.Info("Reference audio not found, skipping audio generation")
			audioStep.Status = database.StatusSkipped
			audioStep.ErrorMsg = "Reference audio not found"
			audioStep.EndTime = database.MyTime{Time: time.Now()}
			audioStep.Duration = 0
			err = p.dbManager.UpdateProcessStep(audioStep)
			if err != nil {
				p.logger.Error("Failed to update audio step status to skipped", zap.Error(err))
			}
		}
	}

	// 处理步骤2: 场景拆分和图像生成
	imageStep, err := p.dbManager.GetProcessStepByChapterAndName(uint(chapterProcess.ID), "image")
	if err != nil {
		p.logger.Error("Failed to get image step", zap.Error(err))
	}

	if imageStep == nil || imageStep.Status != database.StatusCompleted {
		err = p.dbManager.UpdateChapterProcessStatus(uint(chapterProcess.ID), database.StatusProcessing, "Generating images")
		if err != nil {
			p.logger.Error("Failed to update chapter process status for image generation", zap.Error(err))
			return nil, err
		}

		// 创建或重置图像步骤
		imageStep = &database.ProcessStep{
			ChapterID: uint(chapterProcess.ID),
			StepName:  "image",
			Status:    database.StatusPending,
		}

		if imageStep.ID == 0 {
			err = p.dbManager.CreateProcessStep(imageStep)
			if err != nil {
				p.logger.Error("Failed to create image step", zap.Error(err))
				return nil, err
			}
		}

		imageStep.Status = database.StatusProcessing
		imageStep.StartTime = database.MyTime{Time: time.Now()}
		err = p.dbManager.UpdateProcessStep(imageStep)
		if err != nil {
			p.logger.Error("Failed to update image step status to processing", zap.Error(err))
			return nil, err
		}

		// 执行图像生成 - 先进行场景拆分
		chapterText, err := os.ReadFile(inputFilePath)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to read chapter file: %v", err)
			imageStep.Status = database.StatusFailed
			imageStep.ErrorMsg = errMsg
			imageStep.EndTime = database.MyTime{Time: time.Now()}
			p.dbManager.UpdateProcessStep(imageStep)
			p.dbManager.UpdateChapterProcessStatus(uint(chapterProcess.ID), database.StatusFailed, errMsg)
			return nil, fmt.Errorf(errMsg)
		}

		// 进行场景拆分
		scenes, err := p.splitChapterToScenes(novelName, chapterName, string(chapterText))
		if err != nil {
			errMsg := fmt.Sprintf("Failed to split chapter to scenes: %v", err)
			imageStep.Status = database.StatusFailed
			imageStep.ErrorMsg = errMsg
			imageStep.EndTime = database.MyTime{Time: time.Now()}
			p.dbManager.UpdateProcessStep(imageStep)
			p.dbManager.UpdateChapterProcessStatus(uint(chapterProcess.ID), database.StatusFailed, errMsg)
			return nil, fmt.Errorf(errMsg)
		}

		// 为每个场景生成图像
		for i, scene := range scenes {
			scene.SceneNumber = i + 1
			scene.ChapterID = uint(chapterProcess.ID)
			
			// 获取项目特定的配置
			projectConfig, err := p.getProjectSpecificConfig(novelName)
			if err != nil {
				p.logger.Warn("Failed to get project-specific config, using default", zap.Error(err))
				// 使用默认配置
				defaultConfig, err := p.dbManager.GetOrCreateDefaultDrawthingsConfig()
				if err != nil {
					p.logger.Error("Failed to get default config", zap.Error(err))
					continue
				}
				scene.DrawthingsConfig = *defaultConfig
			} else {
				scene.DrawthingsConfig = *projectConfig
			}
			
			err = p.dbManager.CreateScene(&scene)
			if err != nil {
				p.logger.Error("Failed to create scene record", zap.Error(err))
				continue
			}

			// 生成图像
			sceneImageFile := filepath.Join(chapterDir, fmt.Sprintf("scene_%02d.png", scene.SceneNumber))
			scene.Status = database.StatusProcessing
			scene.StartTime = database.MyTime{Time: time.Now()}
			err = p.dbManager.UpdateScene(&scene)
			if err != nil {
				p.logger.Error("Failed to update scene status to processing", zap.Error(err))
				continue
			}

			// 使用场景的提示词和配置生成图像
			err = p.drawThingsTool.Client.GenerateImageFromText(
				scene.Prompt,
				sceneImageFile,
				scene.DrawthingsConfig.Width,
				scene.DrawthingsConfig.Height,
				false, // 风格已在提示词中处理
			)
			if err != nil {
				errMsg := fmt.Sprintf("Failed to generate image for scene %d: %v", scene.SceneNumber, err)
				scene.Status = database.StatusFailed
				scene.ErrorMsg = errMsg
				scene.EndTime = database.MyTime{Time: time.Now()}
				scene.Duration = int64(time.Since(scene.StartTime.Time) / time.Second)
				p.dbManager.UpdateScene(&scene)
				p.logger.Error(errMsg)
				continue
			}

			scene.Status = database.StatusCompleted
			scene.EndTime = database.MyTime{Time: time.Now()}
			scene.Duration = int64(time.Since(scene.StartTime.Time) / time.Second)
			scene.ImageFile = sceneImageFile
			err = p.dbManager.UpdateScene(&scene)
			if err != nil {
				p.logger.Error("Failed to update scene status to completed", zap.Error(err))
			}
		}

		imageStep.Status = database.StatusCompleted
		imageStep.EndTime = database.MyTime{Time: time.Now()}
		imageStep.Duration = int64(time.Since(imageStep.StartTime.Time) / time.Second)
		err = p.dbManager.UpdateProcessStep(imageStep)
		if err != nil {
			p.logger.Error("Failed to update image step status to completed", zap.Error(err))
		}

		// 获取生成的图像文件
		imageFiles, err := p.fileTool.GetFilesInDir(chapterDir, ".png")
		if err == nil {
			result.ImageFiles = imageFiles
		}
		chapterProcess.ImageGenerated = true
	}

	// 处理步骤3: 字幕生成
	subtitleStep, err := p.dbManager.GetProcessStepByChapterAndName(uint(chapterProcess.ID), "subtitle")
	if err != nil {
		p.logger.Error("Failed to get subtitle step", zap.Error(err))
	}

	if subtitleStep == nil || subtitleStep.Status != database.StatusCompleted {
		err = p.dbManager.UpdateChapterProcessStatus(uint(chapterProcess.ID), database.StatusProcessing, "Generating subtitles")
		if err != nil {
			p.logger.Error("Failed to update chapter process status for subtitle generation", zap.Error(err))
			return nil, err
		}

		// 创建或重置字幕步骤
		subtitleStep = &database.ProcessStep{
			ChapterID: uint(chapterProcess.ID),
			StepName:  "subtitle",
			Status:    database.StatusPending,
		}

		if subtitleStep.ID == 0 {
			err = p.dbManager.CreateProcessStep(subtitleStep)
			if err != nil {
				p.logger.Error("Failed to create subtitle step", zap.Error(err))
				return nil, err
			}
		}

		subtitleStep.Status = database.StatusProcessing
		subtitleStep.StartTime = database.MyTime{Time: time.Now()}
		err = p.dbManager.UpdateProcessStep(subtitleStep)
		if err != nil {
			p.logger.Error("Failed to update subtitle step status to processing", zap.Error(err))
			return nil, err
		}

		// 执行字幕生成
		subtitleFile := filepath.Join(chapterDir, chapterName+".srt")
		audioFile := filepath.Join(chapterDir, chapterName+".wav")

		if _, err := os.Stat(audioFile); err == nil {
			chapterText, err := os.ReadFile(inputFilePath)
			if err != nil {
				errMsg := fmt.Sprintf("Failed to read chapter file: %v", err)
				subtitleStep.Status = database.StatusFailed
				subtitleStep.ErrorMsg = errMsg
				subtitleStep.EndTime = database.MyTime{Time: time.Now()}
				p.dbManager.UpdateProcessStep(subtitleStep)
				p.dbManager.UpdateChapterProcessStatus(uint(chapterProcess.ID), database.StatusFailed, errMsg)
				return nil, fmt.Errorf(errMsg)
			}

			err = p.aegisubTool.ProcessIndextts2OutputWithCustomName(audioFile, string(chapterText), subtitleFile)
			if err != nil {
				errMsg := fmt.Sprintf("Failed to generate subtitles: %v", err)
				subtitleStep.Status = database.StatusFailed
				subtitleStep.ErrorMsg = errMsg
				subtitleStep.EndTime = database.MyTime{Time: time.Now()}
				p.dbManager.UpdateProcessStep(subtitleStep)
				p.dbManager.UpdateChapterProcessStatus(uint(chapterProcess.ID), database.StatusFailed, errMsg)
				return nil, fmt.Errorf(errMsg)
			}

			subtitleStep.Status = database.StatusCompleted
			subtitleStep.EndTime = database.MyTime{Time: time.Now()}
			subtitleStep.Duration = int64(time.Since(subtitleStep.StartTime.Time) / time.Second)
			err = p.dbManager.UpdateProcessStep(subtitleStep)
			if err != nil {
				p.logger.Error("Failed to update subtitle step status to completed", zap.Error(err))
			}

			result.SubtitleFile = subtitleFile
			chapterProcess.SubtitleCreated = true
		} else {
			p.logger.Info("Audio file not found, skipping subtitle generation")
			subtitleStep.Status = database.StatusSkipped
			subtitleStep.ErrorMsg = "Audio file not found"
			subtitleStep.EndTime = database.MyTime{Time: time.Now()}
			subtitleStep.Duration = 0
			err = p.dbManager.UpdateProcessStep(subtitleStep)
			if err != nil {
				p.logger.Error("Failed to update subtitle step status to skipped", zap.Error(err))
			}
		}
	}

	// 处理步骤4: 剪映项目生成
	capcutStep, err := p.dbManager.GetProcessStepByChapterAndName(uint(chapterProcess.ID), "capcut")
	if err != nil {
		p.logger.Error("Failed to get capcut step", zap.Error(err))
	}

	if capcutStep == nil || capcutStep.Status != database.StatusCompleted {
		err = p.dbManager.UpdateChapterProcessStatus(uint(chapterProcess.ID), database.StatusProcessing, "Generating CapCut project")
		if err != nil {
			p.logger.Error("Failed to update chapter process status for capcut generation", zap.Error(err))
			return nil, err
		}

		// 创建或重置剪映项目步骤
		capcutStep = &database.ProcessStep{
			ChapterID: uint(chapterProcess.ID),
			StepName:  "capcut",
			Status:    database.StatusPending,
		}

		if capcutStep.ID == 0 {
			err = p.dbManager.CreateProcessStep(capcutStep)
			if err != nil {
				p.logger.Error("Failed to create capcut step", zap.Error(err))
				return nil, err
			}
		}

		capcutStep.Status = database.StatusProcessing
		capcutStep.StartTime = database.MyTime{Time: time.Now()}
		err = p.dbManager.UpdateProcessStep(capcutStep)
		if err != nil {
			p.logger.Error("Failed to update capcut step status to processing", zap.Error(err))
			return nil, err
		}

		// 执行剪映项目生成
		err = p.GenerateCapcutProject(chapterDir)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to generate CapCut project: %v", err)
			capcutStep.Status = database.StatusFailed
			capcutStep.ErrorMsg = errMsg
			capcutStep.EndTime = database.MyTime{Time: time.Now()}
			p.dbManager.UpdateProcessStep(capcutStep)
			p.dbManager.UpdateChapterProcessStatus(uint(chapterProcess.ID), database.StatusFailed, errMsg)
			return nil, fmt.Errorf(errMsg)
		}

		capcutStep.Status = database.StatusCompleted
		capcutStep.EndTime = database.MyTime{Time: time.Now()}
		capcutStep.Duration = int64(time.Since(capcutStep.StartTime.Time) / time.Second)
		err = p.dbManager.UpdateProcessStep(capcutStep)
		if err != nil {
			p.logger.Error("Failed to update capcut step status to completed", zap.Error(err))
		}

		result.VideoProject = filepath.Join(chapterDir, chapterName+".json")
		chapterProcess.CapcutCreated = true
	}

	// 计算总耗时
	endTime := time.Now()
	chapterProcess.EndTime = database.MyTime{Time: endTime}
	startTime := chapterProcess.CreatedAt.Time
	chapterProcess.Duration = int64(endTime.Sub(startTime) / time.Second)
	
	// 更新章节进度
	chapterProcess.Status = database.StatusCompleted
	err = p.dbManager.UpdateChapterProcess(chapterProcess)
	if err != nil {
		p.logger.Error("Failed to update chapter process status to completed", zap.Error(err))
		return nil, err
	}

	result.Status = string(database.StatusCompleted)
	result.Message = "Chapter processing completed successfully"

	p.logger.Info("Chapter processing completed successfully", zap.String("novel", novelName), zap.String("chapter", chapterName))
	return result, nil
}

// splitChapterToScenes 将章节拆分为多个场景
func (p *Processor) splitChapterToScenes(novelName, chapterName, chapterText string) ([]database.Scene, error) {
	// 获取项目信息以了解小说类型和氛围
	project, err := p.dbManager.GetProjectByName(novelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %v", err)
	}
	
	var scenes []database.Scene
	
	// 使用AI进行场景拆分 - 这里可以使用Ollama进行场景分析
	// 为了演示，我们先实现一个简单的场景拆分逻辑
	// 在实际实现中，这里应该调用Ollama API进行智能拆分
	
	// 简单的实现：将文本按段落拆分，每个段落作为一个场景
	paragraphs := p.splitTextToParagraphs(chapterText)
	
	for i, paragraph := range paragraphs {
		if len(paragraph) < 10 { // 忽略太短的段落
			continue
		}
		
		// 为每个段落生成提示词，结合项目氛围信息
		prompt := p.generateScenePrompt(paragraph, project.Atmosphere, project.Genre)
		
		scene := database.Scene{
			SceneNumber:   i + 1,
			Description:   paragraph,
			Prompt:        prompt,
			IsAIgenerated: true,
			Status:        database.StatusPending,
		}
		
		scenes = append(scenes, scene)
	}
	
	return scenes, nil
}

// splitTextToParagraphs 将文本拆分为段落
func (p *Processor) splitTextToParagraphs(text string) []string {
	// 按双换行符拆分段落
	paragraphs := []string{}
	
	// 按换行符分割
	lines := []string{}
	currentLine := ""
	
	for _, char := range text {
		if char == '\n' {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = ""
			} else {
				// 连续的换行符，表示段落分隔
				if len(lines) > 0 && lines[len(lines)-1] != "" {
					lines = append(lines, "") // 添加空行作为段落分隔
				}
			}
		} else {
			currentLine += string(char)
		}
	}
	
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	
	// 合并连续的非空行形成段落
	currentParagraph := ""
	for _, line := range lines {
		if line == "" {
			if currentParagraph != "" {
				paragraphs = append(paragraphs, currentParagraph)
				currentParagraph = ""
			}
		} else {
			if currentParagraph != "" {
				currentParagraph += " " + line
			} else {
				currentParagraph = line
			}
		}
	}
	
	if currentParagraph != "" {
		paragraphs = append(paragraphs, currentParagraph)
	}
	
	return paragraphs
}

// generateScenePrompt 生成场景提示词
func (p *Processor) generateScenePrompt(description, atmosphere, genre string) string {
	// 根据项目氛围和类型生成提示词
	basePrompt := description
	
	if atmosphere != "" {
		basePrompt = fmt.Sprintf("%s, %s氛围", basePrompt, atmosphere)
	}
	
	if genre != "" {
		switch genre {
		case "悬疑":
			basePrompt = fmt.Sprintf("%s, 悬疑惊悚风格，周围环境模糊成黑影, 空气凝滞,浅景深, 胶片颗粒感, 低饱和度，极致悬疑氛围, 阴沉窒息感", basePrompt)
		case "科幻":
			basePrompt = fmt.Sprintf("%s, 科幻风格，未来主义元素，霓虹灯光，金属质感", basePrompt)
		case "言情":
			basePrompt = fmt.Sprintf("%s, 浪漫温馨风格，柔和光线，温暖色调", basePrompt)
		case "武侠":
			basePrompt = fmt.Sprintf("%s, 武侠风格，古风建筑，江湖气息", basePrompt)
		default:
			basePrompt = fmt.Sprintf("%s, 高质量图像，细节丰富", basePrompt)
		}
	}
	
	return basePrompt
}

// getProjectSpecificConfig 获取项目特定的配置
func (p *Processor) getProjectSpecificConfig(novelName string) (*database.DrawthingsConfig, error) {
	// 获取项目信息以确定特定配置
	project, err := p.dbManager.GetProjectByName(novelName)
	if err != nil {
		return nil, err
	}
	
	if project == nil {
		return nil, fmt.Errorf("project not found: %s", novelName)
	}
	
	// 根据项目类型返回特定配置
	config := &database.DrawthingsConfig{
		Width:           512,
		Height:          896,
		Steps:           20,
		CFGScale:        7.0,
		SamplerName:     "Euler a",
		Seed:            -1,
		BatchSize:       1,
		LoraModel:       "",
		LoraTriggerWord: "",
		LoraWeight:      0.8,
		NegativePrompt:  "low quality, worst quality, deformed, distorted",
	}
	
	// 根据小说类型调整配置
	switch project.Genre {
	case "悬疑":
		config.NegativePrompt = "bright, cheerful, cartoon, anime, low quality, worst quality, deformed, distorted"
	case "科幻":
		config.NegativePrompt = "historical, ancient, medieval, low quality, worst quality, deformed, distorted"
	case "言情":
		config.NegativePrompt = "horror, scary, dark, low quality, worst quality, deformed, distorted"
	case "武侠":
		config.NegativePrompt = "modern, futuristic, sci-fi, low quality, worst quality, deformed, distorted"
	}
	
	return config, nil
}

// GetChapterProcess 获取章节处理记录
func (p *Processor) GetChapterProcess(novelName, chapterName string) (*database.ChapterProcess, error) {
	project, err := p.dbManager.GetProjectByName(novelName)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, fmt.Errorf("project not found: %s", novelName)
	}
	
	return p.dbManager.GetChapterProcessByProjectAndName(uint(project.ID), chapterName)
}

// GetChapterProcesses 获取小说的所有章节处理记录
func (p *Processor) GetChapterProcesses(novelName string) ([]database.ChapterProcess, error) {
	project, err := p.dbManager.GetProjectByName(novelName)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, fmt.Errorf("project not found: %s", novelName)
	}
	
	return p.dbManager.GetChaptersByProjectID(uint(project.ID))
}

// RetryChapterProcess 重试章节处理
func (p *Processor) RetryChapterProcess(novelName, chapterName string) error {
	project, err := p.dbManager.GetProjectByName(novelName)
	if err != nil {
		return err
	}
	if project == nil {
		return fmt.Errorf("project not found: %s - %s", novelName, chapterName)
	}
	
	chapterProcess, err := p.dbManager.GetChapterProcessByProjectAndName(uint(project.ID), chapterName)
	if err != nil {
		return err
	}
	if chapterProcess == nil {
		return fmt.Errorf("chapter process not found: %s - %s", novelName, chapterName)
	}
	
	return p.dbManager.RetryChapterProcess(uint(chapterProcess.ID))
}

// RetryStepForChapter 重试特定步骤
func (p *Processor) RetryStepForChapter(novelName, chapterName, stepName string) error {
	project, err := p.dbManager.GetProjectByName(novelName)
	if err != nil {
		return err
	}
	if project == nil {
		return fmt.Errorf("project not found: %s - %s", novelName, chapterName)
	}
	
	chapterProcess, err := p.dbManager.GetChapterProcessByProjectAndName(uint(project.ID), chapterName)
	if err != nil {
		return err
	}
	if chapterProcess == nil {
		return fmt.Errorf("chapter process not found: %s - %s", novelName, chapterName)
	}
	
	return p.dbManager.ResetStepForRetry(uint(chapterProcess.ID), stepName)
}

// GetProject 获取项目信息
func (p *Processor) GetProject(novelName string) (*database.Project, error) {
	return p.dbManager.GetProjectByName(novelName)
}

// GetProjects 获取所有项目
func (p *Processor) GetProjects() ([]database.Project, error) {
	var projects []database.Project
	result := p.dbManager.GetDB().Find(&projects)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get projects: %v", result.Error)
	}
	return projects, nil
}

// GetScenesByChapter 获取章节的所有场景
func (p *Processor) GetScenesByChapter(novelName, chapterName string) ([]database.Scene, error) {
	chapter, err := p.GetChapterProcess(novelName, chapterName)
	if err != nil {
		return nil, err
	}
	
	return p.dbManager.GetScenesByChapterID(uint(chapter.ID))
}

// UpdateScenePrompt 更新场景提示词
func (p *Processor) UpdateScenePrompt(sceneID uint, newPrompt string) error {
	scene, err := p.dbManager.GetSceneByID(sceneID)
	if err != nil {
		return err
	}
	
	if scene == nil {
		return fmt.Errorf("scene not found: %d", sceneID)
	}
	
	scene.Prompt = newPrompt
	scene.IsAIgenerated = false // 标记为用户自定义
	return p.dbManager.UpdateScene(scene)
}

// UpdateDrawthingsConfig 更新Drawthings配置
func (p *Processor) UpdateDrawthingsConfig(sceneID uint, config *database.DrawthingsConfig) error {
	scene, err := p.dbManager.GetSceneByID(sceneID)
	if err != nil {
		return err
	}
	
	if scene == nil {
		return fmt.Errorf("scene not found: %d", sceneID)
	}
	
	scene.DrawthingsConfig = *config
	return p.dbManager.UpdateScene(scene)
}

// Getter methods to access private fields
func (p *Processor) GetAegisubTool() *aegisub.AegisubIntegration {
	return p.aegisubTool
}

func (p *Processor) GetDrawThingsTool() *drawthings.ChapterImageGenerator {
	return p.drawThingsTool
}

func (p *Processor) GetDbManager() *database.GormManager {
	return p.dbManager
}
