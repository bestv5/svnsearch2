# 本地开发与调试指南

## 目录结构

```
svnsearch/
├── .github/workflows/      # GitHub Actions工作流
│   └── build.yml           # Windows便携版构建
├── scripts/                # 本地构建脚本
│   ├── build-local.sh      # macOS/Linux构建脚本
│   └── build-local.bat     # Windows构建脚本
├── cmd/svnsearch/          # 程序入口
├── internal/               # 内部模块
├── pkg/                    # 公共包
├── configs/                # 配置文件
├── build/                  # 编译输出
└── dist/                   # 发布包输出
```

## 本地调试方法

### 方法1: 直接运行（开发模式）

```bash
# macOS/Linux
go run ./cmd/svnsearch

# Windows
go run .\cmd\svnsearch
```

### 方法2: 本地构建便携版

#### macOS/Linux

```bash
# 运行构建脚本
./scripts/build-local.sh

# 或者手动构建
go mod download
go mod tidy
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H windowsgui" -o build/svnsearch.exe ./cmd/svnsearch
```

#### Windows

```cmd
# 运行构建脚本
scripts\build-local.bat

# 或者手动构建
go mod download
go mod tidy
go build -ldflags="-s -w -H windowsgui" -o build\svnsearch.exe .\cmd\svnsearch
```

### 方法3: 使用 act 工具（模拟GitHub Actions）

#### 安装 act

**macOS**：
```bash
brew install act
```

**Windows**：
```bash
choco install act-cli
# 或者使用scoop
scoop install act
```

**Linux**：
```bash
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
```

#### 使用 act 运行工作流

```bash
# 列出所有工作流
act -l

# 运行Windows构建工作流
act -j build-windows

# 使用特定平台镜像
act -j build-windows -P windows-latest=node:16

# 详细输出模式
act -j build-windows -v

# 干运行（不实际执行）
act -j build-windows -n

# 触发tag推送事件
act push --tag v1.0.0
```

#### act 常用参数

| 参数 | 说明 |
|------|------|
| `-l` | 列出所有工作流 |
| `-j <job>` | 运行特定job |
| `-n` | 干运行模式 |
| `-v` | 详细输出 |
| `-P <platform>` | 指定平台镜像 |
| `--secret <key>=<value>` | 设置密钥 |
| `--env <key>=<value>` | 设置环境变量 |

## 开发流程

### 1. 开发新功能

```bash
# 1. 创建新分支
git checkout -b feature/new-feature

# 2. 开发和测试
go run ./cmd/svnsearch

# 3. 运行测试
go test ./...

# 4. 本地构建测试
./scripts/build-local.sh

# 5. 提交代码
git add .
git commit -m "Add new feature"
git push origin feature/new-feature

# 6. 创建Pull Request
# GitHub Actions会自动运行构建
```

### 2. 发布新版本

```bash
# 1. 更新版本号
# 编辑 Makefile 中的 VERSION

# 2. 提交更改
git add .
git commit -m "Bump version to v1.0.1"

# 3. 创建tag
git tag v1.0.1
git push origin v1.0.1

# 4. GitHub Actions会自动构建并创建Release
```

## 调试技巧

### 1. 调试Go代码

```bash
# 使用Delve调试器
go install github.com/go-delve/delve/cmd/dlv@latest

# 启动调试
dlv debug ./cmd/svnsearch

# 设置断点
(dlv) break main.main
(dlv) continue
```

### 2. 调试GUI界面

```bash
# 设置FYNE_DEBUG环境变量
export FYNE_DEBUG=1
go run ./cmd/svnsearch
```

### 3. 查看构建详情

```bash
# 显示编译详情
go build -v -x ./cmd/svnsearch

# 查看依赖关系
go mod graph

# 查看为什么需要某个依赖
go mod why fyne.io/fyne/v2
```

### 4. 性能分析

```bash
# 启用性能分析
go run -race ./cmd/svnsearch

# 生成CPU profile
go test -cpuprofile=cpu.prof -bench .
go tool pprof cpu.prof

# 生成内存profile
go test -memprofile=mem.prof -bench .
go tool pprof mem.prof
```

## 常见问题

### Q1: 编译时提示找不到包

```bash
# 清理缓存并重新下载
go clean -modcache
go mod download
go mod tidy
```

### Q2: Windows编译失败

```bash
# 确保安装了MinGW
choco install mingw -y

# 或者安装TDM-GCC
# https://jmeubank.github.io/tdm-gcc/
```

### Q3: act运行失败

```bash
# 使用更小的Docker镜像
act -j build-windows -P windows-latest=catthehacker/ubuntu:act-latest

# 或者使用本地shell
act -j build-windows --container off
```

### Q4: GUI无法显示

```bash
# macOS: 确保XQuartz已安装
brew install --cask xquartz

# Linux: 确保X11开发库已安装
sudo apt-get install libgl1-mesa-dev xorg-dev
```

## 性能优化

### 1. 编译优化

```bash
# 减小二进制文件大小
go build -ldflags="-s -w" -trimpath -o svnsearch ./cmd/svnsearch

# 使用UPX进一步压缩
upx --best svnsearch
```

### 2. 构建缓存

```bash
# 使用Go构建缓存
export GOCACHE=~/.cache/go-build

# 使用模块缓存
export GOMODCACHE=~/.go/pkg/mod
```

### 3. 并行编译

```bash
# 设置并行编译数
export GOMAXPROCS=4
go build ./cmd/svnsearch
```

## 测试

### 运行所有测试

```bash
go test -v ./...
```

### 运行特定测试

```bash
# 运行特定包的测试
go test -v ./internal/scanner

# 运行特定测试函数
go test -v -run TestScanRepository ./internal/scanner
```

### 测试覆盖率

```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out ./...

# 查看覆盖率
go tool cover -func=coverage.out

# 在浏览器中查看
go tool cover -html=coverage.out
```

## 代码质量

### 代码格式化

```bash
# 格式化代码
go fmt ./...

# 使用goimports
go install golang.org/x/tools/cmd/goimports@latest
goimports -w .
```

### 代码检查

```bash
# 使用go vet
go vet ./...

# 使用golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run
```

### 静态分析

```bash
# 使用staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

## 持续集成

### 本地CI测试

```bash
# 使用act模拟CI
act pull_request

# 测试特定事件
act push --tag v1.0.0
```

### 查看CI日志

```bash
# 详细模式
act -v

# 保存日志到文件
act -j build-windows 2>&1 | tee build.log
```

## 发布检查清单

- [ ] 更新版本号（Makefile）
- [ ] 更新CHANGELOG.md
- [ ] 运行所有测试
- [ ] 本地构建测试
- [ ] 代码格式化
- [ ] 代码检查（go vet, golangci-lint）
- [ ] 更新README.md
- [ ] 创建git tag
- [ ] 推送tag到GitHub
- [ ] 等待GitHub Actions完成
- [ ] 检查Release页面
- [ ] 下载并测试便携版
