package workflow

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

// internal/workflow/batch.go
func (p *Processor) BatchProcess(ctx context.Context, novelDir string) error {
	// 遍历目录中的所有txt文件
	files, err := filepath.Glob(filepath.Join(novelDir, "*.txt"))
	if err != nil {
		return err
	}

	for i, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			p.logger.Warn("读取文件失败",
				zap.String("文件", file),
				zap.Error(err),
			)
			continue
		}

		_, err = p.ProcessChapter(ctx, ChapterParams{
			Text:   string(content),
			Number: i + 1,
		})

		if err != nil {
			p.logger.Warn("处理章节失败",
				zap.Int("章节", i+1),
				zap.Error(err),
			)
		}
	}

	return nil
}

// 添加HTTP服务器显示处理进度
func (p *Processor) StartMonitor(port string) {
	http.HandleFunc("/progress", func(w http.ResponseWriter, r *http.Request) {
		// 返回JSON格式的进度信息
		json.NewEncoder(w).Encode(p.GetProgress())
	})

	go http.ListenAndServe(":"+port, nil)
}
