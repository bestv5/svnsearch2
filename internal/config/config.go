package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)



type Config struct {
	Repositories []Repository `json:"repositories"`
	Settings     Settings     `json:"settings"`
}

type Repository struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	URL             string    `json:"url"`
	Username        string    `json:"username"`
	Password        string    `json:"password"`
	ScanPaths       []string  `json:"scan_paths"`
	ExcludePatterns []string  `json:"exclude_patterns"`
	LastScanTime    time.Time `json:"last_scan_time"`
	Enabled         bool      `json:"enabled"`
}

type Settings struct {
	AutoScanInterval int    `json:"auto_scan_interval"`
	EFUOutputDir     string `json:"efu_output_dir"`
	MonitorChanges   bool   `json:"monitor_changes"`
	LogLevel         string `json:"log_level"`
	MaxWorkers       int    `json:"max_workers"`
}

var DefaultSettings = Settings{
	AutoScanInterval: 15,
	EFUOutputDir:     "./data/efu",
	MonitorChanges:   true,
	LogLevel:         "INFO",
	MaxWorkers:       50,
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Repositories: []Repository{},
				Settings:     DefaultSettings,
			}, nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	if config.Settings.AutoScanInterval == 0 {
		config.Settings.AutoScanInterval = DefaultSettings.AutoScanInterval
	}
	if config.Settings.EFUOutputDir == "" {
		config.Settings.EFUOutputDir = DefaultSettings.EFUOutputDir
	}
	if config.Settings.MaxWorkers == 0 {
		config.Settings.MaxWorkers = DefaultSettings.MaxWorkers
	}
	if config.Settings.LogLevel == "" {
		config.Settings.LogLevel = DefaultSettings.LogLevel
	}

	// 直接返回，不进行密码解密
	return &config, nil
}

func SaveConfig(path string, config *Config) error {
	configCopy := *config
	configCopy.Repositories = make([]Repository, len(config.Repositories))
	copy(configCopy.Repositories, config.Repositories)

	// 直接保存，不进行密码加密

	data, err := json.MarshalIndent(configCopy, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

func (c *Config) AddRepository(repo Repository) {
	repo.ID = uuid.New().String()
	c.Repositories = append(c.Repositories, repo)
}

func (c *Config) UpdateRepository(repo Repository) error {
	for i, r := range c.Repositories {
		if r.ID == repo.ID {
			c.Repositories[i] = repo
			return nil
		}
	}
	return fmt.Errorf("仓库不存在: %s", repo.ID)
}

func (c *Config) DeleteRepository(id string) error {
	for i, r := range c.Repositories {
		if r.ID == id {
			c.Repositories = append(c.Repositories[:i], c.Repositories[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("仓库不存在: %s", id)
}

func (c *Config) GetRepository(id string) (*Repository, error) {
	for _, r := range c.Repositories {
		if r.ID == id {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("仓库不存在: %s", id)
}

func (c *Config) GetAllRepositories() []Repository {
	return c.Repositories
}
