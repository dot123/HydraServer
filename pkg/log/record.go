package log

import (
	"encoding/json"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	globalID  = 0
	logFields = make(map[int][]string)
	rlMap     = make(map[int]*rotatelogs.RotateLogs)
)

// Register 初始化记录器并注册日志字段
func Register(fileName, field string, rotationTime int, isJSON bool) int {
	var fields []string

	if isJSON {
		fields = strings.Split(field, ",")
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}
	}
	outputFile := "./log/" + fileName + ".log"
	dir := filepath.Dir(outputFile)
	ext := filepath.Ext(outputFile)
	_ = os.MkdirAll(dir, 0777)
	r, err := rotatelogs.New(strings.TrimSuffix(outputFile, ext)+"_%Y-%m-%d-%H-%M"+ext,
		rotatelogs.WithLinkName(outputFile),
		rotatelogs.WithRotationTime(time.Duration(rotationTime)*time.Second),
		rotatelogs.WithRotationCount(uint(48)))
	if err != nil {
		return 0
	}

	globalID++
	rlMap[globalID] = r
	logFields[globalID] = fields
	return globalID
}

// Write 根据日志ID和参数写入日志数据
func Write(logID int, args ...interface{}) {
	if _, ok := logFields[logID]; !ok {
		fmt.Println("Invalid log ID:", logID)
		return
	}

	now := time.Now()
	date := now.Format("2006-01-02 15:04:05")
	fields := logFields[logID]

	var b []byte
	if len(fields) > 0 {
		jsonTable := make(map[string]interface{}, len(fields))
		if len(args) != len(fields) {
			fmt.Println("Invalid number of arguments for log ID:", logID)
			return
		}

		for i, field := range fields {
			jsonTable[field] = args[i]
		}
		jsonTable["time"] = now.Unix()
		jsonTable["date"] = date

		var err error
		b, err = json.Marshal(jsonTable)
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)
			return
		}
		b = []byte(fmt.Sprintf("%s,\"%s\"\n", date, string(b)))
	} else {
		var strArgs []string
		for _, arg := range args {
			if str, ok := arg.(string); ok {
				strArgs = append(strArgs, fmt.Sprintf("\"%s\"", str))
			} else {
				strArgs = append(strArgs, fmt.Sprintf("%v", arg))
			}
		}
		b = []byte(fmt.Sprintf("%s,%s\n", date, strings.Join(strArgs, ",")))
	}

	if rl, ok := rlMap[logID]; ok {
		if runtime.GOOS == "windows" {
			fmt.Print(string(b))
		}
		if _, err := rl.Write(b); err != nil {
			fmt.Println("Error writing to log:", err)
			return
		}
	}
}
