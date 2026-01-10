package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	mcp_server "github.com/mark3labs/mcp-go/server"
	mcp "github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
	"novel-video-workflow/internal/tools"
	"novel-video-workflow/internal/tools/indextts2"
	"novel-video-workflow/internal/workflow"
	"os"
	"path/filepath"
	"time"
)

// Handler processes MCP requests
type Handler struct {
	server    *mcp_server.MCPServer
	processor *workflow.Processor
	logger    *zap.Logger
	toolNames []string
}

// NewHandler creates a new handler
func NewHandler(server *mcp_server.MCPServer, processor *workflow.Processor, logger *zap.Logger) *Handler {
	h := &Handler{
		server:    server,
		processor: processor,
		logger:    logger,
		toolNames: make([]string, 0),
	}

	return h
}

// RegisterTools registers all tools with the MCP server
func (h *Handler) RegisterTools() {
	// Register process_chapter tool
	processChapterTool := mcp.NewTool("process_chapter",
		mcp.WithDescription("Process a single novel chapter"),
		mcp.WithString("chapter_text", mcp.Required(), mcp.Description("The text of the chapter to process")),
		mcp.WithNumber("chapter_number", mcp.Required(), mcp.Description("The number of the chapter")),
	)

	h.server.AddTool(processChapterTool, h.handleProcessChapter)
	h.toolNames = append(h.toolNames, "process_chapter")

	// Register generate_audio tool
	generateAudioTool := mcp.NewTool("generate_audio",
		mcp.WithDescription("Generate audio file (TTS)"),
		mcp.WithString("text", mcp.Required(), mcp.Description("The text to convert to speech")),
		mcp.WithString("reference_audio", mcp.Description("Reference audio file path for voice cloning")),
		mcp.WithString("output_file", mcp.Description("Output audio file path")),
	)

	h.server.AddTool(generateAudioTool, h.handleGenerateAudio)
	h.toolNames = append(h.toolNames, "generate_audio")

	// Register generate_indextts2_audio tool - 新增的Indextts2 TTS工具
	generateIndextts2AudioTool := mcp.NewTool("generate_indextts2_audio",
		mcp.WithDescription("Generate audio file using IndexTTS2 with advanced voice cloning capabilities"),
		mcp.WithString("text", mcp.Required(), mcp.Description("The text to convert to speech")),
		mcp.WithString("reference_audio", mcp.Required(), mcp.Description("Reference audio file path for voice cloning")),
		mcp.WithString("output_file", mcp.Description("Output audio file path")),
	)

	h.server.AddTool(generateIndextts2AudioTool, h.handleGenerateIndextts2Audio)
	h.toolNames = append(h.toolNames, "generate_indextts2_audio")

	h.logger.Info("MCP tools registered",
		zap.Int("tool_count", len(h.toolNames)))
}

// handleProcessChapter handles single chapter processing
func (h *Handler) handleProcessChapter(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	chapterText, err := request.RequireString("chapter_text")
	if err != nil {
		h.logger.Error("Missing chapter_text parameter", zap.Error(err))
		return mcp.NewToolResultError("Missing required parameter: chapter_text"), nil
	}

	chapterNumber, err := request.RequireFloat("chapter_number")
	if err != nil {
		h.logger.Error("Missing chapter_number parameter", zap.Error(err))
		return mcp.NewToolResultError("Missing required parameter: chapter_number"), nil
	}

	result, err := h.processor.ProcessChapter(ctx, workflow.ChapterParams{
		Text:   chapterText,
		Number: int(chapterNumber),
	})

	if err != nil {
		h.logger.Error("Failed to process chapter", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to process chapter: %v", err)), nil
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		h.logger.Error("Failed to serialize result", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

// handleGenerateAudio generates audio
func (h *Handler) handleGenerateAudio(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	text, err := request.RequireString("text")
	if err != nil {
		h.logger.Error("Missing text parameter", zap.Error(err))
		return mcp.NewToolResultError("Missing required parameter: text"), nil
	}

	// 获取可选参数
	referenceAudio := request.GetString("reference_audio", "")
	outputFile := request.GetString("output_file", "")

	ttsTool := tools.NewTTSProcessor(h.logger)
	result, err := ttsTool.Generate(text, outputFile, referenceAudio)
	if err != nil {
		h.logger.Error("Failed to generate audio", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to generate audio: %v", err)), nil
	}

	response := map[string]interface{}{
		"success": result.Success,
		"file":    result.OutputFile,
	}

	if !result.Success {
		response["error"] = result.Error
	}

	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		h.logger.Error("Failed to serialize response", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(responseJSON)), nil
}

// handleGenerateIndextts2Audio generates audio using the new Indextts2 client
func (h *Handler) handleGenerateIndextts2Audio(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	text, err := request.RequireString("text")
	if err != nil {
		h.logger.Error("Missing text parameter", zap.Error(err))
		return mcp.NewToolResultError("Missing required parameter: text"), nil
	}

	referenceAudio, err := request.RequireString("reference_audio")
	if err != nil {
		h.logger.Error("Missing reference_audio parameter", zap.Error(err))
		return mcp.NewToolResultError("Missing required parameter: reference_audio"), nil
	}

	// 获取可选参数
	outputFile := request.GetString("output_file", "")

	// 如果outputFile为空，生成默认路径
	if outputFile == "" {
		outputFile = fmt.Sprintf("output/indextts2_output_%d.wav", time.Now().Unix())
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		h.logger.Error("Failed to create output directory", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create output directory: %v", err)), nil
	}

	// 使用新的Indextts2客户端
	client := indextts2.NewIndexTTS2Client(h.logger, "http://localhost:7860")

	// 调用Indextts2客户端生成音频
	var result tools.TTSResult
	err = client.GenerateTTSWithAudio(referenceAudio, text, outputFile)
	if err != nil {
		h.logger.Error("Failed to generate audio with Indextts2", zap.Error(err))
		result = tools.TTSResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to generate audio with Indextts2: %v", err),
		}
	} else {
		result = tools.TTSResult{
			Success:    true,
			OutputFile: outputFile,
		}
	}

	response := map[string]interface{}{
		"success": result.Success,
		"file":    result.OutputFile,
		"engine":  "indextts2",
		"text":    text,
		"reference_audio": referenceAudio,
	}

	if !result.Success {
		response["error"] = result.Error
	}

	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		h.logger.Error("Failed to serialize Indextts2 response", zap.Error(err))
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(responseJSON)), nil
}

// GetToolNames gets all tool names
func (h *Handler) GetToolNames() []string {
	return h.toolNames
}