package api

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"pixiv-tailor/backend/internal/logger"
	"pixiv-tailor/backend/pkg/paths"
)

// DownloadConfig 下载配置
type DownloadConfig struct {
	SaveDir     string        // 保存目录
	MaxFileSize int64         // 最大文件大小
	RetryCount  int           // 重试次数
	RetryDelay  time.Duration // 重试延迟
}

// DownloadProgress 下载进度回调
type DownloadProgress func(url, filename string, downloaded, total int64, percent float64)

// Downloader 下载器接口（统一版）
type Downloader interface {
	// DownloadFile 下载文件（统一接口）
	// 如果 taskID 为空，则使用 savePath 作为完整路径
	// 如果 taskID 不为空，则自动生成任务目录：[时间]_任务类型_哈希值/{filename}
	DownloadFile(url, savePath string, taskID string, progress DownloadProgress) error

	// SetTaskInfo 设置任务信息（用于生成任务目录名）
	SetTaskInfo(taskType string, createdAt time.Time)

	// SetConfig 设置配置
	SetConfig(config DownloadConfig)

	// GetConfig 获取配置
	GetConfig() DownloadConfig
}

// downloaderImpl 下载器实现
type downloaderImpl struct {
	config     DownloadConfig
	httpClient HTTPClient
	// 任务信息（用于生成任务目录名）
	taskType  string
	createdAt time.Time
}

// NewDownloader 创建新的下载器
func NewDownloader(httpClient HTTPClient, config DownloadConfig) Downloader {
	return &downloaderImpl{
		config:     config,
		httpClient: httpClient,
	}
}

// SetTaskInfo 设置任务信息
func (d *downloaderImpl) SetTaskInfo(taskType string, createdAt time.Time) {
	d.taskType = taskType
	d.createdAt = createdAt
}

// DownloadFile 下载文件（统一接口）
func (d *downloaderImpl) DownloadFile(url, savePath string, taskID string, progress DownloadProgress) error {
	var finalPath string

	if taskID != "" {
		// 使用 PathManager 生成任务目录：{SaveDir}/[时间]_任务类型_哈希值/
		pathManager := paths.GetPathManager()
		var taskDir string

		if d.taskType != "" && !d.createdAt.IsZero() {
			// 使用新格式：[2025-10-28_21:16]_任务类型_哈希值
			taskDir = pathManager.GetTaskImagesDir(taskID, d.taskType, d.createdAt)
		} else {
			// 回退到旧格式：task_{taskID}
			taskDir = filepath.Join(d.config.SaveDir, fmt.Sprintf("task_%s", taskID))
		}

		// 如果 savePath 为空，从 URL 中提取文件名
		if savePath == "" {
			logger.Warnf("DownloadFile: savePath is empty, extracting filename from URL: %s", url)
			// 从 URL 中提取文件名和扩展名
			rawFilename := filepath.Base(url)
			// 如果 URL 中没有文件名或扩展名，使用默认值
			if rawFilename == "" || rawFilename == "/" || rawFilename == "." {
				logger.Warnf("DownloadFile: unable to extract filename from URL, using default: image.jpg")
				rawFilename = "image.jpg"
			} else {
				logger.Infof("DownloadFile: extracted filename from URL: %s", rawFilename)
			}
			savePath = rawFilename
		}

		finalPath = filepath.Join(taskDir, filepath.Base(savePath))
		logger.Infof("DownloadFile: final save path: %s", finalPath)
	} else {
		logger.Error("DownloadFile: taskID is empty")
	}

	var lastErr error

	// 重试逻辑
	for attempt := 0; attempt <= d.config.RetryCount; attempt++ {
		if attempt > 0 {
			// 指数退避
			backoffDelay := time.Duration(1<<uint(attempt-1)) * d.config.RetryDelay
			if backoffDelay > 30*time.Second {
				backoffDelay = 30 * time.Second
			}
			time.Sleep(backoffDelay)
		}

		err := d.downloadFileOnce(url, finalPath, progress)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("下载失败，已重试 %d 次，最后错误: %v", d.config.RetryCount, lastErr)
}

// downloadFileOnce 执行一次下载
func (d *downloaderImpl) downloadFileOnce(url, savePath string, progress DownloadProgress) error {
	// 创建请求
	req, err := d.httpClient.CreateRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// 执行请求
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	// 检查文件大小
	fileSize := resp.ContentLength
	if fileSize > 0 && d.config.MaxFileSize > 0 && fileSize > d.config.MaxFileSize {
		return fmt.Errorf("文件过大: %d > %d", fileSize, d.config.MaxFileSize)
	}

	// 创建目录
	dir := filepath.Dir(savePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 创建文件
	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 下载文件
	var written int64
	if fileSize > 0 && progress != nil {
		// 有Content-Length时使用进度跟踪
		progressWriter := &progressWriter{
			total:    fileSize,
			writer:   file,
			url:      url,
			filename: filepath.Base(savePath),
			callback: progress,
		}
		written, err = io.Copy(progressWriter, resp.Body)
	} else {
		// 没有Content-Length时直接下载
		written, err = io.Copy(file, resp.Body)
	}

	if err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}

	// 验证下载完整性
	if fileSize > 0 && written != fileSize {
		return fmt.Errorf("文件大小不匹配: 期望 %d，实际 %d", fileSize, written)
	}

	return nil
}

// SetConfig 设置配置
func (d *downloaderImpl) SetConfig(config DownloadConfig) {
	d.config = config
}

// GetConfig 获取配置
func (d *downloaderImpl) GetConfig() DownloadConfig {
	return d.config
}

// progressWriter 进度写入器
type progressWriter struct {
	total    int64
	written  int64
	writer   io.Writer
	url      string
	filename string
	callback DownloadProgress
}

// Write 实现io.Writer接口
func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if err != nil {
		return n, err
	}

	pw.written += int64(n)

	if pw.callback != nil {
		percent := float64(pw.written) / float64(pw.total) * 100
		pw.callback(pw.url, pw.filename, pw.written, pw.total, percent)
	}

	return n, nil
}

// DefaultDownloadConfig 创建默认下载配置
func DefaultDownloadConfig() DownloadConfig {
	return DownloadConfig{
		SaveDir:     "data/images",
		MaxFileSize: 100 * 1024 * 1024, // 100MB
		RetryCount:  3,
		RetryDelay:  2 * time.Second,
	}
}
