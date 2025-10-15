package config

import (
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	service.ServiceConf

	ListenOn string

	Mysql struct {
		DataSource string
	}

	Cache cache.CacheConf

	// WebRTC 配置
	WebRTC struct {
		// ICE服务器配置
		IceServers []struct {
			URLs       []string `json:"URLs,optional"`
			Username   string   `json:"Username,optional"`
			Credential string   `json:"Credential,optional"`
		}

		// 媒体配置
		Media struct {
			// 视频配置
			Video struct {
				Width     int `json:"Width,optional"`
				Height    int `json:"Height,optional"`
				FrameRate int `json:"FrameRate,optional"`
				Bitrate   int `json:"Bitrate,optional"`
			}

			// 音频配置
			Audio struct {
				SampleRate int `json:"SampleRate,optional"`
				Channels   int `json:"Channels,optional"`
				Bitrate    int `json:"Bitrate,optional"`
			}
		}
	}

	// SFU 配置
	SFU struct {
		// 最大房间数
		MaxRooms int
		// 每个房间最大用户数
		MaxUsersPerRoom int
		// 房间超时时间(秒)
		RoomTimeout int
		// 用户连接超时时间(秒)
		UserTimeout int
	}

	// 信令服务器配置
	Signaling struct {
		// WebSocket 配置
		WebSocket struct {
			ReadBufferSize  int
			WriteBufferSize int
			CheckOrigin     bool
		}

		// 消息队列配置
		MessageQueue struct {
			BufferSize  int
			WorkerCount int
		}
	}

	// 各个微服务的RPC配置
	SocialRpc zrpc.RpcClientConf
	UserRpc   zrpc.RpcClientConf
	ImRpc     zrpc.RpcClientConf
}
