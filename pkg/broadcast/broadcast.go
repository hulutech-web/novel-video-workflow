package broadcast

import (
	"novel-video-workflow/pkg/types"
	"sync"
	"time"
)

// GlobalBroadcastService 全局广播服务
var GlobalBroadcastService *BroadcastService

// BroadcastService 广播服务结构
type BroadcastService struct {
	broadcastChan chan types.MCPLog
	clients       map[*Client]bool
	register      chan *Client
	unregister    chan *Client // 通道用于注销特定客户端
	shutdown      chan bool    // 通道用于关闭整个服务
	mutex         sync.Mutex
}

// Client 表示一个WebSocket客户端
type Client struct {
	Conn interface{}       // WebSocket连接
	Send chan types.MCPLog // 通道用于发送消息
}

// NewBroadcastService 创建新的广播服务
func NewBroadcastService() *BroadcastService {
	//这里有重复定义的风险，需要判断是否重复如果重复了直接返回，做成一个单例
	if GlobalBroadcastService != nil {
		return GlobalBroadcastService
	}
	return &BroadcastService{
		broadcastChan: make(chan types.MCPLog, 100),
		clients:       make(map[*Client]bool),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		shutdown:      make(chan bool),
	}
}

// Start 启动广播服务
func (b *BroadcastService) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case client := <-b.register:
			b.mutex.Lock()
			b.clients[client] = true
			b.mutex.Unlock()
		case client := <-b.unregister:
			b.mutex.Lock()
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client.Send)
			}
			b.mutex.Unlock()
		case <-b.shutdown:
			b.mutex.Lock()
			for client := range b.clients {
				delete(b.clients, client)
				close(client.Send)
			}
			b.mutex.Unlock()
			return
		case message := <-b.broadcastChan:
			b.mutex.Lock()
			// 发送给所有注册的客户端
			for client := range b.clients {
				select {
				case client.Send <- message:
				default:
					// 如果发送失败，移除客户端
					delete(b.clients, client)
					close(client.Send)
				}
			}
			b.mutex.Unlock()
		}
	}
}

// SendLog 发送日志消息
func (b *BroadcastService) SendLog(Name string, msg string, timestamp string) {
	b.broadcastChan <- types.MCPLog{
		ToolName:  Name,
		Type:      "log",
		Message:   msg,
		Timestamp: timestamp,
	}
}

// SendMessage 发送普通消息
func (b *BroadcastService) SendMessage(Name string, msg string, timestamp string) {
	b.broadcastChan <- types.MCPLog{
		ToolName:  Name,
		Type:      "message",
		Message:   msg,
		Timestamp: timestamp,
	}
}

// RegisterClient 注册客户端
func (b *BroadcastService) RegisterClient(conn interface{}) chan types.MCPLog {
	client := &Client{
		Conn: conn,
		Send: make(chan types.MCPLog, 256), // 缓冲通道，避免阻塞
	}
	b.register <- client
	return client.Send
}

// UnregisterClient 注销客户端
func (b *BroadcastService) UnregisterClient(client *Client) {
	b.unregister <- client
}

// Close 关闭广播服务
func (b *BroadcastService) Close() {
	b.shutdown <- true
}

func GetTimeStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
