//go:build linux
// +build linux

package utils

import (
	"fmt"
)

func SetConsoleTitle(title string) {
	// 使用 ANSI 转义序列设置窗口标题
	fmt.Printf("\033]0;%s\007", title)
}
