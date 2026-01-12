package drawthings

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// DrawThingsClient 封装 DrawThings API 调用
type DrawThingsClient struct {
	BaseURL      string
	Logger       *zap.Logger
	HTTPClient   *http.Client
	APIAvailable bool // 记录API是否可用
}

// NewDrawThingsClient 创建新的客户端实例
func NewDrawThingsClient(logger *zap.Logger, baseURL string) *DrawThingsClient {
	if baseURL == "" {
		baseURL = "http://localhost:7861" // 默认地址
	}

	client := &DrawThingsClient{
		BaseURL: baseURL,
		Logger:  logger,
		HTTPClient: &http.Client{
			Timeout: 300 * time.Second, // 图像生成可能需要较长时间
		},
		APIAvailable: false, // 初始状态假设API不可用
	}

	// 检查API可用性
	client.CheckAPIAvailability()

	return client
}

// Txt2ImgRequest 文生图请求参数
type Txt2ImgRequest struct {
	Prompt            string   `json:"prompt"`
	NegativePrompt    string   `json:"negative_prompt"`
	Width             int      `json:"width"`
	Height            int      `json:"height"`
	Steps             int      `json:"steps"`
	Seed              int      `json:"seed"`
	SamplerName       string   `json:"sampler"`
	GuidanceScale     float64  `json:"cfg_scale"`
	BatchSize         int      `json:"batch_size"`
	EnableHr          bool     `json:"enable_hr,omitempty"`
	HrScale           float64  `json:"hr_scale,omitempty"`
	HrUpscaler        string   `json:"hr_upscaler,omitempty"`
	HrSecondPassSteps int      `json:"hr_second_pass_steps,omitempty"`
	DenoisingStrength *float64 `json:"denoising_strength,omitempty"`
	Tiling            bool     `json:"tiling,omitempty"`
	Model             string   `json:"model,omitempty"`
}

// Img2ImgRequest 图生图请求参数
type Img2ImgRequest struct {
	InitImages     []string `json:"init_images"`
	Strength       float64  `json:"strength"`
	Prompt         string   `json:"prompt"`
	NegativePrompt string   `json:"negative_prompt"`
	Width          int      `json:"width"`
	Height         int      `json:"height"`
	Steps          int      `json:"steps"`
	SamplerName    string   `json:"sampler"`
	GuidanceScale  float64  `json:"cfg_scale"`
	BatchSize      int      `json:"batch_size"`
	Model          string   `json:"model,omitempty"`
}

// Txt2ImgResponse 文生图响应
type Txt2ImgResponse struct {
	Images     []string               `json:"images"` // Base64编码的图像数据
	Parameters map[string]interface{} `json:"parameters"`
	Info       string                 `json:"info"`
}

// Img2ImgResponse 图生图响应
type Img2ImgResponse struct {
	Images     []string               `json:"images"` // Base64编码的图像数据
	Parameters map[string]interface{} `json:"parameters"`
	Info       string                 `json:"info"`
}

// CheckAPIAvailability 检查API是否可用
func (c *DrawThingsClient) CheckAPIAvailability() bool {
	testEndpoint := c.BaseURL + "" // 使用进度API作为连通性测试

	req, err := http.NewRequest("GET", testEndpoint, nil)
	if err != nil {
		c.Logger.Error("创建API可用性检查请求失败", zap.Error(err))
		c.APIAvailable = false
		return false
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Logger.Info("DrawThings API不可用", zap.String("url", testEndpoint), zap.Error(err))
		c.APIAvailable = false
		return false
	}
	resp.Body.Close()

	// 不检查响应状态，只要能连接就认为可用
	c.Logger.Info("DrawThings API可用", zap.String("url", testEndpoint))
	c.APIAvailable = true
	return true
}

// Txt2Img 生成图像
func (c *DrawThingsClient) Txt2Img(params Txt2ImgRequest) (*Txt2ImgResponse, error) {
	// 先检查API是否可用
	if !c.APIAvailable {
		if !c.CheckAPIAvailability() {
			return nil, fmt.Errorf("DrawThings API不可用，请确保Stable Diffusion WebUI正在运行在 %s", c.BaseURL)
		}
	}

	endpoint := c.BaseURL + "/sdapi/v1/txt2img"

	payload, err := json.Marshal(params)
	if err != nil {
		c.Logger.Error("序列化请求参数失败", zap.Error(err))
		return nil, fmt.Errorf("序列化请求参数失败: %v", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		c.Logger.Error("创建请求失败", zap.Error(err))
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	c.Logger.Info("发送文生图请求",
		zap.String("endpoint", endpoint),
		zap.String("prompt", params.Prompt),
		zap.String("sampler", params.SamplerName),
		zap.Int("width", params.Width),
		zap.Int("height", params.Height),
		zap.String("model", params.Model))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Logger.Error("发送请求失败", zap.Error(err))
		// 更新API可用性状态
		c.APIAvailable = false
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.Logger.Error("API返回错误状态码",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)))
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var result Txt2ImgResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.Logger.Error("解析响应失败", zap.Error(err))
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	c.Logger.Info("文生图请求成功", zap.Int("images_count", len(result.Images)))

	return &result, nil
}

// Img2Img 图生图
func (c *DrawThingsClient) Img2Img(params Img2ImgRequest) (*Img2ImgResponse, error) {
	// 先检查API是否可用
	if !c.APIAvailable {
		if !c.CheckAPIAvailability() {
			return nil, fmt.Errorf("DrawThings API不可用，请确保Stable Diffusion WebUI正在运行在 %s", c.BaseURL)
		}
	}

	endpoint := c.BaseURL + "/sdapi/v1/img2img"

	payload, err := json.Marshal(params)
	if err != nil {
		c.Logger.Error("序列化请求参数失败", zap.Error(err))
		return nil, fmt.Errorf("序列化请求参数失败: %v", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		c.Logger.Error("创建请求失败", zap.Error(err))
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	c.Logger.Info("发送图生图请求",
		zap.String("endpoint", endpoint),
		zap.String("prompt", params.Prompt),
		zap.Int("init_images_count", len(params.InitImages)))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Logger.Error("发送请求失败", zap.Error(err))
		// 更新API可用性状态
		c.APIAvailable = false
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.Logger.Error("API返回错误状态码",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)))
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var result Img2ImgResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.Logger.Error("解析响应失败", zap.Error(err))
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	c.Logger.Info("图生图请求成功", zap.Int("images_count", len(result.Images)))

	return &result, nil
}

// SaveImageFromBase64 将Base64编码的图像数据保存到文件
func (c *DrawThingsClient) SaveImageFromBase64(base64Data, filePath string) error {
	// 移除Base64数据前缀（如果有）
	if len(base64Data) > 22 && base64Data[:22] == "data:image/png;base64," {
		base64Data = base64Data[22:]
	}

	// 解码Base64数据
	imgData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		c.Logger.Error("解码Base64图像数据失败", zap.Error(err))
		return fmt.Errorf("解码Base64图像数据失败: %v", err)
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(filePath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		c.Logger.Error("创建输出目录失败", zap.Error(err))
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, imgData, 0644); err != nil {
		c.Logger.Error("保存图像文件失败", zap.String("file", filePath), zap.Error(err))
		return fmt.Errorf("保存图像文件失败: %v", err)
	}

	c.Logger.Info("图像保存成功", zap.String("file", filePath))

	return nil
}

// GenerateImageFromText 根据文本生成图像
func (c *DrawThingsClient) GenerateImageFromText(text, outputFile string, width, height int, isSuspense bool) error {
	// 先检查API是否可用
	if !c.APIAvailable {
		if !c.CheckAPIAvailability() {
			return fmt.Errorf("无法连接到DrawThings API，请确保Stable Diffusion WebUI正在运行在 %s 并且可以通过该地址访问", c.BaseURL)
		}
	}

	// 生成提示词，结合文本内容和悬疑风格
	prompt := text
	if isSuspense {
		// 添加悬疑风格描述
		suspenseStyle := ", 周围环境模糊成黑影, 空气凝滞,浅景深, 胶片颗粒感, 低饱和度，极致悬疑氛围, 阴沉窒息感, 夏季，环境阴霾，其他部分模糊不可见"
		prompt = text + suspenseStyle
	}

	// 设置denoisingStrength值
	strengthValue := 1.0

	// 默认参数值
	params := Txt2ImgRequest{
		Prompt:            prompt,
		NegativePrompt:    "人脸特写，半身像，模糊，比例失调，原参考图背景，比例失调，缺肢",
		Width:             width,
		Height:            height,
		Steps:             8, // 快速生成
		Seed:              -1,
		SamplerName:       "DPM++ 2M Trailing", // 使用API支持的标准采样器
		GuidanceScale:     1.0,                 // 使用较低的引导系数
		BatchSize:         1,
		EnableHr:          false, // 关闭高清修复
		HrScale:           0,
		HrUpscaler:        "",
		HrSecondPassSteps: 0,
		Tiling:            false,
		Model:             "z_image_turbo_1.0_q6p.ckpt", // 使用z-image turbo模型
		DenoisingStrength: &strengthValue,
	}

	response, err := c.Txt2Img(params)
	if err != nil {
		return fmt.Errorf("生成图像失败: %v", err)
	}

	if len(response.Images) == 0 {
		return fmt.Errorf("API返回的图像数量为0")
	}

	// 保存第一张图像
	return c.SaveImageFromBase64(response.Images[0], outputFile)
}

// GenerateImageFromImage 根据参考图像生成新图像
func (c *DrawThingsClient) GenerateImageFromImage(initImagePath, text, outputFile string, width, height int, isSuspense bool) error {
	// 读取参考图像并编码为Base64
	initImageBytes, err := os.ReadFile(initImagePath)
	if err != nil {
		return fmt.Errorf("读取参考图像失败: %v", err)
	}

	initImageBase64 := base64.StdEncoding.EncodeToString(initImageBytes)

	// 生成提示词
	prompt := text
	if isSuspense {
		// 添加悬疑风格描述
		suspenseStyle := ", 参考图面部特征，极致悬疑氛围, 阴沉窒息感, 夏季，环境阴霾"
		prompt = "(" + text + ":1.5)" + suspenseStyle
	}

	// 默认参数值
	params := Img2ImgRequest{
		InitImages:     []string{initImageBase64},
		Strength:       0.7, // 关键：突破原图人脸构图限制
		Prompt:         prompt,
		NegativePrompt: "人脸特写，半身像，原参考图背景，比例失调，缺肢",
		Width:          width,
		Height:         height,
		Steps:          8,
		SamplerName:    "DPM++ 2M Trailing",
		GuidanceScale:  1.0,
		BatchSize:      1,
		Model:          "z_image_turbo_1.0_q6p.ckpt", // 使用z-image turbo模型
	}

	response, err := c.Img2Img(params)
	if err != nil {
		return fmt.Errorf("图生图失败: %v", err)
	}

	if len(response.Images) == 0 {
		return fmt.Errorf("API返回的图像数量为0")
	}

	// 保存第一张图像
	return c.SaveImageFromBase64(response.Images[0], outputFile)
}
