package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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
var broadcast = make(chan MCPLog)

// MCPLog 结构存储MCP工具的日志信息
type MCPLog struct {
	ToolName  string `json:"toolName"`
	Message   string `json:"message"`
	Type      string `json:"type"` // "info", "success", "error"
	Timestamp string `json:"timestamp"`
}

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

func handleLogs() {
	for {
		logData := <-broadcast
		for client := range clients {
			err := client.WriteJSON(logData)
			if err != nil {
				log.Printf("Error sending log to client: %v", err)
				delete(clients, client)
				client.Close()
			}
		}
	}
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

	// 添加客户端到映射
	clients[ws] = true
	defer func() {
		// 确保在函数退出时从客户端映射中删除该客户端
		delete(clients, ws)
	}()

	for {
		var msg map[string]interface{}
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			// 客户端已经断开连接，不需要额外处理，defer会自动删除
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
			broadcast <- MCPLog{
				ToolName:  toolName,
				Message:   "工具不存在: " + toolName,
				Type:      "error",
				Timestamp: time.Now().Format(time.RFC3339),
			}
			return
		}

		broadcast <- MCPLog{
			ToolName:  toolName,
			Message:   "开始执行工具...",
			Type:      "info",
			Timestamp: time.Now().Format(time.RFC3339),
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
				broadcast <- MCPLog{
					ToolName:  toolName,
					Message:   fmt.Sprintf("创建输出目录失败: %v", err),
					Type:      "error",
					Timestamp: time.Now().Format(time.RFC3339),
				}
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
						broadcast <- MCPLog{
							ToolName:  toolName,
							Message:   fmt.Sprintf("找到参考音频: %s", referenceAudio),
							Type:      "info",
							Timestamp: time.Now().Format(time.RFC3339),
						}
						break
					}
				}

				if !found {
					broadcast <- MCPLog{
						ToolName:  toolName,
						Message:   "找不到参考音频文件，请确保存在默认音频文件",
						Type:      "error",
						Timestamp: time.Now().Format(time.RFC3339),
					}
					return
				}
			}

			broadcast <- MCPLog{
				ToolName:  toolName,
				Message:   fmt.Sprintf("使用参考音频: %s", referenceAudio),
				Type:      "info",
				Timestamp: time.Now().Format(time.RFC3339),
			}

			broadcast <- MCPLog{
				ToolName:  toolName,
				Message:   fmt.Sprintf("输入文本: %s", text),
				Type:      "info",
				Timestamp: time.Now().Format(time.RFC3339),
			}

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
						broadcast <- MCPLog{
							ToolName:  toolName,
							Message:   fmt.Sprintf("工具执行失败: %v", err),
							Type:      "error",
							Timestamp: time.Now().Format(time.RFC3339),
						}
						return
					}

					// 检查结果
					if success, ok := result["success"].(bool); ok && success {
						broadcast <- MCPLog{
							ToolName:  toolName,
							Message:   fmt.Sprintf("音频生成成功，输出文件: %s", outputFile),
							Type:      "success",
							Timestamp: time.Now().Format(time.RFC3339),
						}
					} else {
						errorMsg := "未知错误"
						if result["error"] != nil {
							if errStr, ok := result["error"].(string); ok {
								errorMsg = errStr
							}
						}
						broadcast <- MCPLog{
							ToolName:  toolName,
							Message:   fmt.Sprintf("工具执行失败: %s", errorMsg),
							Type:      "error",
							Timestamp: time.Now().Format(time.RFC3339),
						}
					}
				} else {
					broadcast <- MCPLog{
						ToolName:  toolName,
						Message:   "错误: MCP处理器未初始化",
						Type:      "error",
						Timestamp: time.Now().Format(time.RFC3339),
					}
				}
			} else {
				// 如果没有MCP服务器实例，给出提示
				broadcast <- MCPLog{
					ToolName:  toolName,
					Message:   "错误: MCP服务器未启动。请确保服务已正确初始化。",
					Type:      "error",
					Timestamp: time.Now().Format(time.RFC3339),
				}
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
							broadcast <- MCPLog{
								ToolName:  toolName,
								Message:   fmt.Sprintf("创建输出目录失败: %v", err),
								Type:      "error",
								Timestamp: time.Now().Format(time.RFC3339),
							}
							return
						}

						broadcast <- MCPLog{
							ToolName:  toolName,
							Message:   fmt.Sprintf("开始处理章节文本，输出目录: %s", outputDir),
							Type:      "info",
							Timestamp: time.Now().Format(time.RFC3339),
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
							broadcast <- MCPLog{
								ToolName:  toolName,
								Message:   fmt.Sprintf("生成图像失败: %v", err),
								Type:      "error",
								Timestamp: time.Now().Format(time.RFC3339),
							}
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
						broadcast <- MCPLog{
							ToolName:  toolName,
							Message:   fmt.Sprintf("暂不支持直接调用工具: %s", toolName),
							Type:      "error",
							Timestamp: time.Now().Format(time.RFC3339),
						}
						return
					}

					if err != nil {
						broadcast <- MCPLog{
							ToolName:  toolName,
							Message:   fmt.Sprintf("工具执行失败: %v", err),
							Type:      "error",
							Timestamp: time.Now().Format(time.RFC3339),
						}
						return
					}

					// 记录执行结果
					broadcast <- MCPLog{
						ToolName:  toolName,
						Message:   fmt.Sprintf("工具执行完成，结果: %+v", result),
						Type:      "info",
						Timestamp: time.Now().Format(time.RFC3339),
					}
				} else {
					broadcast <- MCPLog{
						ToolName:  toolName,
						Message:   "错误: MCP处理器未初始化",
						Type:      "error",
						Timestamp: time.Now().Format(time.RFC3339),
					}
				}
			} else {
				broadcast <- MCPLog{
					ToolName:  toolName,
					Message:   "错误: MCP服务器未启动。请确保服务已正确初始化。",
					Type:      "error",
					Timestamp: time.Now().Format(time.RFC3339),
				}
			}
		}

		broadcast <- MCPLog{
			ToolName:  toolName,
			Message:   "工具执行完成",
			Type:      "success",
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}()

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Tool execution started"})
}

func apiExecuteAllHandler(c *gin.Context) {
	// 执行所有MCP工具
	go func() {
		for _, tool := range mcpTools {
			broadcast <- MCPLog{
				ToolName:  tool.Name,
				Message:   "开始执行工具...",
				Type:      "info",
				Timestamp: time.Now().Format(time.RFC3339),
			}

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
						broadcast <- MCPLog{
							ToolName:  tool.Name,
							Message:   fmt.Sprintf("暂不支持直接调用工具: %s", tool.Name),
							Type:      "error",
							Timestamp: time.Now().Format(time.RFC3339),
						}
						continue
					}

					if err != nil {
						broadcast <- MCPLog{
							ToolName:  tool.Name,
							Message:   fmt.Sprintf("工具执行失败: %v", err),
							Type:      "error",
							Timestamp: time.Now().Format(time.RFC3339),
						}
						continue
					}

					// 记录执行结果
					broadcast <- MCPLog{
						ToolName:  tool.Name,
						Message:   fmt.Sprintf("工具执行完成，结果: %+v", result),
						Type:      "info",
						Timestamp: time.Now().Format(time.RFC3339),
					}
				} else {
					broadcast <- MCPLog{
						ToolName:  tool.Name,
						Message:   "错误: MCP处理器未初始化",
						Type:      "error",
						Timestamp: time.Now().Format(time.RFC3339),
					}
				}
			} else {
				// 如果没有MCP服务器实例，给出提示
				broadcast <- MCPLog{
					ToolName:  tool.Name,
					Message:   "错误: MCP服务器未启动。请确保服务已正确初始化。",
					Type:      "error",
					Timestamp: time.Now().Format(time.RFC3339),
				}
			}

			broadcast <- MCPLog{
				ToolName:  tool.Name,
				Message:   tool.Name + " 执行完成",
				Type:      "success",
				Timestamp: time.Now().Format(time.RFC3339),
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "All tools execution started"})
}

func apiProcessFolderHandler(c *gin.Context) {
	// 处理上传的文件夹
	go func() {
		broadcast <- MCPLog{
			ToolName:  "工作流",
			Message:   "开始文件夹处理工作流...",
			Type:      "info",
			Timestamp: time.Now().Format(time.RFC3339),
		}

		// 检查MCP服务器实例是否存在
		if mcpServerInstance != nil {
			// 获取处理器并直接调用工具
			handler := mcpServerInstance.GetHandler()
			if handler != nil {
				// 模拟工作流处理
				for _, tool := range mcpTools {
					broadcast <- MCPLog{
						ToolName:  tool.Name,
						Message:   "使用 " + tool.Name + " 处理...",
						Type:      "info",
						Timestamp: time.Now().Format(time.RFC3339),
					}

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
						broadcast <- MCPLog{
							ToolName:  tool.Name,
							Message:   fmt.Sprintf("暂不支持直接调用工具: %s", tool.Name),
							Type:      "error",
							Timestamp: time.Now().Format(time.RFC3339),
						}
						continue
					}

					if err != nil {
						broadcast <- MCPLog{
							ToolName:  tool.Name,
							Message:   fmt.Sprintf("工具执行失败: %v", err),
							Type:      "error",
							Timestamp: time.Now().Format(time.RFC3339),
						}
					} else {
						// 记录执行结果
						broadcast <- MCPLog{
							ToolName:  tool.Name,
							Message:   fmt.Sprintf("工具执行完成，结果: %+v", result),
							Type:      "info",
							Timestamp: time.Now().Format(time.RFC3339),
						}
					}

					broadcast <- MCPLog{
						ToolName:  tool.Name,
						Message:   tool.Name + " 完成",
						Type:      "success",
						Timestamp: time.Now().Format(time.RFC3339),
					}
				}
			} else {
				broadcast <- MCPLog{
					ToolName:  "工作流",
					Message:   "错误: MCP处理器未初始化",
					Type:      "error",
					Timestamp: time.Now().Format(time.RFC3339),
				}
			}
		} else {
			broadcast <- MCPLog{
				ToolName:  "工作流",
				Message:   "错误: MCP服务器未启动。请确保服务已正确初始化。",
				Type:      "error",
				Timestamp: time.Now().Format(time.RFC3339),
			}
		}

		broadcast <- MCPLog{
			ToolName:  "工作流",
			Message:   "文件夹处理完成",
			Type:      "success",
			Timestamp: time.Now().Format(time.RFC3339),
		}
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
	if !strings.HasPrefix(filepath.Clean(dir), "./input") && !strings.HasPrefix(filepath.Clean(dir), "./output") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied", "status": "error"})
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
		broadcast <- MCPLog{
			ToolName:  b.toolName,
			Message:   fmt.Sprintf("日志编码失败: %v", err2),
			Type:      "error",
			Timestamp: entry.Time.Format(time.RFC3339),
		}
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

	broadcast <- MCPLog{
		ToolName:  b.toolName,
		Message:   message,
		Type:      logType,
		Timestamp: entry.Time.Format(time.RFC3339),
	}

	return err
}

func main() {
	loadToolsList()

	go handleLogs()

	// 设置Gin为发布模式以获得更好的性能
	gin.SetMode(gin.ReleaseMode)
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
	// 添加文件管理API端点
	r.GET("/api/files/list", fileListHandler)
	r.GET("/api/files/content", fileContentHandler)
	r.DELETE("/api/files/delete", fileDeleteHandler)
	r.POST("/api/files/upload", fileUploadHandler)

	// 添加静态文件服务，用于提供input和output目录的文件访问
	// 使用项目根路径确保正确访问input和output目录
	inputPath := filepath.Join(projectRoot, "input")
	outputPath := filepath.Join(projectRoot, "output")

	// 确保目录存在
	os.MkdirAll(inputPath, 0755)
	os.MkdirAll(outputPath, 0755)

	r.Static("/files/input", inputPath)
	r.Static("/files/output", outputPath)

	// 从环境变量获取端口，默认为8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("服务器启动在 :" + port)
	r.Run(":" + port)
}
