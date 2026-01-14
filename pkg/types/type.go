package types

// MCPLog 结构存储MCP工具的日志信息
type MCPLog struct {
	ToolName  string `json:"toolName"`
	Message   string `json:"message"`
	Type      string `json:"type"` // "info", "success", "error"
	Timestamp string `json:"timestamp"`
}
