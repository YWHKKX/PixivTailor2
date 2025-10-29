package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// PathManager 路径管理器
type PathManager struct {
	rootDir       string
	configDir     string
	dataDir       string
	logsDir       string
	imagesDir     string
	modelsDir     string
	tagsDir       string
	posesDir      string
	charactersDir string
	configsDir    string
	webuiDir      string
	webuiBat      string
	mu            sync.RWMutex
}

var (
	globalPathManager *PathManager
	once              sync.Once
)

// InitPathManager 初始化路径管理器
func InitPathManager(rootDir string) error {
	var err error
	once.Do(func() {
		// 如果没有指定根目录，自动检测项目根目录
		if rootDir == "" {
			rootDir, err = detectProjectRoot()
			if err != nil {
				return
			}
		}

		globalPathManager = &PathManager{
			rootDir: rootDir,
		}

		// 初始化所有子目录路径
		globalPathManager.initPaths()

		// 确保所有必要的目录存在
		err = globalPathManager.ensureDirectories()
	})
	return err
}

// detectProjectRoot 自动检测项目根目录
func detectProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// 从当前目录向上查找，直到找到包含 go.mod 的目录
	for {
		// 检查当前目录是否包含 go.mod 文件
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return currentDir, nil
		}

		// 检查是否已经到达文件系统根目录
		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			// 如果找不到 go.mod，使用当前工作目录作为项目根目录
			return os.Getwd()
		}
		currentDir = parent
	}
}

// GetPathManager 获取全局路径管理器
func GetPathManager() *PathManager {
	return globalPathManager
}

// initPaths 初始化所有路径
func (pm *PathManager) initPaths() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.configDir = filepath.Join(pm.rootDir, "backend", "global_configs")
	pm.dataDir = filepath.Join(pm.rootDir, "backend", "data")
	pm.logsDir = filepath.Join(pm.rootDir, "backend", "data", "logs")
	pm.imagesDir = filepath.Join(pm.rootDir, "backend", "data", "images")
	pm.modelsDir = filepath.Join(pm.rootDir, "backend", "data", "models")
	pm.tagsDir = filepath.Join(pm.rootDir, "backend", "data", "tags")
	pm.posesDir = filepath.Join(pm.rootDir, "backend", "data", "poses")
	pm.charactersDir = filepath.Join(pm.rootDir, "backend", "data", "characters")
	pm.configsDir = filepath.Join(pm.rootDir, "backend", "data", "configs")

	// WebUI 路径配置
	pm.webuiDir = "D:\\PythonProject\\stable-diffusion-webui"
	pm.webuiBat = filepath.Join(pm.webuiDir, "webui.bat")
}

// ensureDirectories 确保所有必要的目录存在
func (pm *PathManager) ensureDirectories() error {
	pm.mu.RLock()
	dirs := []string{
		pm.configDir,
		pm.dataDir,
		pm.logsDir,
		pm.imagesDir,
		pm.modelsDir,
		pm.tagsDir,
		pm.posesDir,
		pm.charactersDir,
		pm.configsDir,
	}
	pm.mu.RUnlock()

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// GetRootDir 获取根目录
func (pm *PathManager) GetRootDir() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.rootDir
}

// GetConfigDir 获取配置目录
func (pm *PathManager) GetConfigDir() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.configDir
}

// GetDataDir 获取数据目录
func (pm *PathManager) GetDataDir() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.dataDir
}

// GetLogsDir 获取日志目录
func (pm *PathManager) GetLogsDir() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.logsDir
}

// GetImagesDir 获取图片目录
func (pm *PathManager) GetImagesDir() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.imagesDir
}

// GetModelsDir 获取模型目录
func (pm *PathManager) GetModelsDir() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.modelsDir
}

// GetTagsDir 获取标签目录
func (pm *PathManager) GetTagsDir() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.tagsDir
}

// GetPosesDir 获取姿势目录
func (pm *PathManager) GetPosesDir() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.posesDir
}

// GetConfigPath 获取配置文件路径
func (pm *PathManager) GetConfigPath(filename string) string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return filepath.Join(pm.configDir, filename)
}

// GetDataPath 获取数据文件路径
func (pm *PathManager) GetDataPath(filename string) string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return filepath.Join(pm.dataDir, filename)
}

// GetLogPath 获取日志文件路径
func (pm *PathManager) GetLogPath(filename string) string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return filepath.Join(pm.logsDir, filename)
}

// GetImagePath 获取图片文件路径
func (pm *PathManager) GetImagePath(filename string) string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return filepath.Join(pm.imagesDir, filename)
}

// GetModelPath 获取模型文件路径
func (pm *PathManager) GetModelPath(filename string) string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return filepath.Join(pm.modelsDir, filename)
}

// GetTagPath 获取标签文件路径
func (pm *PathManager) GetTagPath(filename string) string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return filepath.Join(pm.tagsDir, filename)
}

// GetPosePath 获取姿势文件路径
func (pm *PathManager) GetPosePath(filename string) string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return filepath.Join(pm.posesDir, filename)
}

// GetCharactersDir 获取角色配置文件目录
func (pm *PathManager) GetCharactersDir() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.charactersDir
}

// GetConfigsDir 获取配置目录 (backend/data/configs)
func (pm *PathManager) GetConfigsDir() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.configsDir
}

// GetTaskImagesDir 获取任务图片目录
// 新格式：2025-10-28_21-16_任务类型_哈希值
// 注意：冒号替换为连字符，避免Windows文件名限制
func (pm *PathManager) GetTaskImagesDir(taskID string, taskType string, createdAt time.Time) string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// 格式化时间为 YYYY-MM-DD_HH-MM 格式（冒号替换为连字符）
	timeStr := createdAt.Format("2006-01-02_15-04")

	// 组合为：2025-10-28_21-16_任务类型_哈希值
	dirName := fmt.Sprintf("%s_%s_%s", timeStr, taskType, taskID)

	return filepath.Join(pm.imagesDir, dirName)
}

// Join 连接路径（相对于根目录）
func (pm *PathManager) Join(elem ...string) string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return filepath.Join(append([]string{pm.rootDir}, elem...)...)
}

// Rel 获取相对于根目录的路径
func (pm *PathManager) Rel(targetPath string) (string, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return filepath.Rel(pm.rootDir, targetPath)
}

// Abs 获取绝对路径（相对于根目录）
func (pm *PathManager) Abs(elem ...string) string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return filepath.Join(append([]string{pm.rootDir}, elem...)...)
}

// 预定义的配置文件路径
func (pm *PathManager) GetMainConfigPath() string {
	return pm.GetConfigPath("config.json")
}

func (pm *PathManager) GetCrawlerConfigPath() string {
	return pm.GetConfigPath("crawler_config.json")
}

func (pm *PathManager) GetDatabasePath() string {
	return pm.GetDataPath("database/pixiv_tailor.db")
}

func (pm *PathManager) GetMainLogPath() string {
	return pm.GetLogPath("pixiv-tailor.log")
}

// GetWebUIDir 获取WebUI目录
func (pm *PathManager) GetWebUIDir() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.webuiDir
}

// GetWebUIBat 获取WebUI批处理文件路径
func (pm *PathManager) GetWebUIBat() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.webuiBat
}

// ResolvePath 解析相对路径为绝对路径
// 支持的前缀: images/, tags/, cache/, logs/, models/, poses/
// 如果不是这些前缀，则相对于根目录
func (pm *PathManager) ResolvePath(relativePath string) string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// 如果已经是绝对路径，直接返回
	if filepath.IsAbs(relativePath) {
		return relativePath
	}

	// 处理不同的路径前缀
	switch {
	case strings.HasPrefix(relativePath, "images/"):
		// 图片目录: images/task_xxx -> backend/data/images/task_xxx
		imageName := relativePath[7:] // len("images/") = 7
		return filepath.Join(pm.imagesDir, imageName)
	case strings.HasPrefix(relativePath, "tags/"):
		// 标签目录: tags/ 或 tags/subdir -> backend/data/tags/ 或 backend/data/tags/subdir
		tagName := relativePath[5:] // len("tags/") = 5
		// 如果 tagName 为空，返回 tagsDir 的完整路径
		if tagName == "" {
			return pm.tagsDir
		}
		return filepath.Join(pm.tagsDir, tagName)
	case strings.HasPrefix(relativePath, "logs/"):
		// 日志目录
		logName := relativePath[5:] // len("logs/") = 5
		if logName == "" {
			return pm.logsDir
		}
		return filepath.Join(pm.logsDir, logName)
	case strings.HasPrefix(relativePath, "models/"):
		// 模型目录
		modelName := relativePath[7:] // len("models/") = 7
		if modelName == "" {
			return pm.modelsDir
		}
		return filepath.Join(pm.modelsDir, modelName)
	case strings.HasPrefix(relativePath, "poses/"):
		// 姿势目录
		poseName := relativePath[6:] // len("poses/") = 6
		if poseName == "" {
			return pm.posesDir
		}
		return filepath.Join(pm.posesDir, poseName)
	default:
		// 默认相对于根目录
		return filepath.Join(pm.rootDir, relativePath)
	}
}
