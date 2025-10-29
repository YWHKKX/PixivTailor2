package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"pixiv-tailor/backend/internal/logger"
	"pixiv-tailor/backend/pkg/models"
)

// PixivConfig Pixiv API配置
type PixivConfig struct {
	BaseURL        string
	SearchTagURL   string
	UserWorksURL   string
	IllustPagesURL string
	ArtworksURL    string
	HTTPClient     HTTPClient
	Cache          Cache
	Downloader     Downloader
}

// PixivAPI Pixiv API接口
type PixivAPI interface {
	// SearchByTag 按标签搜索
	SearchByTag(query, order, mode string, limit int) ([]*models.PixivImage, error)

	// GetUserWorks 获取用户作品
	GetUserWorks(userID, limit int) ([]*models.PixivImage, error)

	// GetIllustPages 获取插画页面
	GetIllustPages(illustID int) ([]*models.PixivImage, error)

	// DownloadImage 下载图片
	DownloadImage(imageURL, savePath string) error

	// SetConfig 设置配置
	SetConfig(config PixivConfig)

	// GetConfig 获取配置
	GetConfig() PixivConfig
}

// pixivAPIImpl Pixiv API实现
type pixivAPIImpl struct {
	config PixivConfig
}

// NewPixivAPI 创建新的Pixiv API
func NewPixivAPI(config PixivConfig) PixivAPI {
	return &pixivAPIImpl{
		config: config,
	}
}

// SearchByTag 按标签搜索
func (p *pixivAPIImpl) SearchByTag(query, order, mode string, limit int) ([]*models.PixivImage, error) {
	// 处理多个标签：Pixiv 支持空格分隔的多个标签进行 AND 搜索
	// 例如: "girl,elsa" -> "girl elsa", "エルザ,Re:ゼロ" -> "エルザ Re:ゼロ"
	normalizedQuery := query
	if strings.Contains(query, ",") {
		// 将逗号替换为空格（Pixiv 使用空格进行 AND 搜索）
		normalizedQuery = strings.ReplaceAll(query, ",", " ")
		// 移除多余空格
		for strings.Contains(normalizedQuery, "  ") {
			normalizedQuery = strings.ReplaceAll(normalizedQuery, "  ", " ")
		}
		normalizedQuery = strings.TrimSpace(normalizedQuery)
		logger.Debugf("原始查询: %s -> 标准化: %s", query, normalizedQuery)
	}

	// URL编码标签
	// url.QueryEscape 会将空格编码为 '+'，但 Pixiv API 需要 '%20'
	encodedQuery := url.QueryEscape(normalizedQuery)
	encodedQuery = strings.ReplaceAll(encodedQuery, "+", "%20")

	// 构建搜索URL
	// 格式: https://www.pixiv.net/ajax/search/artworks/TAG?word=TAG&order=DATE&mode=all&p=1&s_mode=s_tag_full&type=all&lang=zh
	targetURL := fmt.Sprintf("%s/%s?word=%s&order=%s&mode=%s&p=1&s_mode=s_tag_full&type=%s&lang=zh",
		p.config.SearchTagURL, encodedQuery, encodedQuery, order, mode, mode)

	// 输出生成的 URL 用于调试
	logger.Debugf("生成的搜索 URL: %s", targetURL)

	// 创建请求
	req, err := p.config.HTTPClient.CreateRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	// 执行请求
	resp, err := p.config.HTTPClient.DoWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	// 读取响应
	body, err := p.config.HTTPClient.ReadResponseBody(resp)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var tagResp PixivTagResponse
	if err := json.Unmarshal(body, &tagResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	if tagResp.Error {
		return nil, fmt.Errorf("API返回错误")
	}

	// 处理图片数据
	var images []*models.PixivImage
	for i, data := range tagResp.Body.IllustManga.Data {
		if i >= limit {
			break
		}

		// 获取图片URL列表
		imageURLs, err := p.getImageURLs(data.ID)
		if err != nil {
			continue
		}

		if len(imageURLs) == 0 {
			continue
		}

		// 将插画ID转换为int64
		illustIDInt, err := strconv.ParseInt(data.ID, 10, 64)
		if err != nil {
			continue
		}

		// 为每个图片URL创建单独的PixivImage对象
		for _, imageURL := range imageURLs {
			image := &models.PixivImage{
				BaseModel: models.BaseModel{
					ID:        illustIDInt,
					CreatedAt: time.Now(),
				},
				Title:        data.Title,
				Author:       "",
				AuthorID:     0,
				Tags:         []string{query},
				URL:          imageURL,
				ThumbnailURL: "",
				Width:        0,
				Height:       0,
				Bookmarks:    0,
				Views:        0,
			}

			images = append(images, image)
		}
	}

	return images, nil
}

// GetUserWorks 获取用户作品
func (p *pixivAPIImpl) GetUserWorks(userID, limit int) ([]*models.PixivImage, error) {
	// 构建用户作品URL
	targetURL := fmt.Sprintf("%s/%d/profile/all", p.config.UserWorksURL, userID)

	// 创建请求
	req, err := p.config.HTTPClient.CreateRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	// 执行请求
	resp, err := p.config.HTTPClient.DoWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	// 读取响应
	body, err := p.config.HTTPClient.ReadResponseBody(resp)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var userResp PixivUserResponse
	if err := json.Unmarshal(body, &userResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	if userResp.Error {
		return nil, fmt.Errorf("API返回错误: %s", userResp.Message)
	}

	// 处理插画数据
	var images []*models.PixivImage
	count := 0

	// 检查illusts字段的类型并处理
	switch illusts := userResp.Body.Illusts.(type) {
	case map[string]interface{}:
		for illustID, illustData := range illusts {
			if count >= limit {
				break
			}

			// 获取标题和标签
			var title string
			var tags []string

			if illustData == nil {
				title = fmt.Sprintf("插画 %s", illustID)
				tags = []string{}
			} else {
				if illustMap, ok := illustData.(map[string]interface{}); ok {
					if titleVal, exists := illustMap["title"]; exists {
						title = fmt.Sprintf("%v", titleVal)
					} else {
						title = fmt.Sprintf("插画 %s", illustID)
					}

					if tagsVal, exists := illustMap["tags"]; exists {
						if tagsSlice, ok := tagsVal.([]interface{}); ok {
							for _, tag := range tagsSlice {
								tags = append(tags, fmt.Sprintf("%v", tag))
							}
						}
					}
				} else {
					title = fmt.Sprintf("插画 %s", illustID)
					tags = []string{}
				}
			}

			// 获取图片URL列表
			imageURLs, err := p.getImageURLs(illustID)
			if err != nil {
				continue
			}

			if len(imageURLs) == 0 {
				continue
			}

			// 将插画ID转换为int64
			illustIDInt, err := strconv.ParseInt(illustID, 10, 64)
			if err != nil {
				continue
			}

			// 为每个图片URL创建单独的PixivImage对象
			for _, imageURL := range imageURLs {
				image := &models.PixivImage{
					BaseModel: models.BaseModel{
						ID:        illustIDInt,
						CreatedAt: time.Now(),
					},
					Title:        title,
					Author:       "",
					AuthorID:     userID,
					Tags:         tags,
					URL:          imageURL,
					ThumbnailURL: "",
					Width:        0,
					Height:       0,
					Bookmarks:    0,
					Views:        0,
				}

				images = append(images, image)
				count++
			}
		}
	}

	return images, nil
}

// GetIllustPages 获取插画页面
func (p *pixivAPIImpl) GetIllustPages(illustID int) ([]*models.PixivImage, error) {
	illustIDStr := fmt.Sprintf("%d", illustID)

	// 获取图片URL列表
	imageURLs, err := p.getImageURLs(illustIDStr)
	if err != nil {
		return nil, fmt.Errorf("获取图片URL失败: %v", err)
	}

	if len(imageURLs) == 0 {
		return nil, fmt.Errorf("未找到图片")
	}

	// 创建图片对象
	var images []*models.PixivImage
	for _, imageURL := range imageURLs {
		image := &models.PixivImage{
			BaseModel: models.BaseModel{
				ID:        int64(illustID),
				CreatedAt: time.Now(),
			},
			Title:        "",
			Author:       "",
			AuthorID:     0,
			Tags:         []string{},
			URL:          imageURL,
			ThumbnailURL: "",
			Width:        0,
			Height:       0,
			Bookmarks:    0,
			Views:        0,
		}
		images = append(images, image)
	}

	return images, nil
}

// DownloadImage 下载图片
func (p *pixivAPIImpl) DownloadImage(imageURL, savePath string) error {
	return p.config.Downloader.DownloadFile(imageURL, savePath, "", nil)
}

// getImageURLs 获取插画的图片URL列表
func (p *pixivAPIImpl) getImageURLs(illustID string) ([]string, error) {
	targetURL := fmt.Sprintf("%s/%s/pages", p.config.IllustPagesURL, illustID)

	// 创建请求
	req, err := p.config.HTTPClient.CreateRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	// 执行请求
	resp, err := p.config.HTTPClient.DoWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	// 读取响应
	body, err := p.config.HTTPClient.ReadResponseBody(resp)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var imageResp PixivImageResponse
	if err := json.Unmarshal(body, &imageResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	if imageResp.Error {
		return nil, fmt.Errorf("API返回错误: %s", imageResp.Message)
	}

	// 提取URL列表
	var urls []string
	for _, img := range imageResp.Body {
		urls = append(urls, img.URLs.Original)
	}

	return urls, nil
}

// SetConfig 设置配置
func (p *pixivAPIImpl) SetConfig(config PixivConfig) {
	p.config = config
}

// GetConfig 获取配置
func (p *pixivAPIImpl) GetConfig() PixivConfig {
	return p.config
}

// Pixiv API响应结构
type PixivTagResponse struct {
	Error bool `json:"error"`
	Body  struct {
		IllustManga struct {
			Data []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
				URL   string `json:"url"`
			} `json:"data"`
		} `json:"illustManga"`
	} `json:"body"`
}

type PixivUserResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Body    struct {
		Illusts interface{} `json:"illusts"`
		Manga   interface{} `json:"manga"`
	} `json:"body"`
}

type PixivImageResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Body    []struct {
		URLs struct {
			ThumbMini string `json:"thumb_mini"`
			Small     string `json:"small"`
			Regular   string `json:"regular"`
			Original  string `json:"original"`
		} `json:"urls"`
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"body"`
}

// DefaultPixivConfig 创建默认Pixiv配置
func DefaultPixivConfig(httpClient HTTPClient, cache Cache, downloader Downloader) PixivConfig {
	// 创建一个空的缓存实现（不需要缓存功能）
	var noCache Cache
	if cache != nil {
		noCache = cache
	} else {
		// 使用nil缓存，表示不使用缓存
		noCache = nil
	}

	return PixivConfig{
		BaseURL:        "https://www.pixiv.net",
		SearchTagURL:   "https://www.pixiv.net/ajax/search/artworks",
		UserWorksURL:   "https://www.pixiv.net/ajax/user",
		IllustPagesURL: "https://www.pixiv.net/ajax/illust",
		ArtworksURL:    "https://www.pixiv.net/artworks",
		HTTPClient:     httpClient,
		Cache:          noCache,
		Downloader:     downloader,
	}
}
