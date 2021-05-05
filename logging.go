package gologger

import (
	"fmt"
	"io"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	LOG_DEBUG = iota
	LOG_INFO
	LOG_WARN
	LOG_ERROR
)

type Logger struct {
	writer io.WriteCloser
	name   string
	lock   sync.Mutex
	config *logConfig
}

type logConfig struct {
	simple     bool
	writeLevel int
	tagAlign   int
}

func NewLogger() *Logger {
	return &Logger{
		writer: nil,
		name:   "main",
		config: &logConfig{
			simple:     false,
			writeLevel: LOG_DEBUG,
			tagAlign:   0,
		},
	}
}

func (l *Logger) Module(name string) *Logger {
	return &Logger{
		writer: l.writer,
		name:   name,
		config: l.config,
	}
}

func (l *Logger) SetLevel(level int) *Logger {
	l.config.writeLevel = level
	return l
}

func (l *Logger) SetSimple(simple bool) *Logger {
	l.config.simple = simple
	return l
}

func (l *Logger) SetFile(path string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	l.writer = file
	return nil
}

func (l *Logger) SetWriter(writer io.WriteCloser) *Logger {
	l.writer = writer
	return l
}

func (l *Logger) SetTagAlign(n int) *Logger {
	l.config.tagAlign = n
	return l
}

func formatLevel(level int) string {
	switch level {
	case LOG_DEBUG:
		return "\033[92m[Debug]\033[0m"
	case LOG_INFO:
		return "\033[96m[Info]\033[0m"
	case LOG_WARN:
		return "\033[93m[Warn]\033[0m"
	case LOG_ERROR:
		return "\033[91m[Error]\033[0m"
	default:
		return "UNKNOWN"
	}
}

func (l *Logger) formatMessage(level int, message []interface{}) string {
	now := time.Now()
	var builder strings.Builder
	builder.WriteRune('[')
	if l.config.simple {
		builder.WriteString(now.Format("15:04:05"))
	} else {
		builder.WriteString(now.Format("2006/01/02 15:04:05.0000"))
	}
	builder.WriteString("] [")
	if l.config.tagAlign != 0 && l.config.tagAlign > len(l.name) {
		builder.WriteString(strings.Repeat(" ", int(math.Floor(float64(l.config.tagAlign-len(l.name))/2))))
		builder.WriteString(l.name)
		builder.WriteString(strings.Repeat(" ", int(math.Ceil(float64(l.config.tagAlign-len(l.name))/2))))
	} else {
		builder.WriteString(l.name)
	}
	builder.WriteString("] ")
	builder.WriteString(formatLevel(level))
	builder.WriteRune(' ')
	if !l.config.simple {
		builder.WriteRune('[')
		pc, _, line, _ := runtime.Caller(3)
		builder.WriteString(runtime.FuncForPC(pc).Name())
		builder.WriteRune(':')
		builder.WriteString(strconv.Itoa(line))
		builder.WriteString("] ")
	}
	for i, d := range message {
		builder.WriteString(fmt.Sprintf("%v", d))
		if i != len(message)-1 {
			builder.WriteRune(' ')
		}
	}
	return builder.String()
}

func (l *Logger) Log(level int, message ...interface{}) {
	if level < l.config.writeLevel {
		return
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	mess := l.formatMessage(level, message)
	fmt.Println(mess)
	if l.writer != nil {
		l.writer.Write([]byte(mess + "\n"))
	}
}

func (l *Logger) Debug(message ...interface{}) {
	l.Log(LOG_DEBUG, message...)
}

func (l *Logger) Info(message ...interface{}) {
	l.Log(LOG_INFO, message...)
}

func (l *Logger) Warn(message ...interface{}) {
	l.Log(LOG_WARN, message...)
}

func (l *Logger) Error(message ...interface{}) {
	l.Log(LOG_ERROR, message...)
}

func (l *Logger) Close() {
	if l.writer != nil {
		l.writer.Close()
	}
}
