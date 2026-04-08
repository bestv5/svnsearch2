package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"svnsearch/internal/config"
	"svnsearch/internal/generator"
	"svnsearch/internal/scanner"
	"svnsearch/pkg/logger"
	"svnsearch/pkg/utils"
)

var (
	Version   = "1.0.0"
	BuildTime = "unknown"
)

func main() {
	configPath := getConfigPath()

	for {
		printMenu()
		choice := getUserInput("请选择操作: ")
		switch choice {
		case "1":
			doAdd(configPath)
		case "2":
			doScanMenu(configPath)
		case "3":
			doList(configPath)
		case "4":
			doRemove(configPath)
		case "5":
			doEnableMenu(configPath, true)
		case "6":
			doEnableMenu(configPath, false)
		case "7":
			fmt.Printf("svnsearch v%s (构建时间: %s)\n", Version, BuildTime)
		case "8":
			fmt.Println("退出程序...")
			return
		default:
			fmt.Println("无效的选择，请重新输入")
		}
		fmt.Println()
		getUserInput("按回车键继续...")
	}
}

func printMenu() {
	fmt.Printf("=====================================\n")
	fmt.Printf("SVN索引管理器 v%s\n", Version)
	fmt.Printf("=====================================\n")
	fmt.Println("1. 添加仓库")
	fmt.Println("2. 扫描仓库")
	fmt.Println("3. 列出仓库")
	fmt.Println("4. 删除仓库")
	fmt.Println("5. 启用仓库")
	fmt.Println("6. 禁用仓库")
	fmt.Println("7. 显示版本")
	fmt.Println("8. 退出")
	fmt.Printf("=====================================\n")
}

func getUserInput(prompt string) string {
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}

func getConfigPath() string {
	execPath, err := os.Executable()
	if err != nil {
		execPath = "."
	}
	configDir := filepath.Join(filepath.Dir(execPath), "configs")
	os.MkdirAll(configDir, 0755)
	return filepath.Join(configDir, "config.json")
}

func loadConfig(configPath string) *config.Config {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

func saveConfig(configPath string, cfg *config.Config) {
	if err := config.SaveConfig(configPath, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "保存配置失败: %v\n", err)
		os.Exit(1)
	}
}

func doAdd(configPath string) {
	fmt.Println("=====================================")
	fmt.Println("添加仓库")
	fmt.Println("=====================================")

	name := getUserInput("仓库名称: ")
	if name == "" {
		fmt.Println("错误: 仓库名称不能为空")
		return
	}

	url := getUserInput("SVN地址: ")
	if url == "" {
		fmt.Println("错误: SVN地址不能为空")
		return
	}

	user := getUserInput("用户名 (可选): ")
	pass := getUserInput("密码 (可选): ")
	pathsInput := getUserInput("扫描路径 (默认为/, 多个路径用逗号分隔): ")
	if pathsInput == "" {
		pathsInput = "/"
	}

	cfg := loadConfig(configPath)

	for _, r := range cfg.Repositories {
		if r.Name == name {
			fmt.Printf("错误: 仓库 '%s' 已存在\n", name)
			return
		}
	}

	scanPaths := parsePaths(pathsInput)
	repo := config.Repository{
		ID:        utils.GenerateID(),
		Name:      name,
		URL:       url,
		Username:  user,
		Password:  pass,
		ScanPaths: scanPaths,
		Enabled:   true,
	}

	cfg.Repositories = append(cfg.Repositories, repo)
	saveConfig(configPath, cfg)
	fmt.Printf("✓ 已添加仓库: %s\n", name)
}

func doScanMenu(configPath string) {
	fmt.Println("=====================================")
	fmt.Println("扫描仓库")
	fmt.Println("=====================================")
	fmt.Println("1. 扫描所有启用的仓库")
	fmt.Println("2. 扫描指定仓库")
	choice := getUserInput("请选择: ")

	switch choice {
	case "1":
		doScan(configPath, true, "")
	case "2":
		name := getUserInput("仓库名称: ")
		if name == "" {
			fmt.Println("错误: 仓库名称不能为空")
			return
		}
		doScan(configPath, false, name)
	default:
		fmt.Println("无效的选择")
	}
}

func doScan(configPath string, scanAll bool, name string) {
	cfg := loadConfig(configPath)

	logDir := filepath.Join(filepath.Dir(configPath), "logs")
	if err := logger.InitLogger(logDir, cfg.Settings.LogLevel); err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		return
	}
	defer logger.Close()

	var repos []config.Repository
	if scanAll {
		for _, r := range cfg.Repositories {
			if r.Enabled {
				repos = append(repos, r)
			}
		}
	} else {
		for _, r := range cfg.Repositories {
			if r.Name == name {
				repos = append(repos, r)
				break
			}
		}
		if len(repos) == 0 {
			fmt.Printf("错误: 未找到仓库 '%s'\n", name)
			return
		}
	}

	if len(repos) == 0 {
		fmt.Println("没有需要扫描的仓库")
		return
	}

	sc := scanner.NewScanner("svn", cfg.Settings.MaxWorkers)
	ctx := context.Background()

	for _, repo := range repos {
		fmt.Printf("正在扫描: %s ...\n", repo.Name)
		result, err := sc.ScanRepository(ctx, &repo)
		if err != nil {
			fmt.Printf("  ✗ 扫描失败: %v\n", err)
			continue
		}

		efuPath := filepath.Join(cfg.Settings.EFUOutputDir, repo.Name+".efu")
		gen := generator.NewEFUGenerator(efuPath)
		if err := gen.Generate(result.Files, repo.Name, repo.URL); err != nil {
			fmt.Printf("  ✗ 生成EFU失败: %v\n", err)
			continue
		}

		for i, r := range cfg.Repositories {
			if r.ID == repo.ID {
				cfg.Repositories[i].LastScanTime = time.Now()
				break
			}
		}

		fmt.Printf("  ✓ 完成: %d 个文件, 耗时 %v\n", len(result.Files), result.Duration)
		fmt.Printf("  EFU文件: %s\n", efuPath)
	}

	saveConfig(configPath, cfg)
}

func doList(configPath string) {
	cfg := loadConfig(configPath)

	if len(cfg.Repositories) == 0 {
		fmt.Println("没有配置任何仓库")
		return
	}

	fmt.Println("=====================================")
	fmt.Println("仓库列表")
	fmt.Println("=====================================")

	for i, r := range cfg.Repositories {
		status := "禁用"
		if r.Enabled {
			status = "启用"
		}
		lastScan := "从未扫描"
		if !r.LastScanTime.IsZero() {
			lastScan = r.LastScanTime.Format("2006-01-02 15:04:05")
		}
		fmt.Printf("%d. %s\n", i+1, r.Name)
		fmt.Printf("   URL: %s\n", r.URL)
		fmt.Printf("   状态: %s\n", status)
		fmt.Printf("   最后扫描: %s\n", lastScan)
		fmt.Printf("   扫描路径: %v\n", r.ScanPaths)
		fmt.Println("-------------------------------------")
	}
}

func doRemove(configPath string) {
	cfg := loadConfig(configPath)

	if len(cfg.Repositories) == 0 {
		fmt.Println("没有配置任何仓库")
		return
	}

	fmt.Println("=====================================")
	fmt.Println("删除仓库")
	fmt.Println("=====================================")

	for i, r := range cfg.Repositories {
		fmt.Printf("%d. %s\n", i+1, r.Name)
	}

	choice := getUserInput("请选择要删除的仓库编号: ")
	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(cfg.Repositories) {
		fmt.Println("无效的选择")
		return
	}

	repo := cfg.Repositories[index-1]
	confirm := getUserInput(fmt.Sprintf("确定要删除仓库 '%s' 吗? (y/n): ", repo.Name))
	if confirm != "y" && confirm != "Y" {
		fmt.Println("操作取消")
		return
	}

	cfg.Repositories = append(cfg.Repositories[:index-1], cfg.Repositories[index:]...)
	saveConfig(configPath, cfg)
	fmt.Printf("✓ 已删除仓库: %s\n", repo.Name)
}

func doEnableMenu(configPath string, enabled bool) {
	cfg := loadConfig(configPath)

	if len(cfg.Repositories) == 0 {
		fmt.Println("没有配置任何仓库")
		return
	}

	action := "启用"
	if !enabled {
		action = "禁用"
	}

	fmt.Printf("=====================================\n")
	fmt.Printf("%s仓库\n", action)
	fmt.Printf("=====================================\n")

	for i, r := range cfg.Repositories {
		fmt.Printf("%d. %s\n", i+1, r.Name)
	}

	choice := getUserInput("请选择要操作的仓库编号: ")
	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(cfg.Repositories) {
		fmt.Println("无效的选择")
		return
	}

	repo := &cfg.Repositories[index-1]
	repo.Enabled = enabled
	saveConfig(configPath, cfg)
	fmt.Printf("✓ 已%s仓库: %s\n", action, repo.Name)
}

func parsePaths(s string) []string {
	var paths []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			p := s[start:i]
			if p != "" {
				paths = append(paths, p)
			}
			start = i + 1
		}
	}
	if len(paths) == 0 {
		paths = []string{"/"}
	}
	return paths
}
