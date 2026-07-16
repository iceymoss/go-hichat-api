package sfu

import "github.com/pion/rtp"

// rtpReader 上行 RTP 源的最小抽象。生产中由 *webrtc.TrackRemote 适配（其 ReadRTP 多返回
// interceptor.Attributes，用一层薄封装丢弃即可）；测试用假源。
type rtpReader interface {
	ReadRTP() (*rtp.Packet, error)
}

// pump 收流泵：不断从上行源 src 读 RTP 包，按 (pubUID, trackID) 路由进房间做 fan-out，
// 直到 src 返回错误（EOF / 连接关闭）才退出。生产中每个 OnTrack 起一个 goroutine 跑它。
func pump(room *Room, pubUID, trackID, rid string, src rtpReader, audioLevelExtensionID uint8, activeSpeakerLimit int) {
	if audioLevelExtensionID > 0 {
		room.ManageAudio(pubUID)
	}
	for {
		pkt, err := src.ReadRTP()
		if err != nil {
			return
		}
		if level, ok := audioLevel(pkt, audioLevelExtensionID); ok {
			room.ObserveAudioLevel(pubUID, level, activeSpeakerLimit)
		}
		room.RouteRTPLayer(pubUID, trackID, rid, pkt)
	}
}
