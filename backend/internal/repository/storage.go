package repository

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Storage 定义数据存储接口
type Storage interface {
	Close() error
	CreateTask(task *Task) error
	GetTask(id string) (*Task, error)
	UpdateTaskStatus(id, status string) error
	UpdateTaskProgress(id string, progress int) error
	UpdateTaskError(id, errorMsg string) error
	UpdateTaskImagesFound(id string, count int) error
	UpdateTaskImagesDownloaded(id string, count int) error
	UpdateTaskResult(id string, result string) error
	ListTasks(status, taskType string, limit, offset int) ([]*Task, error)
	CountTasks(status, taskType string) (int, error)
	DeleteTask(id string) error
	CleanupTasksByStatus(status string) (int, error)
	CleanupAllTasks() (int, error)
	GetConfig(module string) (string, error)
	SetConfig(module, config string) error
	GetCrawlResults(limit, offset int) ([]*CrawlResult, error)
	CountCrawlResults() (int, error)
	AddCrawlResult(result *CrawlResult) error
	DeleteCrawlResultsByTaskID(taskID string) error
	GetGeneratedImages(limit, offset int) ([]*GeneratedImage, error)
	CountGeneratedImages() (int, error)
	AddGeneratedImage(image *GeneratedImage) error
	GetTrainedModels(limit, offset int) ([]*TrainedModel, error)
	CountTrainedModels() (int, error)
}

// Task 任务结构
type Task struct {
	ID               string    `json:"id"`
	Type             string    `json:"type"`
	Status           string    `json:"status"`
	Config           string    `json:"config"`
	Progress         int       `json:"progress"`
	ErrorMessage     string    `json:"error_message"`
	Result           string    `json:"result"`            // 任务结果（JSON格式）
	ImagesFound      int       `json:"images_found"`      // 获取到的图片数量
	ImagesDownloaded int       `json:"images_downloaded"` // 下载的图片数量
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CrawlResult 爬取结果结构
type CrawlResult struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Tags      string    `json:"tags"` // JSON 字符串
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
}

// GeneratedImage 生成图像结构
type GeneratedImage struct {
	ID        string    `json:"id"`
	Prompt    string    `json:"prompt"`
	ImageURL  string    `json:"image_url"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
}

// TrainedModel 训练模型结构
type TrainedModel struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

// SQLiteStorage SQLite存储实现
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage 创建SQLite存储实例
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	storage := &SQLiteStorage{db: db}
	if err := storage.initTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("初始化表失败: %v", err)
	}

	return storage, nil
}

// Close 关闭数据库连接
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// GetDB 获取数据库连接
func (s *SQLiteStorage) GetDB() *sql.DB {
	return s.db
}

// initTables 初始化数据库表
func (s *SQLiteStorage) initTables() error {
	// 先检查并添加新字段（数据库迁移）
	if err := s.migrateTables(); err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			config TEXT,
			progress INTEGER DEFAULT 0,
			error_message TEXT,
			result TEXT,
			images_found INTEGER DEFAULT 0,
			images_downloaded INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS configs (
			module TEXT PRIMARY KEY,
			config TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS crawl_results (
			id TEXT PRIMARY KEY,
			url TEXT,
			title TEXT,
			author TEXT,
			tags TEXT,
			image_url TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS generated_images (
			id TEXT PRIMARY KEY,
			prompt TEXT,
			image_url TEXT,
			model TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS trained_models (
			id TEXT PRIMARY KEY,
			name TEXT,
			type TEXT,
			path TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("执行SQL失败: %v", err)
		}
	}

	return nil
}

// migrateTables 数据库迁移
func (s *SQLiteStorage) migrateTables() error {
	// 检查并添加 images_found 字段
	if err := s.addColumnIfNotExists("tasks", "images_found", "INTEGER DEFAULT 0"); err != nil {
		return err
	}

	// 检查并添加 images_downloaded 字段
	if err := s.addColumnIfNotExists("tasks", "images_downloaded", "INTEGER DEFAULT 0"); err != nil {
		return err
	}

	// 检查并添加 result 字段
	if err := s.addColumnIfNotExists("tasks", "result", "TEXT"); err != nil {
		return err
	}

	return nil
}

// addColumnIfNotExists 如果列不存在则添加列
func (s *SQLiteStorage) addColumnIfNotExists(tableName, columnName, columnType string) error {
	// 检查列是否存在
	var count int
	query := `SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`
	err := s.db.QueryRow(query, tableName, columnName).Scan(&count)
	if err != nil {
		return fmt.Errorf("检查列是否存在失败: %v", err)
	}

	// 如果列不存在，则添加
	if count == 0 {
		alterQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnType)
		_, err := s.db.Exec(alterQuery)
		if err != nil {
			return fmt.Errorf("添加列 %s 失败: %v", columnName, err)
		}
	}

	return nil
}

// CreateTask 创建任务
func (s *SQLiteStorage) CreateTask(task *Task) error {
	query := `INSERT INTO tasks (id, type, status, config, progress, images_found, images_downloaded, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, task.ID, task.Type, task.Status, task.Config, task.Progress, task.ImagesFound, task.ImagesDownloaded, task.CreatedAt, task.UpdatedAt)
	return err
}

// GetTask 获取任务
func (s *SQLiteStorage) GetTask(id string) (*Task, error) {
	query := `SELECT id, type, status, config, progress, error_message, result, images_found, images_downloaded, created_at, updated_at 
			  FROM tasks WHERE id = ?`
	row := s.db.QueryRow(query, id)

	task := &Task{}
	var errorMessage sql.NullString
	var result sql.NullString
	err := row.Scan(&task.ID, &task.Type, &task.Status, &task.Config, &task.Progress,
		&errorMessage, &result, &task.ImagesFound, &task.ImagesDownloaded, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if errorMessage.Valid {
		task.ErrorMessage = errorMessage.String
	}
	if result.Valid {
		task.Result = result.String
	}

	return task, nil
}

// UpdateTaskStatus 更新任务状态
func (s *SQLiteStorage) UpdateTaskStatus(id, status string) error {
	query := `UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.Exec(query, status, id)
	return err
}

// UpdateTaskProgress 更新任务进度
func (s *SQLiteStorage) UpdateTaskProgress(id string, progress int) error {
	query := `UPDATE tasks SET progress = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.Exec(query, progress, id)
	return err
}

// UpdateTaskImagesFound 更新获取到的图片数量
func (s *SQLiteStorage) UpdateTaskImagesFound(id string, count int) error {
	query := `UPDATE tasks SET images_found = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.Exec(query, count, id)
	return err
}

// UpdateTaskImagesDownloaded 更新下载的图片数量
func (s *SQLiteStorage) UpdateTaskImagesDownloaded(id string, count int) error {
	query := `UPDATE tasks SET images_downloaded = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.Exec(query, count, id)
	return err
}

// UpdateTaskResult 更新任务结果
func (s *SQLiteStorage) UpdateTaskResult(id string, result string) error {
	query := `UPDATE tasks SET result = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.Exec(query, result, id)
	return err
}

// UpdateTaskError 更新任务错误信息
func (s *SQLiteStorage) UpdateTaskError(id, errorMsg string) error {
	query := `UPDATE tasks SET error_message = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.Exec(query, errorMsg, id)
	return err
}

// ListTasks 列出任务
func (s *SQLiteStorage) ListTasks(status, taskType string, limit, offset int) ([]*Task, error) {
	query := `SELECT id, type, status, config, progress, error_message, result, images_found, images_downloaded, created_at, updated_at 
			  FROM tasks WHERE 1=1`
	args := []interface{}{}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if taskType != "" {
		query += " AND type = ?"
		args = append(args, taskType)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		var errorMessage sql.NullString
		var result sql.NullString
		err := rows.Scan(&task.ID, &task.Type, &task.Status, &task.Config, &task.Progress,
			&errorMessage, &result, &task.ImagesFound, &task.ImagesDownloaded, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if errorMessage.Valid {
			task.ErrorMessage = errorMessage.String
		}
		if result.Valid {
			task.Result = result.String
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetConfig 获取配置
func (s *SQLiteStorage) GetConfig(module string) (string, error) {
	query := `SELECT config FROM configs WHERE module = ?`
	var config string
	err := s.db.QueryRow(query, module).Scan(&config)
	if err != nil {
		if err == sql.ErrNoRows {
			return "{}", nil // 返回空JSON对象
		}
		return "", err
	}
	return config, nil
}

// SetConfig 设置配置
func (s *SQLiteStorage) SetConfig(module, config string) error {
	query := `INSERT OR REPLACE INTO configs (module, config, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)`
	_, err := s.db.Exec(query, module, config)
	return err
}

// GetCrawlResults 获取爬取结果
func (s *SQLiteStorage) GetCrawlResults(limit, offset int) ([]*CrawlResult, error) {
	query := `SELECT id, url, title, author, tags, image_url, created_at 
			  FROM crawl_results ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*CrawlResult
	for rows.Next() {
		result := &CrawlResult{}
		err := rows.Scan(&result.ID, &result.URL, &result.Title, &result.Author,
			&result.Tags, &result.ImageURL, &result.CreatedAt)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// GetGeneratedImages 获取生成的图像
func (s *SQLiteStorage) GetGeneratedImages(limit, offset int) ([]*GeneratedImage, error) {
	query := `SELECT id, prompt, image_url, model, created_at 
			  FROM generated_images ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []*GeneratedImage
	for rows.Next() {
		image := &GeneratedImage{}
		err := rows.Scan(&image.ID, &image.Prompt, &image.ImageURL, &image.Model, &image.CreatedAt)
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	return images, nil
}

// GetTrainedModels 获取训练的模型
func (s *SQLiteStorage) GetTrainedModels(limit, offset int) ([]*TrainedModel, error) {
	query := `SELECT id, name, type, path, created_at 
			  FROM trained_models ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []*TrainedModel
	for rows.Next() {
		model := &TrainedModel{}
		err := rows.Scan(&model.ID, &model.Name, &model.Type, &model.Path, &model.CreatedAt)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, nil
}

// CountTasks 统计任务数量
func (s *SQLiteStorage) CountTasks(status, taskType string) (int, error) {
	query := `SELECT COUNT(*) FROM tasks WHERE 1=1`
	args := []interface{}{}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if taskType != "" {
		query += " AND type = ?"
		args = append(args, taskType)
	}

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

// CountCrawlResults 统计爬取结果数量
func (s *SQLiteStorage) CountCrawlResults() (int, error) {
	query := `SELECT COUNT(*) FROM crawl_results`
	var count int
	err := s.db.QueryRow(query).Scan(&count)
	return count, err
}

// CountGeneratedImages 统计生成图像数量
func (s *SQLiteStorage) CountGeneratedImages() (int, error) {
	query := `SELECT COUNT(*) FROM generated_images`
	var count int
	err := s.db.QueryRow(query).Scan(&count)
	return count, err
}

// CountTrainedModels 统计训练模型数量
func (s *SQLiteStorage) CountTrainedModels() (int, error) {
	query := `SELECT COUNT(*) FROM trained_models`
	var count int
	err := s.db.QueryRow(query).Scan(&count)
	return count, err
}

// CleanupTasksByStatus 根据状态清理任务
func (s *SQLiteStorage) CleanupTasksByStatus(status string) (int, error) {
	query := `DELETE FROM tasks WHERE status = ?`
	result, err := s.db.Exec(query, status)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}

// CleanupAllTasks 清理所有任务
func (s *SQLiteStorage) CleanupAllTasks() (int, error) {
	query := `DELETE FROM tasks`
	result, err := s.db.Exec(query)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}

// DeleteTask 删除指定任务
func (s *SQLiteStorage) DeleteTask(id string) error {
	query := `DELETE FROM tasks WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}

// AddCrawlResult 添加爬取结果
func (s *SQLiteStorage) AddCrawlResult(result *CrawlResult) error {
	query := `INSERT INTO crawl_results (id, url, title, author, tags, image_url, created_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, result.ID, result.URL, result.Title, result.Author, result.Tags, result.ImageURL, result.CreatedAt)
	return err
}

// AddGeneratedImage 添加生成的图像
func (s *SQLiteStorage) AddGeneratedImage(image *GeneratedImage) error {
	query := `INSERT INTO generated_images (id, prompt, image_url, model, created_at) 
			  VALUES (?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, image.ID, image.Prompt, image.ImageURL, image.Model, image.CreatedAt)
	return err
}

// DeleteCrawlResultsByTaskID 根据任务ID删除爬取结果
func (s *SQLiteStorage) DeleteCrawlResultsByTaskID(taskID string) error {
	query := `DELETE FROM crawl_results WHERE id LIKE ?`
	_, err := s.db.Exec(query, taskID+"_%")
	return err
}
