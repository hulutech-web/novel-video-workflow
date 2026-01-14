package workflow

import (
	"encoding/json"
	"net/http"
)

// 添加HTTP服务器显示处理进度
func (p *Processor) StartMonitor(port string) {
	http.HandleFunc("/progress", func(w http.ResponseWriter, r *http.Request) {
		// 返回JSON格式的进度信息
		json.NewEncoder(w).Encode(p.GetProgress())
	})

	go http.ListenAndServe(":"+port, nil)
}
