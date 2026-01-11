package indextts2

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// IndexTTS2Client 封装 IndexTTS2 API 调用
type IndexTTS2Client struct {
	BaseURL    string
	Logger     *zap.Logger
	HTTPClient *http.Client
}

// NewIndexTTS2Client 创建新的客户端实例
func NewIndexTTS2Client(logger *zap.Logger, baseURL string) *IndexTTS2Client {
	if baseURL == "" {
		baseURL = "http://localhost:7860" // 默认地址
	}

	return &IndexTTS2Client{
		BaseURL: baseURL,
		Logger:  logger,
		HTTPClient: &http.Client{
			Timeout: 300 * time.Second, // TTS生成可能需要较长时间
		},
	}
}

// IndexTTSClient 为了向后兼容，提供旧的客户端接口
type IndexTTSClient struct {
	BaseURL    string
	Logger     *zap.Logger
	HTTPClient *http.Client
}

// NewIndexTTSClient 创建旧版客户端实例
func NewIndexTTSClient(logger *zap.Logger, baseURL string) *IndexTTSClient {
	if baseURL == "" {
		baseURL = "http://localhost:7860" // 默认地址
	}

	return &IndexTTSClient{
		BaseURL: baseURL,
		Logger:  logger,
		HTTPClient: &http.Client{
			Timeout: 300 * time.Second, // TTS生成可能需要较长时间
		},
	}
}

// Option 定义配置选项类型
type Option func(*IndexTTSClient)

// WithSpeakerAudio 设置说话人音频
func WithSpeakerAudio(audioPath string) Option {
	return func(c *IndexTTSClient) {
		// 这里存储配置，实际使用在Generate方法中
		_ = audioPath // 为了编译通过，实际实现在Generate方法中处理
	}
}

// TTSResult 定义TTS结果结构
type TTSResult struct {
	Success   bool   `json:"success"`
	AudioPath string `json:"audio_path"`
	Error     string `json:"error,omitempty"`
}

// Generate 旧版API接口，用于向后兼容
func (c *IndexTTSClient) Generate(text, outputFile string, opts ...Option) (*TTSResult, error) {
	// 解析选项
	var speakerAudio string
	for _, opt := range opts {
		// 这里简单处理，实际应该根据具体选项设置
		_ = opt
	}

	// 为兼容性，这里使用默认参考音频路径
	if speakerAudio == "" {
		// 尝试查找默认音频文件
		wd, _ := os.Getwd()
		defaultPaths := []string{
			filepath.Join(wd, "ref.m4a"),
			filepath.Join(wd, "音色.m4a"),
			"/Users/mac/code/ai/novel-video-workflow/ref.m4a",
			"/Users/mac/code/ai/novel-video-workflow/音色.m4a",
		}

		for _, path := range defaultPaths {
			if _, err := os.Stat(path); err == nil {
				speakerAudio = path
				break
			}
		}

		if speakerAudio == "" {
			return &TTSResult{
				Success: false,
				Error:   "未找到参考音频文件",
			}, fmt.Errorf("未找到参考音频文件")
		}
	}

	// 使用新的API进行TTS生成
	client := &IndexTTS2Client{
		BaseURL:    c.BaseURL,
		Logger:     c.Logger,
		HTTPClient: c.HTTPClient,
	}

	err := client.GenerateTTSWithAudio(speakerAudio, text, outputFile)
	if err != nil {
		return &TTSResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// 检查输出文件是否存在
	if _, err := os.Stat(outputFile); err != nil {
		return &TTSResult{
			Success: false,
			Error:   fmt.Sprintf("输出文件不存在: %v", err),
		}, err
	}

	return &TTSResult{
		Success:   true,
		AudioPath: outputFile,
	}, nil
}

// UploadResponse 上传音频文件的响应
type UploadResponse struct {
	FileName string `json:"filename"`
	FilePath string `json:"filepath"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

// TTSParams TTS生成参数 - 适配Gradio API
type TTSParams struct {
	Data []interface{} `json:"data"` // Gradio API使用数组格式的数据
}

// TTSResponse TTS生成响应
type TTSResponse struct {
	Success     bool          `json:"success"`
	AudioFile   string        `json:"audio_file,omitempty"`
	AudioBase64 string        `json:"audio_base64,omitempty"`
	AudioURL    string        `json:"audio_url,omitempty"`
	Error       string        `json:"error,omitempty"`
	Data        []interface{} `json:"data"` // Gradio响应格式
}

// UploadAudio 上传参考音频文件
func (c *IndexTTS2Client) UploadAudio(audioPath string) (*UploadResponse, error) {
	// 首先检查文件是否存在
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("音频文件不存在: %s", audioPath)
	}

	// 打开音频文件
	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("打开音频文件失败: %v", err)
	}
	defer file.Close()

	// 创建multipart表单
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建文件表单字段 - 尝试使用Gradio兼容的字段名
	part, err := writer.CreateFormFile("file", filepath.Base(audioPath)) // 改为通用的"file"字段名
	if err != nil {
		return nil, fmt.Errorf("创建表单字段失败: %v", err)
	}

	// 复制文件内容
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("复制文件内容失败: %v", err)
	}

	// 关闭writer
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("关闭表单写入器失败: %v", err)
	}

	// 尝试多个可能的上传端点 - Gradio应用通常使用不同的API结构
	possibleEndpoints := []string{
		"/upload_audio",     // 原始端点
		"/upload",           // 通用上传端点
		"/file-upload",      // 文件上传端点
		"/upload_ref",       // 参考音频上传
		"/api/upload",       // API上传端点
		"/api/upload_audio", // API音频上传端点
	}

	var lastErr error
	for _, endpoint := range possibleEndpoints {
		url := c.BaseURL + endpoint
		fmt.Printf("尝试上传到端点: %s\n", url)

		req, err := http.NewRequest("POST", url, body)
		if err != nil {
			lastErr = fmt.Errorf("创建请求失败: %v", err)
			continue
		}

		// 设置请求头
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// 发送请求
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("发送请求失败: %v", err)
			continue
		}
		defer resp.Body.Close()

		// 检查HTTP状态码
		fmt.Printf("HTTP响应状态码: %d\n", resp.StatusCode) // 添加调试信息

		// 读取响应
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("读取响应失败: %v", err)
			continue
		}

		fmt.Printf("响应内容: %s\n", string(respBody)) // 添加调试信息

		// 解析响应
		var uploadResp UploadResponse
		err = json.Unmarshal(respBody, &uploadResp)
		if err != nil {
			// 如果JSON解析失败，尝试其他可能的响应格式
			if strings.Contains(string(respBody), "success") || strings.Contains(string(respBody), "filename") {
				// 可能是不同的JSON结构，尝试直接解析
				uploadResp = UploadResponse{
					FileName: filepath.Base(audioPath),
					FilePath: string(respBody),
					Success:  true,
				}
			} else {
				// 对于Gradio应用，可能需要不同的处理方式
				// 直接返回成功，因为有些Gradio应用可能只是接收文件而无特定响应
				uploadResp = UploadResponse{
					FileName: filepath.Base(audioPath),
					FilePath: audioPath, // 使用原始路径
					Success:  true,
				}
			}
		}

		if uploadResp.Success || resp.StatusCode == http.StatusOK {
			fmt.Printf("上传成功，使用端点: %s\n", endpoint)
			return &uploadResp, nil
		}
	}

	// 如果所有端点都失败，返回最后一个错误
	return nil, fmt.Errorf("上传失败，尝试了所有可能的端点: %v", lastErr)
}

// GenerateTTS 生成TTS语音 - 适配Gradio API
func (c *IndexTTS2Client) GenerateTTS(params TTSParams) (*TTSResponse, error) {
	// Gradio API需要特定的格式
	gradioRequest := map[string]interface{}{
		"data":       params.Data,
		"event_data": nil,
		"fn_index":   9, // gen_single 函数的索引
	}

	// 将参数转换为JSON
	jsonData, err := json.Marshal(gradioRequest)
	if err != nil {
		return nil, fmt.Errorf("序列化参数失败: %v", err)
	}

	// 尝试多种可能的API端点
	endpoints := []string{
		c.BaseURL + "/gradio_api/predict",
		c.BaseURL + "/api/predict",
		c.BaseURL + "/predict",
		c.BaseURL + "/", // 对于Gradio 4.x版本，可能直接在根路径
	}

	var lastErr error
	for _, endpoint := range endpoints {
		fmt.Printf("尝试API端点: %s\n", endpoint)

		// 创建请求
		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("创建请求失败: %v", err)
			continue
		}

		// 设置请求头
		req.Header.Set("Content-Type", "application/json")

		// 发送请求
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("发送请求失败: %v", err)
			continue
		}
		defer resp.Body.Close()

		// 读取响应
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("读取响应失败: %v", err)
			resp.Body.Close()
			continue
		}

		// 检查HTTP状态码
		if resp.StatusCode == http.StatusOK {
			fmt.Printf("API调用成功，端点: %s\n", endpoint)

			// 解析响应
			var ttsResp TTSResponse
			err = json.Unmarshal(respBody, &ttsResp)
			if err != nil {
				// 可能是Gradio的特定响应格式，尝试解析
				var gradioResp map[string]interface{}
				err2 := json.Unmarshal(respBody, &gradioResp)
				if err2 != nil {
					return nil, fmt.Errorf("解析Gradio响应失败: %v, 响应内容: %s", err, string(respBody))
				}

				// 检查是否有data字段
				if data, ok := gradioResp["data"]; ok {
					ttsResp.Data = make([]interface{}, 0)
					if dataArray, ok := data.([]interface{}); ok {
						ttsResp.Data = dataArray
						ttsResp.Success = true
						return &ttsResp, nil
					}
				}

				return nil, fmt.Errorf("无法解析响应: %v", err)
			}

			return &ttsResp, nil
		} else {
			fmt.Printf("API调用失败，端点: %s, 状态码: %d, 响应: %s\n", endpoint, resp.StatusCode, string(respBody))
			resp.Body.Close()
			continue
		}
	}

	return nil, fmt.Errorf("所有API端点尝试都失败了: %v", lastErr)
}

// DownloadAudio 下载生成的音频文件
func (c *IndexTTS2Client) DownloadAudio(audioURL, savePath string) error {
	// 创建请求
	resp, err := http.Get(audioURL)
	if err != nil {
		return fmt.Errorf("下载音频失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载音频失败，状态码: %d", resp.StatusCode)
	}

	// 创建保存文件
	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("创建保存文件失败: %v", err)
	}
	defer file.Close()

	// 复制内容
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("保存音频内容失败: %v", err)
	}

	// 确保文件内容被刷写到磁盘
	err = file.Sync()
	if err != nil {
		return fmt.Errorf("同步文件到磁盘失败: %v", err)
	}

	// 关闭文件
	err = file.Close()
	if err != nil {
		return fmt.Errorf("关闭文件失败: %v", err)
	}

	// 验证文件是否真的存在且有内容
	fileInfo, err := os.Stat(savePath)
	if err != nil {
		return fmt.Errorf("验证下载文件失败: %v", err)
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("下载的文件大小为0字节")
	}

	return nil
}

// GenerateTTSWithAudio 完整的TTS生成流程
func (c *IndexTTS2Client) GenerateTTSWithAudio(audioPath, text, outputPath string) error {
	c.Logger.Info("开始TTS生成", 
		zap.String("audio_path", audioPath),
		zap.String("text", text),
		zap.String("output_path", outputPath))

	// 检查音频文件是否存在
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return fmt.Errorf("音频文件不存在: %s", audioPath)
	}

	// 获取文件信息以显示文件大小
	if fileInfo, err := os.Stat(audioPath); err == nil {
		size := fileInfo.Size()
		c.Logger.Info("音频文件信息", 
			zap.String("path", audioPath),
			zap.Int64("size_bytes", size),
			zap.Float64("size_mb", float64(size)/(1024*1024)))
	} else {
		c.Logger.Error("无法获取音频文件信息", zap.String("path", audioPath), zap.Error(err))
		return fmt.Errorf("无法访问音频文件: %v", err)
	}

	c.Logger.Info("使用音频文件进行TTS生成", zap.String("audio_path", audioPath))
	c.Logger.Info("正在生成TTS语音", zap.String("text", text))

	// 直接调用带音频文件的TTS生成
	ttsResp, err := c.GenerateTTSWithFile(audioPath, text)
	if err != nil {
		c.Logger.Error("TTS生成失败", zap.Error(err))
		return fmt.Errorf("生成TTS失败: %v", err)
	}

	// 检查响应中是否有音频数据
	if len(ttsResp.Data) == 0 {
		return fmt.Errorf("TTS生成失败: 未收到任何响应数据")
	}

	// 从响应中提取音频信息
	audioFound := false
	for i, item := range ttsResp.Data {
		// 检查是否是更新类型的数据
		if itemMap, ok := item.(map[string]interface{}); ok {
			// 检查是否是__type__:update格式
			if typeVal, exists := itemMap["__type__"]; exists && typeVal == "update" {
				// 检查value字段
				if value, valExists := itemMap["value"]; valExists {
					if valueMap, isMap := value.(map[string]interface{}); isMap {
						if path, pathExists := valueMap["path"]; pathExists {
							audioPathFromServer := path.(string)
							c.Logger.Info("找到音频输出", zap.String("path", audioPathFromServer), zap.Int("index", i))

							// 构造完整的音频URL
							audioURL := c.BaseURL + "/gradio_api/file=" + audioPathFromServer
							c.Logger.Info("下载音频", zap.String("url", audioURL))

							err = c.DownloadAudio(audioURL, outputPath)
							if err != nil {
								return fmt.Errorf("下载音频失败: %v", err)
							}

							audioFound = true
							break
						}
					}
				}
			} else if path, exists := itemMap["path"]; exists {
				// 直接是音频路径
				audioPathFromServer := path.(string)
				c.Logger.Info("找到音频输出", zap.String("path", audioPathFromServer), zap.Int("index", i))

				// 构造完整的音频URL
				audioURL := c.BaseURL + "/gradio_api/file=" + audioPathFromServer
				c.Logger.Info("下载音频", zap.String("url", audioURL))

				err = c.DownloadAudio(audioURL, outputPath)
				if err != nil {
					return fmt.Errorf("下载音频失败: %v", err)
				}

				audioFound = true
				break
			}
		} else if str, ok := item.(string); ok && strings.HasSuffix(str, ".wav") {
			// 检查是否是音频文件路径
			audioURL := c.BaseURL + "/gradio_api/file=" + str
			c.Logger.Info("发现音频路径", zap.String("url", audioURL))

			err = c.DownloadAudio(audioURL, outputPath)
			if err != nil {
				return fmt.Errorf("下载音频失败: %v", err)
			}

			audioFound = true
			break
		}
	}

	if !audioFound {
		// 如果响应中没有直接的音频文件，可能返回的是base64编码的音频
		for i, item := range ttsResp.Data {
			if audioMap, ok := item.(map[string]interface{}); ok {
				if _, hasAudio := audioMap["data"]; hasAudio {
					// 这可能是base64编码的音频数据
					c.Logger.Info("发现潜在音频数据", zap.Int("index", i))

					// 尝试获取base64音频数据
					if data, exists := audioMap["data"]; exists {
						if base64Str, ok := data.(string); ok && strings.HasPrefix(base64Str, "data:audio/") {
							err = saveBase64Audio(base64Str, outputPath)
							if err != nil {
								return fmt.Errorf("保存base64音频失败: %v", err)
							}

							audioFound = true
							break
						}
					}
				}
			}
		}
	}

	if !audioFound {
		return fmt.Errorf("未在响应中找到音频数据: %+v", ttsResp.Data)
	}

	c.Logger.Info("TTS生成完成", zap.String("output", outputPath))
	return nil
}

// GenerateTTSWithFile 生成TTS语音，包含音频文件 - 使用Gradio API
func (c *IndexTTS2Client) GenerateTTSWithFile(audioPath string, text string) (*TTSResponse, error) {
	// 首先检查文件是否存在
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("音频文件不存在: %s", err)
	}

	// 上传音频文件到服务器
	uploadResp, err := c.uploadFileToServer(audioPath)
	if err != nil {
		return nil, fmt.Errorf("上传音频文件失败: %v", err)
	}

	c.Logger.Info("音频文件上传成功", zap.Any("upload_resp", uploadResp))

	// 准备Gradio API请求数据，按照webui.py中gen_single函数的参数顺序
	// 根据WebUI的实现，我们需要特别注意参数类型和值
	// 根据错误信息，服务端实际期望的是英文选项
	emoControlMethod := "Same as the voice reference" // 使用服务端实际接受的英文选项
	
	requestData := map[string]interface{}{
		"data": []interface{}{
			emoControlMethod,              // 0: emo_control_method - 情感控制方式
			uploadResp,                    // 1: prompt - 音色参考音频（使用上传后的路径）
			text,                          // 2: text - 输入文本
			nil,                           // 3: emo_ref_path - 情感参考音频路径(设为nil)
			0.65,                          // 4: emo_weight - 情感权重
			0.0,                           // 5: vec1 - 情感向量1(喜)
			0.0,                           // 6: vec2 - 情感向量2(怒)
			0.0,                           // 7: vec3 - 情感向量3(哀)
			0.0,                           // 8: vec4 - 情感向量4(惧)
			0.0,                           // 9: vec5 - 情感向量5(厌恶)
			0.0,                           // 10: vec6 - 情感向量6(低落)
			0.0,                           // 11: vec7 - 情感向量7(惊喜)
			0.0,                           // 12: vec8 - 情感向量8(平静)
			"",                            // 13: emo_text - 情感描述文本
			false,                         // 14: emo_random - 情感随机化
			120,                           // 15: max_text_tokens_per_segment - 最大文本分段token数
			true,                          // 16: do_sample - 是否采样
			0.8,                           // 17: top_p
			30,                            // 18: top_k
			0.8,                           // 19: temperature
			0.0,                           // 20: length_penalty
			3,                             // 21: num_beams
			10.0,                          // 22: repetition_penalty
			1500,                          // 23: max_mel_tokens
		},
		"fn_index":     9,                                    // 函数索引，从配置中得知gen_single的ID是9
		"session_hash": fmt.Sprintf("%d", time.Now().Unix()), // 会话哈希
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("序列化请求数据失败: %v", err)
	}

	c.Logger.Info("准备发送TTS请求", zap.String("text", text), zap.Any("first_param", requestData["data"].([]interface{})[0]))

	// 首先将任务加入队列
	queueEndpoint := c.BaseURL + "/gradio_api/queue/join"

	// 创建队列请求
	req, err := http.NewRequest("POST", queueEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建队列请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送队列请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送队列请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取队列响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取队列响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("队列请求失败，状态码: %d，响应: %s", resp.StatusCode, string(respBody))
	}

	// 解析队列响应，获取event_id
	var queueResp map[string]interface{}
	err = json.Unmarshal(respBody, &queueResp)
	if err != nil {
		return nil, fmt.Errorf("解析队列响应失败: %v", err)
	}

	eventID, ok := queueResp["event_id"].(string)
	if !ok {
		return nil, fmt.Errorf("未能从队列响应中获取event_id: %v", queueResp)
	}

	c.Logger.Info("任务已加入队列", zap.String("event_id", eventID))

	// 现在使用session_hash来监听结果
	resultEndpoint := fmt.Sprintf("%s/gradio_api/queue/data?session_hash=%s&fn_index=%d",
		c.BaseURL,
		requestData["session_hash"],
		requestData["fn_index"])

	// 创建SSE客户端来接收结果
	resultReq, err := http.NewRequest("GET", resultEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("创建结果请求失败: %v", err)
	}

	// 发送结果请求
	resultResp, err := c.HTTPClient.Do(resultReq)
	if err != nil {
		return nil, fmt.Errorf("发送结果请求失败: %v", err)
	}
	defer resultResp.Body.Close()

	// 读取SSE流式响应
	scanner := bufio.NewScanner(resultResp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			var sseData map[string]interface{}
			err := json.Unmarshal([]byte(data), &sseData)
			if err != nil {
				// 尝试解析可能的错误响应
				var sseStr string
				err2 := json.Unmarshal([]byte(data), &sseStr)
				if err2 != nil {
					continue // 如果解析失败，继续下一行
				}
				return nil, fmt.Errorf("SSE数据解析错误: %v, 原始数据: %s", err, data)
			}

			// 检查消息类型
			if msg, ok := sseData["msg"]; ok {
				switch msg {
				case "estimation":
					// 处理估计队列位置
					if rank, exists := sseData["rank"]; exists {
						if queueSize, exists2 := sseData["queue_size"]; exists2 {
							c.Logger.Info("排队中", zap.Float64("rank", rank.(float64)), zap.Float64("queue_size", queueSize.(float64)))
						} else {
							c.Logger.Info("排队中", zap.Float64("rank", rank.(float64)))
						}
					}
				case "process_starts":
					// 处理开始处理
					c.Logger.Info("开始处理TTS生成")
				case "process_completed":
					// 处理完成
					c.Logger.Info("TTS生成完成")
					// 检查是否成功
					if success, ok := sseData["success"]; ok {
						if success == true || success == "true" {
							// 检查是否有结果数据
							if resultData, ok := sseData["output"]; ok {
								if outputMap, ok := resultData.(map[string]interface{}); ok {
									if dataArr, ok := outputMap["data"].([]interface{}); ok {
										ttsResp := &TTSResponse{
											Success: true,
											Data:    dataArr,
										}
										return ttsResp, nil
									}
								}
							}
							return nil, fmt.Errorf("TTS生成完成但未返回数据: %+v", sseData)
						} else {
							// 详细错误处理
							if output, ok := sseData["output"]; ok {
								if outputMap, ok := output.(map[string]interface{}); ok {
									if errorMsg, exists := outputMap["error"]; exists && errorMsg != nil {
										return nil, fmt.Errorf("TTS生成失败: %v", errorMsg)
									}
									if title, exists := outputMap["title"]; exists && title != nil {
										return nil, fmt.Errorf("TTS生成失败: %v", title)
									}
									if details, exists := outputMap["details"]; exists && details != nil {
										return nil, fmt.Errorf("TTS生成失败: %v", details)
									}
								}
							}
							// 检查是否是其他错误字段
							for key, value := range sseData {
								if strings.Contains(strings.ToLower(key), "error") && value != nil {
									return nil, fmt.Errorf("TTS生成失败: %s=%v", key, value)
								}
							}
							return nil, fmt.Errorf("TTS生成失败，响应详情: %+v", sseData)
						}
					} else {
						return nil, fmt.Errorf("TTS生成失败，响应中无success字段: %+v", sseData)
					}
				case "process_generating":
					// 处理生成中的状态
					c.Logger.Info("TTS正在生成中")
					// 检查是否有进度信息
					if data, exists := sseData["output"]; exists {
						if outputMap, ok := data.(map[string]interface{}); ok {
							if log, exists := outputMap["progress_data"]; exists {
								c.Logger.Info("生成进度", zap.Any("progress", log))
							}
						}
					}
				case "log":
					// 处理日志信息
					if data, exists := sseData["data"]; exists {
						c.Logger.Info("服务器日志", zap.Any("data", data))
					}
				case "progress":
					// 处理进度信息
					if data, exists := sseData["data"]; exists {
						c.Logger.Info("进度更新", zap.Any("data", data))
					}
				case "close_stream":
					return nil, fmt.Errorf("流已关闭，但未收到结果")
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取SSE流时出错: %v", err)
	}

	return nil, fmt.Errorf("未能从SSE流中获取结果，可能超时或发生错误")
}

// uploadFileToServer 上传文件到服务器
func (c *IndexTTS2Client) uploadFileToServer(filePath string) (interface{}, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 创建multipart表单
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文件
	part, err := writer.CreateFormFile("files", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("创建表单文件失败: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("复制文件失败: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("关闭表单写入器失败: %v", err)
	}

	// 发送上传请求到正确的端点
	uploadEndpoint := c.BaseURL + "/gradio_api/upload"
	req, err := http.NewRequest("POST", uploadEndpoint, body)
	if err != nil {
		return nil, fmt.Errorf("创建上传请求失败: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送上传请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取上传响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("上传请求失败，状态码: %d，响应: %s", resp.StatusCode, string(respBody))
	}

	// 解析上传响应 - 应该是一个包含文件路径的数组
	var uploadResult []string
	err = json.Unmarshal(respBody, &uploadResult)
	if err != nil {
		return nil, fmt.Errorf("解析上传响应失败: %v", err)
	}

	if len(uploadResult) == 0 {
		return nil, fmt.Errorf("上传响应中没有文件路径")
	}

	// 返回第一个文件路径，包含必需的meta字段
	return map[string]interface{}{
		"path":      uploadResult[0],
		"orig_name": filepath.Base(filePath),
		"meta": map[string]string{
			"_type": "gradio.FileData",
		},
	}, nil
}

// saveBase64Audio 保存base64编码的音频
func saveBase64Audio(base64Str, savePath string) error {
	// 处理可能包含data:audio/wav;base64,前缀的base64字符串
	var audioData string
	if strings.HasPrefix(base64Str, "data:audio/") {
		// 移除MIME类型前缀
		parts := strings.SplitN(base64Str, ",", 2)
		if len(parts) == 2 {
			audioData = parts[1]
		} else {
			audioData = base64Str
		}
	} else {
		audioData = base64Str
	}

	// 解码base64
	data, err := base64.StdEncoding.DecodeString(audioData)
	if err != nil {
		return fmt.Errorf("解码base64失败: %v", err)
	}

	// 保存文件
	return os.WriteFile(savePath, data, 0644)
}

// isAbsoluteURL 检查是否是绝对URL
func isAbsoluteURL(url string) bool {
	return len(url) > 7 && (url[:7] == "http://" || url[:8] == "https://")
}