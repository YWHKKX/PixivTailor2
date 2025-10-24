package api

import (
	"fmt"
	"path/filepath"

	"pixiv-tailor/backend/pkg/models"
	"pixiv-tailor/backend/pkg/paths"
)

// ExampleUsage 展示如何使用新的API
func ExampleUsage() {
	// 1. 创建HTTP客户端
	httpConfig := DefaultHTTPClientConfig()
	httpConfig.ProxyEnabled = true
	httpConfig.ProxyURL = "http://127.0.0.1:7890"

	httpClient := NewHTTPClient(httpConfig)

	// 2. 创建缓存
	cacheConfig := DefaultCacheConfig()
	pathManager := paths.GetPathManager()
	if pathManager != nil {
		cacheConfig.CacheDir = pathManager.GetDataPath("cache")
	}

	cache := NewFileCache(cacheConfig)

	// 3. 创建下载器
	downloadConfig := DefaultDownloadConfig()
	if pathManager != nil {
		downloadConfig.SaveDir = pathManager.GetImagesDir()
	}

	downloader := NewDownloader(httpClient, downloadConfig)

	// 4. 创建Pixiv API
	pixivConfig := DefaultPixivConfig(httpClient, cache, downloader)
	pixivAPI := NewPixivAPI(pixivConfig)

	// 5. 使用Pixiv API进行搜索
	images, err := pixivAPI.SearchByTag("1girl", "popular", "all", 10)
	if err != nil {
		fmt.Printf("搜索失败: %v\n", err)
		return
	}

	fmt.Printf("找到 %d 张图片\n", len(images))

	// 6. 下载图片
	for i, image := range images {
		if i >= 3 { // 只下载前3张作为示例
			break
		}

		filename := fmt.Sprintf("image_%d.jpg", i+1)
		savePath := filepath.Join(downloadConfig.SaveDir, filename)

		err := pixivAPI.DownloadImage(image.URL, savePath)
		if err != nil {
			fmt.Printf("下载失败 %s: %v\n", image.URL, err)
		} else {
			fmt.Printf("下载成功: %s\n", savePath)
		}
	}
}

// ExampleCrawlerRefactor 展示如何重构crawler.go
func ExampleCrawlerRefactor() {
	// 原来的crawler.go中的代码可以简化为：

	// 1. 初始化API
	httpClient := NewHTTPClient(DefaultHTTPClientConfig())
	cache := NewFileCache(DefaultCacheConfig())
	downloader := NewDownloader(httpClient, DefaultDownloadConfig())
	pixivAPI := NewPixivAPI(DefaultPixivConfig(httpClient, cache, downloader))

	// 2. 按标签爬取 - 原来需要100+行代码，现在只需要几行
	images, err := pixivAPI.SearchByTag("anime", "popular", "all", 50)
	if err != nil {
		fmt.Printf("爬取失败: %v\n", err)
		return
	}

	// 3. 下载图片 - 原来需要复杂的重试和进度跟踪，现在自动处理
	for _, image := range images {
		filename := fmt.Sprintf("artworks_%d.jpg", image.ID)
		savePath := filepath.Join("data/images", filename)

		if err := pixivAPI.DownloadImage(image.URL, savePath); err != nil {
			fmt.Printf("下载失败: %v\n", err)
		}
	}
}

// ExampleTaskServiceRefactor 展示如何重构task_service.go
func ExampleTaskServiceRefactor() {
	// 原来的task_service.go中的executeCrawlTask可以简化为：

	// 1. 初始化API
	httpClient := NewHTTPClient(DefaultHTTPClientConfig())
	cache := NewFileCache(DefaultCacheConfig())
	downloader := NewDownloader(httpClient, DefaultDownloadConfig())
	pixivAPI := NewPixivAPI(DefaultPixivConfig(httpClient, cache, downloader))

	// 2. 根据任务类型执行爬取
	var images []*models.PixivImage
	var err error

	// 这些复杂的switch case可以简化为：
	query := "anime"
	order := "popular"
	mode := "all"
	limit := 50
	userID := 107022296
	illustID := 12345678

	switch taskType := "tag"; taskType {
	case "tag":
		images, err = pixivAPI.SearchByTag(query, order, mode, limit)
	case "user":
		images, err = pixivAPI.GetUserWorks(userID, limit)
	case "illust":
		images, err = pixivAPI.GetIllustPages(illustID)
	}

	if err != nil {
		fmt.Printf("爬取失败: %v\n", err)
		return
	}

	// 3. 下载图片 - 原来需要复杂的循环和进度更新，现在自动处理
	for _, image := range images {
		filename := fmt.Sprintf("artworks_%d.jpg", image.ID)
		savePath := filepath.Join("data/images", filename)

		if err := pixivAPI.DownloadImage(image.URL, savePath); err != nil {
			fmt.Printf("下载失败: %v\n", err)
		}
	}
}
