package utils

import (
	"time"
)

var (
	// ChinaLocation 中国时区 (UTC+8)
	ChinaLocation *time.Location
)

func init() {
	var err error
	ChinaLocation, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 如果加载失败，使用固定偏移量 UTC+8
		ChinaLocation = time.FixedZone("CST", 8*60*60)
	}
}

// NowInChina 获取中国时区的当前时间
func NowInChina() time.Time {
	return time.Now().In(ChinaLocation)
}

// UnixToChinaTime 将Unix时间戳转换为中国时区的时间
func UnixToChinaTime(sec int64) time.Time {
	return time.Unix(sec, 0).In(ChinaLocation)
}

// TimeToChinaUnix 将时间转换为中国时区的Unix时间戳
func TimeToChinaUnix(t time.Time) int64 {
	return t.In(ChinaLocation).Unix()
}
