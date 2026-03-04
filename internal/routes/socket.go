package routes

import (
	"go-gin-crud/internal/socket"

	"github.com/gin-gonic/gin"
)

var socketServer *socket.Server

func init() {
	socketServer = socket.NewServer()
}

// RegisterSocketRoutes 註冊 Socket.IO 路由，掛載在 /socket.io/
// 前端需使用 socket.io-client 連到相同 path（例如 ws://localhost:8080/socket.io/）
func RegisterSocketRoutes(r *gin.Engine) {
	// Socket.IO 需要處理 /socket.io/ 前綴的所有請求（含 GET/POST 等）
	r.GET("/socket.io/", gin.WrapH(socketServer.HttpHandler()))
	r.POST("/socket.io/", gin.WrapH(socketServer.HttpHandler()))
}

// GetSocketServer 供其他套件需要廣播時取得 Socket.IO 伺服器
func GetSocketServer() *socket.Server {
	return socketServer
}
