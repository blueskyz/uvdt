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
type uvdtLoger struct {
	Logger *log.Logger
}

var Logger uvdtLoger

func LoggerInit(logFile string) error {
	if len(logFile) == 0 {
		fmt.Printf("Open log file err: logFile is empty\n")
		return errors.New("Open log file err: logFile is empty\n")
	}
	file, err := os.OpenFile(logFile,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		666)
	if err != nil {
		fmt.Printf("Open log file err: %s, %s\n", err, logFile)
		return err
	}

	Logger.Logger = log.New(io.MultiWriter(file, os.Stderr),
		"",
		log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}

func (l *uvdtLoger) Printf(format string, v ...interface{}) {
	l.Logger.Output(3, fmt.Sprintf(format, v...))
}

func Info(msg string) {
	Logger.Printf("[I]: %s\n", msg)
}

func Err(msg string) error {
	Logger.Printf("[E]: %s\n", msg)
	return errors.New("msg")
}
