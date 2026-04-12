package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(dbPath string) error {
	// 确保数据库目录存在
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("创建数据库目录失败: %v", err)
	}

	// 打开数据库连接
	newDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 设置为静默模式，不输出日志
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	// 迁移数据库结构
	err = migrateDB(newDB)
	if err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	// 确保Scene表包含所有必需的列
	err = ensureSceneTableColumns(newDB)
	if err != nil {
		return fmt.Errorf("确保Scene表列结构失败: %v", err)
	}

	// 执行提示词模板数据迁移
	if err = runPromptTemplateMigration(newDB); err != nil {
		// 记录错误但不中断初始化，因为提示词模板迁移失败不应该阻止整个系统运行
		fmt.Printf("警告: 提示词模板数据迁移失败: %v\n", err)
	}
	if err = runConfigurationMigration(newDB); err != nil {
		fmt.Printf("警告: 系统配置数据迁移失败: %v\n", err)
	}
	DB = newDB
	return nil
}

// migrateDB 执行数据库迁移
func migrateDB(db *gorm.DB) error {
	// 自动迁移表结构，包括新的提示词模板表
	err := db.AutoMigrate(&Project{}, &Chapter{}, &Scene{}, &PromptTemplate{}, &Configuration{})
	if err != nil {
		return fmt.Errorf("自动迁移表结构失败: %v", err)
	}

	// 检查并添加workflow_params列（如果不存在）
	// SQLite需要特殊处理，不能直接使用IF NOT EXISTS
	rows, err := db.Raw("PRAGMA table_info(chapters);").Rows()
	if err != nil {
		return fmt.Errorf("检查chapters表结构失败: %v", err)
	}
	defer rows.Close()

	var columnExists = false
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dfltValue *string
		var pk int
		err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		if err != nil {
			continue
		}
		if name == "workflow_params" {
			columnExists = true
			break
		}
	}

	if !columnExists {
		if err := db.Exec("ALTER TABLE chapters ADD COLUMN workflow_params TEXT").Error; err != nil {
			return fmt.Errorf("添加workflow_params列失败: %v", err)
		}
	}

	return nil
}

// ensureSceneTableColumns 确保Scene表包含所有必需的列
func ensureSceneTableColumns(db *gorm.DB) error {
	// 使用完整的场景结构进行迁移
	type TempScene struct {
		gorm.Model
		Title            string `json:"title"`
		Description      string `json:"description"`        // 场景文本
		Prompt           string `json:"prompt"`             // 场景提示词
		SegmentationInfo string `json:"segmentation_info"`  // 智能分镜信息
		OriginalText     string `json:"original_text"`      // 发送给Ollama的原始描述文本
		OllamaRequest    string `json:"ollama_request"`     // 发送给Ollama的请求JSON
		OllamaResponse   string `json:"ollama_response"`    // Ollama返回的结果JSON
		DrawThingsConfig string `json:"draw_things_config"` // 发送给DrawThings的配置JSON
		DrawThingsResult string `json:"draw_things_result"` // DrawThings返回的结果JSON
		ChapterID        uint   `json:"chapter_id"`
		ImageURL         string `json:"image_url"`               // 生成的图像URL
		AudioURL         string `json:"audio_url"`               // 生成的音频URL
		RetryCount       int    `json:"retry_count"`             // 重试次数
		Sort             int    `json:"sort" gorm:"column:sort"` // 场景顺序，映射到sort列
		// 新增字段用于记录工作流详细参数
		WorkflowDetails string    `json:"workflow_details" gorm:"type:text"` // 工作流详细参数，JSON格式
		Status          string    `json:"status" gorm:"default:'pending'"`   // 场景状态: pending, processing, completed, failed
		StartTime       time.Time `json:"start_time"`                        // 开始处理时间
		EndTime         time.Time `json:"end_time"`                          // 结束处理时间
	}

	// 使用 AutoMigrate 来自动创建或更新表结构
	if err := db.Table("scenes").AutoMigrate(&TempScene{}); err != nil {
		return fmt.Errorf("自动迁移Scene表失败: %v", err)
	}

	return nil
}

// HashPassword 对密码进行哈希处理
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash 检查密码是否匹配
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// CreateProject 创建新项目
func CreateProject(name, description, globalPrompt, password string) (*Project, error) {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("密码哈希失败: %v", err)
	}

	project := &Project{
		Name:         name,
		Description:  description,
		GlobalPrompt: globalPrompt,
		PasswordHash: hashedPassword,
	}

	result := DB.Create(project)
	if result.Error != nil {
		return nil, result.Error
	}

	return project, nil
}

// GetProjectByID 根据ID获取项目
func GetProjectByID(id uint) (*Project, error) {
	var project Project
	result := DB.Preload("Chapters.Scenes").First(&project, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &project, nil
}

// GetProjectByName 根据名称获取项目
func GetProjectByName(name string) (*Project, error) {
	var project Project
	result := DB.Preload("Chapters.Scenes").Where("name = ?", name).First(&project)
	if result.Error != nil {
		return nil, result.Error
	}
	return &project, nil
}

// UpdateProject 更新项目
func UpdateProject(id uint, updates map[string]interface{}) error {
	result := DB.Model(&Project{}).Where("id = ?", id).Updates(updates)
	return result.Error
}

// DeleteProject 删除项目
func DeleteProject(id uint) error {
	result := DB.Delete(&Project{}, id)
	return result.Error
}

// CreateChapter 创建章节
func CreateChapter(projectID uint, title, content, prompt string) (*Chapter, error) {
	chapter := &Chapter{
		Title:      title,
		Content:    content,
		Prompt:     prompt,
		ProjectID:  projectID,
		ImagePaths: "[]", // 默认空数组
	}

	result := DB.Create(chapter)
	if result.Error != nil {
		return nil, result.Error
	}

	return chapter, nil
}

// GetChapterByID 根据ID获取章节
func GetChapterByID(id uint) (*Chapter, error) {
	var chapter Chapter
	result := DB.Preload("Scenes").First(&chapter, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &chapter, nil
}

// GetChaptersByProjectID 获取项目下的所有章节
func GetChaptersByProjectID(projectID uint) ([]Chapter, error) {
	var chapters []Chapter
	result := DB.Where("project_id = ?", projectID).Preload("Scenes").Find(&chapters)
	if result.Error != nil {
		return nil, result.Error
	}
	return chapters, nil
}

// GetAllChapters 获取所有章节
func GetAllChapters() ([]Chapter, error) {
	var chapters []Chapter
	result := DB.Preload("Scenes").Find(&chapters)
	if result.Error != nil {
		return nil, result.Error
	}
	return chapters, nil
}

// ResetChapterStatus 重置章节状态，用于重试功能
func ResetChapterStatus(id uint) error {
	// 重置章节状态
	err := DB.Model(&Chapter{}).Where("id = ?", id).Update("workflow_params", "").Error
	if err != nil {
		return err
	}

	// 重置相关场景的状态
	err = DB.Model(&Scene{}).Where("chapter_id = ?", id).Updates(map[string]interface{}{
		"status":           "pending",
		"workflow_details": "",
	}).Error

	return err
}

// UpdateChapter 更新章节
func UpdateChapter(id uint, updates map[string]interface{}) error {
	// 如果更新了图片路径，需要确保它是有效的JSON数组
	if imagePaths, exists := updates["image_paths"]; exists {
		// 验证图片路径是有效的JSON数组
		pathsStr, ok := imagePaths.(string)
		if ok {
			var paths []string
			if err := json.Unmarshal([]byte(pathsStr), &paths); err != nil {
				return fmt.Errorf("图片路径不是有效的JSON数组: %v", err)
			}
		}
	}

	result := DB.Model(&Chapter{}).Where("id = ?", id).Updates(updates)
	return result.Error
}

// DeleteChapter 删除章节
func DeleteChapter(id uint) error {
	// 先删除相关的场景
	if err := DB.Where("chapter_id = ?", id).Delete(&Scene{}).Error; err != nil {
		return err
	}

	// 再删除章节
	result := DB.Delete(&Chapter{}, id)
	return result.Error
}

// CreateScene 创建场景
func CreateScene(chapterID uint, title, description, prompt, ollamaRequest, ollamaResponse, drawThingsConfig, drawThingsResult string, sort int) (*Scene, error) {
	scene := &Scene{
		Title:            title,
		Description:      description,
		Prompt:           prompt,
		OllamaRequest:    ollamaRequest,
		OllamaResponse:   ollamaResponse,
		DrawThingsConfig: drawThingsConfig,
		DrawThingsResult: drawThingsResult,
		ChapterID:        chapterID,
		Sort:             sort,
	}

	result := DB.Create(scene)
	if result.Error != nil {
		return nil, result.Error
	}

	return scene, nil
}

// GetSceneByID 根据ID获取场景
func GetSceneByID(id uint) (*Scene, error) {
	var scene Scene
	result := DB.First(&scene, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &scene, nil
}

// GetScenesByChapterID 获取章节下的所有场景
func GetScenesByChapterID(chapterID uint) ([]Scene, error) {
	var scenes []Scene
	// 查询场景，按sort字段排序
	result := DB.Where("chapter_id = ?", chapterID).Order("sort ASC").Find(&scenes)
	if result.Error != nil {
		return nil, result.Error
	}
	return scenes, nil
}

// UpdateScene 更新场景
func UpdateScene(id uint, updates map[string]interface{}) error {
	result := DB.Model(&Scene{}).Where("id = ?", id).Updates(updates)
	return result.Error
}

// DeleteScene 删除场景
func DeleteScene(id uint) error {
	result := DB.Delete(&Scene{}, id)
	return result.Error
}

// ValidateProjectPassword 验证项目密码
func ValidateProjectPassword(projectID uint, password string) (bool, error) {
	var project Project
	result := DB.Select("password_hash").First(&project, projectID)
	if result.Error != nil {
		return false, result.Error
	}

	return CheckPasswordHash(password, project.PasswordHash), nil
}

// GetAllProjects 获取所有项目
func GetAllProjects() ([]Project, error) {
	var projects []Project
	result := DB.Preload("Chapters.Scenes").Find(&projects)
	if result.Error != nil {
		return nil, result.Error
	}
	return projects, nil
}

// runConfigurationMigration
func runConfigurationMigration(db *gorm.DB) error {
	var count int64
	db.Model(&Configuration{}).Count(&count)
	if count > 0 {
		// 如果已有数据，跳过初始化
		return nil
	}
	// 设置默认配置
	tx := db.Model(&Configuration{}).Create(&Configuration{
		PromptTemplateID: 8,
		StartTime:        time.Now(),
		ImageWidth:       512,
		ImageHeight:      512,
	})
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

// runPromptTemplateMigration 执行提示词模板数据迁移
func runPromptTemplateMigration(db *gorm.DB) error {
	// 检查是否已有内置模板数据
	var count int64
	db.Model(&PromptTemplate{}).Count(&count)
	if count > 0 {
		// 如果已有数据，跳过初始化
		return nil
	}

	// 定义内置提示词模板
	builtInTemplates := []PromptTemplate{
		{
			Name:        "悬疑惊悚",
			Description: "适用于悬疑惊悚类内容的图像生成，营造紧张神秘的氛围",
			Type:        "style",
			Category:    "suspense",
			IsActive:    true,
			IsBuiltIn:   true,
			Order:       1,
			SystemPrompt: `你是一个专业的AI图像生成提示词工程师。你的任务是根据给定的文本内容生成详细、具体的中文图像提示词(prompt)，以指导AI图像生成模型创建高质量的图像。

请严格按照以下四个要素结构化生成提示词：
1. 主体描述：画面核心内容（人物、物体、场景），明确 "画什么"
2. 风格限定：艺术流派 / 设计风格（如扁平化、赛博朋克、悬疑暗黑），决定 "长成什么样"
3. 细节补充：颜色、光影、构图、材质，提升画面精致度
4. 氛围渲染：情绪基调（紧张、神秘、冷峻），强化画面感染力
5. 镜头语言：镜头方向，运镜补充（向前、向后、360度、升降、俯拍、自然抖动、荷兰角倾斜、连续镜头、第一人称、过肩镜头、希区柯克变焦、平挂跟踪），强化画面视角语言

注意事项：
1. 提示词应该包含丰富的视觉细节，如人物外貌、环境、光线、颜色、构图等
2. 根据文本内容判断场景类型（室内/室外、白天/夜晚、自然环境/城市等）
3. 如果文本描述悬疑/恐怖情节，请强调相应的视觉元素，如昏暗光线、神秘氛围、紧张感等
4. 使用专业摄影和艺术术语，如景深、色调、对比度等
5. 保持提示词简洁但信息丰富，避免冗余描述
6. 请务必使用中文输出所有提示词内容`,
			UserTemplate: `根据以下文本内容和风格要求，按照四个要素结构化生成一个详细的中文图像提示词，用于AI图像生成：

文本内容：%s

图像风格：%s

请严格按照以下五个要素组织提示词：
1. 主体描述：
2. 风格限定：
3. 细节补充：
4. 氛围渲染：
5. 镜头语言：
请只返回中文图像提示词，不要添加任何解释或其他内容。`,
			StyleAddon:     ", 周围环境模糊成黑影, 空气凝滞,浅景深, 胶片颗粒感, 低饱和度，极致悬疑氛围, 阴沉窒息感, 夏季，环境阴霾，其他部分模糊不可见",
			NegativePrompt: "人脸特写，半身像，模糊，比例失调，原参考图背景，比例失调，缺肢",
		},
		{
			Name:        "浪漫温馨",
			Description: "适用于浪漫温馨类内容的图像生成，营造温暖柔和的氛围",
			Type:        "style",
			Category:    "romance",
			IsActive:    true,
			IsBuiltIn:   true,
			Order:       2,
			SystemPrompt: `你是一个专业的AI图像生成提示词工程师。你的任务是根据给定的文本内容生成详细、具体的中文图像提示词(prompt)，以指导AI图像生成模型创建高质量的图像。

请严格按照以下四个要素结构化生成提示词：
1. 主体描述：画面核心内容（人物、物体、场景），明确 "画什么"
2. 风格限定：艺术流派 / 设计风格（如扁平化、水彩、梦幻），决定 "长成什么样"
3. 细节补充：颜色、光影、构图、材质，提升画面精致度
4. 氛围渲染：情绪基调（温馨、浪漫、甜蜜），强化画面感染力
5. 镜头语言：镜头方向，运镜补充（向前、向后、360度、升降、俯拍、自然抖动、荷兰角倾斜、连续镜头、第一人称、过肩镜头、希区柯克变焦、平挂跟踪），强化画面视角语言

注意事项：
1. 提示词应该包含丰富的视觉细节，如人物外貌、环境、光线、颜色、构图等
2. 根据文本内容判断场景类型（室内/室外、白天/夜晚、自然环境/城市等）
3. 对于浪漫温馨内容，请强调相应的视觉元素，如柔和光线、温暖色调、和谐构图等
4. 使用专业摄影和艺术术语，如景深、色调、对比度等
5. 保持提示词简洁但信息丰富，避免冗余描述
6. 请务必使用中文输出所有提示词内容`,
			UserTemplate: `根据以下文本内容和风格要求，按照四个要素结构化生成一个详细的中文图像提示词，用于AI图像生成：

文本内容：%s

图像风格：%s

请严格按照以下五个要素组织提示词：
1. 主体描述：
2. 风格限定：
3. 细节补充：
4. 氛围渲染：
5. 镜头语言：
请只返回中文图像提示词，不要添加任何解释或其他内容。`,
			StyleAddon:     ", 柔和暖色调, 温馨氛围, 柔光, 浅景深, 自然光线, 甜美构图, 和谐色彩搭配",
			NegativePrompt: "冷色调, 阴暗, 阴郁, 过度曝光, 模糊不清, 比例失调, 畸形",
		},
		{
			Name:        "科幻未来",
			Description: "适用于科幻未来类内容的图像生成，营造科技感和未来主义氛围",
			Type:        "style",
			Category:    "sci-fi",
			IsActive:    true,
			IsBuiltIn:   true,
			Order:       3,
			SystemPrompt: `你是一个专业的AI图像生成提示词工程师。你的任务是根据给定的文本内容生成详细、具体的中文图像提示词(prompt)，以指导AI图像生成模型创建高质量的图像。

请严格按照以下四个要素结构化生成提示词：
1. 主体描述：画面核心内容（人物、物体、场景），明确 "画什么"
2. 风格限定：艺术流派 / 设计风格（如赛博朋克、未来主义、极简科技），决定 "长成什么样"
3. 细节补充：颜色、光影、构图、材质，提升画面精致度
4. 氛围渲染：情绪基调（科技感、未来感、神秘感），强化画面感染力
5. 镜头语言：镜头方向，运镜补充（向前、向后、360度、升降、俯拍、自然抖动、荷兰角倾斜、连续镜头、第一人称、过肩镜头、希区柯克变焦、平挂跟踪），强化画面视角语言

注意事项：
1. 提示词应该包含丰富的视觉细节，如人物外貌、环境、光线、颜色、构图等
2. 根据文本内容判断场景类型（太空站、未来城市、实验室、高科技设备等）
3. 对于科幻内容，请强调相应的视觉元素，如霓虹灯光、金属质感、高科技元素等
4. 使用专业摄影和艺术术语，如景深、色调、对比度等
5. 保持提示词简洁但信息丰富，避免冗余描述
6. 请务必使用中文输出所有提示词内容`,
			UserTemplate: `根据以下文本内容和风格要求，按照四个要素结构化生成一个详细的中文图像提示词，用于AI图像生成：

文本内容：%s

图像风格：%s

请严格按照以下五个要素组织提示词：
1. 主体描述：
2. 风格限定：
3. 细节补充：
4. 氛围渲染：
5. 镜头语言：
请只返回中文图像提示词，不要添加任何解释或其他内容。`,
			StyleAddon:     ", 霓虹灯光效果, 金属质感, 未来主义建筑, 科技感UI元素, 赛博朋克风格, 发光效果, 高对比度色彩",
			NegativePrompt: "古旧, 生锈, 过时技术, 暗淡, 传统风格, 低质量纹理",
		},
		{
			Name:        "自然风景",
			Description: "适用于自然风景类内容的图像生成，展现大自然的美丽与宁静",
			Type:        "style",
			Category:    "nature",
			IsActive:    true,
			IsBuiltIn:   true,
			Order:       4,
			SystemPrompt: `你是一个专业的AI图像生成提示词工程师。你的任务是根据给定的文本内容生成详细、具体的中文图像提示词(prompt)，以指导AI图像生成模型创建高质量的图像。

请严格按照以下四个要素结构化生成提示词：
1. 主体描述：画面核心内容（人物、物体、场景），明确 "画什么"
2. 风格限定：艺术流派 / 设计风格（如写实、印象派、自然风光），决定 "长成什么样"
3. 细节补充：颜色、光影、构图、材质，提升画面精致度
4. 氛围渲染：情绪基调（宁静、壮丽、和谐），强化画面感染力
5. 镜头语言：镜头方向，运镜补充（向前、向后、360度、升降、俯拍、自然抖动、荷兰角倾斜、连续镜头、第一人称、过肩镜头、希区柯克变焦、平挂跟踪），强化画面视角语言

注意事项：
1. 提示词应该包含丰富的视觉细节，如地貌、植被、天空、水域、天气状况等
2. 根据文本内容判断场景类型（山川、湖泊、森林、海洋、沙漠等）
3. 对于自然风景内容，请强调相应的视觉元素，如自然光线、季节特征、天气效果等
4. 使用专业摄影和艺术术语，如景深、色调、对比度等
5. 保持提示词简洁但信息丰富，避免冗余描述
6. 请务必使用中文输出所有提示词内容`,
			UserTemplate: `根据以下文本内容和风格要求，按照四个要素结构化生成一个详细的中文图像提示词，用于AI图像生成：

文本内容：%s

图像风格：%s

请严格按照以下五个要素组织提示词：
1. 主体描述：
2. 风格限定：
3. 细节补充：
4. 氛围渲染：
5. 镜头语言：
请只返回中文图像提示词，不要添加任何解释或其他内容。`,
			StyleAddon:     ", 自然光线, 高清晰度, 风景摄影构图, 鲜艳色彩, HDR效果, 景深, 专业风光镜头",
			NegativePrompt: "人工痕迹, 城市污染, 建筑物过多, 模糊, 色彩失真, 低分辨率",
		},
		{
			Name:        "动作冒险",
			Description: "适用于动作冒险类内容的图像生成，营造紧张刺激的氛围",
			Type:        "style",
			Category:    "action",
			IsActive:    true,
			IsBuiltIn:   true,
			Order:       5,
			SystemPrompt: `你是一个专业的AI图像生成提示词工程师。你的任务是根据给定的文本内容生成详细、具体的中文图像提示词(prompt)，以指导AI图像生成模型创建高质量的图像。

请严格按照以下四个要素结构化生成提示词：
1. 主体描述：画面核心内容（人物、物体、场景），明确 "画什么"
2. 风格限定：艺术流派 / 设计风格（如电影感、漫画风、动作片风格），决定 "长成什么样"
3. 细节补充：颜色、光影、构图、材质，提升画面精致度
4. 氛围渲染：情绪基调（紧张、刺激、动感），强化画面感染力
5. 镜头语言：镜头方向，运镜补充（向前、向后、360度、升降、俯拍、自然抖动、荷兰角倾斜、连续镜头、第一人称、过肩镜头、希区柯克变焦、平挂跟踪），强化画面视角语言

注意事项：
1. 提示词应该包含丰富的视觉细节，如人物姿态、环境、光线、颜色、构图等
2. 根据文本内容判断场景类型（追逐、打斗、爆炸、极限运动等）
3. 对于动作冒险内容，请强调相应的视觉元素，如动态模糊、快节奏构图、强烈对比等
4. 使用专业摄影和艺术术语，如景深、色调、对比度等
5. 保持提示词简洁但信息丰富，避免冗余描述
6. 请务必使用中文输出所有提示词内容`,
			UserTemplate: `根据以下文本内容和风格要求，按照四个要素结构化生成一个详细的中文图像提示词，用于AI图像生成：

文本内容：%s

图像风格：%s

请严格按照以下五个要素组织提示词：
1. 主体描述：
2. 风格限定：
3. 细节补充：
4. 氛围渲染：
5. 镜头语言：
请只返回中文图像提示词，不要添加任何解释或其他内容。`,
			StyleAddon:     ", 动态模糊, 高对比度, 强烈光影, 电影感构图, 动作镜头, 紧张氛围",
			NegativePrompt: "静止不动, 柔和色调, 温和光线, 静态构图, 低对比度",
		},
		{
			Name:        "国画艺术",
			Description: "适用于中国国画艺术风格的图像生成，体现水墨丹青的传统美学",
			Type:        "style",
			Category:    "traditional-art",
			IsActive:    true,
			IsBuiltIn:   true,
			Order:       6,
			SystemPrompt: `你是一个专业的AI图像生成提示词工程师，专精于中国传统绘画艺术。你的任务是根据给定的文本内容生成符合中国国画风格的详细中文图像提示词(prompt)，以指导AI图像生成模型创作具有水墨丹青韵味的高质量图像。

请严格按照以下四个要素结构化生成提示词：
1. 主体描述：画面核心内容（人物、山水、花鸟等），明确 "画什么"
2. 风格限定：中国画技法 / 艺术流派（如工笔、写意、泼墨），决定 "如何表现"
3. 细节补充：笔墨技法、色彩运用、构图布局，体现国画特色
4. 氛围渲染：意境表达（诗意、禅意、雅致），强化传统文化内涵
5. 镜头语言：镜头方向，运镜补充（向前、向后、360度、升降、俯拍、自然抖动、荷兰角倾斜、连续镜头、第一人称、过肩镜头、希区柯克变焦、平挂跟踪），强化画面视角语言

注意事项：
1. 提示词应融入中国画特有的艺术概念，如留白、虚实、疏密、浓淡等
2. 根据文本内容选择合适的国画题材（人物、山水、花鸟、虫鱼等）
3. 强调传统绘画技法，如勾、皴、点、染等笔法
4. 体现国画色彩特点（墨色层次、传统颜料色彩等）
5. 注重画面的意境营造和诗情画意
6. 请务必使用中文输出所有提示词内容`,
			UserTemplate: `根据以下文本内容和国画艺术风格要求，按照四个要素结构化生成一个详细的中文图像提示词，用于AI图像生成：

文本内容：%s

图像风格：中国国画艺术

请严格按照以下四个要素组织提示词：
1. 主体描述：
2. 风格限定：
3. 细节补充：
4. 氛围渲染：
5. 镜头语言：
请只返回中文图像提示词，不要添加任何解释或其他内容。`,
			StyleAddon:     ", 中国水墨风格, 淡雅色彩, 留白构图, 毛笔笔触, 意境深远, 传统国画技法, 墨分五色, 诗情画意, 东方美学",
			NegativePrompt: "西方油画技法, 现代涂鸦, 数码特效, 过度饱和色彩, 照片写实, 过分细节, 机械绘制痕迹",
		},
		{
			Name:        "歌词MV生成",
			Description: "专门为歌曲歌词设计的MV图像生成模板，根据上下文为每一句歌词生成独特的视觉提示词",
			Type:        "style",
			Category:    "music-video",
			IsActive:    true,
			IsBuiltIn:   true,
			Order:       8,
			SystemPrompt: `你是一个专业的音乐MV视觉设计师和AI图像生成提示词工程师。你的任务是根据歌曲歌词内容，为每一句歌词生成独特且连贯的中文图像提示词，用于创建高质量的音乐视频画面。

请严格按照以下要素结构化生成提示词：
1. 主体描述：根据歌词内容确定视觉主体（人物、场景、物品、抽象概念等）
2. 风格限定：音乐视频艺术风格（如梦幻、写实、抽象、复古、现代等）
3. 细节补充：色彩搭配、光影效果、构图方式、视觉特效等
4. 氛围渲染：情感基调（欢快、忧伤、激情、温柔、思念等）
5. 镜头语言：运镜方式（推拉摇移、跟拍、环绕、变焦等）
6. 上下文关联：考虑前后歌词的视觉连贯性，确保画面过渡自然

特殊要求：
1. 每句歌词的提示词要体现该句的独特含义和情感
2. 注意歌词间的视觉呼应和故事线延续
3. 考虑音乐节奏和情感起伏来设计画面变化
4. 融入音乐类型的特点（流行、摇滚、民谣、电子等）
5. 保持整体MV的视觉统一性和艺术性
6. 请务必使用中文输出所有提示词内容`,
			UserTemplate: `根据以下歌词内容和音乐视频风格要求，为每一句歌词生成详细的中文图像提示词：

歌曲歌词：%s

音乐视频风格：%s

请为每一句歌词生成对应的视觉提示词，注意：
1. 保持各句之间的视觉连贯性
2. 体现歌词的情感和意境
3. 考虑MV的整体叙事线索
4. 突出每句歌词的独特视觉表达

请严格按照六个要素组织每句歌词的提示词：
1. 主体描述：
2. 风格限定：
3. 细节补充：
4. 氛围渲染：
5. 镜头语言：
6. 上下文关联：
请只返回中文图像提示词，不要添加任何解释或其他内容。`,
			StyleAddon:     ", 音乐视频艺术风格, 动态视觉效果, 色彩渐变, 光影流动, 节奏感强烈的构图, 现代数字艺术技法, 视觉冲击力强, 情感表达丰富",
			NegativePrompt: "静态照片, 模糊不清, 色彩单调, 构图呆板, 缺乏艺术感, 过度写实, 没有视觉创意, 平淡无奇",
			BackgroundText: "mv背景：一个写给写给父亲的故事，传递了对父辈精神的传承与敬意，同时饱含对未来的期许",
		},
	}

	// 插入内置模板
	for _, template := range builtInTemplates {
		// 检查是否已存在同名模板
		var existingTemplate PromptTemplate
		result := db.Where("name = ?", template.Name).First(&existingTemplate)
		if result.Error != nil {
			// 模板不存在，插入新模板
			if err := db.Create(&template).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func GetConfig(db *gorm.DB) (Configuration, error) {
	var config Configuration
	tx := DB.Model(&Configuration{}).First(&config)
	return config, tx.Error
}
func GetConfigAddStr(db *gorm.DB) (string, error) {
	var config Configuration
	tx := DB.Model(&Configuration{}).First(&config)
	tmp := PromptTemplate{}
	DB.Model(&PromptTemplate{}).Where("id=?", config.PromptTemplateID).Find(&tmp)
	return tmp.BackgroundText, tx.Error
}

func SetConfig(PromptTemplateID uint, ImageWidth int, ImageHeight int) error {
	var config Configuration
	tx := DB.Model(&Configuration{}).First(&config)
	config.PromptTemplateID = PromptTemplateID
	config.ImageWidth = ImageWidth
	config.ImageHeight = ImageHeight
	tx = DB.Model(&Configuration{}).Save(&config)
	return tx.Error
}
