/*
	logger 日志记录组件
*/
package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

//---------------------------------------------------------------
// Logger 设置
type uvdtLogger struct {
	Logger *log.Logger
}

var Logger uvdtLogger

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

	// Logger.Logger = log.New(io.MultiWriter(file, os.Stderr),
	Logger.Logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}

func (l *uvdtLogger) Printf(format string, v ...interface{}) {
	l.Logger.Output(3, fmt.Sprintf(format, v...))
}

func NewAgent() LogAgent {
	logAgent := LogAgent{
		buffers: []string{},
		status:  "[I]: ",
	}
	return logAgent
}

// log 记录代理，缓存单次请求过程中的日志，在处理函数结束时记录日志
type LogAgent struct {
	logger_impl *uvdtLogger
	buffers     []string
	status      string
}

func (logAgent *LogAgent) Info(msg string) {
	logAgent.buffers = append(logAgent.buffers, fmt.Sprintf("%s", msg))
}

func (logAgent *LogAgent) Err(msg string) error {
	logAgent.buffers = append(logAgent.buffers, fmt.Sprintf("[E] %s", msg))
	logAgent.status = "[E]: "
	return errors.New("msg")
}

func (logAgent *LogAgent) EndLog() {
	msg := logAgent.status + strings.Join(logAgent.buffers, " | ")
	Logger.Printf("%s\n", msg)
	logAgent.buffers = []string{}
	logAgent.status = "[I]: "
}
