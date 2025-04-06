package storage

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"GolemCore/internal/game"
)

var (
	ErrSnapshotNotFound = errors.New("快照不存在")
	ErrChecksumMismatch = errors.New("校验和不匹配")
)

// 系统状态快照
type SystemSnapshot struct {
	Timestamp  time.Time
	Grid       *game.Grid
	TaskQueue  []string // 待处理任务ID列表
	WorkerList []string // 在线节点地址列表
	Version    string   // 系统版本
}

type SnapshotConfig struct {
	StoragePath     string        // 存储目录
	RetentionPolicy time.Duration // 保留策略
	Compress        bool          // 启用压缩
}

type SnapshotManager struct {
	config    SnapshotConfig
	logger    Logger
	mu        sync.RWMutex
	latestSHA []byte
}

func NewSnapshotManager(config SnapshotConfig, logger Logger) *SnapshotManager {
	return &SnapshotManager{
		config: config,
		logger: logger,
	}
}

// 保存系统快照
func (m *SnapshotManager) Save(snap *SystemSnapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	// 注册游戏网格类型
	gob.Register(&game.Grid{})

	if err := enc.Encode(snap); err != nil {
		return fmt.Errorf("编码失败: %w", err)
	}

	// 计算校验和
	hash := sha256.Sum256(buf.Bytes())
	data := buf.Bytes()

	if m.config.Compress {
		var compressed bytes.Buffer
		gz := gzip.NewWriter(&compressed)
		if _, err := gz.Write(data); err != nil {
			return err
		}
		gz.Close()
		data = compressed.Bytes()
	}

	// 生成文件名
	timestamp := snap.Timestamp.Format("20060102-150405")
	filename := filepath.Join(m.config.StoragePath,
		fmt.Sprintf("snapshot-%s.gob%s", timestamp, m.getExtension()))

	if err := os.MkdirAll(m.config.StoragePath, 0755); err != nil {
		return fmt.Errorf("创建存储目录失败: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	// 更新最新校验和
	m.latestSHA = hash[:]

	// 清理旧快照
	go m.cleanupExpired()

	return nil
}

// 加载最新快照
func (m *SnapshotManager) LoadLatest() (*SystemSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries, err := os.ReadDir(m.config.StoragePath)
	if err != nil {
		return nil, ErrSnapshotNotFound
	}

	var latestFile string
	var latestTime time.Time

	for _, entry := range entries {
		if info, _ := entry.Info(); info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestFile = entry.Name()
		}
	}

	if latestFile == "" {
		return nil, ErrSnapshotNotFound
	}

	return m.loadFromFile(filepath.Join(m.config.StoragePath, latestFile))
}

// 从指定文件加载快照
func (m *SnapshotManager) loadFromFile(path string) (*SystemSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	if m.config.Compress {
		gz, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		data, err = io.ReadAll(gz)
		if err != nil {
			return nil, err
		}
	}

	// 验证校验和
	hash := sha256.Sum256(data)
	if !bytes.Equal(hash[:], m.latestSHA) {
		return nil, ErrChecksumMismatch
	}

	var snap SystemSnapshot
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&snap); err != nil {
		return nil, fmt.Errorf("解码失败: %w", err)
	}

	return &snap, nil
}

// 清理过期快照
func (m *SnapshotManager) cleanupExpired() {
	cutoff := time.Now().Add(-m.config.RetentionPolicy)

	entries, err := os.ReadDir(m.config.StoragePath)
	if err != nil {
		m.logger.Error("snapshot", "清理失败", Fields{"error": err.Error()})
		return
	}

	for _, entry := range entries {
		info, _ := entry.Info()
		if info.ModTime().Before(cutoff) {
			path := filepath.Join(m.config.StoragePath, entry.Name())
			if err := os.Remove(path); err != nil {
				m.logger.Warn("snapshot", "删除旧快照失败", Fields{
					"file":  entry.Name(),
					"error": err.Error(),
				})
			}
		}
	}
}

func (m *SnapshotManager) getExtension() string {
	if m.config.Compress {
		return ".gz"
	}
	return ""
}
