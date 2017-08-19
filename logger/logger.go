/*
	logger 日志记录组件
*/
package logger

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

//---------------------------------------------------------------
// Logger 设置
var Logger *log.Logger

func LoggerInit(logFile string) {
	file, err := os.OpenFile(logFile,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		666)
	if err != nil {
		fmt.Printf("Open log file err: %s\n", logFile)
	}

	Logger = log.New(io.MultiWriter(file, os.Stderr),
		"",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func Info(msg string) {
	Logger.Printf("[I]: %s\n", msg)
}

func Err(msg string) error {
	Logger.Printf("[E]: %s\n", msg)
	return errors.New("msg")
}
