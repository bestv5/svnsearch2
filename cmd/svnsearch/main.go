package main

import (
	"fmt"
	"os"
	"path/filepath"

	"svnsearch/internal/gui"
	"svnsearch/pkg/logger"
)

var (
	Version   = "1.0.0"
	BuildTime = "unknown"
)

func main() {
	configPath := getConfigPath()

	app, err := gui.NewApp(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化应用失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	fmt.Printf("SVN索引管理器 v%s (构建时间: %s)\n", Version, BuildTime)
	app.Run()
}

func getConfigPath() string {
	execPath, err := os.Executable()
	if err != nil {
		execPath = "."
	}

	configDir := filepath.Join(filepath.Dir(execPath), "configs")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		configDir = "."
	}

	return filepath.Join(configDir, "config.json")
}
