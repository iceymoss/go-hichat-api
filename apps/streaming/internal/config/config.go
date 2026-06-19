package config

import (
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	service.ServiceConf

	ListenOn string

	// JwtAuth 与 im/user 共用同一 AccessSecret，校验 ws 升级与 HTTP 接口的 JWT
	JwtAuth struct {
		AccessSecret string
		AccessExpire int64
	}

	Mysql struct {
		DataSource string
	}

	Cache cache.CacheConf

	// WebRTC 配置
	WebRTC struct {
		// ICE服务器配置
		IceServers []struct {
			URLs       []string `yaml:"URLs"`
			Username   string   `yaml:"Username,omitempty"`
			Credential string   `yaml:"Credential,omitempty"`
		} `yaml:"IceServers"`

		// 媒体配置
		Media struct {
			// 视频配置
			Video struct {
				Width     int `yaml:"Width"`
				Height    int `yaml:"Height"`
				FrameRate int `yaml:"FrameRate"`
				Bitrate   int `yaml:"Bitrate"`
			} `yaml:"Video"`

			// 音频配置
			Audio struct {
				SampleRate int `yaml:"SampleRate"`
				Channels   int `yaml:"Channels"`
				Bitrate    int `yaml:"Bitrate"`
			} `yaml:"Audio"`
		} `yaml:"Media"`
	}

	// SFU 配置
	SFU struct {
		// 最大房间数
		MaxRooms int `yaml:"MaxRooms"`
		// 每个房间最大用户数
		MaxUsersPerRoom int `yaml:"MaxUsersPerRoom"`
		// 房间超时时间(秒)
		RoomTimeout int `yaml:"RoomTimeout"`
		// 用户连接超时时间(秒)
		UserTimeout int `yaml:"UserTimeout"`
	}

	// 信令服务器配置
	Signaling struct {
		// WebSocket 配置
		WebSocket struct {
			ReadBufferSize  int  `yaml:"ReadBufferSize"`
			WriteBufferSize int  `yaml:"WriteBufferSize"`
			CheckOrigin     bool `yaml:"CheckOrigin"`
		} `yaml:"WebSocket"`

		// 消息队列配置
		MessageQueue struct {
			BufferSize  int `yaml:"BufferSize"`
			WorkerCount int `yaml:"WorkerCount"`
		} `yaml:"MessageQueue"`
	}

	// 各个微服务的RPC配置
	SocialRpc zrpc.RpcClientConf
	UserRpc   zrpc.RpcClientConf
	ImRpc     zrpc.RpcClientConf
}
