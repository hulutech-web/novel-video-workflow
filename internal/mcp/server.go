package mcp

import (
	"context"
	"novel-video-workflow/internal/workflow"

	mcp_server "github.com/mark3labs/mcp-go/server"

	"go.uber.org/zap"
)

type Server struct {
	server    *mcp_server.MCPServer
	processor *workflow.Processor
	logger    *zap.Logger
	handler   *Handler
}

func NewServer(processor *workflow.Processor, logger *zap.Logger) (*Server, error) {
	// 创建新的MCP服务器
	mcpServer := mcp_server.NewMCPServer(
		"novel-video-workflow-server",
		"1.0.0",
		mcp_server.WithToolCapabilities(true),
		//mcp_server.WithResourceCapabilities(false, false),
		mcp_server.WithRecovery(),
	)

	s := &Server{
		server:    mcpServer,
		processor: processor,
		logger:    logger,
	}

	// 创建 Handler 实例
	s.handler = NewHandler(s.server, processor, logger)

	// 注册所有工具到MCP服务器
	s.handler.RegisterTools()

	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	// 启动MCP服务器，使用标准输入输出
	if err := mcp_server.ServeStdio(s.server); err != nil {
		s.logger.Error("Failed to start MCP server", zap.Error(err))
		return err
	}
	return nil
}

func (s *Server) GetToolNames() []string {
	return s.handler.GetToolNames()
}
