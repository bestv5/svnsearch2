#!/bin/bash

# 本地构建脚本 - 模拟GitHub Actions工作流

set -e

echo "========================================="
echo "SVN索引管理器 - 本地构建脚本"
echo "========================================="

# 设置变量
VERSION="1.0.0-local-$(date +%Y%m%d-%H%M%S)"
BUILD_TIME=$(date +"%Y-%m-%d_%H:%M:%S")
PACKAGE_NAME="svnsearch-portable-windows-v$VERSION"

echo ""
echo "构建版本: $VERSION"
echo "构建时间: $BUILD_TIME"
echo ""

# 步骤1: 下载依赖
echo ">>> 步骤1: 下载Go模块依赖..."
go mod download
go mod tidy
echo "✓ 依赖下载完成"

# 步骤2: 编译Windows版本
echo ""
echo ">>> 步骤2: 编译Windows可执行文件..."
GOOS=windows GOARCH=amd64 go build \
    -ldflags="-s -w -H windowsgui -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME" \
    -o build/svnsearch.exe \
    ./cmd/svnsearch
echo "✓ 编译完成"

# 步骤3: 创建便携版目录结构
echo ""
echo ">>> 步骤3: 创建便携版目录结构..."
mkdir -p "dist/$PACKAGE_NAME"/{configs,data/efu,logs}
echo "✓ 目录创建完成"

# 步骤4: 复制文件
echo ""
echo ">>> 步骤4: 复制文件..."
cp build/svnsearch.exe "dist/$PACKAGE_NAME/"
cp configs/config.json "dist/$PACKAGE_NAME/configs/"
cp README.md "dist/$PACKAGE_NAME/"
echo "✓ 文件复制完成"

# 步骤5: 创建启动脚本
echo ""
echo ">>> 步骤5: 创建启动脚本..."
cat > "dist/$PACKAGE_NAME/start.bat" << 'EOF'
@echo off
cd /d "%~dp0"
start "" svnsearch.exe
EOF
echo "✓ 启动脚本创建完成"

# 步骤6: 创建使用说明
echo ""
echo ">>> 步骤6: 创建使用说明..."
cat > "dist/$PACKAGE_NAME/使用说明.txt" << EOF
SVN索引管理器 - 便携版
====================

使用说明：
1. 双击 start.bat 或 svnsearch.exe 启动程序
2. 添加SVN仓库配置
3. 扫描仓库生成EFU文件
4. 在Everything中加载EFU文件

目录说明：
- configs/     : 配置文件目录
- data/efu/    : EFU文件存储目录
- logs/        : 日志文件目录

版本: v$VERSION
构建时间: $(date +"%Y-%m-%d %H:%M:%S")
EOF
echo "✓ 使用说明创建完成"

# 步骤7: 创建ZIP压缩包
echo ""
echo ">>> 步骤7: 创建ZIP压缩包..."
cd dist
zip -r "$PACKAGE_NAME.zip" "$PACKAGE_NAME"
cd ..
echo "✓ ZIP压缩包创建完成"

# 步骤8: 计算校验和
echo ""
echo ">>> 步骤8: 计算SHA256校验和..."
CHECKSUM=$(shasum -a 256 "dist/$PACKAGE_NAME.zip" | cut -d' ' -f1)
echo "$CHECKSUM  $PACKAGE_NAME.zip" > dist/checksum.sha256
echo "✓ 校验和计算完成"
echo "SHA256: $CHECKSUM"

# 完成
echo ""
echo "========================================="
echo "✓ 构建完成！"
echo "========================================="
echo ""
echo "输出文件:"
echo "  - dist/$PACKAGE_NAME.zip"
echo "  - dist/checksum.sha256"
echo ""
echo "便携版目录: dist/$PACKAGE_NAME/"
echo ""
