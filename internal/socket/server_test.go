package socket

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestMessageBroadcastExcludesSender(t *testing.T) {
	server := httptest.NewServer(NewServer().HttpHandler())
	t.Cleanup(server.Close)

	connect := func() (*websocket.Conn, string) {
		t.Helper()
		conn, _, err := websocket.DefaultDialer.Dial(
			"ws"+strings.TrimPrefix(server.URL, "http")+"/socket.io/?EIO=4&transport=websocket", nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = conn.Close() })
		require.NoError(t, conn.SetReadDeadline(time.Now().Add(10*time.Second)))
		_, packet, err := conn.ReadMessage()
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(string(packet), "0{"), string(packet))
		require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte("40")))
		_, packet, err = conn.ReadMessage()
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(string(packet), "40{"), string(packet))
		var handshake struct {
			SID string `json:"sid"`
		}
		require.NoError(t, json.Unmarshal(packet[2:], &handshake))
		require.NotEmpty(t, handshake.SID)
		return conn, handshake.SID
	}
	readEvent := func(conn *websocket.Conn) []string {
		t.Helper()
		_, packet, err := conn.ReadMessage()
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(string(packet), "42["), string(packet))
		var event []string
		require.NoError(t, json.Unmarshal(packet[2:], &event))
		return event
	}
	send := func(conn *websocket.Conn, event ...string) {
		t.Helper()
		packet, err := json.Marshal(event)
		require.NoError(t, err)
		require.NoError(t, conn.WriteMessage(websocket.TextMessage, append([]byte("42"), packet...)))
	}
	join := func(conn *websocket.Conn, room string) {
		t.Helper()
		send(conn, "join_room", room)
		require.Equal(t, []string{"joined_room", room}, readEvent(conn))
	}

	sender, senderID := connect()
	peer, _ := connect()
	outsider, _ := connect()
	join(sender, "test-room")
	join(peer, "test-room")

	send(sender, "message", "test-room", "hello")
	require.Equal(t, []string{"message", senderID, "hello"}, readEvent(peer))
	// 後續事件的回覆作為屏障，若前面收到多餘廣播便立即失敗，不靠等待逾時。
	join(sender, "sender-barrier")
	join(outsider, "outsider-barrier")

	send(peer, "leave_room", "test-room")
	require.Equal(t, []string{"left_room", "test-room"}, readEvent(peer))
	send(sender, "message", "test-room", "after leave")
	join(sender, "after-leave-barrier")
	join(peer, "peer-barrier")

	// 空房間參數仍應送往預設大廳，並排除發送者。
	send(sender, "message", "", "lobby message")
	for _, conn := range []*websocket.Conn{peer, outsider} {
		require.Equal(t, []string{"message", senderID, "lobby message"}, readEvent(conn))
	}
	join(sender, "lobby-barrier")
}
