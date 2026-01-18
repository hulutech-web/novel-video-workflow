package web_server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"novel-video-workflow/pkg/broadcast"
	"novel-video-workflow/pkg/capcut"
	"novel-video-workflow/pkg/tools/aegisub"
	"novel-video-workflow/pkg/tools/file"
	"novel-video-workflow/pkg/tools/indextts2"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	mcp_pkg "novel-video-workflow/pkg/mcp"
	"novel-video-workflow/pkg/tools/drawthings"
	workflow_pkg "novel-video-workflow/pkg/workflow"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 存储WebSocket连接
var clients = make(map[*websocket.Conn]bool)

// ToolInfo 结构存储MCP工具的信息
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
}

// 存储所有可用的MCP工具
var mcpTools []ToolInfo
var mcpServerInstance *mcp_pkg.Server

// 启动MCP服务器
func startMCPServer() error {
	// 创建logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 创建工作流处理器
	processor, err := workflow_pkg.NewProcessor(logger)
	if err != nil {
		return fmt.Errorf("创建工作流处理器失败: %v", err)
	}

	// 创建MCP服务器
	server, err := mcp_pkg.NewServer(processor, logger)
	if err != nil {
		return fmt.Errorf("创建MCP服务器失败: %v", err)
	}
	mcpServerInstance = server

	// 获取可用工具列表
	availableTools := server.GetHandler().GetToolNames()
	log.Printf("MCP服务器启动成功，加载了 %d 个工具", len(availableTools))

	// 为每个工具创建描述信息
	for _, toolName := range availableTools {
		description := getToolDescription(toolName) // 获取工具描述
		mcpTools = append(mcpTools, ToolInfo{
			Name:        toolName,
			Description: description,
			Path:        fmt.Sprintf("./mcp-tools/%s.yaml", toolName),
		})
	}

	log.Printf("Loaded %d MCP tools", len(mcpTools))

	return nil
}

func loadToolsList() {
	// 启动MCP服务器
	if err := startMCPServer(); err != nil {
		log.Printf("启动MCP服务器失败: %v", err)
		// 即使启动失败，也要提供默认工具列表
		fallbackToolList()
		return
	}
}

// fallbackToolList 提供备用工具列表
func fallbackToolList() {
	descriptions := map[string]string{
		"generate_indextts2_audio":                    "使用IndexTTS2生成音频文件，具有高级语音克隆功能",
		"generate_subtitles_from_indextts2":           "使用Aegisub从IndexTTS2音频和提供的文本生成字幕(SRT)",
		"file_split_novel_into_chapters":              "根据章节标记将小说文件拆分为单独的章节文件夹和文件",
		"generate_image_from_text":                    "使用DrawThings API根据文本生成图像，采用悬疑风格",
		"generate_image_from_image":                   "使用DrawThings API根据参考图像生成图像，采用悬疑风格",
		"generate_images_from_chapter":                "使用DrawThings API根据章节文本生成图像，采用悬疑风格",
		"generate_images_from_chapter_with_ai_prompt": "使用AI生成提示词和DrawThings API根据章节文本生成图像，采用悬疑风格",
	}

	defaultTools := []string{
		"generate_indextts2_audio",
		"generate_subtitles_from_indextts2",
		"file_split_novel_into_chapters",
		"generate_image_from_text",
		"generate_image_from_image",
		"generate_images_from_chapter",
		"generate_images_from_chapter_with_ai_prompt",
	}

	for _, toolName := range defaultTools {
		description, exists := descriptions[toolName]
		if !exists {
			description = fmt.Sprintf("MCP工具: %s", toolName)
		}
		mcpTools = append(mcpTools, ToolInfo{
			Name:        toolName,
			Description: description,
			Path:        fmt.Sprintf("./mcp-tools/%s.yaml", toolName),
		})
	}

	log.Printf("加载了 %d 个备用工具", len(mcpTools))
}

// getToolDescription 根据工具名称获取描述
func getToolDescription(toolName string) string {
	descriptions := map[string]string{
		"generate_indextts2_audio":                    "使用IndexTTS2生成音频文件，具有高级语音克隆功能",
		"generate_subtitles_from_indextts2":           "使用Aegisub从IndexTTS2音频和提供的文本生成字幕(SRT)",
		"file_split_novel_into_chapters":              "根据章节标记将小说文件拆分为单独的章节文件夹和文件",
		"generate_image_from_text":                    "使用DrawThings API根据文本生成图像，采用悬疑风格",
		"generate_image_from_image":                   "使用DrawThings API根据参考图像生成图像，采用悬疑风格",
		"generate_images_from_chapter":                "使用DrawThings API根据章节文本生成图像，采用悬疑风格",
		"generate_images_from_chapter_with_ai_prompt": "使用AI生成提示词和DrawThings API根据章节文本生成图像，采用悬疑风格",
	}

	if desc, exists := descriptions[toolName]; exists {
		return desc
	}

	return fmt.Sprintf("MCP工具: %s", toolName)
}

// Gin路由处理函数
func homePage(c *gin.Context) {
	tmpl := template.Must(template.ParseFiles("./templates/index.html"))
	tmpl.Execute(c.Writer, nil)
}

func wsEndpoint(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer ws.Close()

	// 添加客户端到全局广播服务
	clientChan := broadcast.GlobalBroadcastService.RegisterClient(ws)
	defer func() {
		// 从全局广播服务注销客户端
		client := &broadcast.Client{Conn: ws, Send: clientChan}
		broadcast.GlobalBroadcastService.UnregisterClient(client)
	}()

	// 启动goroutine处理来自广播服务的消息
	go func() {
		for message := range clientChan {
			// 直接发送消息，因为现在BroadcastMessage已经包含了前端期望的字段
			if err := ws.WriteJSON(message); err != nil {
				log.Printf("Error sending message to client: %v", err)
				return
			}
		}
	}()

	for {
		var msg map[string]interface{}
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}
		// 可以处理从客户端发送的消息，如果需要的话
	}
}

func apiToolsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, mcpTools)
}

func apiExecuteHandler(c *gin.Context) {
	var reqBody map[string]interface{} // 修改为interface{}以支持不同类型参数
	err := json.NewDecoder(c.Request.Body).Decode(&reqBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	toolName, ok := reqBody["toolName"].(string)
	if !ok || toolName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing toolName"})
		return
	}

	// 启动MCP工具执行
	go func() {
		// 检查工具是否存在
		toolExists := false
		for _, tool := range mcpTools {
			if tool.Name == toolName {
				toolExists = true
				break
			}
		}

		if !toolExists {
			return
		}
		// 对于generate_indextts2_audio工具，处理文本输入和音频生成
		if toolName == "generate_indextts2_audio" {
			text, ok := reqBody["text"].(string)
			if !ok || text == "" {
				text = "这是一个默认的测试文本。" // 默认文本
			}

			referenceAudio, ok := reqBody["reference_audio"].(string)
			if !ok || referenceAudio == "" {
				referenceAudio = "./assets/ref_audio/ref.m4a" // 默认参考音频
			}

			outputFile, ok := reqBody["output_file"].(string)
			if !ok || outputFile == "" {
				outputFile = fmt.Sprintf("./output/audio_%d.wav", time.Now().Unix()) // 默认输出文件
			}

			// 确保输出目录存在
			outputDir := filepath.Dir(outputFile)
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				broadcast.GlobalBroadcastService.SendLog("indextts2", fmt.Sprintf("[%s] 创建输出目录失败: %v", toolName, err), broadcast.GetTimeStr())
				return
			}

			// 检查参考音频是否存在
			if _, err := os.Stat(referenceAudio); os.IsNotExist(err) {
				// 尝试其他可能的默认路径
				possiblePaths := []string{
					"./ref.m4a",
					"./音色.m4a",
					"./assets/ref_audio/ref.m4a",
					"./assets/ref_audio/音色.m4a",
				}

				found := false
				for _, path := range possiblePaths {
					if _, err := os.Stat(path); err == nil {
						referenceAudio = path
						found = true
						break
					}
				}

				if !found {
					broadcast.GlobalBroadcastService.SendLog("indextts2", fmt.Sprintf("[%s] 找不到参考音频文件，请确保存在默认音频文件", toolName), broadcast.GetTimeStr())

					return
				}
			}
			broadcast.GlobalBroadcastService.SendLog("indextts2", fmt.Sprintf("[%s] 使用参考音频: %s", toolName, referenceAudio), broadcast.GetTimeStr())
			broadcast.GlobalBroadcastService.SendLog("indextts2", fmt.Sprintf("[%s] 输入文本: %s", toolName, text), broadcast.GetTimeStr())

			// 检查MCP服务器实例是否存在
			if mcpServerInstance != nil {
				// 获取处理器并直接调用工具
				handler := mcpServerInstance.GetHandler()
				if handler != nil {
					// 创建MockRequest对象
					mockRequest := &mcp_pkg.MockRequest{
						Params: map[string]interface{}{
							"text":            text,
							"reference_audio": referenceAudio,
							"output_file":     outputFile,
						},
					}

					// 调用特定工具处理函数
					result, err := handler.HandleGenerateIndextts2AudioDirect(mockRequest)
					if err != nil {
						return
					}

					// 检查结果
					if success, ok := result["success"].(bool); ok && success {
					} else {
						errorMsg := "未知错误"
						if result["error"] != nil {
							if errStr, ok := result["error"].(string); ok {
								errorMsg = errStr
							}
						}
						broadcast.GlobalBroadcastService.SendLog("indextts2", fmt.Sprintf("[%s] 工具执行失败: %s", toolName, errorMsg), broadcast.GetTimeStr())

					}
				} else {
					broadcast.GlobalBroadcastService.SendLog("indextts2", fmt.Sprintf("[%s] 错误: MCP处理器未初始化", toolName), broadcast.GetTimeStr())

				}
			} else {
				// 如果没有MCP服务器实例，给出提示
				broadcast.GlobalBroadcastService.SendLog("indextts2", fmt.Sprintf("[%s] 错误: MCP服务器未启动。请确保服务已正确初始化。", toolName), broadcast.GetTimeStr())
			}
		} else {
			// 其他工具的处理 - 也需要类似处理
			if mcpServerInstance != nil {
				// 获取处理器并直接调用工具
				handler := mcpServerInstance.GetHandler()
				if handler != nil {
					// 根据工具名称调用相应的处理函数
					var result map[string]interface{}
					var err error

					// 为其他工具传递参数
					params, ok := reqBody["params"].(map[string]interface{})
					if !ok {
						params = make(map[string]interface{})
					}

					// 这里需要根据工具名称调用相应的处理函数
					switch toolName {
					case "generate_subtitles_from_indextts2":
						mockRequest := &mcp_pkg.MockRequest{Params: params}
						result, err = handler.HandleGenerateSubtitlesFromIndextts2Direct(mockRequest)
					case "file_split_novel_into_chapters":
						mockRequest := &mcp_pkg.MockRequest{Params: params}
						result, err = handler.HandleFileSplitNovelIntoChaptersDirect(mockRequest)
					case "generate_image_from_text":
						mockRequest := &mcp_pkg.MockRequest{Params: params}
						result, err = handler.HandleGenerateImageFromTextDirect(mockRequest)
					case "generate_image_from_image":
						mockRequest := &mcp_pkg.MockRequest{Params: params}
						result, err = handler.HandleGenerateImageFromImageDirect(mockRequest)
					case "generate_images_from_chapter_with_ai_prompt":
						// 处理章节图像生成（使用AI提示词）
						chapterText, ok := reqBody["chapter_text"].(string)
						if !ok {
							chapterText = "这是一个默认的章节文本。"
						}

						outputDir, ok := reqBody["output_dir"].(string)
						if !ok {
							outputDir = fmt.Sprintf("./output/chapter_images_%d", time.Now().Unix())
						}

						widthFloat, ok := reqBody["width"].(float64)
						var width int
						if ok {
							width = int(widthFloat)
						} else {
							width = 512 // 默认宽度
						}

						heightFloat, ok := reqBody["height"].(float64)
						var height int
						if ok {
							height = int(heightFloat)
						} else {
							height = 896 // 默认高度
						}

						// 确保输出目录存在
						if err := os.MkdirAll(outputDir, 0755); err != nil {
							return
						}

						// 创建一个自定义的日志记录器，将内部日志广播到前端
						logger, _ := zap.NewProduction()
						defer logger.Sync()

						// 使用自定义的广播日志适配器
						encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
						writeSyncer := zapcore.AddSync(os.Stdout) // 输出到标准输出，同时也会被广播
						broadcastLogger := NewBroadcastLoggerAdapter(toolName, encoder, writeSyncer)
						broadcaster := zap.New(broadcastLogger)

						// 使用带广播功能的日志记录器创建章节图像生成器
						generator := drawthings.NewChapterImageGenerator(broadcaster)

						// 直接调用图像生成方法，而不是通过MCP处理器
						results, err := generator.GenerateImagesFromChapter(chapterText, outputDir, width, height, true)
						if err != nil {
							return
						}

						// 准备结果
						imageFiles := make([]string, len(results))
						paragraphs := make([]string, len(results))
						prompts := make([]string, len(results))

						for i, result := range results {
							imageFiles[i] = result.ImageFile
							paragraphs[i] = result.ParagraphText
							prompts[i] = result.ImagePrompt
						}

						result = map[string]interface{}{
							"success":               true,
							"output_dir":            outputDir,
							"chapter_text_length":   len(chapterText),
							"generated_image_count": len(results),
							"image_files":           imageFiles,
							"paragraphs":            paragraphs,
							"prompts":               prompts,
							"width":                 width,
							"height":                height,
							"is_suspense":           true,
							"tool":                  "drawthings_chapter_txt2img_with_ai_prompt",
						}
					default:
						return
					}

					if err != nil {
						return
					}

					// 记录执行结果
					broadcast.GlobalBroadcastService.SendLog("indextts2", fmt.Sprintf("[%s] 工具执行完成，结果: %+v", toolName, result), broadcast.GetTimeStr())
				} else {
				}
			} else {
			}
		}

	}()

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Tool execution started"})
}

func apiExecuteAllHandler(c *gin.Context) {
	// 执行所有MCP工具
	go func() {
		for _, tool := range mcpTools {

			// 检查MCP服务器实例是否存在
			if mcpServerInstance != nil {
				// 获取处理器并直接调用工具
				handler := mcpServerInstance.GetHandler()
				if handler != nil {
					// 根据工具名称调用相应的处理函数
					var result map[string]interface{}
					var err error

					// 为工具传递默认参数
					params := make(map[string]interface{})

					// 这里需要根据工具名称调用相应的处理函数
					switch tool.Name {
					case "generate_subtitles_from_indextts2":
						mockRequest := &mcp_pkg.MockRequest{Params: params}
						result, err = handler.HandleGenerateSubtitlesFromIndextts2Direct(mockRequest)
					case "file_split_novel_into_chapters":
						mockRequest := &mcp_pkg.MockRequest{Params: params}
						result, err = handler.HandleFileSplitNovelIntoChaptersDirect(mockRequest)
					case "generate_image_from_text":
						mockRequest := &mcp_pkg.MockRequest{Params: params}
						result, err = handler.HandleGenerateImageFromTextDirect(mockRequest)
					case "generate_image_from_image":
						mockRequest := &mcp_pkg.MockRequest{Params: params}
						result, err = handler.HandleGenerateImageFromImageDirect(mockRequest)
					case "generate_indextts2_audio":
						// 对于音频生成工具，使用默认参数
						defaultParams := map[string]interface{}{
							"text":            "这是一个测试音频。",
							"reference_audio": "./assets/ref_audio/ref.m4a",
							"output_file":     fmt.Sprintf("./output/test_%d.wav", time.Now().Unix()),
						}
						mockRequest := &mcp_pkg.MockRequest{Params: defaultParams}
						result, err = handler.HandleGenerateIndextts2AudioDirect(mockRequest)
					default:
						continue
					}

					if err != nil {
						broadcast.GlobalBroadcastService.SendLog("indextts2", err.Error(), broadcast.GetTimeStr())

						continue
					}
					//map转 json
					jsonData, _ := json.Marshal(result)
					broadcast.GlobalBroadcastService.SendLog("indextts2", string(jsonData), broadcast.GetTimeStr())

					// 记录执行结果
				} else {
				}

			} else {
				// 如果没有MCP服务器实例，给出提示
				broadcast.GlobalBroadcastService.SendLog("indextts2", "[提示] 请先启动MCP服务器！", broadcast.GetTimeStr())
			}

		}
	}()

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "All tools execution started"})
}

func apiProcessFolderHandler(c *gin.Context) {
	// 处理上传的文件夹
	go func() {
		broadcast.GlobalBroadcastService.SendLog("上传文件夹", "[工作流] 开始文件夹处理工作流...", broadcast.GetTimeStr())

		// 检查MCP服务器实例是否存在
		if mcpServerInstance != nil {
			// 获取处理器并直接调用工具
			handler := mcpServerInstance.GetHandler()
			if handler != nil {
				// 模拟工作流处理
				for _, tool := range mcpTools {
					broadcast.GlobalBroadcastService.SendLog("上传文件夹", fmt.Sprintf("[%s] 使用 %s 处理...", tool.Name, tool.Name), broadcast.GetTimeStr())

					// 根据工具名称调用相应的处理函数
					var result map[string]interface{}
					var err error

					// 为工具传递默认参数
					params := make(map[string]interface{})

					// 这里需要根据工具名称调用相应的处理函数
					switch tool.Name {
					case "generate_subtitles_from_indextts2":
						mockRequest := &mcp_pkg.MockRequest{Params: params}
						result, err = handler.HandleGenerateSubtitlesFromIndextts2Direct(mockRequest)
					case "file_split_novel_into_chapters":
						mockRequest := &mcp_pkg.MockRequest{Params: params}
						result, err = handler.HandleFileSplitNovelIntoChaptersDirect(mockRequest)
					case "generate_image_from_text":
						mockRequest := &mcp_pkg.MockRequest{Params: params}
						result, err = handler.HandleGenerateImageFromTextDirect(mockRequest)
					case "generate_image_from_image":
						mockRequest := &mcp_pkg.MockRequest{Params: params}
						result, err = handler.HandleGenerateImageFromImageDirect(mockRequest)
					case "generate_indextts2_audio":
						// 对于音频生成工具，使用默认参数
						defaultParams := map[string]interface{}{
							"text":            "这是文件夹处理的一部分。",
							"reference_audio": "./assets/ref_audio/ref.m4a",
							"output_file":     fmt.Sprintf("./output/folder_process_%d.wav", time.Now().Unix()),
						}
						mockRequest := &mcp_pkg.MockRequest{Params: defaultParams}
						result, err = handler.HandleGenerateIndextts2AudioDirect(mockRequest)
					default:
						broadcast.GlobalBroadcastService.SendLog("上传文件夹", fmt.Sprintf("[%s] 暂不支持直接调用工具: %s", tool.Name, tool.Name), broadcast.GetTimeStr())

						continue
					}

					if err != nil {
						broadcast.GlobalBroadcastService.SendLog("上传文件夹", fmt.Sprintf("[%s] 工具执行失败: %v", tool.Name, err), broadcast.GetTimeStr())

					} else {
						// 记录执行结果
						broadcast.GlobalBroadcastService.SendLog("上传完毕", fmt.Sprintf("[%s] 工具执行完成，结果: %+v", tool.Name, result), broadcast.GetTimeStr())

					}
				}
			} else {
				broadcast.GlobalBroadcastService.SendLog("[工作流] 错误", "MCP处理器未初始化", broadcast.GetTimeStr())

			}
		} else {
			broadcast.GlobalBroadcastService.SendLog("[工作流] 错误", "MCP服务器未启动。请确保服务已正确初始化", broadcast.GetTimeStr())
		}
		broadcast.GlobalBroadcastService.SendLog("处理完成", "[工作流] 文件夹处理完成", broadcast.GetTimeStr())

	}()

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Folder processing started"})
}

// fileListHandler 返回指定目录中的文件列表
func fileListHandler(c *gin.Context) {
	dir := c.Query("dir")

	// 获取项目根目录
	wd, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取当前工作目录", "status": "error"})
		return
	}

	projectRoot := wd
	if strings.HasSuffix(wd, "/cmd/web_server") {
		projectRoot = filepath.Dir(filepath.Dir(wd)) // 回退两级到项目根目录
	}

	if dir == "" {
		// 默认目录使用项目根路径
		dir = filepath.Join(projectRoot, "input")
	} else {
		// 解码URL参数
		decodedDir, err := url.QueryUnescape(dir)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid directory path", "status": "error"})
			return
		}

		// 如果是相对路径格式（如./input），将其转换为绝对路径
		if strings.HasPrefix(decodedDir, "./") {
			if strings.HasPrefix(decodedDir, "./input") {
				dir = filepath.Join(projectRoot, decodedDir[2:]) // 移除开头的"./"
			} else if strings.HasPrefix(decodedDir, "./output") {
				dir = filepath.Join(projectRoot, decodedDir[2:]) // 移除开头的"./"
			} else {
				c.JSON(http.StatusForbidden, gin.H{"error": "Invalid directory path", "status": "error"})
				return
			}
		} else {
			// 如果已经是绝对路径，直接使用
			dir = decodedDir
		}
	}

	// 确保路径安全，防止路径遍历攻击
	cleanDir := filepath.Clean(dir)

	// 构建允许的路径前缀
	allowedInputPrefix := filepath.Join(projectRoot, "input")
	allowedOutputPrefix := filepath.Join(projectRoot, "output")

	// 检查路径是否在允许的范围内
	isValidPath := strings.HasPrefix(cleanDir, allowedInputPrefix+"/") ||
		strings.HasPrefix(cleanDir, allowedOutputPrefix+"/") ||
		cleanDir == allowedInputPrefix ||
		cleanDir == allowedOutputPrefix

	if !isValidPath {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied", "status": "error"})
		return
	}

	files, err := os.ReadDir(cleanDir)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Directory not found", "status": "error"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": "error"})
		return
	}

	var fileList []map[string]interface{}
	for _, file := range files {
		filePath := filepath.Join(cleanDir, file.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		fileInfo := map[string]interface{}{
			"name":    file.Name(),
			"size":    info.Size(),
			"modTime": info.ModTime().Format(time.RFC3339),
			"isDir":   file.IsDir(),
			"type":    getFileType(file.Name()),
		}
		fileList = append(fileList, fileInfo)
	}

	c.JSON(http.StatusOK, gin.H{"files": fileList, "directory": cleanDir})
}

// fileContentHandler 返回文件的内容
func fileContentHandler(c *gin.Context) {
	pathParam := c.Query("path")

	// 获取项目根目录
	wd, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取当前工作目录", "status": "error"})
		return
	}

	projectRoot := wd
	if strings.HasSuffix(wd, "/cmd/web_server") {
		projectRoot = filepath.Dir(filepath.Dir(wd)) // 回退两级到项目根目录
	}

	if pathParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File path is required", "status": "error"})
		return
	}

	// 解码URL参数
	decodedPath, err := url.QueryUnescape(pathParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file path", "status": "error"})
		return
	}

	// 如果是相对路径格式（如./input/file.txt），将其转换为绝对路径
	var cleanPath string
	if strings.HasPrefix(decodedPath, "./") {
		if strings.HasPrefix(decodedPath, "./input") || strings.HasPrefix(decodedPath, "./output") {
			cleanPath = filepath.Join(projectRoot, decodedPath[2:]) // 移除开头的"./"
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid file path", "status": "error"})
			return
		}
	} else {
		// 如果已经是绝对路径，直接使用
		cleanPath = decodedPath
	}

	// 确保路径安全，防止路径遍历攻击
	cleanPath = filepath.Clean(cleanPath)

	// 构建允许的路径前缀
	allowedInputPrefix := filepath.Join(projectRoot, "input")
	allowedOutputPrefix := filepath.Join(projectRoot, "output")

	// 检查路径是否在允许的范围内
	isValidPath := strings.HasPrefix(cleanPath, allowedInputPrefix+"/") ||
		strings.HasPrefix(cleanPath, allowedOutputPrefix+"/") ||
		cleanPath == allowedInputPrefix ||
		cleanPath == allowedOutputPrefix

	if !isValidPath {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied", "status": "error"})
		return
	}

	// 检查文件类型，只允许预览特定类型的文件
	fileExt := strings.ToLower(filepath.Ext(cleanPath))
	allowedExts := map[string]bool{
		".txt":  true,
		".md":   true,
		".json": true,
		".yaml": true,
		".yml":  true,
		".xml":  true,
		".csv":  true,
		".log":  true,
	}

	if !allowedExts[fileExt] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File type not supported for preview", "status": "error"})
		return
	}

	content, err := os.ReadFile(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found", "status": "error"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": "error"})
		return
	}

	c.Data(http.StatusOK, "text/plain; charset=utf-8", content)
}

// fileDeleteHandler 删除指定的文件或目录
func fileDeleteHandler(c *gin.Context) {
	pathParam := c.Query("path")

	// 获取项目根目录
	wd, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取当前工作目录", "status": "error"})
		return
	}

	projectRoot := wd
	if strings.HasSuffix(wd, "/cmd/web_server") {
		projectRoot = filepath.Dir(filepath.Dir(wd)) // 回退两级到项目根目录
	}

	if pathParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File path is required", "status": "error"})
		return
	}

	// 解码URL参数
	decodedPath, err := url.QueryUnescape(pathParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file path", "status": "error"})
		return
	}

	// 如果是相对路径格式（如./input/file.txt），将其转换为绝对路径
	var cleanPath string
	if strings.HasPrefix(decodedPath, "./") {
		if strings.HasPrefix(decodedPath, "./input") || strings.HasPrefix(decodedPath, "./output") {
			cleanPath = filepath.Join(projectRoot, decodedPath[2:]) // 移除开头的"./"
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid file path", "status": "error"})
			return
		}
	} else {
		// 如果已经是绝对路径，直接使用
		cleanPath = decodedPath
	}

	// 确保路径安全，防止路径遍历攻击
	cleanPath = filepath.Clean(cleanPath)

	// 构建允许的路径前缀
	allowedInputPrefix := filepath.Join(projectRoot, "input")
	allowedOutputPrefix := filepath.Join(projectRoot, "output")

	// 检查路径是否在允许的范围内
	isValidPath := strings.HasPrefix(cleanPath, allowedInputPrefix+"/") ||
		strings.HasPrefix(cleanPath, allowedOutputPrefix+"/") ||
		cleanPath == allowedInputPrefix ||
		cleanPath == allowedOutputPrefix

	if !isValidPath {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied", "status": "error"})
		return
	}

	// 确认文件或目录存在
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File or directory does not exist", "status": "error"})
		return
	}

	err = os.RemoveAll(cleanPath) // 使用RemoveAll可以删除非空目录
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file or directory: " + err.Error(), "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "File or directory deleted successfully"})
}

// fileUploadHandler 上传文件到指定目录
func fileUploadHandler(c *gin.Context) {
	// 解析 multipart form (32MB max)
	err := c.Request.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to parse form", "status": "error"})
		return
	}

	dir := c.PostForm("dir")
	if dir == "" {
		dir = "./input" // 默认目录
	}

	// 确保路径安全，防止路径遍历攻击
	cleanDir := filepath.Clean(dir)

	// 获取当前工作目录作为基础路径
	workDir, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get working directory", "status": "error"})
		return
	}

	// 尝试找到项目根目录
	projectRoot := workDir
	// 如果当前在cmd/web_server目录下，向上两级到达项目根目录
	if strings.HasSuffix(workDir, "/cmd/web_server") {
		projectRoot = filepath.Dir(filepath.Dir(workDir))
	}

	// 检查目录是否在允许的范围内 - 只允许input目录
	allowedInputDir := filepath.Join(projectRoot, "input")

	// 处理相对路径和绝对路径的情况
	var cleanTargetDir string
	if strings.HasPrefix(cleanDir, "./") {
		// 如果是 ./ 开头的相对路径，转换为绝对路径
		cleanTargetDir = filepath.Clean(filepath.Join(projectRoot, cleanDir[2:]))
	} else if strings.HasPrefix(cleanDir, "input/") || cleanDir == "input" {
		// 如果是 input/ 开头的相对路径，转换为绝对路径
		cleanTargetDir = filepath.Clean(filepath.Join(projectRoot, cleanDir))
	} else {
		// 其他情况直接使用 cleanDir
		cleanTargetDir = filepath.Clean(cleanDir)
	}

	// 检查路径是否在允许的目录内 - 只允许上传到input目录
	isInInputDir := strings.HasPrefix(cleanTargetDir, allowedInputDir+string(filepath.Separator)) || cleanTargetDir == allowedInputDir

	if !isInInputDir {
		c.JSON(http.StatusForbidden, gin.H{"error": "文件上传路径不被允许，只能上传到input目录", "status": "error", "details": "目标路径: " + cleanTargetDir})
		return
	}

	// 确保目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create directory", "status": "error"})
		return
	}

	file, handler, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error retrieving file", "status": "error"})
		return
	}
	defer file.Close()

	filePath := filepath.Join(dir, handler.Filename)
	dest, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating file", "status": "error"})
		return
	}
	defer dest.Close()

	_, err = io.Copy(dest, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving file", "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "filename": handler.Filename, "message": "File uploaded successfully"})
}

// getFileType 根据文件扩展名确定文件类型
func getFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp":
		return "image"
	case ".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv":
		return "video"
	case ".mp3", ".wav", ".flac", ".aac", ".ogg":
		return "audio"
	case ".txt", ".md", ".json", ".yaml", ".yml", ".xml", ".csv", ".log":
		return "text"
	case ".pdf":
		return "pdf"
	case ".zip", ".rar", ".tar", ".gz", ".7z":
		return "archive"
	default:
		return "unknown"
	}
}

// BroadcastLoggerAdapter 是一个自定义的zapcore.Core实现，用于将日志广播到WebSocket
type BroadcastLoggerAdapter struct {
	toolName string
	zapcore.Core
}

// NewBroadcastLoggerAdapter 创建一个新的广播日志适配器
func NewBroadcastLoggerAdapter(toolName string, encoder zapcore.Encoder, writeSyncer zapcore.WriteSyncer) *BroadcastLoggerAdapter {
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	return &BroadcastLoggerAdapter{
		toolName: toolName,
		Core:     core,
	}
}

// With 添加字段并返回新的Core
func (b *BroadcastLoggerAdapter) With(fields []zapcore.Field) zapcore.Core {
	newCore := b.Core.With(fields)
	return &BroadcastLoggerAdapter{
		toolName: b.toolName,
		Core:     newCore,
	}
}

// Check 检查日志级别是否启用
func (b *BroadcastLoggerAdapter) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if b.Core.Enabled(entry.Level) {
		return ce.AddCore(entry, b)
	}
	return ce
}

// Write 将日志条目写入并广播到WebSocket
func (b *BroadcastLoggerAdapter) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// 首先让底层core处理日志
	err := b.Core.Write(entry, fields)

	// 构建日志消息
	// 创建一个临时编码器来生成日志消息
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	buffer, err2 := encoder.EncodeEntry(entry, fields)
	if err2 != nil {
		// 如果编码失败，使用简单消息
		return err
	}

	message := strings.TrimSpace(string(buffer.Bytes()))

	// 广播到WebSocket
	logType := "info"
	switch entry.Level {
	case zapcore.ErrorLevel:
		logType = "error"
	case zapcore.WarnLevel:
		logType = "error" // 使用error类型显示警告
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		logType = "error"
	}

	broadcast.GlobalBroadcastService.SendMessage(logType, fmt.Sprintf("[%s] %s", b.toolName, message), broadcast.GetTimeStr())

	return err
}

func webServerMain() {
	loadToolsList()

	// 初始化全局广播服务
	broadcast.GlobalBroadcastService = broadcast.NewBroadcastService()
	var wg sync.WaitGroup
	wg.Add(1)
	go broadcast.GlobalBroadcastService.Start(&wg)

	// 设置Gin为发布模式以获得更好的性能
	gin.SetMode(gin.ReleaseMode)
	//设置gin的超时时间
	r := gin.Default()

	// 获取项目根目录的绝对路径
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("无法获取当前工作目录:", err)
	}
	projectRoot := wd

	// 如果是从子目录运行的，需要调整到项目根目录
	if strings.HasSuffix(wd, "/cmd/web_server") {
		projectRoot = filepath.Dir(filepath.Dir(wd)) // 回退两级到项目根目录
	}

	// 注册路由
	r.GET("/", homePage)
	r.GET("/ws", wsEndpoint)
	r.GET("/api/tools", apiToolsHandler)
	r.POST("/api/execute", apiExecuteHandler)
	r.POST("/api/execute-all", apiExecuteAllHandler)
	r.POST("/api/process-folder", apiProcessFolderHandler)
	r.POST("/api/one-click-film", oneClickFilmHandler)
	// 添加CapCut项目生成API端点
	r.GET("/api/capcut-project", capcutProjectHandler)
	// 添加文件管理API端点
	r.GET("/api/files/list", fileListHandler)
	r.GET("/api/files/content", fileContentHandler)
	r.DELETE("/api/files/delete", fileDeleteHandler)
	r.POST("/api/files/upload", fileUploadHandler)
	// 添加数据库跟踪相关的API端点
	r.GET("/api/chapter-processes", getChapterProcessesHandler)
	r.GET("/api/chapter-process", getChapterProcessHandler)
	r.POST("/api/retry-chapter", retryChapterHandler)
	r.POST("/api/retry-step", retryStepHandler)
	// 添加项目管理相关的API端点
	r.GET("/api/projects", getProjectsHandler)
	r.GET("/api/project", getProjectHandler)
	r.POST("/api/project", createProjectHandler)
	// 添加扫描input目录自动创建项目的API端点
	r.POST("/api/scan-input-projects", scanInputForProjectsHandler)
	// 添加场景管理相关的API端点
	r.GET("/api/scenes", getScenesByChapterHandler)
	r.POST("/api/update-scene-prompt", updateScenePromptHandler)

	// 添加静态文件服务，用于提供input和output目录的文件访问
	// 使用项目根路径确保正确访问input和output目录
	inputPath := filepath.Join(projectRoot, "input")
	outputPath := filepath.Join(projectRoot, "output")
	assetsPath := filepath.Join(projectRoot, "assets")

	// 确保目录存在
	os.MkdirAll(inputPath, 0755)
	os.MkdirAll(outputPath, 0755)
	os.MkdirAll(assetsPath, 0755)

	r.Static("/files/input", inputPath)
	r.Static("/files/output", outputPath)
	r.Static("assets", assetsPath)

	// 从环境变量获取端口，默认为8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 一键出片功能 - 完整工作流处理
	go func() {
		broadcast.GlobalBroadcastService.SendLog("movie", "[一键出片] 服务器启动完成，准备处理一键出片任务", broadcast.GetTimeStr())
	}()

	log.Println("服务器启动在 :" + port)
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  15 * time.Millisecond, // 读取请求头最大耗时
		WriteTimeout: 15 * time.Millisecond, // 写响应最大耗时
		IdleTimeout:  15 * time.Second,      // 空闲连接保持时间
	}
	srv.ListenAndServe()
}

// StartServer 启动Web服务器
func StartServer() {
	webServerMain()
}

// WorkflowProcessor 工作流处理器
type WorkflowProcessor struct {
	logger        *zap.Logger
	fileManager   *file.FileManager
	ttsClient     *indextts2.IndexTTS2Client
	aegisubGen    *aegisub.AegisubGenerator
	drawThingsGen *drawthings.ChapterImageGenerator
}

// generateImagesWithOllamaPrompts 使用Ollama优化的提示词生成图像
func (wp *WorkflowProcessor) generateImagesWithOllamaPrompts(content, imagesDir string, chapterNum int, audioDurationSecs int) error {
	// 使用Ollama分析整个章节内容并生成分镜提示词
	styleDesc := "悬疑惊悚风格，周围环境模糊成黑影, 空气凝滞,浅景深, 胶片颗粒感, 低饱和度，极致悬疑氛围, 阴沉窒息感, 夏季，环境阴霾，其他部分模糊不可见"

	// 使用实际音频时长，如果未提供则估算
	estimatedDurationSecs := audioDurationSecs
	if estimatedDurationSecs <= 0 {
		// 估算音频时长（假设每分钟300字，即每个字符约0.2秒）
		estimatedDurationSecs = len(content) * 2 / 10 // 简化估算，大约每个字符0.2秒
		if estimatedDurationSecs < 60 {               // 最少1分钟
			estimatedDurationSecs = 60
		}
	}

	// 让Ollama分析整个章节并生成分镜
	wp.logger.Info("开始Ollama分镜分析", zap.Int("chapter_num", chapterNum), zap.Int("content_length", len(content)), zap.Int("estimated_duration_secs", estimatedDurationSecs))
	sceneDescriptions, err := wp.drawThingsGen.OllamaClient.AnalyzeScenesAndGeneratePrompts(content, styleDesc, estimatedDurationSecs)
	if err != nil {
		wp.logger.Warn("使用Ollama分析场景并生成分镜提示词失败",
			zap.Error(err))

		// 如果Ollama场景分析失败，回退到原来的段落处理方式
		wp.logger.Info("Ollama分镜分析失败，回退到段落处理方式")
		paragraphs := wp.splitChapterIntoParagraphsWithMerge(content)

		for idx, paragraph := range paragraphs {
			if strings.TrimSpace(paragraph) == "" {
				continue
			}

			optimizedPrompt, err := wp.drawThingsGen.OllamaClient.GenerateImagePrompt(paragraph, styleDesc)
			if err != nil {
				wp.logger.Warn("使用Ollama生成图像提示词失败，使用原始文本",
					zap.Int("paragraph_index", idx),
					zap.String("paragraph", paragraph),
					zap.Error(err))
				optimizedPrompt = paragraph + ", 周围环境模糊成黑影, 空气凝滞,浅景深, 胶片颗粒感, 低饱和度，极致悬疑氛围, 阴沉窒息感, 夏季，环境阴霾，其他部分模糊不可见"
			}

			imageFile := filepath.Join(imagesDir, fmt.Sprintf("paragraph_%02d.png", idx+1))

			err = wp.drawThingsGen.Client.GenerateImageFromText(
				optimizedPrompt,
				imageFile,
				512,   // 缩小宽度
				896,   // 缩小高度
				false, // 风格已在提示词中处理
			)
			if err != nil {
				wp.logger.Warn("生成图像失败", zap.String("paragraph", paragraph[:min(len(paragraph), 50)]), zap.Error(err))
				fmt.Printf("⚠️  段落图像生成失败: %v\n", err)
			} else {
				fmt.Printf("✅ 段落图像生成完成: %s\n", imageFile)
			}
		}

		return nil
	}

	// 如果Ollama分镜分析成功，使用生成的分镜描述生成图像
	wp.logger.Info("Ollama分镜分析成功", zap.Int("scene_count", len(sceneDescriptions)))
	for idx, sceneDesc := range sceneDescriptions {
		imageFile := filepath.Join(imagesDir, fmt.Sprintf("scene_%02d.png", idx+1))

		// 使用分镜描述生成图像
		err = wp.drawThingsGen.Client.GenerateImageFromText(
			sceneDesc,
			imageFile,
			512,   // 缩小宽度
			896,   // 缩小高度
			false, // 风格已在提示词中处理
		)
		if err != nil {
			wp.logger.Warn("生成分镜图像失败", zap.String("scene", sceneDesc[:min(len(sceneDesc), 50)]), zap.Error(err))
			fmt.Printf("⚠️  分镜图像生成失败: %v\n", err)
		} else {
			fmt.Printf("✅ 分镜图像生成完成: %s\n", imageFile)
		}
	}

	return nil
}

// splitChapterIntoParagraphsWithMerge 将章节文本分割为段落，并对短段落进行合并
func (wp *WorkflowProcessor) splitChapterIntoParagraphsWithMerge(text string) []string {
	// 按换行符分割文本
	lines := strings.Split(text, "\n")

	var rawParagraphs []string
	var currentParagraph strings.Builder

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			// 遇到空行，结束当前段落
			if currentParagraph.Len() > 0 {
				rawParagraphs = append(rawParagraphs, strings.TrimSpace(currentParagraph.String()))
				currentParagraph.Reset()
			}
		} else {
			// 添加到当前段落
			if currentParagraph.Len() > 0 {
				currentParagraph.WriteString(" ")
			}
			currentParagraph.WriteString(trimmedLine)
		}
	}

	// 处理最后一个段落
	if currentParagraph.Len() > 0 {
		rawParagraphs = append(rawParagraphs, strings.TrimSpace(currentParagraph.String()))
	}

	// 合并短段落
	var mergedParagraphs []string
	minLength := 50 // 设定最小长度阈值，低于此值的段落将与相邻段落合并

	for i := 0; i < len(rawParagraphs); i++ {
		currentPara := rawParagraphs[i]

		// 如果当前段落太短，考虑与下一个段落合并
		if len(currentPara) < minLength && i < len(rawParagraphs)-1 {
			// 与下一个段落合并
			merged := currentPara + " " + rawParagraphs[i+1]
			mergedParagraphs = append(mergedParagraphs, merged)
			i++ // 跳过下一个段落，因为它已经被合并了
		} else {
			// 检查是否当前段落太短但已经是最后一段
			if len(currentPara) < minLength && len(mergedParagraphs) > 0 {
				// 将其与前一段落合并
				lastIdx := len(mergedParagraphs) - 1
				mergedParagraphs[lastIdx] = mergedParagraphs[lastIdx] + " " + currentPara
			} else {
				// 添加正常段落
				mergedParagraphs = append(mergedParagraphs, currentPara)
			}
		}
	}

	// 过滤掉过短的段落（比如只有标点符号）
	var filtered []string
	for _, para := range mergedParagraphs {
		// 只保留非空且有一定长度的段落
		if len(strings.TrimSpace(para)) > 3 { // 至少3个字符
			filtered = append(filtered, para)
		}
	}

	return filtered
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 一键出片功能 - 完整工作流处理
func oneClickFilmHandler(c *gin.Context) {
	// 直接执行完整工作流处理，不使用goroutine以便调试
	broadcast.GlobalBroadcastService.SendLog("movie", "开始执行一键出片完整工作流...", broadcast.GetTimeStr())

	// 获取项目根目录
	wd, err := os.Getwd()
	if err != nil {
		broadcast.GlobalBroadcastService.SendLog("movie", fmt.Sprintf("[一键出片] 获取工作目录失败: %v", err), broadcast.GetTimeStr())
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": fmt.Sprintf("获取工作目录失败: %v", err)})
		return
	}

	projectRoot := wd
	if strings.HasSuffix(wd, "/cmd/web_server") {
		projectRoot = filepath.Dir(filepath.Dir(wd)) // 回退两级到项目根目录
	}

	inputDir := filepath.Join(projectRoot, "input")
	items, err := os.ReadDir(inputDir)
	if err != nil {
		broadcast.GlobalBroadcastService.SendLog("movie", fmt.Sprintf("[一键出片] ❌ 无法读取input目录: %v", err), broadcast.GetTimeStr())

		c.JSON(http.StatusOK, gin.H{"status": "error", "message": fmt.Sprintf("无法读取input目录: %v", err)})
		return
	}

	if len(items) == 0 {
		broadcast.GlobalBroadcastService.SendLog("movie", "[一键出片] ❌ input目录为空，请在input目录下放置小说文本文件", broadcast.GetTimeStr())
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "input目录为空，请在input目录下放置小说文本文件"})
		return
	}

	// 遍历input目录寻找小说目录
	for _, item := range items {
		if item.IsDir() { // 只处理目录
			novelDir := filepath.Join(inputDir, item.Name())

			// 在小说目录中寻找对应的小说文件
			novelFiles, err := os.ReadDir(novelDir)
			if err != nil {
				broadcast.GlobalBroadcastService.SendLog("movie", fmt.Sprintf("[一键出片] ❌ 无法读取小说目录 %s: %v", item.Name(), err), broadcast.GetTimeStr())
				continue
			}

			// 寻找与目录名匹配的.txt文件（例如 幽灵客栈/幽灵客栈.txt）
			for _, novelFile := range novelFiles {
				expectedFileName := item.Name() + ".txt"
				if !novelFile.IsDir() && strings.EqualFold(novelFile.Name(), expectedFileName) {
					absPath := filepath.Join(novelDir, novelFile.Name())
					broadcast.GlobalBroadcastService.SendLog("movie", "[一键出片] 🧪 开始测试章节编号解析功能...", broadcast.GetTimeStr())

					// 创建FileManager实例
					fm := file.NewFileManager()
					broadcast.GlobalBroadcastService.SendLog("movie", fmt.Sprintf("[一键出片] 📖 处理小说文件: %s", novelFile.Name()), broadcast.GetTimeStr())

					// 读取输入目录中的小说
					_, err = fm.CreateInputChapterStructure(absPath)
					if err != nil {
						broadcast.GlobalBroadcastService.SendLog("movie", fmt.Sprintf("[一键出片] ❌ 处理小说文件失败: %v", err), broadcast.GetTimeStr())

						c.JSON(http.StatusOK, gin.H{"status": "error", "message": fmt.Sprintf("处理小说文件失败: %v", err)})
						return
					}

					// 创建输出目录结构
					fm.CreateOutputChapterStructure(inputDir)
					broadcast.GlobalBroadcastService.SendLog("movie", fmt.Sprintf("[输出目录的名字] 📖 输出目录的名字: %v", inputDir), broadcast.GetTimeStr())

					// 创建logger
					logger, err := zap.NewProduction()
					if err != nil {
						broadcast.GlobalBroadcastService.SendLog("movie", fmt.Sprintf("[一键出片] ❌ 创建logger失败: %v", err), broadcast.GetTimeStr())

						c.JSON(http.StatusOK, gin.H{"status": "error", "message": fmt.Sprintf("创建logger失败: %v", err)})
						return
					}
					defer logger.Sync()

					// 使用数据库跟踪的工作流处理器
					processor, err := workflow_pkg.NewProcessor(logger)
					if err != nil {
						broadcast.GlobalBroadcastService.SendLog("movie", fmt.Sprintf("[一键出片] ❌ 创建工作流处理器失败: %v", err), broadcast.GetTimeStr())
						c.JSON(http.StatusOK, gin.H{"status": "error", "message": fmt.Sprintf("创建工作流处理器失败: %v", err)})
						return
					}

					// 广播开始生成音频
					broadcast.GlobalBroadcastService.SendLog("voice", "[一键出片] 🔊 步骤2 - 开始生成音频...", broadcast.GetTimeStr())

					// 遍历章节处理
					for key, val := range file.ChapterMap {
						outputDir := filepath.Join(projectRoot, "output", item.Name())
						chapterName := fmt.Sprintf("chapter_%02d", key)
						inputFilePath := filepath.Join(novelDir, chapterName, chapterName+".txt")

						// 检查章节输入文件是否存在
						if _, err := os.Stat(inputFilePath); os.IsNotExist(err) {
							// 如果不存在，创建临时文件
							chapterDir := filepath.Join(novelDir, chapterName)
							if err := os.MkdirAll(chapterDir, 0755); err != nil {
								broadcast.GlobalBroadcastService.SendLog("movie", fmt.Sprintf("[一键出片] ❌ 创建章节目录失败: %v", err), broadcast.GetTimeStr())
								continue
							}
							inputFilePath = filepath.Join(chapterDir, chapterName+".txt")
							if err := os.WriteFile(inputFilePath, []byte(val), 0644); err != nil {
								broadcast.GlobalBroadcastService.SendLog("movie", fmt.Sprintf("[一键出片] ❌ 创建章节输入文件失败: %v", err), broadcast.GetTimeStr())
								continue
							}
						}

						// 使用数据库跟踪处理章节
						_, err := processor.ProcessChapterWithTracking(item.Name(), chapterName, inputFilePath, outputDir)
						if err != nil {
							broadcast.GlobalBroadcastService.SendLog("movie", fmt.Sprintf("[一键出片] ❌ 章节处理失败 %s: %v", chapterName, err), broadcast.GetTimeStr())
							c.JSON(http.StatusOK, gin.H{"status": "error", "message": fmt.Sprintf("章节处理失败 %s: %v", chapterName, err)})
							return
						}

						broadcast.GlobalBroadcastService.SendLog("movie", fmt.Sprintf("[一键出片] ✅ 章节处理完成: %s", chapterName), broadcast.GetTimeStr())

						// 检查音频文件是否存在
						audioFile := filepath.Join(outputDir, fmt.Sprintf("chapter_%02d", key), "audio.wav")
						if _, err := os.Stat(audioFile); err == nil {
							// 步骤3: 生成字幕
							broadcast.GlobalBroadcastService.SendLog("aegisub", "[一键出片] 📝 步骤3 - 生成字幕...", broadcast.GetTimeStr())

							// 使用Aegisub生成字幕
							subtitleFile := strings.TrimSuffix(audioFile, ".wav") + ".srt"
							chapterText, err := os.ReadFile(inputFilePath)
							if err != nil {
								broadcast.GlobalBroadcastService.SendLog("aegisub", fmt.Sprintf("[一键出片] ❌ 读取章节文本失败: %v", err), broadcast.GetTimeStr())
							} else {
								err = processor.GetAegisubTool().ProcessIndextts2OutputWithCustomName(audioFile, string(chapterText), subtitleFile)
								if err != nil {
									broadcast.GlobalBroadcastService.SendLog("aegisub", fmt.Sprintf("[一键出片] ⚠️  字幕生成失败: %v", err), broadcast.GetTimeStr())
								} else {
									broadcast.GlobalBroadcastService.SendLog("aegisub", fmt.Sprintf("[一键出片] ✅ 字幕生成完成: %s", subtitleFile), broadcast.GetTimeStr())
								}
							}
						} else {
							broadcast.GlobalBroadcastService.SendLog("aegisub", "[一键出片] ⚠️  由于音频文件不存在，跳过字幕生成", broadcast.GetTimeStr())
						}

						// 步骤4: 生成图像
						broadcast.GlobalBroadcastService.SendLog("image", "[一键出片] 🎨 步骤4 - 生成图像...", broadcast.GetTimeStr())

						imagesDir := filepath.Join(outputDir, fmt.Sprintf("chapter_%02d", key))
						if err := os.MkdirAll(imagesDir, 0755); err != nil {
							broadcast.GlobalBroadcastService.SendLog("image", fmt.Sprintf("[一键出片] ❌ 创建图像目录失败: %v", err), broadcast.GetTimeStr())
							c.JSON(http.StatusOK, gin.H{"status": "error", "message": fmt.Sprintf("创建图像目录失败: %v", err)})
							return
						}

						// 估算音频时长用于分镜生成
						estimatedAudioDuration := 0
						if _, statErr := os.Stat(audioFile); statErr == nil {
							// 基于音频文件大小估算时长
							if fileInfo, err := os.Stat(audioFile); err == nil {
								fileSizeMB := float64(fileInfo.Size()) / (1024 * 1024)
								// 假设平均 1MB ≈ 10秒音频
								estimatedAudioDuration = int(fileSizeMB * 10)
								if estimatedAudioDuration < 30 { // 最少30秒
									estimatedAudioDuration = 30
								}
							}
						} else {
							// 如果没有音频文件，基于文本长度估算
							estimatedAudioDuration = len(val) * 2 / 10 // 每个字符约0.2秒
							if estimatedAudioDuration < 60 {           // 最少1分钟
								estimatedAudioDuration = 60
							}
						}

						// 使用Ollama优化的提示词生成图像
						wp := &WorkflowProcessor{
							logger:        logger,
							drawThingsGen: processor.GetDrawThingsTool(),
						}
						err = wp.generateImagesWithOllamaPrompts(val, imagesDir, key, estimatedAudioDuration)
						if err != nil {
							broadcast.GlobalBroadcastService.SendLog("image", fmt.Sprintf("[一键出片] ⚠️  图像生成失败: %v", err), broadcast.GetTimeStr())
						} else {
							broadcast.GlobalBroadcastService.SendLog("image", fmt.Sprintf("[一键出片] ✅ 图像生成完成，保存在: %s", imagesDir), broadcast.GetTimeStr())
						}

						// 一键出片流程至此完成，所有资源（音频、字幕、图像）已保存到output目录
						// 剪映项目生成留给用户手动操作

						// 步骤5: 生成剪映项目 (CapCut)
						broadcast.GlobalBroadcastService.SendLog("capcut", "[一键出片] 🎬 步骤5 - 生成剪映项目...", broadcast.GetTimeStr())

						// 遵循用户的要求，将input文件夹改为当前项目目录的input
						chapterDir := filepath.Join(projectRoot, "input", item.Name(), fmt.Sprintf("chapter_%02d", key))

						// 检查章节目录是否存在
						if _, err := os.Stat(chapterDir); err == nil {
							// 使用CapCut生成器创建项目
							capcutGenerator := capcut.NewCapcutGenerator(nil) // 传递logger或nil
							err = capcutGenerator.GenerateProject(chapterDir)
							if err != nil {
								broadcast.GlobalBroadcastService.SendLog("capcut", fmt.Sprintf("[一键出片] ⚠️  剪映项目生成失败: %v", err), broadcast.GetTimeStr())
							} else {
								broadcast.GlobalBroadcastService.SendLog("capcut", fmt.Sprintf("[一键出片] ✅ 剪映项目生成完成，章节: %d", key), broadcast.GetTimeStr())
							}
						} else {
							broadcast.GlobalBroadcastService.SendLog("capcut", fmt.Sprintf("[一键出片] ⚠️  章节目录不存在: %s", chapterDir), broadcast.GetTimeStr())
						}
					}

					broadcast.GlobalBroadcastService.SendLog("workflow", "[一键出片] ✅ 一键出片完整工作流执行完成！", broadcast.GetTimeStr())

					return // 处理完一个小说就返回
				}

				broadcast.GlobalBroadcastService.SendLog("workflow", "[一键出片] ✅ 一键出片完整工作流执行完成！", broadcast.GetTimeStr())

				return // 处理完一个小说就返回
			}
		}
	}

	broadcast.GlobalBroadcastService.SendLog("workflow", "[一键出片] ✅ 一键出片完整工作流执行完成！", broadcast.GetTimeStr())

	return // 处理完一个小说就返回
}

// capcutProjectHandler 生成剪映项目
func capcutProjectHandler(c *gin.Context) {
	chapterPath := c.Query("chapter_path")

	if chapterPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing chapter_path parameter", "status": "error"})
		return
	}

	// 获取项目根目录
	wd, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取当前工作目录", "status": "error"})
		return
	}

	projectRoot := wd
	if strings.HasSuffix(wd, "/cmd/web_server") {
		projectRoot = filepath.Dir(filepath.Dir(wd)) // 回退两级到项目根目录
	}

	// 构建实际路径
	var actualPath string
	if strings.HasPrefix(chapterPath, "./") {
		actualPath = filepath.Join(projectRoot, chapterPath[2:]) // 移除开头的"./"
	} else {
		actualPath = filepath.Join(projectRoot, chapterPath)
	}

	// 确保路径安全，防止路径遍历攻击
	cleanPath := filepath.Clean(actualPath)

	// 构建允许的路径前缀
	allowedInputPrefix := filepath.Join(projectRoot, "input")
	allowedOutputPrefix := filepath.Join(projectRoot, "output")

	// 检查路径是否在允许的范围内
	isValidPath := strings.HasPrefix(cleanPath, allowedInputPrefix+"/") ||
		strings.HasPrefix(cleanPath, allowedOutputPrefix+"/") ||
		cleanPath == allowedInputPrefix ||
		cleanPath == allowedOutputPrefix

	if !isValidPath {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied", "status": "error"})
		return
	}

	// 检查目录是否存在
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter directory does not exist", "status": "error"})
		return
	}

	// 提取项目名称（从路径中提取小说名和章节号）
	// 例如: /path/to/output/小说名/chapter_01 -> 小说名_第01章
	relativePath, err := filepath.Rel(projectRoot, cleanPath)
	if err != nil {
		broadcast.GlobalBroadcastService.SendLog("capcut", fmt.Sprintf("[CapCut] 解析相对路径失败: %v", err), broadcast.GetTimeStr())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法解析路径", "status": "error"})
		return
	}

	// 从路径中提取小说名和章节号
	pathParts := strings.Split(relativePath, string(filepath.Separator))
	var projectName string
	if len(pathParts) >= 2 {
		novelName := pathParts[len(pathParts)-2]   // 倒数第二部分是小说名
		chapterName := pathParts[len(pathParts)-1] // 最后一部分是章节名
		projectName = fmt.Sprintf("%s_%s", novelName, chapterName)
	} else {
		projectName = filepath.Base(cleanPath)
	}

	// 启动 goroutine 生成 CapCut 项目
	go func() {
		broadcast.GlobalBroadcastService.SendLog("capcut", fmt.Sprintf("[CapCut] 开始生成剪映项目，路径: %s, 项目名: %s", cleanPath, projectName), broadcast.GetTimeStr())

		capcutGenerator := capcut.NewCapcutGenerator(nil) // 传递logger或nil
		err := capcutGenerator.GenerateAndImportProject(cleanPath, projectName)
		if err != nil {
			broadcast.GlobalBroadcastService.SendLog("capcut", fmt.Sprintf("[CapCut] 生成失败: %v", err), broadcast.GetTimeStr())
		} else {
			broadcast.GlobalBroadcastService.SendLog("capcut", fmt.Sprintf("[CapCut] 项目生成并导入完成: %s", projectName), broadcast.GetTimeStr())
		}
	}()

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "CapCut project generation started"})
}

// getChapterProcessesHandler 获取小说的所有章节处理记录
func getChapterProcessesHandler(c *gin.Context) {
	novelName := c.Query("novel_name")

	if novelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing novel_name parameter", "status": "error"})
		return
	}

	// 创建logger
	logger, err := zap.NewProduction()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create logger", "status": "error"})
		return
	}
	defer logger.Sync()

	// 创建处理器
	processor, err := workflow_pkg.NewProcessor(logger)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create processor", "status": "error"})
		return
	}
	defer processor.GetDbManager().Close() // 确保关闭数据库连接

	processes, err := processor.GetChapterProcesses(novelName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get chapter processes: %v", err), "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"processes": processes, "status": "success"})
}

// getChapterProcessHandler 获取特定章节的处理记录
func getChapterProcessHandler(c *gin.Context) {
	novelName := c.Query("novel_name")
	chapterName := c.Query("chapter_name")

	if novelName == "" || chapterName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing novel_name or chapter_name parameter", "status": "error"})
		return
	}

	// 创建logger
	logger, err := zap.NewProduction()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create logger", "status": "error"})
		return
	}
	defer logger.Sync()

	// 创建处理器
	processor, err := workflow_pkg.NewProcessor(logger)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create processor", "status": "error"})
		return
	}
	defer processor.GetDbManager().Close() // 确保关闭数据库连接

	process, err := processor.GetChapterProcess(novelName, chapterName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get chapter process: %v", err), "status": "error"})
		return
	}

	if process == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter process not found", "status": "not_found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"process": process, "status": "success"})
}

// retryChapterHandler 重试章节处理
func retryChapterHandler(c *gin.Context) {
	var reqBody map[string]interface{}
	err := json.NewDecoder(c.Request.Body).Decode(&reqBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "status": "error"})
		return
	}

	novelName, ok := reqBody["novel_name"].(string)
	chapterName, ok2 := reqBody["chapter_name"].(string)

	if !ok || !ok2 || novelName == "" || chapterName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing novel_name or chapter_name", "status": "error"})
		return
	}

	// 创建logger
	logger, err := zap.NewProduction()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create logger", "status": "error"})
		return
	}
	defer logger.Sync()

	// 创建处理器
	processor, err := workflow_pkg.NewProcessor(logger)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create processor", "status": "error"})
		return
	}
	defer processor.GetDbManager().Close() // 确保关闭数据库连接

	err = processor.RetryChapterProcess(novelName, chapterName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to retry chapter: %v", err), "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Chapter retry initiated"})
}

// retryStepHandler 重试特定步骤
func retryStepHandler(c *gin.Context) {
	var reqBody map[string]interface{}
	err := json.NewDecoder(c.Request.Body).Decode(&reqBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "status": "error"})
		return
	}

	novelName, ok := reqBody["novel_name"].(string)
	chapterName, ok2 := reqBody["chapter_name"].(string)
	stepName, ok3 := reqBody["step_name"].(string)

	if !ok || !ok2 || !ok3 || novelName == "" || chapterName == "" || stepName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing novel_name, chapter_name, or step_name", "status": "error"})
		return
	}

	// 创建logger
	logger, err := zap.NewProduction()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create logger", "status": "error"})
		return
	}
	defer logger.Sync()

	// 创建处理器
	processor, err := workflow_pkg.NewProcessor(logger)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create processor", "status": "error"})
		return
	}
	defer processor.GetDbManager().Close() // 确保关闭数据库连接

	err = processor.RetryStepForChapter(novelName, chapterName, stepName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to retry step: %v", err), "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Step retry initiated"})
}

// updateScenePromptHandler 更新场景提示词
func updateScenePromptHandler(c *gin.Context) {
	var reqBody map[string]interface{}
	err := json.NewDecoder(c.Request.Body).Decode(&reqBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "status": "error"})
		return
	}

	sceneIDFloat, ok := reqBody["scene_id"].(float64) // JSON解码时数字默认为float64
	newPrompt, ok2 := reqBody["new_prompt"].(string)

	if !ok || !ok2 || sceneIDFloat <= 0 || newPrompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing scene_id or new_prompt", "status": "error"})
		return
	}

	sceneID := uint(sceneIDFloat)

	// 创建logger
	logger, err := zap.NewProduction()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create logger", "status": "error"})
		return
	}
	defer logger.Sync()

	// 创建处理器
	processor, err := workflow_pkg.NewProcessor(logger)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create processor", "status": "error"})
		return
	}
	defer processor.GetDbManager().Close() // 确保关闭数据库连接

	err = processor.UpdateScenePrompt(sceneID, newPrompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update scene prompt: %v", err), "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Scene prompt updated successfully"})
}

// getProjectsHandler 获取所有项目
func getProjectsHandler(c *gin.Context) {
	// 创建logger
	logger, err := zap.NewProduction()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create logger", "status": "error"})
		return
	}
	defer logger.Sync()

	// 创建处理器
	processor, err := workflow_pkg.NewProcessor(logger)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create processor", "status": "error"})
		return
	}
	defer processor.GetDbManager().Close() // 确保关闭数据库连接

	projects, err := processor.GetProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get projects: %v", err), "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"projects": projects, "status": "success"})
}

// getProjectHandler 获取特定项目
func getProjectHandler(c *gin.Context) {
	novelName := c.Query("novel_name")

	if novelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing novel_name parameter", "status": "error"})
		return
	}

	// 创建logger
	logger, err := zap.NewProduction()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create logger", "status": "error"})
		return
	}
	defer logger.Sync()

	// 创建处理器
	processor, err := workflow_pkg.NewProcessor(logger)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create processor", "status": "error"})
		return
	}
	defer processor.GetDbManager().Close() // 确保关闭数据库连接

	project, err := processor.GetProject(novelName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get project: %v", err), "status": "error"})
		return
	}

	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found", "status": "not_found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"project": project, "status": "success"})
}

// createProjectHandler 创建新项目
func createProjectHandler(c *gin.Context) {
	var reqBody map[string]interface{}
	err := json.NewDecoder(c.Request.Body).Decode(&reqBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "status": "error"})
		return
	}

	novelName, ok := reqBody["name"].(string)
	description, _ := reqBody["description"].(string)
	genre, _ := reqBody["genre"].(string)
	atmosphere, _ := reqBody["atmosphere"].(string)

	if !ok || novelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing project name", "status": "error"})
		return
	}

	// 创建logger
	logger, err := zap.NewProduction()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create logger", "status": "error"})
		return
	}
	defer logger.Sync()

	// 创建处理器
	processor, err := workflow_pkg.NewProcessor(logger)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create processor", "status": "error"})
		return
	}
	defer processor.GetDbManager().Close() // 确保关闭数据库连接

	project, err := processor.CreateProject(novelName, description, genre, atmosphere)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create project: %v", err), "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"project": project, "status": "success", "message": "Project created successfully"})
}

// getScenesByChapterHandler 获取章节的所有场景
func getScenesByChapterHandler(c *gin.Context) {
	novelName := c.Query("novel_name")
	chapterName := c.Query("chapter_name")

	if novelName == "" || chapterName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing novel_name or chapter_name parameter", "status": "error"})
		return
	}

	// 创建logger
	logger, err := zap.NewProduction()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create logger", "status": "error"})
		return
	}
	defer logger.Sync()

	// 创建处理器
	processor, err := workflow_pkg.NewProcessor(logger)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create processor", "status": "error"})
		return
	}
	defer processor.GetDbManager().Close() // 确保关闭数据库连接

	scenes, err := processor.GetScenesByChapter(novelName, chapterName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get scenes: %v", err), "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"scenes": scenes, "status": "success"})
}

// scanInputForProjectsHandler 扫描input目录并自动创建项目
func scanInputForProjectsHandler(c *gin.Context) {
	// 获取项目根目录
	wd, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取当前工作目录", "status": "error"})
		return
	}

	projectRoot := wd
	if strings.HasSuffix(wd, "/cmd/web_server") {
		projectRoot = filepath.Dir(filepath.Dir(wd)) // 回退两级到项目根目录
	}

	// 获取input目录路径
	inputDir := filepath.Join(projectRoot, "input")

	// 读取input目录
	items, err := os.ReadDir(inputDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取input目录", "status": "error"})
		return
	}

	createdProjects := make([]map[string]interface{}, 0)
	skippedProjects := make([]string, 0)

	// 创建logger
	logger, err := zap.NewProduction()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create logger", "status": "error"})
		return
	}
	defer logger.Sync()

	// 创建处理器
	processor, err := workflow_pkg.NewProcessor(logger)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create processor", "status": "error"})
		return
	}
	defer processor.GetDbManager().Close() // 确保关闭数据库连接

	// 遍历input目录寻找小说目录
	for _, item := range items {
		if item.IsDir() { // 只处理目录
			novelDir := filepath.Join(inputDir, item.Name())

			// 在小说目录中寻找对应的小说文件
			novelFiles, err := os.ReadDir(novelDir)
			if err != nil {
				broadcast.GlobalBroadcastService.SendLog("scan", fmt.Sprintf("[扫描] ❌ 无法读取小说目录 %s: %v", item.Name(), err), broadcast.GetTimeStr())
				continue
			}

			// 寻找与目录名匹配的.txt文件（例如 幽灵客栈/幽灵客栈.txt）
			for _, novelFile := range novelFiles {
				expectedFileName := item.Name() + ".txt"
				if !novelFile.IsDir() && strings.EqualFold(novelFile.Name(), expectedFileName) {
					// 检查项目是否已存在
					existingProject, err := processor.GetProject(item.Name())
					if err != nil {
						broadcast.GlobalBroadcastService.SendLog("scan", fmt.Sprintf("[扫描] 检查项目时出错 %s: %v", item.Name(), err), broadcast.GetTimeStr())
					}
					
					if existingProject == nil {
						// 项目不存在，创建新项目
						project, err := processor.CreateProject(item.Name(), fmt.Sprintf("项目: %s", item.Name()), "", "")
						if err != nil {
							broadcast.GlobalBroadcastService.SendLog("scan", fmt.Sprintf("[扫描] ❌ 创建项目失败 %s: %v", item.Name(), err), broadcast.GetTimeStr())
						} else {
							broadcast.GlobalBroadcastService.SendLog("scan", fmt.Sprintf("[扫描] ✅ 自动创建项目: %s", item.Name()), broadcast.GetTimeStr())
							createdProjects = append(createdProjects, map[string]interface{}{
								"name":        project.Name,
								"description": project.Description,
								"genre":       project.Genre,
								"atmosphere":  project.Atmosphere,
							})
						}
					} else {
						// 项目已存在，跳过
						broadcast.GlobalBroadcastService.SendLog("scan", fmt.Sprintf("[扫描] 📋 项目已存在，跳过: %s", item.Name()), broadcast.GetTimeStr())
						skippedProjects = append(skippedProjects, item.Name())
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "success",
		"message":          fmt.Sprintf("扫描完成，创建了 %d 个项目", len(createdProjects)),
		"created_projects": createdProjects,
		"skipped_projects": skippedProjects,
		"total_found":      len(createdProjects) + len(skippedProjects),
	})
}
