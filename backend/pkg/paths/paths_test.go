package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathManager(t *testing.T) {
	// 创建临时目录作为测试根目录
	tempDir, err := os.MkdirTemp("", "pixiv-tailor-test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 初始化路径管理器
	err = InitPathManager(tempDir)
	if err != nil {
		t.Fatalf("初始化路径管理器失败: %v", err)
	}

	pm := GetPathManager()
	if pm == nil {
		t.Fatal("路径管理器为空")
	}

	// 测试根目录
	if pm.GetRootDir() != tempDir {
		t.Errorf("期望根目录 %s，实际 %s", tempDir, pm.GetRootDir())
	}

	// 测试配置目录
	expectedConfigDir := filepath.Join(tempDir, "backend", "configs")
	if pm.GetConfigDir() != expectedConfigDir {
		t.Errorf("期望配置目录 %s，实际 %s", expectedConfigDir, pm.GetConfigDir())
	}

	// 测试数据目录
	expectedDataDir := filepath.Join(tempDir, "backend", "data")
	if pm.GetDataDir() != expectedDataDir {
		t.Errorf("期望数据目录 %s，实际 %s", expectedDataDir, pm.GetDataDir())
	}

	// 测试图片目录
	expectedImagesDir := filepath.Join(tempDir, "backend", "data", "images")
	if pm.GetImagesDir() != expectedImagesDir {
		t.Errorf("期望图片目录 %s，实际 %s", expectedImagesDir, pm.GetImagesDir())
	}

	// 测试日志目录
	expectedLogsDir := filepath.Join(tempDir, "backend", "data", "logs")
	if pm.GetLogsDir() != expectedLogsDir {
		t.Errorf("期望日志目录 %s，实际 %s", expectedLogsDir, pm.GetLogsDir())
	}

	// 测试路径连接
	testPath := pm.Join("test", "file.txt")
	expectedPath := filepath.Join(tempDir, "test", "file.txt")
	if testPath != expectedPath {
		t.Errorf("期望路径 %s，实际 %s", expectedPath, testPath)
	}

	// 测试配置文件路径
	configPath := pm.GetCrawlerConfigPath()
	expectedConfigPath := filepath.Join(tempDir, "backend", "configs", "crawler_config.json")
	if configPath != expectedConfigPath {
		t.Errorf("期望配置文件路径 %s，实际 %s", expectedConfigPath, configPath)
	}

	// 测试数据库路径
	dbPath := pm.GetDatabasePath()
	expectedDbPath := filepath.Join(tempDir, "backend", "data", "database", "pixiv_tailor.db")
	if dbPath != expectedDbPath {
		t.Errorf("期望数据库路径 %s，实际 %s", expectedDbPath, dbPath)
	}

	// 测试日志文件路径
	logPath := pm.GetMainLogPath()
	expectedLogPath := filepath.Join(tempDir, "backend", "data", "logs", "pixiv-tailor.log")
	if logPath != expectedLogPath {
		t.Errorf("期望日志文件路径 %s，实际 %s", expectedLogPath, logPath)
	}
}

func TestPathManagerWithEmptyRoot(t *testing.T) {
	// 测试使用空字符串作为根目录（应该使用当前工作目录）
	err := InitPathManager("")
	if err != nil {
		t.Fatalf("初始化路径管理器失败: %v", err)
	}

	pm := GetPathManager()
	if pm == nil {
		t.Fatal("路径管理器为空")
	}

	// 验证根目录不为空
	if pm.GetRootDir() == "" {
		t.Error("根目录不应为空")
	}
}
