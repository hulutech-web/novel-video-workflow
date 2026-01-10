package tools

import (
	"encoding/json"
	"fmt"
	"novel-video-workflow/internal/tools/indextts2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

type TTSProcessor struct {
	logger *zap.Logger
}

type TTSResult struct {
	Success    bool   `json:"success"`
	OutputFile string `json:"output_file"`
	Error      string `json:"error,omitempty"`
}

func NewTTSProcessor(logger *zap.Logger) *TTSProcessor {
	return &TTSProcessor{
		logger: logger,
	}
}

func (tp *TTSProcessor) Generate(text, outputFile, referenceAudio string) (*TTSResult, error) {
	// 如果outputFile为空，创建一个默认输出路径
	if outputFile == "" {
		// 获取项目根目录 - 使用绝对路径确保文件保存到正确位置
		wd, err := os.Getwd()
		if err != nil {
			return &TTSResult{
				Success: false,
				Error:   fmt.Sprintf("无法获取当前工作目录: %v", err),
			}, err
		}
		
		// 检查是否在项目根目录或内部子目录中
		var projectRoot string
		if strings.HasSuffix(wd, "/internal/tools") {
			// 如果当前工作目录是internal/tools，向上两级到达项目根目录
			projectRoot = filepath.Join(wd, "..", "..")
		} else {
			// 否则假定当前工作目录是项目根目录
			projectRoot = wd
		}
		
		// 确保输出目录是项目根目录下的output
		outputDir := filepath.Join(projectRoot, "output")
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return &TTSResult{
				Success: false,
				Error:   fmt.Sprintf("创建输出目录失败: %v", err),
			}, err
		}

		// 生成一个唯一的输出文件名
		outputFile = filepath.Join(outputDir, fmt.Sprintf("tts_output_%d.wav", tp.generateTimestamp()))
	} else {
		// 如果提供了outputFile，确保输出目录存在
		outputDir := filepath.Dir(outputFile)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return &TTSResult{
				Success: false,
				Error:   fmt.Sprintf("创建输出目录失败: %v", err),
			}, err
		}
	}

	// 尝试使用IndexTTS2 HTTP API客户端进行TTS生成 (新集成的客户端)
	client := indextts2.NewIndexTTS2Client(tp.logger, "http://localhost:7860")

	// 如果没有提供参考音频，尝试使用项目根目录的默认音频文件
	speakerAudio := referenceAudio
	if speakerAudio == "" {
		// 尝试项目根目录的默认音频文件
		wd, err := os.Getwd()
		if err != nil {
			return &TTSResult{
				Success: false,
				Error:   "无法获取工作目录",
			}, fmt.Errorf("无法获取工作目录: %v", err)
		}
		
		// 计算项目根目录
		var projectRoot string
		if strings.HasSuffix(wd, "/internal/tools") {
			projectRoot = filepath.Join(wd, "..", "..")
		} else {
			projectRoot = wd
		}

		// 优先尝试英文文件名
		defaultAudioPath := filepath.Join(projectRoot, "ref.m4a")
		if _, err := os.Stat(defaultAudioPath); err == nil {
			speakerAudio = defaultAudioPath
		} else {
			// 尝试中文文件名
			defaultAudioPath = filepath.Join(projectRoot, "音色.m4a")
			if _, err := os.Stat(defaultAudioPath); err == nil {
				speakerAudio = defaultAudioPath
			} else {
				// 再尝试使用绝对路径
				defaultAudioPath = "/Users/mac/code/ai/novel-video-workflow/ref.m4a"
				if _, err := os.Stat(defaultAudioPath); err == nil {
					speakerAudio = defaultAudioPath
				} else {
					// 尝试绝对路径的中文文件名
					defaultAudioPath = "/Users/mac/code/ai/novel-video-workflow/音色.m4a"
					if _, err := os.Stat(defaultAudioPath); err == nil {
						speakerAudio = defaultAudioPath
					} else {
						return &TTSResult{
							Success: false,
							Error:   "必须提供参考音频文件路径，或在项目根目录放置ref.m4a或音色.m4a文件",
						}, fmt.Errorf("必须提供参考音频文件路径，或在项目根目录放置ref.m4a或音色.m4a文件")
					}
				}
			}
		}
	} else {
		// 如果提供了参考音频，验证文件存在
		if _, err := os.Stat(referenceAudio); os.IsNotExist(err) {
			return &TTSResult{
				Success: false,
				Error:   fmt.Sprintf("参考音频文件不存在: %s", referenceAudio),
			}, fmt.Errorf("参考音频文件不存在: %s", referenceAudio)
		}
		speakerAudio = referenceAudio
	}

	// 首先尝试新的IndexTTS2客户端
	tp.logger.Info("尝试使用IndexTTS2 HTTP API客户端", 
		zap.String("output_file", outputFile),
		zap.String("speaker_audio", speakerAudio),
		zap.String("text", text))
		
	err := client.GenerateTTSWithAudio(speakerAudio, text, outputFile)
	if err == nil {
		// 如果API调用成功，检查文件是否存在
		tp.logger.Info("IndexTTS2 HTTP API调用完成，检查输出文件", zap.String("output_file", outputFile))
		if _, statErr := os.Stat(outputFile); statErr == nil {
			tp.logger.Info("IndexTTS2 HTTP API调用成功，文件已生成并验证存在", 
				zap.String("output_file", outputFile))
			return &TTSResult{
				Success:    true,
				OutputFile: outputFile,
			}, nil
		} else {
			tp.logger.Warn("IndexTTS2 HTTP API调用返回成功，但输出文件不存在或无法访问", 
				zap.String("output_file", outputFile), 
				zap.Error(statErr))
		}
	} else {
		tp.logger.Warn("IndexTTS2 HTTP API调用失败", zap.Error(err))
	}

	// 如果新的IndexTTS2客户端失败，尝试旧的IndexTTS客户端
	tp.logger.Info("尝试旧版IndexTTS客户端")
	oldClient := indextts2.NewIndexTTSClient(tp.logger, "http://localhost:7860")

	oldResult, err := oldClient.Generate(
		text,
		outputFile,
		indextts2.WithSpeakerAudio(speakerAudio),
	)

	// 检查旧版API调用结果
	if err != nil || (oldResult != nil && !oldResult.Success) {
		// 如果旧版HTTP API也失败，尝试本地调用方式
		tp.logger.Info("旧版IndexTTS HTTP API调用失败，尝试本地调用")
		localResult := tp.useIndexTTS2Directly(text, outputFile, speakerAudio)
		return localResult, nil
	}

	// 如果旧版API调用成功，直接返回
	if oldResult != nil && oldResult.Success {
		tp.logger.Info("旧版IndexTTS客户端调用成功", zap.String("audio_path", oldResult.AudioPath))
		return &TTSResult{
			Success:    true,
			OutputFile: oldResult.AudioPath,
		}, nil
	}

	// 如果所有方式都失败
	return &TTSResult{
		Success: false,
		Error:   "所有TTS引擎均失败",
	}, fmt.Errorf("TTS生成失败")
}

// generateTimestamp 生成时间戳用于创建唯一文件名
func (tp *TTSProcessor) generateTimestamp() int64 {
	// 简单实现，返回当前时间的秒数
	return time.Now().Unix()
}

// useIndexTTS2Directly 直接使用indexTTS2进行TTS生成
func (tp *TTSProcessor) useIndexTTS2Directly(text, outputFile, referenceAudio string) *TTSResult {
	// 首先尝试使用预定义的Python脚本
	scriptPath := "/Users/mac/code/ai/novel-video-workflow/scripts/run_tts.py"
	if _, err := os.Stat(scriptPath); err == nil {
		// 如果脚本存在，使用它
		cmd := exec.Command("python3", scriptPath, "--text", text, "--output", outputFile, "--reference", referenceAudio)

		output, err := cmd.CombinedOutput()
		if err != nil {
			// 解析错误响应
			var result map[string]interface{}
			if err2 := json.Unmarshal(output, &result); err2 == nil {
				if errMsg, ok := result["error"].([]interface{}); ok {
					// 如果是数组，取第一个元素
					if len(errMsg) > 0 {
						if errStr, ok := errMsg[0].(string); ok {
							return &TTSResult{
								Success: false,
								Error:   fmt.Sprintf("TTS本地调用失败: %s", errStr),
							}
						}
					}
				}
				if errMsg, ok := result["error"].(string); ok {
					return &TTSResult{
						Success: false,
						Error:   fmt.Sprintf("TTS本地调用失败: %s", errMsg),
					}
				}
			}

			return &TTSResult{
				Success: false,
				Error:   fmt.Sprintf("TTS本地调用失败: %v, 输出: %s", err, string(output)),
			}
		}

		// 解析成功响应
		var result map[string]interface{}
		if err := json.Unmarshal(output, &result); err == nil {
			if success, ok := result["success"].(bool); ok && success {
				if outputPath, ok := result["output_file"].(string); ok {
					return &TTSResult{
						Success:    true,
						OutputFile: outputPath,
					}
				}
			}
		}

		// 如果JSON解析失败，检查输出
		if strings.Contains(string(output), "SUCCESS") || strings.Contains(string(output), "successfully") {
			if _, err := os.Stat(outputFile); err == nil {
				return &TTSResult{
					Success:    true,
					OutputFile: outputFile,
				}
			}
		}
	} else {
		// 如果脚本不存在，使用内联Python代码
		script := fmt.Sprintf(`
import sys
import os
import json

try:
    sys.path.append('/Users/mac/code/ai/tts/index-tts')
    
    from indextts.infer_v2 import IndexTTS2
    
    # 设置模型目录
    model_dir = '/Users/mac/code/ai/tts/index-tts/checkpoints'
    if not os.path.exists(model_dir):
        raise FileNotFoundError(f"模型目录不存在: {model_dir}")
    
    # 初始化TTS模型
    tts = IndexTTS2(
        model_dir=model_dir,
        cfg_path=os.path.join(model_dir, "config.yaml")
    )
    
    # 生成音频
    result = tts.infer(
        spk_audio_prompt='%s',  # 参考音频
        text='%s',
        output_path='%s',
        verbose=True
    )
    
    print(json.dumps({"success": True, "output_file": "%s"}))
    
except Exception as e:
    print(json.dumps({"success": False, "error": str(e)}))
    sys.exit(1)
`, referenceAudio, text, outputFile, outputFile)

		cmd := exec.Command("python3", "-c", script)

		output, err := cmd.CombinedOutput()
		if err != nil {
			// 尝试解析错误输出
			var result map[string]interface{}
			if err2 := json.Unmarshal(output, &result); err2 == nil {
				if errMsg, ok := result["error"].(string); ok {
					return &TTSResult{
						Success: false,
						Error:   fmt.Sprintf("indexTTS2调用失败: %s", errMsg),
					}
				}
			}

			tp.logger.Warn("indexTTS2调用失败",
				zap.Error(err),
				zap.String("output", string(output)),
				zap.String("script", func() string {
					end := len(script)
					if 200 < end {
						end = 200
					}
					return script[:end] + "..."
				}()))
			return &TTSResult{
				Success: false,
				Error:   fmt.Sprintf("indexTTS2调用失败: %v, 输出: %s", err, string(output)),
			}
		}

		// 解析响应
		var result map[string]interface{}
		if err := json.Unmarshal(output, &result); err == nil {
			if success, ok := result["success"].(bool); ok && success {
				if outputPath, ok := result["output_file"].(string); ok {
					if _, err := os.Stat(outputPath); err == nil {
						return &TTSResult{
							Success:    true,
							OutputFile: outputPath,
						}
					}
				}
			}
		}
	}

	// 如果所有方法都失败，检查文件是否存在
	if _, err := os.Stat(outputFile); err == nil {
		return &TTSResult{
			Success:    true,
			OutputFile: outputFile,
		}
	}

	return &TTSResult{
		Success: false,
		Error:   fmt.Sprintf("音频文件未生成: %s", outputFile),
	}
}