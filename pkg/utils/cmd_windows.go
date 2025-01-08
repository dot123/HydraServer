//go:build windows
// +build windows

package utils

import (
	"syscall"
	"unsafe"
)

var (
	modkernel32  = syscall.NewLazyDLL("kernel32.dll")
	procSetTitle = modkernel32.NewProc("SetConsoleTitleW")
)

func SetConsoleTitle(title string) {
	titleUTF16, err := syscall.UTF16FromString(title)
	if err != nil {
		panic(err)
	}

	ret, _, err := procSetTitle.Call(uintptr(unsafe.Pointer(&titleUTF16[0])))
	if ret == 0 {
		panic(err)
	}
}
