package crawler

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"pixiv-tailor/backend/internal/api"
	"pixiv-tailor/backend/internal/logger"
	"pixiv-tailor/backend/pkg/models"
	"pixiv-tailor/backend/pkg/paths"
)

// CrawlCache 爬取缓存结构
type CrawlCache struct {
	Date      string               `json:"date"`       // 缓存日期 YYYY-MM-DD
	Type      string               `json:"type"`       // 爬取类型 (tag/user/illust)
	Query     string               `json:"query"`      // 查询参数
	UserID    int                  `json:"user_id"`    // 用户ID (仅用户爬取时使用)
	Limit     int                  `json:"limit"`      // 限制数量
	Images    []*models.PixivImage `json:"images"`     // 图片数据
	CreatedAt time.Time            `json:"created_at"` // 创建时间
	UpdatedAt time.Time            `json:"updated_at"` // 更新时间
}

// Crawler 爬虫接口
type Crawler interface {
	// CrawlByTag 按标签爬取
	CrawlByTag(query, order, mode string, limit int) ([]*models.PixivImage, error)

	// CrawlByUser 按用户爬取
	CrawlByUser(userID, limit int) ([]*models.PixivImage, error)

	// CrawlByIllust 按插画ID爬取（返回所有页面）
	CrawlByIllust(illustID int) ([]*models.PixivImage, error)

	// DownloadImage 下载图像（统一接口）
	DownloadImage(url, filepath string, taskID string, progress func(url, filename string, downloaded, total int64, percent float64)) error

	// SetLogCallback 设置日志回调函数
	SetLogCallback(callback func(level, message string))

	// SetProxy 设置代理
	SetProxy(enabled bool, proxyURL string)

	// SetCookie 设置Cookie
	SetCookie(cookie string)

	// SetTaskID 设置任务ID（用于缓存键生成）
	SetTaskID(taskID string)

	// SetTaskInfo 设置任务信息（用于生成任务文件夹名）
	SetTaskInfo(taskType string, createdAt time.Time)
}

// PixivCrawler Pixiv爬虫实现（重构版）
type PixivCrawler struct {
	httpClient  api.HTTPClient
	downloader  api.Downloader
	pixivAPI    api.PixivAPI
	logCallback func(level, message string)
	taskID      string
}

// NewCrawler 创建新的爬虫实例（使用默认配置）
func NewCrawler() (Crawler, error) {
	// 使用路径管理器获取配置
	pathManager := paths.GetPathManager()
	if pathManager == nil {
		return nil, fmt.Errorf("路径管理器未初始化")
	}

	// 创建HTTP客户端
	httpConfig := api.DefaultHTTPClientConfig()
	httpClient := api.NewHTTPClient(httpConfig)

	// 创建下载器
	downloadConfig := api.DefaultDownloadConfig()
	downloadConfig.SaveDir = pathManager.GetImagesDir()
	downloader := api.NewDownloader(httpClient, downloadConfig)

	// 创建Pixiv API（不使用缓存）
	pixivConfig := api.DefaultPixivConfig(httpClient, nil, downloader)
	pixivAPI := api.NewPixivAPI(pixivConfig)

	return &PixivCrawler{
		httpClient: httpClient,
		downloader: downloader,
		pixivAPI:   pixivAPI,
	}, nil
}

// NewCrawlerWithConfig 使用配置创建新的爬虫实例
func NewCrawlerWithConfig(config map[string]interface{}) (Crawler, error) {
	// 创建HTTP客户端
	httpConfig := api.DefaultHTTPClientConfig()

	// 从配置中加载设置
	if pixivConfig, ok := config["pixiv"].(map[string]interface{}); ok {
		if userAgent, ok := pixivConfig["user_agent"].(string); ok {
			httpConfig.UserAgent = userAgent
		}
		if accept, ok := pixivConfig["accept"].(string); ok {
			httpConfig.Accept = accept
		}
		if acceptLang, ok := pixivConfig["accept_language"].(string); ok {
			httpConfig.AcceptLanguage = acceptLang
		}
		if referer, ok := pixivConfig["referer"].(string); ok {
			httpConfig.Referer = referer
		}
		if timeout, ok := pixivConfig["timeout"].(float64); ok {
			httpConfig.Timeout = int(timeout)
		}
		if retryCount, ok := pixivConfig["retry_count"].(float64); ok {
			httpConfig.RetryCount = int(retryCount)
		}
		if defaultDelay, ok := pixivConfig["default_delay"].(float64); ok {
			httpConfig.Delay = time.Duration(defaultDelay) * time.Second
		}

		// 代理设置
		if proxyConfig, ok := pixivConfig["proxy"].(map[string]interface{}); ok {
			if enabled, ok := proxyConfig["enabled"].(bool); ok {
				httpConfig.ProxyEnabled = enabled
			}
			if proxyURL, ok := proxyConfig["url"].(string); ok {
				httpConfig.ProxyURL = proxyURL
			}
		}
	}

	httpClient := api.NewHTTPClient(httpConfig)

	// 创建下载器
	downloadConfig := api.DefaultDownloadConfig()
	pathManager := paths.GetPathManager()
	if pathManager != nil {
		downloadConfig.SaveDir = pathManager.GetImagesDir()
	}
	downloader := api.NewDownloader(httpClient, downloadConfig)

	// 创建Pixiv API（不使用缓存）
	pixivConfig := api.DefaultPixivConfig(httpClient, nil, downloader)
	pixivAPI := api.NewPixivAPI(pixivConfig)

	return &PixivCrawler{
		httpClient: httpClient,
		downloader: downloader,
		pixivAPI:   pixivAPI,
	}, nil
}

// CrawlerConfig 爬虫配置文件结构
type CrawlerConfig struct {
	Pixiv struct {
		Cookie         string  `json:"cookie"`
		UserAgent      string  `json:"user_agent"`
		Accept         string  `json:"accept"`
		AcceptLanguage string  `json:"accept_language"`
		Referer        string  `json:"referer"`
		Timeout        int     `json:"timeout"`
		RetryCount     int     `json:"retry_count"`
		DefaultDelay   float64 `json:"default_delay"`
		Proxy          struct {
			Enabled bool   `json:"enabled"`
			URL     string `json:"url"`
		} `json:"proxy"`
	} `json:"pixiv"`
}

// loadCrawlerConfig 加载爬虫配置文件
func loadCrawlerConfig() (*CrawlerConfig, error) {
	pathManager := paths.GetPathManager()
	if pathManager == nil {
		return nil, fmt.Errorf("路径管理器未初始化")
	}

	configPath := pathManager.GetCrawlerConfigPath()

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Warnf("爬虫配置文件不存在: %s", configPath)
		return &CrawlerConfig{}, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取爬虫配置文件失败: %v", err)
	}

	var config CrawlerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析爬虫配置文件失败: %v", err)
	}

	logger.Infof("成功加载爬虫配置文件: %s", configPath)
	return &config, nil
}

// NewCrawlerForTask 为特定任务创建爬虫实例（使用任务特定的缓存和下载目录）
func NewCrawlerForTask(taskID string, config map[string]interface{}) (Crawler, error) {
	// 获取路径管理器
	pathManager := paths.GetPathManager()
	if pathManager == nil {
		return nil, fmt.Errorf("路径管理器未初始化")
	}

	// 加载爬虫配置文件
	crawlerConfig, err := loadCrawlerConfig()
	if err != nil {
		logger.Warnf("加载爬虫配置文件失败，使用默认配置: %v", err)
		crawlerConfig = &CrawlerConfig{}
	}

	// 创建HTTP客户端配置
	httpConfig := api.DefaultHTTPClientConfig()

	// 使用配置文件中的默认值
	if crawlerConfig.Pixiv.Cookie != "" {
		httpConfig.Cookie = crawlerConfig.Pixiv.Cookie
		if len(crawlerConfig.Pixiv.Cookie) > 50 {
			logger.Infof("使用配置文件中的Cookie: %s...", crawlerConfig.Pixiv.Cookie[:50])
		} else {
			logger.Infof("使用配置文件中的Cookie: %s", crawlerConfig.Pixiv.Cookie)
		}
	}
	if crawlerConfig.Pixiv.UserAgent != "" {
		httpConfig.UserAgent = crawlerConfig.Pixiv.UserAgent
	}
	if crawlerConfig.Pixiv.Accept != "" {
		httpConfig.Accept = crawlerConfig.Pixiv.Accept
	}
	if crawlerConfig.Pixiv.AcceptLanguage != "" {
		httpConfig.AcceptLanguage = crawlerConfig.Pixiv.AcceptLanguage
	}
	if crawlerConfig.Pixiv.Referer != "" {
		httpConfig.Referer = crawlerConfig.Pixiv.Referer
	}
	if crawlerConfig.Pixiv.Timeout > 0 {
		httpConfig.Timeout = crawlerConfig.Pixiv.Timeout
	}
	if crawlerConfig.Pixiv.RetryCount > 0 {
		httpConfig.RetryCount = crawlerConfig.Pixiv.RetryCount
	}
	if crawlerConfig.Pixiv.Proxy.Enabled {
		httpConfig.ProxyEnabled = crawlerConfig.Pixiv.Proxy.Enabled
		httpConfig.ProxyURL = crawlerConfig.Pixiv.Proxy.URL
	}

	// 从任务配置中覆盖设置（任务特定的配置优先）
	if pixivConfig, ok := config["pixiv"].(map[string]interface{}); ok {
		if userAgent, ok := pixivConfig["user_agent"].(string); ok {
			httpConfig.UserAgent = userAgent
		}
		if accept, ok := pixivConfig["accept"].(string); ok {
			httpConfig.Accept = accept
		}
		if cookie, ok := pixivConfig["cookie"].(string); ok {
			httpConfig.Cookie = cookie
		}
		if referer, ok := pixivConfig["referer"].(string); ok {
			httpConfig.Referer = referer
		}
		if acceptLanguage, ok := pixivConfig["accept_language"].(string); ok {
			httpConfig.AcceptLanguage = acceptLanguage
		}
		if proxyEnabled, ok := pixivConfig["proxy_enabled"].(bool); ok {
			httpConfig.ProxyEnabled = proxyEnabled
		}
		if proxyURL, ok := pixivConfig["proxy_url"].(string); ok {
			httpConfig.ProxyURL = proxyURL
		}
	}

	// 从任务配置中读取 delay 参数
	if delay, exists := config["delay"].(float64); exists && delay > 0 {
		httpConfig.Delay = time.Duration(delay) * time.Second
	}

	httpClient := api.NewHTTPClient(httpConfig)

	// 创建任务特定的下载器
	downloadConfig := api.DefaultDownloadConfig()
	downloadConfig.SaveDir = pathManager.GetImagesDir() // 直接使用images目录，让下载器自动添加task_前缀
	downloader := api.NewDownloader(httpClient, downloadConfig)

	// 创建Pixiv API（不使用缓存）
	pixivConfig := api.DefaultPixivConfig(httpClient, nil, downloader)
	pixivAPI := api.NewPixivAPI(pixivConfig)

	return &PixivCrawler{
		httpClient: httpClient,
		downloader: downloader,
		pixivAPI:   pixivAPI,
	}, nil
}

// SetLogCallback 设置日志回调函数
func (c *PixivCrawler) SetLogCallback(callback func(level, message string)) {
	c.logCallback = callback
}

// SetTaskID 设置任务ID
func (c *PixivCrawler) SetTaskID(taskID string) {
	c.taskID = taskID
}

// SetTaskInfo 设置任务信息（用于生成任务文件夹名）
func (c *PixivCrawler) SetTaskInfo(taskType string, createdAt time.Time) {
	if c.downloader != nil {
		c.downloader.SetTaskInfo(taskType, createdAt)
	}
}

// sendLog 发送日志消息
func (c *PixivCrawler) sendLog(level, message string) {
	if c.logCallback != nil {
		c.logCallback(level, message)
	}
}

// SetProxy 设置代理
func (c *PixivCrawler) SetProxy(enabled bool, proxyURL string) {
	// 更新HTTP客户端配置
	config := c.httpClient.GetConfig()
	config.ProxyEnabled = enabled
	if enabled && proxyURL != "" {
		config.ProxyURL = proxyURL
	}
	c.httpClient.SetConfig(config)

	// 测试代理连接
	if enabled {
		if err := c.httpClient.TestProxyConnection(); err != nil {
			c.sendLog("error", fmt.Sprintf("代理连接测试失败: %v", err))
		} else {
			c.sendLog("info", "代理连接测试成功")
		}
	}
}

// SetCookie 设置Cookie
func (c *PixivCrawler) SetCookie(cookie string) {
	// 更新HTTP客户端配置
	config := c.httpClient.GetConfig()
	config.Cookie = cookie
	c.httpClient.SetConfig(config)

	if len(cookie) > 50 {
		c.sendLog("info", fmt.Sprintf("Cookie已设置: %s...", cookie[:50]))
	} else {
		c.sendLog("info", fmt.Sprintf("Cookie已设置: %s", cookie))
	}
}

// CrawlByTag 按标签爬取
func (c *PixivCrawler) CrawlByTag(query, order, mode string, limit int) ([]*models.PixivImage, error) {
	logger.Infof("开始按标签爬取: %s, 排序: %s, 模式: %s, 限制: %d", query, order, mode, limit)

	// 使用Pixiv API进行搜索
	images, err := c.pixivAPI.SearchByTag(query, order, mode, limit)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %v", err)
	}

	logger.Infof("成功爬取 %d 张图片", len(images))

	return images, nil
}

// CrawlByUser 按用户爬取
func (c *PixivCrawler) CrawlByUser(userID, limit int) ([]*models.PixivImage, error) {
	logger.Infof("开始按用户爬取: %d, 限制: %d", userID, limit)
	c.sendLog("info", fmt.Sprintf("开始按用户爬取: %d, 限制: %d", userID, limit))

	// 使用Pixiv API获取用户作品
	images, err := c.pixivAPI.GetUserWorks(userID, limit)
	if err != nil {
		c.sendLog("error", fmt.Sprintf("获取用户作品失败: %v", err))
		return nil, fmt.Errorf("获取用户作品失败: %v", err)
	}

	logger.Infof("成功爬取 %d 张图片", len(images))
	c.sendLog("info", fmt.Sprintf("成功爬取 %d 张图片", len(images)))

	return images, nil
}

// CrawlByIllust 按插画ID爬取（返回所有页面）
func (c *PixivCrawler) CrawlByIllust(illustID int) ([]*models.PixivImage, error) {
	logger.Infof("开始按插画ID爬取: %d", illustID)

	// 使用Pixiv API获取插画页面
	images, err := c.pixivAPI.GetIllustPages(illustID)
	if err != nil {
		return nil, fmt.Errorf("获取插画页面失败: %v", err)
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("未找到图片")
	}

	// 返回所有图片
	logger.Infof("成功获取插画信息，共 %d 张图片", len(images))
	return images, nil
}

// DownloadImage 下载图像（统一接口）
func (c *PixivCrawler) DownloadImage(imageURL, savePath string, taskID string, progress func(url, filename string, downloaded, total int64, percent float64)) error {
	// 使用下载器的统一DownloadFile方法
	return c.downloader.DownloadFile(imageURL, savePath, taskID, progress)
}

// GetConfig 获取当前配置
func (c *PixivCrawler) GetConfig() map[string]interface{} {
	httpConfig := c.httpClient.GetConfig()
	downloadConfig := c.downloader.GetConfig()

	return map[string]interface{}{
		"pixiv": map[string]interface{}{
			"user_agent":      httpConfig.UserAgent,
			"accept":          httpConfig.Accept,
			"accept_language": httpConfig.AcceptLanguage,
			"referer":         httpConfig.Referer,
			"timeout":         httpConfig.Timeout,
			"retry_count":     httpConfig.RetryCount,
			"default_delay":   httpConfig.Delay.Seconds(),
			"proxy": map[string]interface{}{
				"enabled": httpConfig.ProxyEnabled,
				"url":     httpConfig.ProxyURL,
			},
		},
		"download": map[string]interface{}{
			"save_dir": downloadConfig.SaveDir,
		},
	}
}

// UpdateConfig 更新配置
func (c *PixivCrawler) UpdateConfig(config map[string]interface{}) error {
	// 更新HTTP客户端配置
	if pixivConfig, ok := config["pixiv"].(map[string]interface{}); ok {
		httpConfig := c.httpClient.GetConfig()

		if userAgent, ok := pixivConfig["user_agent"].(string); ok {
			httpConfig.UserAgent = userAgent
		}
		if accept, ok := pixivConfig["accept"].(string); ok {
			httpConfig.Accept = accept
		}
		if acceptLang, ok := pixivConfig["accept_language"].(string); ok {
			httpConfig.AcceptLanguage = acceptLang
		}
		if referer, ok := pixivConfig["referer"].(string); ok {
			httpConfig.Referer = referer
		}
		if timeout, ok := pixivConfig["timeout"].(float64); ok {
			httpConfig.Timeout = int(timeout)
		}
		if retryCount, ok := pixivConfig["retry_count"].(float64); ok {
			httpConfig.RetryCount = int(retryCount)
		}
		if defaultDelay, ok := pixivConfig["default_delay"].(float64); ok {
			httpConfig.Delay = time.Duration(defaultDelay) * time.Second
		}

		// 代理设置
		if proxyConfig, ok := pixivConfig["proxy"].(map[string]interface{}); ok {
			if enabled, ok := proxyConfig["enabled"].(bool); ok {
				httpConfig.ProxyEnabled = enabled
			}
			if proxyURL, ok := proxyConfig["url"].(string); ok {
				httpConfig.ProxyURL = proxyURL
			}
		}

		c.httpClient.SetConfig(httpConfig)
	}

	// 更新下载配置
	if downloadConfig, ok := config["download"].(map[string]interface{}); ok {
		config := c.downloader.GetConfig()

		if saveDir, ok := downloadConfig["save_dir"].(string); ok {
			config.SaveDir = saveDir
		}

		c.downloader.SetConfig(config)
	}

	return nil
}
