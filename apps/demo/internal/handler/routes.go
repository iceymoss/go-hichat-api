package handler

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/demo/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/ws",
				Handler: WebSocketHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/status",
				Handler: StatusHandler(serverCtx),
			},
		},
	)
}
