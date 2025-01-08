package utils

import "time"

// IsDifferentDays 判断两个时间是否在不同的日历天上，根据指定时区
func IsDifferentDays(t1, t2 time.Time, name string) bool {
	location, _ := time.LoadLocation(name)
	// 将两个时间都转换到指定时区
	t1InLoc := t1.In(location)
	t2InLoc := t2.In(location)

	// 比较日期部分（年、月、日）
	return t1InLoc.Year() != t2InLoc.Year() ||
		t1InLoc.Month() != t2InLoc.Month() ||
		t1InLoc.Day() != t2InLoc.Day()
}
