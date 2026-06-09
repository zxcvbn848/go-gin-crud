package socket

import (
	"net/http"

	"go-gin-crud/internal/logger"

	socketio "github.com/ismhdez/socket.io-golang/v4"
)

// Server 封裝 Socket.IO 伺服器，提供連線與事件處理
type Server struct {
	io *socketio.Io
}

// NewServer 建立並設定 Socket.IO 伺服器（連線、房間、自訂事件）
func NewServer() *Server {
	io := socketio.New()

	io.OnConnection(func(socket *socketio.Socket) {
		logger.Log.WithField("socket_id", socket.Id).Info("Socket.IO 客戶端連線")

		// 加入預設大廳房間（可依業務改為依 user/room 分房）
		socket.Join("lobby")

		// 自訂事件：加入指定房間
		socket.On("join_room", func(event *socketio.EventPayload) {
			if len(event.Data) > 0 && event.Data[0] != nil {
				if roomName, ok := event.Data[0].(string); ok {
					socket.Join(roomName)
					logger.Log.WithFields(map[string]interface{}{
						"socket_id": socket.Id,
						"room":      roomName,
					}).Info("加入房間")
					_ = socket.Emit("joined_room", roomName)
				}
			}
		})

		// 自訂事件：離開房間
		socket.On("leave_room", func(event *socketio.EventPayload) {
			if len(event.Data) > 0 && event.Data[0] != nil {
				if roomName, ok := event.Data[0].(string); ok {
					socket.Leave(roomName)
					logger.Log.WithFields(map[string]interface{}{
						"socket_id": socket.Id,
						"room":      roomName,
					}).Info("離開房間")
					_ = socket.Emit("left_room", roomName)
				}
			}
		})

		// 自訂事件：聊天訊息（可改為你需要的業務事件）
		socket.On("message", func(event *socketio.EventPayload) {
			var room, text string
			if len(event.Data) > 0 && event.Data[0] != nil {
				if r, ok := event.Data[0].(string); ok {
					room = r
				}
			}
			if len(event.Data) > 1 && event.Data[1] != nil {
				if t, ok := event.Data[1].(string); ok {
					text = t
				}
			}
			if room == "" {
				room = "lobby"
			}
			// 廣播給同房間的其他人（含自己可改為 io.To(room) 後不排除 sender）
			_ = socket.To(room).Emit("message", socket.Id, text)
			logger.Log.WithFields(map[string]interface{}{
				"socket_id": socket.Id,
				"room":      room,
				"text":      text,
			}).Debug("Socket 訊息")
		})

		// 斷線處理
		socket.On("disconnect", func(event *socketio.EventPayload) {
			logger.Log.WithField("socket_id", socket.Id).Info("Socket.IO 客戶端斷線")
		})
	})

	return &Server{io: io}
}

// IO 回傳底層 Socket.IO 伺服器，供廣播或進階操作使用
func (s *Server) IO() *socketio.Io {
	return s.io
}

// HttpHandler 回傳 http.Handler，給 Gin 用 gin.WrapH 掛載
func (s *Server) HttpHandler() http.Handler {
	return s.io.HttpHandler()
}
