package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormManager GORM数据库管理器
type GormManager struct {
	DB *gorm.DB
}

// NewGormManager 创建新的GORM数据库管理器
func NewGormManager() (*GormManager, error) {
	dbPath, err := GetDatabasePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %v", err)
	}

	// 创建GORM配置，启用日志记录
	newLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // 慢SQL阈值
			LogLevel:                  logger.Silent, // 日志级别
			IgnoreRecordNotFoundError: true,          // 忽略ErrRecordNotFound错误
			Colorful:                  false,         // 禁用彩色打印
		},
	)

	// 连接数据库，添加UTF-8编码参数
	dsn := fmt.Sprintf("%s?_busy_timeout=10000&_journal_mode=WAL&_foreign_keys=on", dbPath)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %v", err)
	}

	manager := &GormManager{DB: db}

	// 自动迁移数据库表
	if err := manager.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	return manager, nil
}

// Migrate 执行数据库迁移
func (gm *GormManager) Migrate() error {
	return gm.DB.AutoMigrate(&Project{}, &ChapterProcess{}, &Scene{}, &ProcessStep{}, &DrawthingsConfig{})
}

// Close 关闭数据库连接
func (gm *GormManager) Close() error {
	sqlDB, err := gm.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB 获取数据库实例
func (gm *GormManager) GetDB() *gorm.DB {
	return gm.DB
}

// CreateProject 创建项目
func (gm *GormManager) CreateProject(name, description, genre, atmosphere string) (*Project, error) {
	project := &Project{
		Name:        name,
		Description: description,
		Genre:       genre,
		Atmosphere:  atmosphere,
		Status:      StatusPending,
	}

	result := gm.DB.Create(project)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create project: %v", result.Error)
	}

	return project, nil
}

// GetProjectByName 根据名称获取项目
func (gm *GormManager) GetProjectByName(name string) (*Project, error) {
	var project Project
	result := gm.DB.Preload("Chapters.Scenes.DrawthingsConfig").Preload("Chapters.Steps").First(&project, "name = ?", name)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get project: %v", result.Error)
	}

	return &project, nil
}

// GetProjectByID 根据ID获取项目
func (gm *GormManager) GetProjectByID(id uint) (*Project, error) {
	var project Project
	result := gm.DB.Preload("Chapters.Scenes.DrawthingsConfig").Preload("Chapters.Steps").First(&project, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get project: %v", result.Error)
	}

	return &project, nil
}

// UpdateProjectStatus 更新项目状态
func (gm *GormManager) UpdateProjectStatus(id uint, status ProcessStatus, errorMsg string) error {
	project := &Project{BaseModel: BaseModel{ID: id}}
	result := gm.DB.Model(project).Updates(map[string]interface{}{
		"status":     status,
		"error_msg":  errorMsg,
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update project status: %v", result.Error)
	}

	return nil
}

// UpdateProject 更新项目信息
func (gm *GormManager) UpdateProject(project *Project) error {
	result := gm.DB.Save(project)
	if result.Error != nil {
		return fmt.Errorf("failed to update project: %v", result.Error)
	}

	return nil
}

// CreateChapterProcess 创建章节处理记录
func (gm *GormManager) CreateChapterProcess(chapter *ChapterProcess) error {
	result := gm.DB.Create(chapter)
	if result.Error != nil {
		return fmt.Errorf("failed to create chapter process: %v", result.Error)
	}

	return nil
}

// GetChapterProcessByID 根据ID获取章节处理记录
func (gm *GormManager) GetChapterProcessByID(id uint) (*ChapterProcess, error) {
	var chapter ChapterProcess
	result := gm.DB.Preload("Scenes.DrawthingsConfig").Preload("Steps").First(&chapter, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get chapter process: %v", result.Error)
	}

	return &chapter, nil
}

// GetChapterProcessByProjectAndName 根据项目和章节名称获取章节处理记录
func (gm *GormManager) GetChapterProcessByProjectAndName(projectID uint, chapterName string) (*ChapterProcess, error) {
	var chapter ChapterProcess
	result := gm.DB.Preload("Scenes.DrawthingsConfig").Preload("Steps").Where("project_id = ? AND chapter_name = ?", projectID, chapterName).First(&chapter)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get chapter process: %v", result.Error)
	}

	return &chapter, nil
}

// UpdateChapterProcess 更新章节处理记录
func (gm *GormManager) UpdateChapterProcess(chapter *ChapterProcess) error {
	result := gm.DB.Save(chapter)
	if result.Error != nil {
		return fmt.Errorf("failed to update chapter process: %v", result.Error)
	}

	return nil
}

// UpdateChapterProcessStatus 更新章节处理状态
func (gm *GormManager) UpdateChapterProcessStatus(id uint, status ProcessStatus, errorMsg string) error {
	chapter := &ChapterProcess{BaseModel: BaseModel{ID: id}}
	result := gm.DB.Model(chapter).Updates(map[string]interface{}{
		"status":     status,
		"error_msg":  errorMsg,
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update chapter process status: %v", result.Error)
	}

	return nil
}

// UpdateChapterProgress 更新章节进度
func (gm *GormManager) UpdateChapterProgress(id uint, audioGenerated, imageGenerated, subtitleCreated, capcutCreated bool) error {
	chapter := &ChapterProcess{BaseModel: BaseModel{ID: id}}
	result := gm.DB.Model(chapter).Updates(map[string]interface{}{
		"audio_generated":  audioGenerated,
		"image_generated":  imageGenerated,
		"subtitle_created": subtitleCreated,
		"capcut_created":   capcutCreated,
		"updated_at":       time.Now(),
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update chapter progress: %v", result.Error)
	}

	return nil
}

// CreateProcessStep 创建处理步骤记录
func (gm *GormManager) CreateProcessStep(step *ProcessStep) error {
	result := gm.DB.Create(step)
	if result.Error != nil {
		return fmt.Errorf("failed to create process step: %v", result.Error)
	}

	return nil
}

// GetProcessStepByID 根据ID获取处理步骤记录
func (gm *GormManager) GetProcessStepByID(id uint) (*ProcessStep, error) {
	var step ProcessStep
	result := gm.DB.First(&step, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get process step: %v", result.Error)
	}

	return &step, nil
}

// GetProcessStepByChapterAndName 根据章节ID和步骤名称获取处理步骤
func (gm *GormManager) GetProcessStepByChapterAndName(chapterID uint, stepName string) (*ProcessStep, error) {
	var step ProcessStep
	result := gm.DB.Where("chapter_id = ? AND step_name = ?", chapterID, stepName).First(&step)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get process step: %v", result.Error)
	}

	return &step, nil
}

// UpdateProcessStep 更新处理步骤记录
func (gm *GormManager) UpdateProcessStep(step *ProcessStep) error {
	result := gm.DB.Save(step)
	if result.Error != nil {
		return fmt.Errorf("failed to update process step: %v", result.Error)
	}

	return nil
}

// UpdateProcessStepStatus 更新处理步骤状态
func (gm *GormManager) UpdateProcessStepStatus(id uint, status ProcessStatus, errorMsg string) error {
	step := &ProcessStep{BaseModel: BaseModel{ID: id}}
	result := gm.DB.Model(step).Updates(map[string]interface{}{
		"status":     status,
		"error_msg":  errorMsg,
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update process step status: %v", result.Error)
	}

	return nil
}

// GetChaptersByProjectID 根据项目ID获取所有章节
func (gm *GormManager) GetChaptersByProjectID(projectID uint) ([]ChapterProcess, error) {
	var chapters []ChapterProcess
	result := gm.DB.Preload("Scenes.DrawthingsConfig").Preload("Steps").Where("project_id = ?", projectID).Find(&chapters)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get chapters by project: %v", result.Error)
	}

	return chapters, nil
}

// RetryChapterProcess 重试章节处理
func (gm *GormManager) RetryChapterProcess(id uint) error {
	tx := gm.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新章节状态为待处理
	chapter := &ChapterProcess{BaseModel: BaseModel{ID: id}}
	if err := tx.Model(chapter).Updates(map[string]interface{}{
		"status":     StatusPending,
		"error_msg":  "",
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to reset chapter process status: %v", err)
	}

	// 重置相关步骤状态为待处理
	if err := tx.Model(&ProcessStep{}).Where("chapter_id = ?", id).Updates(map[string]interface{}{
		"status":     StatusPending,
		"error_msg":  "",
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to reset process steps: %v", err)
	}

	// 重置相关场景状态为待处理
	if err := tx.Model(&Scene{}).Where("chapter_id = ?", id).Updates(map[string]interface{}{
		"status":     StatusPending,
		"error_msg":  "",
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to reset scenes: %v", err)
	}

	return tx.Commit().Error
}

// ResetStepForRetry 重试特定步骤
func (gm *GormManager) ResetStepForRetry(chapterID uint, stepName string) error {
	step, err := gm.GetProcessStepByChapterAndName(chapterID, stepName)
	if err != nil {
		return fmt.Errorf("failed to get process step: %v", err)
	}

	if step == nil {
		// 如果步骤不存在，创建一个新的待处理步骤
		newStep := &ProcessStep{
			ChapterID: chapterID,
			StepName:  stepName,
			Status:    StatusPending,
		}
		return gm.CreateProcessStep(newStep)
	}

	// 重置步骤状态
	step.Status = StatusPending
	step.ErrorMsg = ""
	return gm.UpdateProcessStep(step)
}

// CreateScene 创建场景记录
func (gm *GormManager) CreateScene(scene *Scene) error {
	result := gm.DB.Create(scene)
	if result.Error != nil {
		return fmt.Errorf("failed to create scene: %v", result.Error)
	}

	return nil
}

// GetSceneByID 根据ID获取场景记录
func (gm *GormManager) GetSceneByID(id uint) (*Scene, error) {
	var scene Scene
	result := gm.DB.Preload("DrawthingsConfig").First(&scene, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get scene: %v", result.Error)
	}

	return &scene, nil
}

// GetScenesByChapterID 根据章节ID获取所有场景
func (gm *GormManager) GetScenesByChapterID(chapterID uint) ([]Scene, error) {
	var scenes []Scene
	result := gm.DB.Preload("DrawthingsConfig").Where("chapter_id = ?", chapterID).Order("scene_number").Find(&scenes)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get scenes by chapter: %v", result.Error)
	}

	return scenes, nil
}

// UpdateScene 更新场景记录
func (gm *GormManager) UpdateScene(scene *Scene) error {
	result := gm.DB.Save(scene)
	if result.Error != nil {
		return fmt.Errorf("failed to update scene: %v", result.Error)
	}

	return nil
}

// UpdateSceneStatus 更新场景状态
func (gm *GormManager) UpdateSceneStatus(id uint, status ProcessStatus, errorMsg string) error {
	scene := &Scene{BaseModel: BaseModel{ID: id}}
	result := gm.DB.Model(scene).Updates(map[string]interface{}{
		"status":     status,
		"error_msg":  errorMsg,
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update scene status: %v", result.Error)
	}

	return nil
}

// GetDrawthingsConfigByID 根据ID获取Drawthings配置
func (gm *GormManager) GetDrawthingsConfigByID(id uint) (*DrawthingsConfig, error) {
	var config DrawthingsConfig
	result := gm.DB.First(&config, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get drawthings config: %v", result.Error)
	}

	return &config, nil
}

// UpdateDrawthingsConfig 更新Drawthings配置
func (gm *GormManager) UpdateDrawthingsConfig(config *DrawthingsConfig) error {
	result := gm.DB.Save(config)
	if result.Error != nil {
		return fmt.Errorf("failed to update drawthings config: %v", result.Error)
	}

	return nil
}

// GetOrCreateDefaultDrawthingsConfig 获取或创建默认Drawthings配置
func (gm *GormManager) GetOrCreateDefaultDrawthingsConfig() (*DrawthingsConfig, error) {
	var config DrawthingsConfig
	result := gm.DB.First(&config)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 创建默认配置
			defaultConfig := &DrawthingsConfig{
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

			result := gm.DB.Create(defaultConfig)
			if result.Error != nil {
				return nil, fmt.Errorf("failed to create default drawthings config: %v", result.Error)
			}
			return defaultConfig, nil
		}
		return nil, fmt.Errorf("failed to get drawthings config: %v", result.Error)
	}

	return &config, nil
}
