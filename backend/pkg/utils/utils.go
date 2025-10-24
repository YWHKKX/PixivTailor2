package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ============================================================================
// 文件系统操作
// ============================================================================

// FileExists 检查文件是否存在
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// DirExists 检查目录是否存在
func DirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	return !os.IsNotExist(err) && info.IsDir()
}

// EnsureDir 确保目录存在
func EnsureDir(dirname string) error {
	if !DirExists(dirname) {
		return os.MkdirAll(dirname, 0755)
	}
	return nil
}

// GetFileSize 获取文件大小
func GetFileSize(filename string) (int64, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetFileModTime 获取文件修改时间
func GetFileModTime(filename string) (time.Time, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// GetFileExtension 获取文件扩展名
func GetFileExtension(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

// IsImageFile 检查是否为图像文件
func IsImageFile(filename string) bool {
	ext := GetFileExtension(filename)
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".tiff"}

	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}

// SanitizeFilename 清理文件名
func SanitizeFilename(filename string) string {
	// 移除或替换非法字符
	illegalChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range illegalChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}

	// 限制长度
	if len(filename) > 100 {
		filename = filename[:100]
	}

	return filename
}

// ============================================================================
// 文件操作
// ============================================================================

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}

// MoveFile 移动文件
func MoveFile(src, dst string) error {
	if err := CopyFile(src, dst); err != nil {
		return err
	}
	return os.Remove(src)
}

// ============================================================================
// 文件哈希和校验
// ============================================================================

// GetFileHash 获取文件MD5哈希值
func GetFileHash(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// ============================================================================
// 临时文件操作
// ============================================================================

// GetTempDir 获取临时目录
func GetTempDir() string {
	return os.TempDir()
}

// GetTempFile 创建临时文件
func GetTempFile(prefix, suffix string) (*os.File, error) {
	return os.CreateTemp(GetTempDir(), prefix+"*"+suffix)
}

// CleanTempFiles 清理临时文件
func CleanTempFiles(pattern string) error {
	tempDir := GetTempDir()
	files, err := filepath.Glob(filepath.Join(tempDir, pattern))
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			continue // 忽略删除错误
		}
	}

	return nil
}

// ============================================================================
// 字符串操作
// ============================================================================

// IsEmpty 检查字符串是否为空
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty 检查字符串是否不为空
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// Contains 检查切片是否包含元素
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveDuplicates 移除切片中的重复元素
func RemoveDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// ChunkSlice 将切片分块
func ChunkSlice(slice []string, chunkSize int) [][]string {
	var chunks [][]string

	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

// ============================================================================
// 时间操作
// ============================================================================

// GetCurrentTime 获取当前时间
func GetCurrentTime() time.Time {
	return time.Now()
}

// FormatTime 格式化时间
func FormatTime(t time.Time, layout string) string {
	if layout == "" {
		layout = "2006-01-02 15:04:05"
	}
	return t.Format(layout)
}

// ParseTime 解析时间字符串
func ParseTime(timeStr, layout string) (time.Time, error) {
	if layout == "" {
		layout = "2006-01-02 15:04:05"
	}
	return time.Parse(layout, timeStr)
}

// FormatDuration 格式化持续时间
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// ============================================================================
// 格式化工具
// ============================================================================

// FormatFileSize 格式化文件大小
func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// ============================================================================
// 重试机制
// ============================================================================

// Retry 重试函数
func Retry(maxRetries int, fn func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if err = fn(); err == nil {
			return nil
		}
	}
	return err
}

// RetryWithDelay 带延迟的重试函数
func RetryWithDelay(maxRetries int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if err = fn(); err == nil {
			return nil
		}
		if i < maxRetries-1 {
			time.Sleep(delay)
		}
	}
	return err
}
