package sfu

import (
	"sort"
	"sync"
)

// speakerAlpha EMA 平滑系数：越大越跟手（新样本权重高），越小越平滑。
const speakerAlpha = 0.2

// ActiveSpeakers 按 RFC6464 音量的指数滑动平均维护各发言人响度，供 top-N 选路 + 高亮。
// 音量输入 0..127（0=最响，127=静音）；内部转成响度 = 127-level（越大越响）。并发安全。
type ActiveSpeakers struct {
	mu  sync.Mutex
	ema map[string]float64
}

// NewActiveSpeakers 创建空追踪器。
func NewActiveSpeakers() *ActiveSpeakers {
	return &ActiveSpeakers{ema: make(map[string]float64)}
}

// Observe 喂入 uid 的一个音量样本（0..127，0=最响）。
func (a *ActiveSpeakers) Observe(uid string, level uint8) {
	loud := float64(127 - int(level))
	a.mu.Lock()
	if cur, ok := a.ema[uid]; ok {
		a.ema[uid] = speakerAlpha*loud + (1-speakerAlpha)*cur
	} else {
		a.ema[uid] = loud
	}
	a.mu.Unlock()
}

// Remove 移除某发言人（离开/停止发布音频时）。
func (a *ActiveSpeakers) Remove(uid string) {
	a.mu.Lock()
	delete(a.ema, uid)
	a.mu.Unlock()
}

// Top 返回当前响度最高的 n 个 uid（响的在前）；n<=0 返回空，n 过大则返回全部。
func (a *ActiveSpeakers) Top(n int) []string {
	if n <= 0 {
		return nil
	}
	a.mu.Lock()
	type kv struct {
		uid  string
		loud float64
	}
	items := make([]kv, 0, len(a.ema))
	for uid, loud := range a.ema {
		items = append(items, kv{uid, loud})
	}
	a.mu.Unlock()

	sort.Slice(items, func(i, j int) bool {
		if items[i].loud != items[j].loud {
			return items[i].loud > items[j].loud
		}
		return items[i].uid < items[j].uid // 稳定排序：响度相同按 uid
	})
	if n > len(items) {
		n = len(items)
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, items[i].uid)
	}
	return out
}
