package main

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"time"
)

// simple timestamp'd logger
func LogRaw(prefix string, args ...interface{}) {
	pc, _, line, _ := runtime.Caller(2)
	fun := runtime.FuncForPC(pc)
	now := time.Now().Format("2006-01-02 15:04:05")
	buffer := new(bytes.Buffer)
	buffer.WriteString(fmt.Sprintf("%s [%5s] :: [", now, prefix))
	buffer.WriteString(fun.Name())
	buffer.WriteString(":")
	buffer.WriteString(strconv.Itoa(line))
	buffer.WriteString("]: ")
	for _, arg := range args {
		fmt.Fprint(buffer, arg)
	}
	fmt.Println(buffer.String())
}

func INFO(args ...interface{}) {
	LogRaw("INFO", args...)
}

func ERROR(args ...interface{}) {
	LogRaw("ERROR", args...)
}
