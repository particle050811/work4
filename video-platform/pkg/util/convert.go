package util

import (
	"strconv"
	"strings"
	"time"
)

// ParseUint 将字符串解析为 uint 类型
// 会自动去除首尾空格，支持 10 进制正整数
func ParseUint(s string) (uint, error) {
	u64, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(u64), nil
}

// NormalizePage 标准化分页参数
// 返回: pageNum(≥1), pageSize(10-50), offset
func NormalizePage(pageNum, pageSize int32) (int, int, int) {
	pn := int(pageNum)
	ps := int(pageSize)
	if pn < 1 {
		pn = 1
	}
	if ps <= 0 {
		ps = 10
	}
	if ps > 50 {
		ps = 50
	}
	return pn, ps, (pn - 1) * ps
}

// ParseUnixRange 将 Unix 时间戳范围转换为时间指针
// 支持秒级和毫秒级时间戳，0 或负数表示不限制
func ParseUnixRange(fromDate, toDate int64) (*time.Time, *time.Time) {
	var from, to *time.Time
	if fromDate > 0 {
		t := UnixToTime(fromDate)
		from = &t
	}
	if toDate > 0 {
		t := UnixToTime(toDate)
		to = &t
	}
	return from, to
}

// UnixToTime 将 Unix 时间戳转换为 time.Time
// 自动兼容毫秒级（>1e12）和秒级时间戳
func UnixToTime(ts int64) time.Time {
	// 兼容毫秒/秒时间戳
	if ts > 1_000_000_000_000 {
		return time.UnixMilli(ts)
	}
	return time.Unix(ts, 0)
}
