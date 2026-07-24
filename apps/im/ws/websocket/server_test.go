package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	gorilla "github.com/gorilla/websocket"
	"github.com/iceymoss/go-hichat-api/pkg/presence"
	"github.com/stretchr/testify/require"
)

type queryAuth struct{}

func (queryAuth) Auth(http.ResponseWriter, *http.Request) bool { return true }
func (queryAuth) UserId(r *http.Request) string                { return r.URL.Query().Get("uid") }

func dialTestConn(t *testing.T, url, uid string) *gorilla.Conn {
	t.Helper()
	conn, _, err := gorilla.DefaultDialer.Dial(strings.Replace(url, "http", "ws", 1)+"?uid="+uid, nil)
	require.NoError(t, err)
	return conn
}
func waitConnCount(t *testing.T, s *Server, uid string, want int) {
	t.Helper()
	require.Eventually(t, func() bool { return len(s.GetConn([]string{uid})) == want }, time.Second, 10*time.Millisecond)
}

func TestServerSupportsMultipleConnectionsPerUser(t *testing.T) {
	srv := NewServer("", WithAuthentication(queryAuth{}), WithMaxConnectionIdle(time.Minute))
	httpServer := httptest.NewServer(http.HandlerFunc(srv.ServerWs))
	defer httpServer.Close()
	first := dialTestConn(t, httpServer.URL, "7")
	defer first.Close()
	second := dialTestConn(t, httpServer.URL, "7")
	defer second.Close()
	waitConnCount(t, srv, "7", 2)
	require.Equal(t, []string{"7"}, srv.GetUsers(nil))
	require.NoError(t, srv.SendByUserId(&Message{Method: "test", Data: "hello"}, "7"))
	for _, client := range []*gorilla.Conn{first, second} {
		_, body, err := client.ReadMessage()
		require.NoError(t, err)
		require.Contains(t, string(body), "hello")
	}
	require.NoError(t, first.Close())
	waitConnCount(t, srv, "7", 1)
}

func TestServerSendIsolatesFailedConnection(t *testing.T) {
	srv := NewServer("", WithAuthentication(queryAuth{}), WithMaxConnectionIdle(time.Minute))
	httpServer := httptest.NewServer(http.HandlerFunc(srv.ServerWs))
	defer httpServer.Close()
	first := dialTestConn(t, httpServer.URL, "9")
	defer first.Close()
	second := dialTestConn(t, httpServer.URL, "9")
	defer second.Close()
	waitConnCount(t, srv, "9", 2)
	connections := srv.GetConn([]string{"9", "9"})
	require.Len(t, connections, 2)
	srv.opt.writeTimeout = 0
	require.NoError(t, connections[0].Conn.SetWriteDeadline(time.Now().Add(-time.Second)))
	err := srv.SendByUserId(&Message{Method: "test", Data: "survives"}, "9")
	require.Error(t, err)
	received := false
	for _, client := range []*gorilla.Conn{first, second} {
		_ = client.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		_, body, readErr := client.ReadMessage()
		if readErr == nil && strings.Contains(string(body), "survives") {
			received = true
		}
	}
	require.True(t, received)
	require.Eventually(t, func() bool { return len(srv.GetConn([]string{"9"})) == 1 }, time.Second, 10*time.Millisecond)
}

func TestServerPresenceUsesFirstAndLastConnection(t *testing.T) {
	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	srv := NewServer("", WithAuthentication(queryAuth{}), WithPresence("node-a", presence.New(client), time.Minute, 10*time.Second), WithMaxConnectionIdle(time.Minute))
	httpServer := httptest.NewServer(http.HandlerFunc(srv.ServerWs))
	defer httpServer.Close()
	first := dialTestConn(t, httpServer.URL, "8")
	second := dialTestConn(t, httpServer.URL, "8")
	waitConnCount(t, srv, "8", 2)
	value, err := redisServer.Get(presence.Key("8"))
	require.NoError(t, err)
	require.Equal(t, "node-a", value)
	require.NoError(t, first.Close())
	waitConnCount(t, srv, "8", 1)
	require.True(t, redisServer.Exists(presence.Key("8")))
	require.NoError(t, second.Close())
	waitConnCount(t, srv, "8", 0)
	require.Eventually(t, func() bool { return !redisServer.Exists(presence.Key("8")) }, time.Second, 10*time.Millisecond)
}
