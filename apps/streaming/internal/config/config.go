package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

// ApplyEnvironment 用部署环境变量覆盖公网 WebRTC 配置，避免 coturn 与 streaming 分别维护密钥。
func (c *Config) ApplyEnvironment() error {
	if value := strings.TrimSpace(os.Getenv("PUBLIC_IP")); value != "" {
		c.Public.IP = value
	}
	var err error
	if c.Public.UDPPortMin, err = envInt("SFU_UDP_MIN_PORT", c.Public.UDPPortMin); err != nil {
		return err
	}
	if c.Public.UDPPortMax, err = envInt("SFU_UDP_MAX_PORT", c.Public.UDPPortMax); err != nil {
		return err
	}
	if (c.Public.UDPPortMin > 0 || c.Public.UDPPortMax > 0) &&
		(c.Public.UDPPortMin < 1 || c.Public.UDPPortMax > 65535 || c.Public.UDPPortMin > c.Public.UDPPortMax) {
		return fmt.Errorf("invalid SFU UDP port range %d-%d", c.Public.UDPPortMin, c.Public.UDPPortMax)
	}
	if value := strings.TrimSpace(os.Getenv("TURN_URLS")); value != "" {
		c.Turn.URLs = nil
		for _, raw := range strings.Split(value, ",") {
			if url := strings.TrimSpace(raw); url != "" {
				c.Turn.URLs = append(c.Turn.URLs, url)
			}
		}
	}
	if value := os.Getenv("TURN_SECRET"); value != "" {
		c.Turn.Secret = value
	}
	if value := strings.TrimSpace(os.Getenv("TURN_TTL_SECONDS")); value != "" {
		ttl, parseErr := strconv.ParseInt(value, 10, 64)
		if parseErr != nil || ttl <= 0 {
			return fmt.Errorf("invalid TURN_TTL_SECONDS %q", value)
		}
		c.Turn.TTLSeconds = ttl
	}
	return nil
}

func envInt(name string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q", name, value)
	}
	return parsed, nil
}

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

	// Diagnostics 将浏览器和服务端 SFU 状态汇总为本地 JSONL，便于复盘多人通话。
	Diagnostics struct {
		Enabled bool
		Path    string
		MaxMB   int
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

	// Ws im ws 地址：streaming 以系统 root 身份连入，推送通话控制信令（push.call）
	Ws struct {
		Host string
	}

	// Call 通话相关参数
	Call struct {
		// RingTimeoutSeconds 振铃超时（秒），超时未接转“未接听”。默认 30。
		RingTimeoutSeconds int
	}

	// Public SFU 公网化（Phase 1）：跨 NAT 真实可用所需
	Public struct {
		// IP 公网 IP（NAT1To1），SFU 对外宣告候选用；空则用本机采集的候选（本地/局域网）
		IP string
		// UDPPortMin/Max 媒体 UDP 端口范围，需在防火墙/docker 放行；<=0 用 pion 默认随机端口
		UDPPortMin int
		UDPPortMax int
	}

	// Turn coturn 中继（Phase 1）：REST(use-auth-secret) 短期凭证，穿透对称 NAT
	Turn struct {
		// URLs 下发给前端的 turn 地址，如 turn:example.com:3478?transport=udp
		URLs []string
		// Secret coturn static-auth-secret，签发短期凭证用；走配置不写死、不入库
		Secret string
		// TTLSeconds 短期凭证有效期（秒），默认 3600
		TTLSeconds int64
	}
}
