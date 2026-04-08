#!/bin/bash

set -e

echo "========================================="
echo "SVN索引管理器 - 本地构建脚本"
echo "========================================="

VERSION="1.0.0-local-$(date +%Y%m%d-%H%M%S)"
BUILD_TIME=$(date +"%Y-%m-%d_%H:%M:%S")
PACKAGE_NAME="svnsearch-v$VERSION"

echo ""
echo "构建版本: $VERSION"
echo "构建时间: $BUILD_TIME"
echo ""

echo ">>> 步骤1: 下载Go模块依赖..."
go mod download
go mod tidy
echo "✓ 依赖下载完成"

echo ""
echo ">>> 步骤2: 编译可执行文件..."
go build \
    -ldflags="-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME" \
    -o build/svnsearch \
    ./cmd/svnsearch
echo "✓ 编译完成"

echo ""
echo ">>> 步骤3: 创建发布目录..."
mkdir -p "dist/$PACKAGE_NAME"/{configs,data/efu,logs}
echo "✓ 目录创建完成"

echo ""
echo ">>> 步骤4: 复制文件..."
cp build/svnsearch "dist/$PACKAGE_NAME/"
cp configs/config.json "dist/$PACKAGE_NAME/configs/"
cp README.md "dist/$PACKAGE_NAME/" 2>/dev/null || true
echo "✓ 文件复制完成"

echo ""
echo ">>> 步骤5: 创建ZIP压缩包..."
cd dist
zip -r "$PACKAGE_NAME.zip" "$PACKAGE_NAME"
cd ..
echo "✓ ZIP压缩包创建完成"

echo ""
echo ">>> 步骤6: 计算SHA256校验和..."
CHECKSUM=$(shasum -a 256 "dist/$PACKAGE_NAME.zip" | cut -d ' ' -f1)
echo "$CHECKSUM  $PACKAGE_NAME.zip" > dist/checksum.sha256
echo "✓ 校验和计算完成"
echo "SHA256: $CHECKSUM"

echo ""
echo "========================================="
echo "✓ 构建完成！"
echo "========================================="
echo ""
echo "输出文件:"
echo "  - dist/$PACKAGE_NAME.zip"
echo "  - dist/checksum.sha256"
echo ""
