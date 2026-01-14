package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	mcp_pkg "novel-video-workflow/pkg/mcp"
	"novel-video-workflow/pkg/tools/drawthings"
	workflow_pkg "novel-video-workflow/pkg/workflow"
	"os"
	"path/filepath"
	"strings"
	"time"

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

func homePage(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
	<title>MCP 工作流控制台</title>
	<style>
		body {
			font-family: Arial, sans-serif;
			margin: 0;
			padding: 20px;
			background-color: #f5f5f5;
		}
		.container {
			max-width: 1200px;
			margin: 0 auto;
		}
		h1 {
			color: #333;
			text-align: center;
		}
		.nav-tabs {
			display: flex;
			margin-bottom: 20px;
			background: white;
			border-radius: 5px;
			box-shadow: 0 2px 5px rgba(0,0,0,0.1);
		}
		.nav-tab {
			padding: 15px 30px;
			cursor: pointer;
			border-right: 1px solid #eee;
		}
		.nav-tab:last-child {
			border-right: none;
		}
		.nav-tab.active {
			background: #007bff;
			color: white;
		}
		.tab-content {
			display: none;
			background: white;
			padding: 20px;
			border-radius: 5px;
			box-shadow: 0 2px 5px rgba(0,0,0,0.1);
		}
		.tab-content.active {
			display: block;
		}
		.tool-card {
			border: 1px solid #ddd;
			padding: 15px;
			margin: 10px 0;
			border-radius: 5px;
			background: #f9f9f9;
		}
		.console {
			height: 400px;
			overflow-y: scroll;
			background: #000;
			color: #00ff00;
			padding: 10px;
			font-family: monospace;
			margin-top: 20px;
		}
		.console-line {
			margin: 2px 0;
		}
		.info { color: #00ff00; }
		.success { color: #00ff00; font-weight: bold; }
		.error { color: #ff5555; }
		.upload-area {
			border: 2px dashed #ccc;
			padding: 20px;
			text-align: center;
			margin: 20px 0;
			border-radius: 5px;
			cursor: pointer;
		}
		.upload-area.drag-over {
			border-color: #007bff;
			background-color: #f0f8ff;
		}
		button {
			background: #007bff;
			color: white;
			border: none;
			padding: 10px 20px;
			border-radius: 3px;
			cursor: pointer;
			margin: 5px;
		}
		button:hover {
			background: #0056b3;
		}
		input, select, textarea {
			padding: 8px;
			margin: 5px;
			border: 1px solid #ccc;
			border-radius: 3px;
			width: 100%;
			box-sizing: border-box;
		}
		.tutorial {
			line-height: 1.6;
		}
		.tutorial h3 {
			color: #007bff;
			margin-top: 20px;
		}
		.tutorial ul {
			padding-left: 20px;
		}
		.tutorial li {
			margin: 10px 0;
		}
		/* 专门针对音频生成工具的样式 */
		.audio-tool-form {
			background: #e7f3ff;
			padding: 15px;
			border-radius: 5px;
			margin-top: 10px;
			display: none;
		}
		.audio-tool-form.active {
			display: block;
		}
		.form-group {
			margin-bottom: 15px;
		}
		.form-group label {
			display: block;
			margin-bottom: 5px;
			font-weight: bold;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>MCP 工作流控制台</h1>
		
		<div class="nav-tabs">
			<div class="nav-tab active" onclick="switchTab('dashboard')">仪表板</div>
			<div class="nav-tab" onclick="switchTab('tools')">MCP 工具</div>
			<div class="nav-tab" onclick="switchTab('upload')">上传并处理</div>
			<div class="nav-tab" onclick="switchTab('tutorial')">教程</div>
		</div>
		
		<div id="dashboard" class="tab-content active">
			<h2>工作流仪表板</h2>
			<p>当前状态: <span id="current-status">空闲</span></p>
			<button onclick="executeAll()">执行所有 MCP 工具</button>
			<button onclick="stopExecution()">停止执行</button>
		</div>
		
		<div id="tools" class="tab-content">
			<h2>MCP 工具</h2>
			<div id="tools-list"></div>
		</div>
		
		<div id="upload" class="tab-content">
			<h2>上传文件夹并处理</h2>
			<div class="upload-area" id="uploadArea" ondrop="handleDrop(event)" ondragover="handleDragOver(event)">
				<p>拖放文件夹到此处或点击选择</p>
				<input type="file" id="folderInput" webkitdirectory directory multiple style="display: none;" />
			</div>
			<button onclick="processFolder()">处理上传的文件夹</button>
			<p id="uploadStatus"></p>
		</div>
		
		<div id="tutorial" class="tab-content">
			<h2>教程</h2>
			<div class="tutorial">
				<h3>入门指南</h3>
				<ul>
					<li>此控制台允许您管理和执行 MCP 服务</li>
					<li>使用 MCP 工具标签页查看和执行单个工具</li>
					<li>使用上传并处理标签页上传文件夹并运行整个工作流</li>
					<li>在每个部分底部的控制台中监视进度</li>
				</ul>
				
				<h3>执行单个工具</h3>
				<ul>
					<li>导航到 MCP 工具标签页</li>
					<li>您将看到所有可用 MCP 工具的列表</li>
					<li>点击任何工具上的"执行"按钮单独运行它</li>
					<li>在控制台中监视执行日志</li>
				</ul>
				
				<h3>处理整个文件夹</h3>
				<ul>
					<li>转到上传并处理标签页</li>
					<li>拖放文件夹或点击选择一个文件夹</li>
					<li>点击"处理上传的文件夹"开始工作流</li>
					<li>在控制台中监视进度</li>
				</ul>
				
				<h3>监控进度</h3>
				<ul>
					<li>控制台显示来自 MCP 工具的实时日志</li>
					<li>绿色文本表示信息消息</li>
					<li>亮绿色文本表示成功消息</li>
					<li>红色文本表示错误消息</li>
				</ul>
			</div>
		</div>
		
		<div class="console" id="console"></div>
	</div>

	<script>
		// WebSocket连接
		const socket = new WebSocket('ws://' + window.location.host + '/ws');
		
		socket.onopen = function(event) {
			console.log('Connected to WebSocket');
		};
		
		socket.onmessage = function(event) {
			const logData = JSON.parse(event.data);
			const consoleDiv = document.getElementById('console');
			
			const timestamp = new Date().toLocaleTimeString();
			const lineDiv = document.createElement('div');
			lineDiv.className = 'console-line ' + logData.type;
			lineDiv.textContent = '[' + timestamp + '] [' + logData.toolName + '] ' + logData.message;
			
			consoleDiv.appendChild(lineDiv);
			consoleDiv.scrollTop = consoleDiv.scrollHeight;
		};
		
		function switchTab(tabName) {
			// 隐藏所有标签内容
			const tabs = document.getElementsByClassName('tab-content');
			for (let i = 0; i < tabs.length; i++) {
				tabs[i].classList.remove('active');
			}
			
			// 移除所有标签的激活状态
			const navTabs = document.getElementsByClassName('nav-tab');
			for (let i = 0; i < navTabs.length; i++) {
				navTabs[i].classList.remove('active');
			}
			
			// 显示选中的标签内容
			document.getElementById(tabName).classList.add('active');
			
			// 激活选中的标签
			event.target.classList.add('active');
			
			// 如果切换到工具标签，则加载工具列表
			if (tabName === 'tools') {
				loadToolsList();
			}
		}
		
		function loadToolsList() {
			fetch('/api/tools')
				.then(response => response.json())
				.then(tools => {
					const toolsListDiv = document.getElementById('tools-list');
					toolsListDiv.innerHTML = '';
					
					tools.forEach(function(tool) {
						const toolCard = document.createElement('div');
						toolCard.className = 'tool-card';
						
						// 为generate_indextts2_audio工具添加特殊处理
						let buttonHtml = '';
						if (tool.name === 'generate_indextts2_audio') {
							// 为音频生成工具添加表单
							buttonHtml = 
								'<h3>' + tool.name + '</h3>' +
								'<p>' + tool.description + '</p>' +
								'<button onclick="toggleAudioForm(\'' + tool.name + '\')">执行工具</button>' +
								'<div id="form_' + tool.name + '" class="audio-tool-form">' +
									'<div class="form-group">' +
										'<label for="textInput_' + tool.name + '">输入文本:</label>' +
										'<textarea id="textInput_' + tool.name + '" placeholder="请输入要转换为语音的文本" rows="4"></textarea>' +
									'</div>' +
									'<div class="form-group">' +
										'<label for="outputDir_' + tool.name + '">输出目录:</label>' +
										'<input type="text" id="outputDir_' + tool.name + '" value="./output/" placeholder="请输入输出目录">' +
									'</div>' +
									'<button onclick="executeAudioTool(\'' + tool.name + '\')">生成音频</button>' +
									'<button onclick="hideAudioForm(\'' + tool.name + '\')" style="background: #6c757d;">取消</button>' +
								'</div>';
						} else if (tool.name === 'generate_images_from_chapter_with_ai_prompt') {
							// 为图像生成工具添加表单
							buttonHtml = 
								'<h3>' + tool.name + '</h3>' +
								'<p>' + tool.description + '</p>' +
								'<button onclick="toggleImageForm(\'' + tool.name + '\')">执行工具</button>' +
								'<div id="form_' + tool.name + '" class="audio-tool-form">' +
									'<div class="form-group">' +
										'<label for="chapterText_' + tool.name + '">章节文本:</label>' +
										'<textarea id="chapterText_' + tool.name + '" placeholder="请输入要生成图像的章节文本" rows="6"></textarea>' +
									'</div>' +
									'<div class="form-group">' +
										'<label for="outputDir_' + tool.name + '">输出目录:</label>' +
										'<input type="text" id="outputDir_' + tool.name + '" value="./output/images_" placeholder="请输入输出目录">' +
									'</div>' +
									'<div class="form-group">' +
										'<label for="imageWidth_' + tool.name + '">图像宽度:</label>' +
										'<input type="number" id="imageWidth_' + tool.name + '" value="512" min="256" max="2048">' +
									'</div>' +
									'<div class="form-group">' +
										'<label for="imageHeight_' + tool.name + '">图像高度:</label>' +
										'<input type="number" id="imageHeight_' + tool.name + '" value="896" min="256" max="2048">' +
									'</div>' +
									'<button onclick="executeImageTool(\'' + tool.name + '\')">生成图像</button>' +
									'<button onclick="hideImageForm(\'' + tool.name + '\')" style="background: #6c757d;">取消</button>' +
								'</div>';
						} else {
							// 其他工具使用普通按钮
							buttonHtml = 
								'<h3>' + tool.name + '</h3>' +
								'<p>' + tool.description + '</p>' +
								'<button onclick="executeTool(\'' + tool.name + '\')">执行工具</button>';
						}
						
						toolCard.innerHTML = buttonHtml;
						toolsListDiv.appendChild(toolCard);
					});
				})
				.catch(error => {
					console.error('Error loading tools:', error);
					// 添加错误提示
					const toolsListDiv = document.getElementById('tools-list');
					toolsListDiv.innerHTML = '<p style="color: red;">加载工具列表失败: ' + error.message + '</p>';
				});
		}
		
		// 显示音频生成表单
		function toggleAudioForm(toolName) {
			const form = document.getElementById('form_' + toolName);
			if (form) {
				form.classList.toggle('active');
			}
		}
		
		// 隐藏音频生成表单
		function hideAudioForm(toolName) {
			const form = document.getElementById('form_' + toolName);
			if (form) {
				form.classList.remove('active');
			}
		}

		// 显示图像生成表单
		function toggleImageForm(toolName) {
			const form = document.getElementById('form_' + toolName);
			if (form) {
				form.classList.toggle('active');
			}
		}

		// 隐藏图像生成表单
		function hideImageForm(toolName) {
			const form = document.getElementById('form_' + toolName);
			if (form) {
				form.classList.remove('active');
			}
		}

		// 执行音频生成工具
		function executeAudioTool(toolName) {
			const textInput = document.getElementById('textInput_' + toolName).value;
			const outputDir = document.getElementById('outputDir_' + toolName).value;
			
			if (!textInput || textInput.trim() === '') {
				alert('请输入要转换为语音的文本');
				return;
			}
			
			// 生成输出文件路径
			const timestamp = new Date().getTime();
			const outputFile = outputDir + '/audio_' + timestamp + '.wav';
			
			const params = {
				toolName: toolName,
				text: textInput,
				output_file: outputFile
			};
			
			fetch('/api/execute', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify(params)
			})
			.then(response => response.json())
			.then(data => {
				console.log('Audio tool execution initiated:', data);
				// 隐藏表单
				hideAudioForm(toolName);
			})
			.catch(error => {
				console.error('Error executing audio tool:', error);
			});
		}

		// 执行图像生成工具
		function executeImageTool(toolName) {
			const chapterText = document.getElementById('chapterText_' + toolName).value;
			const outputDir = document.getElementById('outputDir_' + toolName).value;
			const imageWidth = parseInt(document.getElementById('imageWidth_' + toolName).value);
			const imageHeight = parseInt(document.getElementById('imageHeight_' + toolName).value);
			
			if (!chapterText || chapterText.trim() === '') {
				alert('请输入要生成图像的章节文本');
				return;
			}
			
			if (!outputDir || outputDir.trim() === '') {
				alert('请输入输出目录');
				return;
			}
			
			// 生成输出目录路径
			const timestamp = new Date().getTime();
			const outputDirectory = outputDir + timestamp;
			
			const params = {
				toolName: toolName,
				chapter_text: chapterText,
				output_dir: outputDirectory,
				width: imageWidth,
				height: imageHeight
			};
			
			fetch('/api/execute', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify(params)
			})
			.then(response => response.json())
			.then(data => {
				console.log('Image tool execution initiated:', data);
				// 隐藏表单
				hideImageForm(toolName);
			})
			.catch(error => {
				console.error('Error executing image tool:', error);
			});
		}
		
		function executeTool(toolName) {
			fetch('/api/execute', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({toolName: toolName})
			})
			.then(response => response.json())
			.then(data => {
				console.log('Tool execution initiated:', data);
			})
			.catch(error => {
				console.error('Error executing tool:', error);
			});
		}
		
		function executeAll() {
			fetch('/api/execute-all', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({})
			})
			.then(response => response.json())
			.then(data => {
				console.log('All tools execution initiated:', data);
			})
			.catch(error => {
				console.error('Error executing all tools:', error);
			});
		}
		
		function stopExecution() {
			fetch('/api/stop', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({})
			})
			.then(response => response.json())
			.then(data => {
				console.log('Execution stopped:', data);
			})
			.catch(error => {
				console.error('Error stopping execution:', error);
			});
		}
		
		function handleDragOver(e) {
			e.preventDefault();
			e.stopPropagation();
			document.getElementById('uploadArea').classList.add('drag-over');
		}
		
		function handleDrop(e) {
			e.preventDefault();
			e.stopPropagation();
			document.getElementById('uploadArea').classList.remove('drag-over');
			
			const files = e.dataTransfer.files;
			if (files.length > 0) {
				// 简单显示上传状态
				document.getElementById('uploadStatus').innerText = '已放置文件: ' + files.length + ' 个文件';
			}
		}
		
		function processFolder() {
			fetch('/api/process-folder', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({})
			})
			.then(response => response.json())
			.then(data => {
				console.log('Folder processing initiated:', data);
			})
			.catch(error => {
				console.error('Error processing folder:', error);
			});
		}
		
		// 初始化 - 加载工具列表
		window.onload = function() {
			// 在初始状态下加载工具列表
			if (document.querySelector('.nav-tab.active').textContent.trim() === 'MCP 工具') {
				loadToolsList();
			}
		};
	</script>
</body>
</html>`
	fmt.Fprintf(w, "%s", html)
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
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

func apiToolsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mcpTools)
}

func apiExecuteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody map[string]interface{} // 修改为interface{}以支持不同类型参数
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	toolName, ok := reqBody["toolName"].(string)
	if !ok || toolName == "" {
		http.Error(w, "Missing toolName", http.StatusBadRequest)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Tool execution started"})
}

func apiExecuteAllHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "All tools execution started"})
}

func apiProcessFolderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Folder processing started"})
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

	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpoint)
	http.HandleFunc("/api/tools", apiToolsHandler)
	http.HandleFunc("/api/execute", apiExecuteHandler)
	http.HandleFunc("/api/execute-all", apiExecuteAllHandler)
	http.HandleFunc("/api/process-folder", apiProcessFolderHandler)

	log.Println("服务器启动在 :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
