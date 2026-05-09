package helper

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// WsUpgrader WebSocket 升级器（全局共享）
// 安全策略：仅允许同源连接，禁止跨域 WebSocket
var WsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://"+r.Host || origin == "https://"+r.Host
	},
}
