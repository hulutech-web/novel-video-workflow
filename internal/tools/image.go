/*图片生辰*/
package tools

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// ImageGenerator 图片生成器
type ImageGenerator struct {
	logger     *zap.Logger
	httpClient *http.Client
}

// ImageParams 图片生成参数
type ImageParams struct {
	Prompt         string  `json:"prompt"`
	NegativePrompt string  `json:"negative_prompt,omitempty"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	Steps          int     `json:"steps"`
	CFGScale       float64 `json:"cfg_scale"`
	Seed           int64   `json:"seed,omitempty"`
	StylePreset    string  `json:"style_preset,omitempty"`
}

// ImageResult 图片生成结果
type ImageResult struct {
	Success   bool    `json:"success"`
	ImageFile string  `json:"image_file,omitempty"`
	Prompt    string  `json:"prompt,omitempty"`
	TimeCost  float64 `json:"time_cost,omitempty"`
	Error     string  `json:"error,omitempty"`
}

// SceneImage 场景图片
type SceneImage struct {
	ParagraphIndex int      `json:"paragraph_index"`
	ImageFile      string   `json:"image_file"`
	Prompt         string   `json:"prompt"`
	Keywords       []string `json:"keywords"`
}

// NewImageGenerator 创建图片生成器
func NewImageGenerator(logger *zap.Logger) *ImageGenerator {
	return &ImageGenerator{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 300 * time.Second, // 图片生成可能较慢
		},
	}
}

// GenerateSceneImages 为文本生成场景图片
func (ig *ImageGenerator) GenerateSceneImages(text, outputDir string, maxImages int) ([]SceneImage, error) {
	ig.logger.Info("开始生成场景图片",
		zap.String("输出目录", outputDir),
		zap.Int("最大图片数", maxImages),
	)

	// 1. 分割文本为段落
	paragraphs := ig.splitTextToParagraphs(text)
	if len(paragraphs) == 0 {
		return nil, fmt.Errorf("没有可用的段落")
	}

	// 限制生成图片数量
	if len(paragraphs) > maxImages {
		paragraphs = paragraphs[:maxImages]
	}

	// 2. 确定使用的图片生成引擎
	engine := viper.GetString("image.engine")
	ig.logger.Info("使用图片生成引擎", zap.String("引擎", engine))

	// 3. 为每个段落生成图片
	var sceneImages []SceneImage
	var mu sync.Mutex  // 添加互斥锁解决数据竞争
	semaphore := make(chan struct{}, 2) // 并发限制

	for i, paragraph := range paragraphs {
		semaphore <- struct{}{}

		go func(idx int, para string) {
			defer func() { <-semaphore }()

			// 从段落提取提示词
			prompt := ig.extractPromptFromParagraph(para)

			// 生成图片文件名
			imageFile := filepath.Join(outputDir, fmt.Sprintf("scene_%03d.png", idx+1))

			// 根据引擎生成图片
			var result *ImageResult

			switch engine {
			case "drawthings":
				result = ig.generateWithDrawThings(prompt, imageFile)
			case "stable_diffusion":
				result = ig.generateWithStableDiffusion(prompt, imageFile)
			case "dall_e":
				result = ig.generateWithDallE(prompt, imageFile)
			default:
				result = &ImageResult{
					Success: false,
					Error:   fmt.Sprintf("不支持的引擎: %s", engine),
				}
			}

			if result.Success {
				sceneImage := SceneImage{
					ParagraphIndex: idx + 1,
					ImageFile:      imageFile,
					Prompt:         prompt,
					Keywords:       ig.extractKeywords(para),
				}

				ig.logger.Info("场景图片生成成功",
					zap.Int("段落", idx+1),
					zap.String("图片", imageFile),
				)

				// 添加到结果列表（线程安全）
				mu.Lock()
				sceneImages = append(sceneImages, sceneImage)
				mu.Unlock()
			} else {
				ig.logger.Warn("场景图片生成失败",
					zap.Int("段落", idx+1),
					zap.String("错误", result.Error),
				)
			}
		}(i, paragraph)

		// 避免请求过于频繁
		time.Sleep(1 * time.Second)
	}

	// 等待所有goroutine完成
	for i := 0; i < cap(semaphore); i++ {
		select {
		case semaphore <- struct{}{}:
		case <-time.After(10 * time.Second): // 防止死锁
		}
	}
	close(semaphore)  // 关闭通道

	ig.logger.Info("场景图片生成完成",
		zap.Int("成功数量", len(sceneImages)),
	)

	return sceneImages, nil
}

// generateWithDrawThings 使用DrawThings生成图片
func (ig *ImageGenerator) generateWithDrawThings(prompt, outputFile string) *ImageResult {
	startTime := time.Now()

	// 创建AppleScript脚本
	appleScript := fmt.Sprintf(`
tell application "DrawThings"
    activate
    delay 2
    
    -- 检查是否已打开项目
    try
        tell application "System Events"
            tell process "DrawThings"
                -- 点击新建按钮
                click button "New" of window 1
                delay 1
                
                -- 输入提示词
                keystroke "a" using command down
                keystroke "%s"
                
                -- 设置参数
                -- 宽度
                set value of text field "Width" of window 1 to "%d"
                -- 高度
                set value of text field "Height" of window 1 to "%d"
                -- 步数
                set value of text field "Steps" of window 1 to "%d"
                
                -- 开始生成
                click button "Generate" of window 1
                
                -- 等待生成完成（根据步数调整）
                delay %d
                
                -- 保存图片
                click menu item "Export Image…" of menu "File" of menu bar 1
                delay 1
                
                -- 输入保存路径
                keystroke "%s"
                delay 1
                keystroke return
                
                -- 等待保存完成
                delay 3
            end tell
        end tell
        
        return true
    on error errMsg
        return "错误: " & errMsg
    end try
end tell
`,
		strings.ReplaceAll(prompt, "\"", "\\\""),
		viper.GetInt("image.width"),
		viper.GetInt("image.height"),
		viper.GetInt("image.steps"),
		viper.GetInt("image.steps")*2+10, // 预估等待时间
		outputFile,
	)

	// 执行AppleScript
	cmd := exec.Command("osascript", "-e", appleScript)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &ImageResult{
			Success:  false,
			Error:    fmt.Sprintf("AppleScript执行失败: %v, 输出: %s", err, output),
			TimeCost: time.Since(startTime).Seconds(),
		}
	}

	// 检查文件是否生成
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		return &ImageResult{
			Success:  false,
			Error:    fmt.Sprintf("图片文件未生成: %v", err),
			TimeCost: time.Since(startTime).Seconds(),
		}
	}

	return &ImageResult{
		Success:   true,
		ImageFile: outputFile,
		Prompt:    prompt,
		TimeCost:  time.Since(startTime).Seconds(),
	}
}

// generateWithStableDiffusion 使用Stable Diffusion API生成图片
func (ig *ImageGenerator) generateWithStableDiffusion(prompt, outputFile string) *ImageResult {
	startTime := time.Now()

	// 构建API请求
	sdURL := viper.GetString("image.sd_api_url")
	if sdURL == "" {
		sdURL = "http://localhost:7860"
	}

	requestBody := map[string]interface{}{
		"prompt":          prompt,
		"negative_prompt": viper.GetString("image.negative_prompt"),
		"width":           viper.GetInt("image.width"),
		"height":          viper.GetInt("image.height"),
		"steps":           viper.GetInt("image.steps"),
		"cfg_scale":       viper.GetFloat64("image.cfg_scale"),
		"seed":            -1, // 随机种子
		"sampler_name":    "DPM++ 2M Karras",
		"enable_hr":       false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return &ImageResult{
			Success:  false,
			Error:    fmt.Sprintf("构建请求失败: %v", err),
			TimeCost: time.Since(startTime).Seconds(),
		}
	}

	// 发送请求
	resp, err := ig.httpClient.Post(
		sdURL+"/sdapi/v1/txt2img",
		"application/json",
		strings.NewReader(string(jsonData)),
	)

	if err != nil {
		return &ImageResult{
			Success:  false,
			Error:    fmt.Sprintf("API请求失败: %v", err),
			TimeCost: time.Since(startTime).Seconds(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &ImageResult{
			Success:  false,
			Error:    fmt.Sprintf("API返回错误: %s, 响应: %s", resp.Status, body),
			TimeCost: time.Since(startTime).Seconds(),
		}
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return &ImageResult{
			Success:  false,
			Error:    fmt.Sprintf("解析响应失败: %v", err),
			TimeCost: time.Since(startTime).Seconds(),
		}
	}

	// 提取图片数据
	images, ok := response["images"].([]interface{})
	if !ok || len(images) == 0 {
		return &ImageResult{
			Success:  false,
			Error:    "API响应中没有图片数据",
			TimeCost: time.Since(startTime).Seconds(),
		}
	}

	// 解码Base64图片
	imageData, ok := images[0].(string)
	if !ok {
		return &ImageResult{
			Success:  false,
			Error:    "图片数据格式错误",
			TimeCost: time.Since(startTime).Seconds(),
		}
	}

	// 保存图片
	decoded, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return &ImageResult{
			Success:  false,
			Error:    fmt.Sprintf("解码图片失败: %v", err),
			TimeCost: time.Since(startTime).Seconds(),
		}
	}

	if err := os.WriteFile(outputFile, decoded, 0644); err != nil {
		return &ImageResult{
			Success:  false,
			Error:    fmt.Sprintf("保存图片失败: %v", err),
			TimeCost: time.Since(startTime).Seconds(),
		}
	}

	return &ImageResult{
		Success:   true,
		ImageFile: outputFile,
		Prompt:    prompt,
		TimeCost:  time.Since(startTime).Seconds(),
	}
}

// generateWithDallE 使用DALL-E API生成图片
func (ig *ImageGenerator) generateWithDallE(prompt, outputFile string) *ImageResult {
	startTime := time.Now()

	apiKey := viper.GetString("image.dalle_api_key")
	if apiKey == "" {
		return &ImageResult{
			Success:  false,
			Error:    "未配置DALL-E API密钥",
			TimeCost: time.Since(startTime).Seconds(),
		}
	}

	requestBody := map[string]interface{}{
		"prompt":          prompt,
		"n":               1,
		"size":            "1024x1024",
		"response_format": "b64_json",
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return &ImageResult{
			Success:  false,
			Error:    fmt.Sprintf("构建请求失败: %v", err),
			TimeCost: time.Since(startTime).Seconds(),
		}
	}

	req, err := http.NewRequest(
		"POST",
		"https://api.openai.com/v1/images/generations",
		strings.NewReader(string(jsonData)),
	)

	if err != nil {
		return &ImageResult{
			Success:  false,
			Error:    fmt.Sprintf("创建请求失败: %v", err),
			TimeCost: time.Since(startTime).Seconds(),
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := ig.httpClient.Do(req)
	if err != nil {
		return &ImageResult{
			Success:  false,
			Error:    fmt.Sprintf("API请求失败: %v", err),
			TimeCost: time.Since(startTime).Seconds(),
		}
	}
	defer resp.Body.Close()

	// 解析响应
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return &ImageResult{
			Success:  false,
			Error:    fmt.Sprintf("解析响应失败: %v", err),
			TimeCost: time.Since(startTime).Seconds(),
		}
	}

	// 提取图片数据
	images, ok := response["data"].([]interface{})
	if !ok || len(images) == 0 {
		return &ImageResult{
			Success:  false,
			Error:    "API响应中没有图片数据",
			TimeCost: time.Since(startTime).Seconds(),
		}
	}

	// 解码Base64图片
	imageData, ok := images[0].(map[string]interface{})["b64_json"].(string)
	if !ok {
		return &ImageResult{
			Success:  false,
			Error:    "图片数据格式错误",
			TimeCost: time.Since(startTime).Seconds(),
		}
	}

	// 保存图片
	decoded, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return &ImageResult{
			Success:  false,
			Error:    fmt.Sprintf("解码图片失败: %v", err),
			TimeCost: time.Since(startTime).Seconds(),
		}
	}

	if err := os.WriteFile(outputFile, decoded, 0644); err != nil {
		return &ImageResult{
			Success:  false,
			Error:    fmt.Sprintf("保存图片失败: %v", err),
			TimeCost: time.Since(startTime).Seconds(),
		}
	}

	return &ImageResult{
		Success:   true,
		ImageFile: outputFile,
		Prompt:    prompt,
		TimeCost:  time.Since(startTime).Seconds(),
	}
}

// splitTextToParagraphs 将文本分割为段落
func (ig *ImageGenerator) splitTextToParagraphs(text string) []string {
	// 按空行分割
	rawParagraphs := strings.Split(text, "\n\n")

	var paragraphs []string
	for _, p := range rawParagraphs {
		trimmed := strings.TrimSpace(p)
		if len(trimmed) > 0 {
			paragraphs = append(paragraphs, trimmed)
		}
	}

	return paragraphs
}

// extractPromptFromParagraph 从段落中提取图片提示词
func (ig *ImageGenerator) extractPromptFromParagraph(paragraph string) string {
	// 1. 移除标点符号
	cleanText := strings.Map(func(r rune) rune {
		if strings.ContainsRune(".,;:!?。，；：！？", r) {
			return -1
		}
		return r
	}, paragraph)

	// 2. 提取关键词（简单实现）
	words := strings.Fields(cleanText)

	// 3. 限制关键词数量
	maxKeywords := 15
	if len(words) > maxKeywords {
		words = words[:maxKeywords]
	}

	// 4. 组合成提示词
	prompt := strings.Join(words, ", ")

	// 5. 添加风格后缀
	stylePreset := viper.GetString("image.style_preset")
	switch stylePreset {
	case "anime":
		prompt += ", anime style, high quality, masterpiece"
	case "realistic":
		prompt += ", photorealistic, detailed, 8k"
	case "fantasy":
		prompt += ", fantasy art, magical, ethereal"
	default:
		prompt += ", digital art, detailed, cinematic"
	}

	return prompt
}

// extractKeywords 从文本中提取关键词
func (ig *ImageGenerator) extractKeywords(text string) []string {
	// 简单的关键词提取（可替换为更复杂的NLP处理）
	stopWords := map[string]bool{
		"的": true, "了": true, "在": true, "是": true, "我": true,
		"有": true, "和": true, "就": true, "不": true, "人": true,
		"都": true, "一": true, "一个": true, "上": true, "也": true,
		"很": true, "到": true, "说": true, "要": true, "去": true,
		"你": true, "会": true, "着": true, "没有": true, "看": true,
		"好": true, "自己": true, "这": true, "那": true,
	}

	words := strings.Fields(text)
	var keywords []string

	for _, word := range words {
		if len(word) > 1 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	// 去重并限制数量
	return ig.removeDuplicates(keywords[:min(10, len(keywords))])
}

// removeDuplicates 移除重复元素
func (ig *ImageGenerator) removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, val := range slice {
		if !seen[val] {
			seen[val] = true
			result = append(result, val)
		}
	}

	return result
}

// min 返回最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}