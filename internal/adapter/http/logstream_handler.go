package handler

import (
	"net/http"

	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/registry"
	"github.com/Haevnen/audit-logging-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type LogStreamHandler struct {
	Pubsub service.PubSub
}

func newLogStreamHandler(r *registry.Registry) LogStreamHandler {
	return LogStreamHandler{Pubsub: r.PubSub()}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// StreamLogs implements GET /api/v1/logs/stream
func (h LogStreamHandler) StreamLogs(c *gin.Context, params api_service.StreamLogsParams) {
	// 1. Determine channel
	ctx := c.Request.Context()
	tenantId := getClaimTenant(c)

	channel := "logs"
	if tenantId != "" {
		channel = "logs:" + tenantId
	}

	// 2. Upgrade to WebSocket and subscribe
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	sub := h.Pubsub.Subscribe(ctx, channel)
	defer sub.Close()

	ch := sub.Channel(redis.WithChannelSize(100))
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			if msg == nil {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				return
			}
		}
	}
}
