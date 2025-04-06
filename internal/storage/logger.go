package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// 日志级别类型
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var levelStrings = [...]string{"DEBUG", "INFO", "WARN", "ERROR"}

// 日志条目结构
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     LogLevel  `json:"level"`
	Component string    `json:"component"`
	Message   string    `json:"message"`
	Extra     Fields    `json:"extra,omitempty"`
}

type Fields map[string]interface{}

// 日志记录器接口
type Logger interface {
	Debug(component, message string, fields Fields)
	Info(component, message string, fields Fields)
	Warn(component, message string, fields Fields)
	Error(component, message string, fields Fields)
	Close() error
}

// 实现旋转文件日志记录器
type RotatingFileLogger struct {
	currentFile *os.File
	mu          sync.Mutex
	config      LogConfig
	queue       chan LogEntry
	done        chan struct{}
}

type LogConfig struct {
	Directory    string        // 日志目录
	MaxSize      int64         // 单个文件最大大小（字节）
	RotatePeriod time.Duration // 旋转周期
	QueueSize    int           // 异步队列缓冲区
}

func NewRotatingFileLogger(config LogConfig) (*RotatingFileLogger, error) {
	if err := os.MkdirAll(config.Directory, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	l := &RotatingFileLogger{
		config: config,
		queue:  make(chan LogEntry, config.QueueSize),
		done:   make(chan struct{}),
	}

	if err := l.rotate(); err != nil {
		return nil, err
	}

	go l.processEntries()
	return l, nil
}

func (l *RotatingFileLogger) log(level LogLevel, component, message string, fields Fields) {
	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     level,
		Component: component,
		Message:   message,
		Extra:     fields,
	}

	select {
	case l.queue <- entry:
	default:
		// 队列满时直接输出到stderr
		fmt.Fprintf(os.Stderr, "日志队列已满，丢弃条目: %+v\n", entry)
	}
}

func (l *RotatingFileLogger) processEntries() {
	for entry := range l.queue {
		data, _ := json.Marshal(entry)
		data = append(data, '\n')

		l.mu.Lock()
		if stats, _ := l.currentFile.Stat(); stats.Size() > l.config.MaxSize {
			l.rotate()
		}
		l.currentFile.Write(data)
		l.mu.Unlock()
	}
	close(l.done)
}

func (l *RotatingFileLogger) rotate() error {
	if l.currentFile != nil {
		l.currentFile.Close()
	}

	pattern := time.Now().Format("20060102-150405")
	filename := filepath.Join(l.config.Directory, fmt.Sprintf("golem-%s.log", pattern))

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("创建日志文件失败: %w", err)
	}

	l.currentFile = file
	return nil
}

// 实现Logger接口方法
func (l *RotatingFileLogger) Debug(component, message string, fields Fields) {
	l.log(DEBUG, component, message, fields)
}

func (l *RotatingFileLogger) Info(component, message string, fields Fields) {
	l.log(INFO, component, message, fields)
}

func (l *RotatingFileLogger) Warn(component, message string, fields Fields) {
	l.log(WARN, component, message, fields)
}

func (l *RotatingFileLogger) Error(component, message string, fields Fields) {
	l.log(ERROR, component, message, fields)
}

func (l *RotatingFileLogger) Close() error {
	close(l.queue)
	<-l.done
	return l.currentFile.Close()
}
