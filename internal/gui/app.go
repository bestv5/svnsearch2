package gui

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"svnsearch/internal/config"
	"svnsearch/internal/generator"
	"svnsearch/internal/scanner"
	"svnsearch/internal/scheduler"
	"svnsearch/pkg/logger"
	"svnsearch/pkg/utils"
)

type App struct {
	fyneApp    fyne.App
	mainWindow fyne.Window
	config     *config.Config
	configPath string
	scanner    *scanner.Scanner
	scheduler  *scheduler.Scheduler
	logger     *logger.Logger
	cancelFunc context.CancelFunc
	mu         sync.Mutex
}

func NewApp(configPath string) (*App, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	logDir := filepath.Join(filepath.Dir(configPath), "logs")
	if err := logger.InitLogger(logDir, cfg.Settings.LogLevel); err != nil {
		return nil, fmt.Errorf("初始化日志失败: %w", err)
	}

	app := &App{
		fyneApp:    app.New(),
		config:     cfg,
		configPath: configPath,
		scanner:    scanner.NewScanner("svn", cfg.Settings.MaxWorkers),
		scheduler:  scheduler.NewScheduler(),
	}

	app.mainWindow = app.createMainWindow()
	app.setupAutoScan()

	return app, nil
}

func (a *App) Run() {
	a.mainWindow.ShowAndRun()
}

func (a *App) createMainWindow() fyne.Window {
	w := a.fyneApp.NewWindow("SVN索引管理器")
	w.Resize(fyne.NewSize(800, 600))

	var selectedIndex int

	repoList := widget.NewList(
		func() int {
			return len(a.config.Repositories)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewCheck("", nil),
				widget.NewLabel("仓库名称"),
				widget.NewLabel("SVN地址"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(a.config.Repositories) {
				return
			}
			repo := a.config.Repositories[id]
			hbox := obj.(*fyne.Container)
			check := hbox.Objects[0].(*widget.Check)
			nameLabel := hbox.Objects[1].(*widget.Label)
			urlLabel := hbox.Objects[2].(*widget.Label)

			check.Checked = repo.Enabled
			check.OnChanged = func(checked bool) {
				repo.Enabled = checked
				a.saveConfig()
			}

			nameLabel.SetText(repo.Name)
			urlLabel.SetText(repo.URL)
		},
	)

	repoList.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
	}

	repoList.OnUnselected = func(id widget.ListItemID) {
		selectedIndex = -1
	}

	addBtn := widget.NewButton("添加", func() {
		a.showConfigDialog(nil)
	})

	editBtn := widget.NewButton("编辑", func() {
		if len(a.config.Repositories) == 0 {
			dialog.ShowInformation("提示", "没有可编辑的仓库", w)
			return
		}
		if selectedIndex < 0 || selectedIndex >= len(a.config.Repositories) {
			dialog.ShowInformation("提示", "请先选择一个仓库", w)
			return
		}
		repo := a.config.Repositories[selectedIndex]
		a.showConfigDialog(&repo)
	})

	deleteBtn := widget.NewButton("删除", func() {
		if len(a.config.Repositories) == 0 {
			dialog.ShowInformation("提示", "没有可删除的仓库", w)
			return
		}
		if selectedIndex < 0 || selectedIndex >= len(a.config.Repositories) {
			dialog.ShowInformation("提示", "请先选择一个仓库", w)
			return
		}
		repo := a.config.Repositories[selectedIndex]
		dialog.ShowConfirm("确认删除", fmt.Sprintf("确定要删除仓库 '%s' 吗?", repo.Name),
			func(confirmed bool) {
				if confirmed {
					a.config.DeleteRepository(repo.ID)
					a.saveConfig()
					repoList.Refresh()
				}
			}, w)
	})

	scanSelectedBtn := widget.NewButton("扫描选中", func() {
		a.scanRepositories(false)
	})

	scanAllBtn := widget.NewButton("扫描全部", func() {
		a.scanRepositories(true)
	})

	stopScanBtn := widget.NewButton("停止扫描", func() {
		a.stopScan()
	})

	autoScanCheck := widget.NewCheck("自动扫描", func(checked bool) {
		a.config.Settings.MonitorChanges = checked
		a.saveConfig()
		if checked {
			a.setupAutoScan()
		} else {
			a.scheduler.Stop()
		}
	})
	autoScanCheck.Checked = a.config.Settings.MonitorChanges

	intervalEntry := widget.NewEntry()
	intervalEntry.SetText(fmt.Sprintf("%d", a.config.Settings.AutoScanInterval))
	intervalEntry.OnChanged = func(text string) {
		var interval int
		fmt.Sscanf(text, "%d", &interval)
		if interval > 0 {
			a.config.Settings.AutoScanInterval = interval
			a.saveConfig()
			a.setupAutoScan()
		}
	}

	logEntry := widget.NewMultiLineEntry()
	logEntry.Wrapping = fyne.TextWrapWord
	logEntry.SetPlaceHolder("扫描日志将显示在这里...")

	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	statusLabel := widget.NewLabel("就绪")

	topBox := container.NewVBox(
		widget.NewLabel("仓库列表"),
		repoList,
		container.NewHBox(addBtn, editBtn, deleteBtn),
		widget.NewSeparator(),
		widget.NewLabel("扫描控制"),
		container.NewHBox(scanSelectedBtn, scanAllBtn, stopScanBtn),
		container.NewHBox(autoScanCheck, widget.NewLabel("间隔(分钟):"), intervalEntry),
		widget.NewSeparator(),
		widget.NewLabel("扫描日志"),
	)

	content := container.NewBorder(
		topBox,
		container.NewVBox(progressBar, statusLabel),
		nil, nil,
		logEntry,
	)

	w.SetContent(content)
	w.SetMaster()

	return w
}

func (a *App) showConfigDialog(repo *config.Repository) {
	var editingRepo config.Repository
	if repo != nil {
		editingRepo = *repo
	} else {
		editingRepo = config.Repository{
			ID:        utils.GenerateID(),
			Enabled:   true,
			ScanPaths: []string{"/"},
		}
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetText(editingRepo.Name)
	nameEntry.SetPlaceHolder("输入仓库名称")

	urlEntry := widget.NewEntry()
	urlEntry.SetText(editingRepo.URL)
	urlEntry.SetPlaceHolder("svn://example.com/repo")

	usernameEntry := widget.NewEntry()
	usernameEntry.SetText(editingRepo.Username)
	usernameEntry.SetPlaceHolder("SVN用户名")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetText(editingRepo.Password)
	passwordEntry.SetPlaceHolder("SVN密码")

	scanPathsEntry := widget.NewMultiLineEntry()
	scanPathsEntry.SetText(joinPaths(editingRepo.ScanPaths))
	scanPathsEntry.SetPlaceHolder("每行一个路径，如：\n/trunk\n/branches")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "仓库名称", Widget: nameEntry},
			{Text: "SVN地址", Widget: urlEntry},
			{Text: "用户名", Widget: usernameEntry},
			{Text: "密码", Widget: passwordEntry},
			{Text: "扫描路径", Widget: scanPathsEntry},
		},
		OnSubmit: func() {
			editingRepo.Name = nameEntry.Text
			editingRepo.URL = urlEntry.Text
			editingRepo.Username = usernameEntry.Text
			editingRepo.Password = passwordEntry.Text
			editingRepo.ScanPaths = parsePaths(scanPathsEntry.Text)

			if repo == nil {
				a.config.AddRepository(editingRepo)
			} else {
				a.config.UpdateRepository(editingRepo)
			}
			a.saveConfig()
		},
	}

	dialog.ShowForm("配置仓库", "保存", "取消", form.Items, func(confirmed bool) {
		if confirmed {
			form.OnSubmit()
		}
	}, a.mainWindow)
}

func (a *App) scanRepositories(scanAll bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var repos []config.Repository
	if scanAll {
		repos = a.config.Repositories
	} else {
		for _, repo := range a.config.Repositories {
			if repo.Enabled {
				repos = append(repos, repo)
			}
		}
	}

	if len(repos) == 0 {
		dialog.ShowInformation("提示", "没有需要扫描的仓库", a.mainWindow)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	a.cancelFunc = cancel

	go func() {
		for _, repo := range repos {
			select {
			case <-ctx.Done():
				return
			default:
				a.scanRepository(ctx, &repo)
			}
		}
	}()
}

func (a *App) scanRepository(ctx context.Context, repo *config.Repository) {
	logger.Info("开始扫描仓库: %s", repo.Name)

	result, err := a.scanner.ScanRepository(ctx, repo)
	if err != nil {
		logger.Error("扫描仓库 %s 失败: %v", repo.Name, err)
		return
	}

	efuPath := filepath.Join(a.config.Settings.EFUOutputDir, fmt.Sprintf("%s.efu", repo.Name))
	gen := generator.NewEFUGenerator(efuPath)
	if err := gen.Generate(result.Files, repo.Name, repo.URL); err != nil {
		logger.Error("生成EFU文件失败: %v", err)
		return
	}

	repo.LastScanTime = time.Now()
	a.saveConfig()

	logger.Info("扫描完成: %s, 共 %d 个文件, 耗时 %v", repo.Name, len(result.Files), result.Duration)
}

func (a *App) stopScan() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancelFunc != nil {
		a.cancelFunc()
		a.cancelFunc = nil
		logger.Info("扫描已停止")
	}
}

func (a *App) setupAutoScan() {
	a.scheduler.Stop()

	if !a.config.Settings.MonitorChanges {
		return
	}

	interval := time.Duration(a.config.Settings.AutoScanInterval) * time.Minute
	for _, repo := range a.config.Repositories {
		if repo.Enabled {
			repoCopy := repo
			a.scheduler.AddJob(repo.ID, interval, func() {
				ctx := context.Background()
				a.scanRepository(ctx, &repoCopy)
			})
		}
	}
}

func (a *App) saveConfig() {
	if err := config.SaveConfig(a.configPath, a.config); err != nil {
		logger.Error("保存配置失败: %v", err)
	}
}

func joinPaths(paths []string) string {
	result := ""
	for _, path := range paths {
		result += path + "\n"
	}
	return result
}

func parsePaths(text string) []string {
	var paths []string
	for _, line := range splitLines(text) {
		line = trimSpace(line)
		if line != "" {
			paths = append(paths, line)
		}
	}
	return paths
}

func splitLines(text string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			lines = append(lines, text[start:i])
			start = i + 1
		}
	}
	if start < len(text) {
		lines = append(lines, text[start:])
	}
	return lines
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
