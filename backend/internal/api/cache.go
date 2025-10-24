package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CacheConfig 缓存配置
type CacheConfig struct {
	CacheDir      string
	MaxAge        time.Duration
	MaxSize       int64
	Compress      bool
	Encrypt       bool
	EncryptionKey string
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	CreatedAt time.Time   `json:"created_at"`
	ExpiresAt time.Time   `json:"expires_at"`
	Size      int64       `json:"size"`
}

// Cache 缓存接口
type Cache interface {
	// Get 获取缓存
	Get(key string, value interface{}) error

	// Set 设置缓存
	Set(key string, value interface{}, ttl time.Duration) error

	// Delete 删除缓存
	Delete(key string) error

	// Exists 检查缓存是否存在
	Exists(key string) bool

	// Clear 清空缓存
	Clear() error

	// GetStats 获取缓存统计
	GetStats() CacheStats

	// SetConfig 设置缓存配置
	SetConfig(config CacheConfig)

	// GetConfig 获取缓存配置
	GetConfig() CacheConfig
}

// CacheStats 缓存统计
type CacheStats struct {
	TotalKeys   int
	TotalSize   int64
	HitRate     float64
	MissRate    float64
	OldestEntry time.Time
	NewestEntry time.Time
}

// fileCacheImpl 文件缓存实现
type fileCacheImpl struct {
	config CacheConfig
	stats  CacheStats
}

// NewFileCache 创建新的文件缓存
func NewFileCache(config CacheConfig) Cache {
	// 确保缓存目录存在
	if err := os.MkdirAll(config.CacheDir, 0755); err != nil {
		panic(fmt.Sprintf("创建缓存目录失败: %v", err))
	}

	return &fileCacheImpl{
		config: config,
	}
}

// Get 获取缓存
func (c *fileCacheImpl) Get(key string, value interface{}) error {
	cachePath := c.getCachePath(key)

	// 检查文件是否存在
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		c.stats.MissRate++
		return fmt.Errorf("缓存不存在")
	}

	// 读取缓存文件
	data, err := os.ReadFile(cachePath)
	if err != nil {
		c.stats.MissRate++
		return fmt.Errorf("读取缓存文件失败: %v", err)
	}

	// 解析缓存条目
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		c.stats.MissRate++
		return fmt.Errorf("解析缓存文件失败: %v", err)
	}

	// 检查是否过期
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		c.stats.MissRate++
		os.Remove(cachePath) // 删除过期缓存
		return fmt.Errorf("缓存已过期")
	}

	// 反序列化值
	if err := json.Unmarshal([]byte(entry.Value.(string)), value); err != nil {
		c.stats.MissRate++
		return fmt.Errorf("反序列化缓存值失败: %v", err)
	}

	c.stats.HitRate++
	return nil
}

// Set 设置缓存
func (c *fileCacheImpl) Set(key string, value interface{}, ttl time.Duration) error {
	cachePath := c.getCachePath(key)

	// 序列化值
	valueData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化值失败: %v", err)
	}

	// 创建缓存条目
	entry := CacheEntry{
		Key:       key,
		Value:     string(valueData),
		CreatedAt: time.Now(),
		Size:      int64(len(valueData)),
	}

	// 设置过期时间
	if ttl > 0 {
		entry.ExpiresAt = time.Now().Add(ttl)
	}

	// 序列化缓存条目
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化缓存条目失败: %v", err)
	}

	// 写入缓存文件
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("写入缓存文件失败: %v", err)
	}

	c.stats.TotalKeys++
	c.stats.TotalSize += entry.Size
	c.stats.NewestEntry = entry.CreatedAt

	return nil
}

// Delete 删除缓存
func (c *fileCacheImpl) Delete(key string) error {
	cachePath := c.getCachePath(key)

	if err := os.Remove(cachePath); err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，认为删除成功
		}
		return fmt.Errorf("删除缓存文件失败: %v", err)
	}

	c.stats.TotalKeys--
	return nil
}

// Exists 检查缓存是否存在
func (c *fileCacheImpl) Exists(key string) bool {
	cachePath := c.getCachePath(key)
	_, err := os.Stat(cachePath)
	return err == nil
}

// Clear 清空缓存
func (c *fileCacheImpl) Clear() error {
	// 删除缓存目录下的所有文件
	files, err := filepath.Glob(filepath.Join(c.config.CacheDir, "*.json"))
	if err != nil {
		return fmt.Errorf("获取缓存文件列表失败: %v", err)
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return fmt.Errorf("删除缓存文件失败: %v", err)
		}
	}

	// 重置统计
	c.stats = CacheStats{}
	return nil
}

// GetStats 获取缓存统计
func (c *fileCacheImpl) GetStats() CacheStats {
	return c.stats
}

// SetConfig 设置缓存配置
func (c *fileCacheImpl) SetConfig(config CacheConfig) {
	c.config = config
	// 确保新的缓存目录存在
	os.MkdirAll(config.CacheDir, 0755)
}

// GetConfig 获取缓存配置
func (c *fileCacheImpl) GetConfig() CacheConfig {
	return c.config
}

// getCachePath 获取缓存文件路径
func (c *fileCacheImpl) getCachePath(key string) string {
	// 使用key的hash作为文件名，避免特殊字符问题
	filename := fmt.Sprintf("%s.json", key)
	return filepath.Join(c.config.CacheDir, filename)
}

// DefaultCacheConfig 创建默认缓存配置
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		CacheDir: "data/cache",
		MaxAge:   24 * time.Hour,
		MaxSize:  100 * 1024 * 1024, // 100MB
		Compress: false,
		Encrypt:  false,
	}
}
