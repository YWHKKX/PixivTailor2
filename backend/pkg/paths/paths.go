package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// PathManager 路径管理器
type PathManager struct {
	rootDir   string
	configDir string
	dataDir   string
	logsDir   string
	imagesDir string
	modelsDir string
	tagsDir   string
	posesDir  string
	cacheDir  string
	mu        sync.RWMutex
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

	pm.configDir = filepath.Join(pm.rootDir, "backend", "configs")
	pm.dataDir = filepath.Join(pm.rootDir, "backend", "data")
	pm.logsDir = filepath.Join(pm.rootDir, "backend", "data", "logs")
	pm.imagesDir = filepath.Join(pm.rootDir, "backend", "data", "images")
	pm.modelsDir = filepath.Join(pm.rootDir, "backend", "data", "models")
	pm.tagsDir = filepath.Join(pm.rootDir, "backend", "data", "tags")
	pm.posesDir = filepath.Join(pm.rootDir, "backend", "data", "poses")
	pm.cacheDir = filepath.Join(pm.rootDir, "backend", "data", "cache")
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
		pm.cacheDir,
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

// GetCacheDir 获取缓存目录
func (pm *PathManager) GetCacheDir() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.cacheDir
}

// GetTaskCacheDir 获取任务缓存目录
func (pm *PathManager) GetTaskCacheDir(taskID string) string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return filepath.Join(pm.cacheDir, fmt.Sprintf("task_%s", taskID))
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
