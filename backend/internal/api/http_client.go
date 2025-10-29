package api

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// HTTPClientConfig HTTP客户端配置
type HTTPClientConfig struct {
	UserAgent      string
	Accept         string
	AcceptLanguage string
	Referer        string
	Timeout        int
	RetryCount     int
	Delay          time.Duration
	ProxyEnabled   bool
	ProxyURL       string
	Cookie         string
}

// HTTPClient HTTP客户端接口
type HTTPClient interface {
	// CreateRequest 创建HTTP请求
	CreateRequest(method, targetURL string, body io.Reader) (*http.Request, error)

	// Do 执行HTTP请求
	Do(req *http.Request) (*http.Response, error)

	// DoWithRetry 执行HTTP请求（带重试）
	DoWithRetry(req *http.Request) (*http.Response, error)

	// ReadResponseBody 读取响应体
	ReadResponseBody(resp *http.Response) ([]byte, error)

	// TestProxyConnection 测试代理连接
	TestProxyConnection() error

	// SetConfig 更新配置
	SetConfig(config HTTPClientConfig)

	// GetConfig 获取当前配置
	GetConfig() HTTPClientConfig
}

// httpClientImpl HTTP客户端实现
type httpClientImpl struct {
	config         HTTPClientConfig
	client         *http.Client
	requestLimiter chan struct{}
}

// NewHTTPClient 创建新的HTTP客户端
func NewHTTPClient(config HTTPClientConfig) HTTPClient {
	impl := &httpClientImpl{
		config:         config,
		requestLimiter: make(chan struct{}, 1),
	}
	impl.createClient()
	return impl
}

// createClient 创建HTTP客户端
func (h *httpClientImpl) createClient() {
	// 创建传输配置
	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 1,
		IdleConnTimeout:     10 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
		DisableKeepAlives:   false,
		DisableCompression:  false,
		DialContext: (&net.Dialer{
			Timeout:   8 * time.Second,
			KeepAlive: 8 * time.Second,
		}).DialContext,
		MaxConnsPerHost: 1,
	}

	// 配置代理
	if h.config.ProxyEnabled && h.config.ProxyURL != "" {
		if proxyURL, err := url.Parse(h.config.ProxyURL); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
			transport.DialContext = (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 10 * time.Second,
			}).DialContext
		}
	}

	h.client = &http.Client{
		Transport: transport,
		Timeout:   time.Duration(h.config.Timeout) * time.Second,
	}
}

// CreateRequest 创建HTTP请求
func (h *httpClientImpl) CreateRequest(method, targetURL string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, targetURL, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header = http.Header{
		"Accept":                    []string{h.config.Accept},
		"Accept-Encoding":           []string{"gzip, deflate, br"},
		"Accept-Language":           []string{h.config.AcceptLanguage},
		"Cache-Control":             []string{"no-cache"},
		"Connection":                []string{"keep-alive"},
		"Cookie":                    []string{h.config.Cookie},
		"DNT":                       []string{"1"},
		"Pragma":                    []string{"no-cache"},
		"Referer":                   []string{h.config.Referer},
		"Sec-Fetch-Dest":            []string{"image"},
		"Sec-Fetch-Mode":            []string{"no-cors"},
		"Sec-Fetch-Site":            []string{"cross-site"},
		"Sec-Ch-Ua":                 []string{`"Microsoft Edge";v="135", "Chromium";v="135", "Not_A Brand";v="99"`},
		"Sec-Ch-Ua-Mobile":          []string{"?0"},
		"Sec-Ch-Ua-Platform":        []string{`"Windows"`},
		"Upgrade-Insecure-Requests": []string{"1"},
		"User-Agent":                []string{h.config.UserAgent},
	}

	return req, nil
}

// Do 执行HTTP请求
func (h *httpClientImpl) Do(req *http.Request) (*http.Response, error) {
	// 使用请求限制器
	h.requestLimiter <- struct{}{}
	defer func() { <-h.requestLimiter }()

	// 应用配置的延迟
	if h.config.Delay > 0 {
		time.Sleep(h.config.Delay)
	}

	return h.client.Do(req)
}

// DoWithRetry 执行HTTP请求（带重试）
func (h *httpClientImpl) DoWithRetry(req *http.Request) (*http.Response, error) {
	var lastErr error
	maxRetries := h.config.RetryCount

	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			// 指数退避
			backoffDelay := time.Duration(1<<uint(i-1)) * time.Second
			if backoffDelay > 30*time.Second {
				backoffDelay = 30 * time.Second
			}
			time.Sleep(backoffDelay)
		}

		resp, err := h.Do(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf("请求失败，已重试 %d 次，最后错误: %v", maxRetries, lastErr)
}

// ReadResponseBody 读取响应体
func (h *httpClientImpl) ReadResponseBody(resp *http.Response) ([]byte, error) {
	var reader io.Reader = resp.Body

	// 检查是否为gzip压缩
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("创建gzip读取器失败: %v", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	return io.ReadAll(reader)
}

// TestProxyConnection 测试代理连接
func (h *httpClientImpl) TestProxyConnection() error {
	if !h.config.ProxyEnabled || h.config.ProxyURL == "" {
		return nil
	}

	req, err := h.CreateRequest("GET", "https://httpbin.org/ip", nil)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := h.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("代理测试失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// SetConfig 更新配置
func (h *httpClientImpl) SetConfig(config HTTPClientConfig) {
	h.config = config
	h.createClient()
}

// GetConfig 获取当前配置
func (h *httpClientImpl) GetConfig() HTTPClientConfig {
	return h.config
}

// DefaultHTTPClientConfig 创建默认HTTP客户端配置
func DefaultHTTPClientConfig() HTTPClientConfig {
	return HTTPClientConfig{
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0",
		Accept:         "image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8",
		AcceptLanguage: "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
		Referer:        "https://www.pixiv.net/",
		Timeout:        30,
		RetryCount:     3,
		Delay:          5 * time.Second,
		ProxyEnabled:   false,
		ProxyURL:       "http://127.0.0.1:7890",
		Cookie:         "",
	}
}
