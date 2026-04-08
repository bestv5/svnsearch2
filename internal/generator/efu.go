package generator

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"svnsearch/internal/scanner"
)

type EFUGenerator struct {
	outputPath string
}

func NewEFUGenerator(outputPath string) *EFUGenerator {
	return &EFUGenerator{
		outputPath: outputPath,
	}
}

func (g *EFUGenerator) Generate(files []scanner.FileInfo, repoName, repoURL string) error {
	dir := filepath.Dir(g.outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	file, err := os.Create(g.outputPath)
	if err != nil {
		return fmt.Errorf("创建EFU文件失败: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := g.writeHeader(writer); err != nil {
		return fmt.Errorf("写入EFU文件头失败: %w", err)
	}

	for _, f := range files {
		if err := g.writeFileEntry(writer, f, repoName, repoURL); err != nil {
			return fmt.Errorf("写入文件条目失败: %w", err)
		}
	}

	return nil
}

func (g *EFUGenerator) writeHeader(writer *csv.Writer) error {
	header := []string{"Filename", "Size", "Date Modified", "Date Created", "Attributes"}
	return writer.Write(header)
}

func (g *EFUGenerator) writeFileEntry(writer *csv.Writer, file scanner.FileInfo, repoName, repoURL string) error {
	cleanURL := repoURL
	cleanURL = strings.TrimPrefix(cleanURL, "svn://")
	cleanURL = strings.TrimPrefix(cleanURL, "http://")
	cleanURL = strings.TrimPrefix(cleanURL, "https://")
	cleanURL = strings.ReplaceAll(cleanURL, "/", "\\")
	
	fullPath := fmt.Sprintf("SVN:\\%s\\%s\\%s", repoName, cleanURL, strings.ReplaceAll(file.Filename, "/", "\\"))

	record := []string{
		fullPath,
		fmt.Sprintf("%d", file.Size),
		fmt.Sprintf("%d", unixToFileTime(file.DateModified)),
		fmt.Sprintf("%d", unixToFileTime(file.DateModified)),
		fmt.Sprintf("%d", calculateAttributes(file.IsDirectory)),
	}

	return writer.Write(record)
}

func unixToFileTime(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return (t.Unix() + 11644473600) * 10000000
}

func calculateAttributes(isDirectory bool) int {
	if isDirectory {
		return 16
	}
	return 32
}
