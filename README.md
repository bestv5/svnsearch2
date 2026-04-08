# SVN索引管理器

一个基于Go语言开发的SVN仓库索引管理工具，用于生成EFU文件列表并集成到Everything搜索引擎中，实现SVN文件的快速搜索。

## 功能特性

- **多仓库管理**：支持同时管理多个SVN仓库
- **高性能扫描**：使用Go并发技术，扫描速度快
- **仓库标题标识**：每个仓库都有标题，方便识别
- **EFU文件生成**：生成符合Everything规范的EFU文件列表
- **自动扫描**：支持定时自动扫描
- **实时日志**：显示扫描进度和结果
- **现代化界面**：使用Fyne GUI框架，界面美观

## 技术栈

- **语言**：Go 1.21+
- **GUI框架**：Fyne
- **并发模型**：Goroutine + Channel
- **数据存储**：JSON配置文件

## 安装说明

### 系统要求

- Windows 7+ 或 macOS
- SVN命令行工具（TortoiseSVN或CollabNet SVN）
- Everything搜索引擎（v1.4+）

### 安装步骤

1. **安装依赖**：
   - 安装SVN命令行工具（确保`svn`命令在系统PATH中）
   - 安装Everything搜索引擎

2. **下载程序**：
   - 从`build`目录获取`svnsearch.exe`（Windows）或`svnsearch`（macOS）

3. **运行程序**：
   - 双击`svnsearch.exe`启动程序
   - 首次运行会自动创建配置文件

## 使用指南

### 1. 添加SVN仓库

1. 点击"添加"按钮
2. 填写仓库信息：
   - **仓库名称**：用于标识仓库的标题（如"项目A"）
   - **SVN地址**：完整的SVN仓库URL（如`http://svn.example.com/repo`）
   - **用户名/密码**：SVN认证信息
   - **扫描路径**：要扫描的路径（如`/trunk`、`/branches`）
3. 点击"保存"按钮

### 2. 扫描仓库

- **扫描选中**：只扫描已启用的仓库
- **扫描全部**：扫描所有仓库
- **停止扫描**：中断正在进行的扫描

### 3. 自动扫描设置

- 勾选"自动扫描"启用定时扫描
- 设置扫描间隔（分钟）

### 4. Everything集成

1. 扫描完成后，EFU文件会生成在`data/efu`目录
2. 打开Everything → 工具 → 选项 → 文件列表
3. 点击"添加"按钮，选择生成的EFU文件
4. 点击"确定"后，Everything会自动索引EFU文件中的内容

### 5. 搜索SVN文件

在Everything搜索框中：
- 搜索仓库标题：`项目A`
- 搜索文件名：`main.py`
- 搜索特定仓库的文件：`项目A main.py`

## 路径格式

EFU文件中的路径格式：

```
SVN://仓库标题/完整SVN URL/文件路径
```

示例：
```
SVN://项目A/http://192.168.1.100/svn/repo/trunk/src/main.py
```

**特点**：
- 右键复制完整路径时，得到的是完整的SVN URL
- 仓库标题清晰可见，方便识别
- 符合URL格式规范

## 配置文件

配置文件位于`configs/config.json`，包含：
- 仓库配置（URL、认证信息等）
- 全局设置（扫描间隔、日志级别等）

## 日志文件

日志文件位于`logs`目录，记录程序运行状态和错误信息。

## 性能指标

- **1000个文件**：约1分钟
- **10000个文件**：约10分钟
- **100000个文件**：约1.5小时

## 常见问题

### SVN命令未找到
- 确保SVN命令行工具已安装
- 确保`svn`命令在系统PATH中

### 扫描失败
- 检查SVN URL是否正确
- 检查用户名和密码是否正确
- 检查网络连接是否正常

### Everything未索引EFU文件
- 确保EFU文件格式正确
- 确保Everything版本支持EFU文件
- 检查Everything的文件列表设置

## 开发说明

### 构建项目

```bash
# 安装依赖
go mod download

# 构建Windows版本
make build

# 构建Linux版本
make build-linux

# 运行开发版本
make run
```

### 项目结构

```
svnsearch/
├── cmd/svnsearch/         # 程序入口
├── internal/              # 内部模块
│   ├── config/            # 配置管理
│   ├── scanner/           # SVN扫描器
│   ├── generator/         # EFU生成器
│   ├── scheduler/         # 定时调度器
│   └── gui/               # GUI界面
├── pkg/                   # 公共包
│   ├── logger/            # 日志工具
│   └── utils/             # 辅助函数
├── configs/               # 配置文件
├── data/efu/              # EFU文件
└── logs/                  # 日志文件
```

## 版本历史

- **v1.0.0**：初始版本
  - 支持多仓库管理
  - 高性能并发扫描
  - EFU文件生成
  - 自动扫描功能
  - 现代化GUI界面
