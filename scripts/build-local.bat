@echo off
REM 本地构建脚本 - Windows版本

setlocal enabledelayedexpansion

echo =========================================
echo SVN索引管理器 - 本地构建脚本
echo =========================================

REM 设置变量
for /f "tokens=2 delims==" %%I in ('wmic os get localdatetime /value') do set datetime=%%I
set VERSION=1.0.0-local-%datetime:~0,8%-%datetime:~8,6%
set BUILD_TIME=%date% %time%
set PACKAGE_NAME=svnsearch-portable-windows-v%VERSION%

echo.
echo 构建版本: %VERSION%
echo 构建时间: %BUILD_TIME%
echo.

REM 步骤1: 下载依赖
echo ^>^>^> 步骤1: 下载Go模块依赖...
go mod download
go mod tidy
echo √ 依赖下载完成

REM 步骤2: 编译Windows版本
echo.
echo ^>^>^> 步骤2: 编译Windows可执行文件...
go build -ldflags="-s -w -H windowsgui -X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME%" -o build\svnsearch.exe .\cmd\svnsearch
echo √ 编译完成

REM 步骤3: 创建便携版目录结构
echo.
echo ^>^>^> 步骤3: 创建便携版目录结构...
if not exist "dist\%PACKAGE_NAME%" mkdir "dist\%PACKAGE_NAME%"
if not exist "dist\%PACKAGE_NAME%\configs" mkdir "dist\%PACKAGE_NAME%\configs"
if not exist "dist\%PACKAGE_NAME%\data\efu" mkdir "dist\%PACKAGE_NAME%\data\efu"
if not exist "dist\%PACKAGE_NAME%\logs" mkdir "dist\%PACKAGE_NAME%\logs"
echo √ 目录创建完成

REM 步骤4: 复制文件
echo.
echo ^>^>^> 步骤4: 复制文件...
copy build\svnsearch.exe "dist\%PACKAGE_NAME%\" >nul
copy configs\config.json "dist\%PACKAGE_NAME%\configs\" >nul
copy README.md "dist\%PACKAGE_NAME%\" >nul
echo √ 文件复制完成

REM 步骤5: 创建启动脚本
echo.
echo ^>^>^> 步骤5: 创建启动脚本...
(
echo @echo off
echo cd /d "%%~dp0"
echo start "" svnsearch.exe
) > "dist\%PACKAGE_NAME%\start.bat"
echo √ 启动脚本创建完成

REM 步骤6: 创建使用说明
echo.
echo ^>^>^> 步骤6: 创建使用说明...
(
echo SVN索引管理器 - 便携版
echo ====================
echo.
echo 使用说明：
echo 1. 双击 start.bat 或 svnsearch.exe 启动程序
echo 2. 添加SVN仓库配置
echo 3. 扫描仓库生成EFU文件
echo 4. 在Everything中加载EFU文件
echo.
echo 目录说明：
echo - configs/     : 配置文件目录
echo - data/efu/    : EFU文件存储目录
echo - logs/        : 日志文件目录
echo.
echo 版本: v%VERSION%
echo 构建时间: %date% %time%
) > "dist\%PACKAGE_NAME%\使用说明.txt"
echo √ 使用说明创建完成

REM 步骤7: 创建ZIP压缩包
echo.
echo ^>^>^> 步骤7: 创建ZIP压缩包...
powershell -Command "Compress-Archive -Path 'dist\%PACKAGE_NAME%' -DestinationPath 'dist\%PACKAGE_NAME%.zip' -Force"
echo √ ZIP压缩包创建完成

REM 步骤8: 计算校验和
echo.
echo ^>^>^> 步骤8: 计算SHA256校验和...
for /f "tokens=*" %%I in ('certutil -hashfile "dist\%PACKAGE_NAME%.zip" SHA256 ^| find /v ":" ^| find /v "CertUtil"') do set CHECKSUM=%%I
set CHECKSUM=%CHECKSUM: =%
echo %CHECKSUM%  %PACKAGE_NAME%.zip > dist\checksum.sha256
echo √ 校验和计算完成
echo SHA256: %CHECKSUM%

REM 完成
echo.
echo =========================================
echo √ 构建完成！
echo =========================================
echo.
echo 输出文件:
echo   - dist\%PACKAGE_NAME%.zip
echo   - dist\checksum.sha256
echo.
echo 便携版目录: dist\%PACKAGE_NAME%\
echo.

pause
