package database

import (
	"gorm.io/gorm"
)

// BaseModel 包含公共字段
type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt MyTime         `json:"created_at"`
	UpdatedAt MyTime         `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Project 项目模型 - 代表每个小说项目
type Project struct {
	BaseModel
	Name        string           `json:"name"`                    // 项目名称（小说名称）
	Description string           `json:"description,omitempty"`   // 项目描述
	Genre       string           `json:"genre"`                   // 小说类型（悬疑、惊悚、科幻等）
	Atmosphere  string           `json:"atmosphere"`              // 整体氛围设定
	Status      ProcessStatus    `json:"status" gorm:"default:pending"` // 项目整体状态
	ErrorMsg    string           `json:"error_msg,omitempty"`     // 错误信息
	TotalChapters int            `json:"total_chapters"`          // 总章节数
	ProcessedChapters int        `json:"processed_chapters"`      // 已处理章节数
	Chapters    []ChapterProcess `json:"chapters" gorm:"foreignKey:ProjectID"`
}

// ChapterProcess 章节处理记录模型
type ChapterProcess struct {
	BaseModel
	ProjectID     uint          `json:"project_id"`              // 关联项目ID
	Project       Project       `json:"project" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // 项目关联
	ChapterName   string        `json:"chapter_name"`            // 章节名称
	ChapterText   string        `json:"chapter_text"`            // 章节文本内容
	InputFilePath string        `json:"input_file_path"`         // 输入文件路径
	OutputDir     string        `json:"output_dir"`              // 输出目录
	Status        ProcessStatus `json:"status" gorm:"default:pending"` // 处理状态
	ErrorMsg      string        `json:"error_msg,omitempty"`     // 错误信息
	AudioGenerated  bool        `json:"audio_generated" gorm:"default:false"`  // 音频是否已生成
	ImageGenerated  bool        `json:"image_generated" gorm:"default:false"`  // 图像是否已生成
	SubtitleCreated bool        `json:"subtitle_created" gorm:"default:false"` // 字幕是否已创建
	CapcutCreated   bool        `json:"capcut_created" gorm:"default:false"`   // 剪映项目是否已创建
	StartTime     MyTime        `json:"start_time,omitempty"`    // 开始时间
	EndTime       MyTime        `json:"end_time,omitempty"`      // 结束时间
	Duration      int64         `json:"duration"`                // 处理耗时（秒）
	Scenes        []Scene       `json:"scenes" gorm:"foreignKey:ChapterID"` // 章节场景
	Steps         []ProcessStep `json:"steps" gorm:"foreignKey:ChapterID"` // 处理步骤
}

// Scene 章节场景模型 - 用于存储AI拆分的场景或用户自定义场景
type Scene struct {
	BaseModel
	ChapterID     uint          `json:"chapter_id"`              // 关联章节ID
	Chapter       ChapterProcess `json:"chapter" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // 章节关联
	SceneNumber   int           `json:"scene_number"`            // 场景编号
	Description   string        `json:"description"`             // 场景描述
	Prompt        string        `json:"prompt"`                  // 生成提示词
	IsAIgenerated bool          `json:"is_ai_generated"`         // 是否为AI生成
	Status        ProcessStatus `json:"status" gorm:"default:pending"` // 处理状态
	ErrorMsg      string        `json:"error_msg,omitempty"`     // 错误信息
	ImageFile     string        `json:"image_file,omitempty"`    // 生成的图像文件路径
	StartTime     MyTime        `json:"start_time,omitempty"`    // 开始时间
	EndTime       MyTime        `json:"end_time,omitempty"`      // 结束时间
	Duration      int64         `json:"duration"`                // 处理耗时（秒）
	DrawthingsConfig DrawthingsConfig `json:"drawthings_config" gorm:"embedded"` // Drawthings配置
}

// DrawthingsConfig Drawthings配置模型 - 用于存储图像生成参数
type DrawthingsConfig struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	Width           int    `json:"width" gorm:"default:512"`              // 图像宽度
	Height          int    `json:"height" gorm:"default:896"`             // 图像高度
	Steps           int    `json:"steps" gorm:"default:20"`               // 生成步数
	CFGScale        float64 `json:"cfg_scale" gorm:"default:7.0"`         // CFG缩放
	SamplerName     string `json:"sampler_name" gorm:"default:'Euler a'"` // 采样器名称
	Seed            int    `json:"seed" gorm:"default:-1"`                // 随机种子
	BatchSize       int    `json:"batch_size" gorm:"default:1"`           // 批次大小
	LoraModel       string `json:"lora_model"`                            // LoRA模型名称
	LoraTriggerWord string `json:"lora_trigger_word"`                     // LoRA触发词
	LoraWeight      float64 `json:"lora_weight" gorm:"default:0.8"`       // LoRA权重
	NegativePrompt  string `json:"negative_prompt"`                       // 负面提示词
	CreatedAt       MyTime `json:"created_at"`
	UpdatedAt       MyTime `json:"updated_at"`
}

// ProcessStep 处理步骤记录模型
type ProcessStep struct {
	BaseModel
	ChapterID   uint        `json:"chapter_id"`        // 关联章节ID
	Chapter     ChapterProcess `json:"chapter" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // 章节关联
	StepName    string      `json:"step_name"`         // 步骤名称：split, audio, image, subtitle, capcut等
	Status      ProcessStatus `json:"status" gorm:"default:pending"` // 步骤状态
	ErrorMsg    string      `json:"error_msg,omitempty"` // 错误信息
	StartTime   MyTime      `json:"start_time,omitempty"` // 开始时间
	EndTime     MyTime      `json:"end_time,omitempty"`   // 结束时间
	Duration    int64       `json:"duration"`             // 耗时（秒）
	Details     string      `json:"details,omitempty"`    // 详细信息
}