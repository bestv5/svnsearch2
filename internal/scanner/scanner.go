package scanner

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"svnsearch/internal/config"
	"svnsearch/pkg/logger"
)

type FileInfo struct {
	Filename     string
	Size         int64
	DateModified time.Time
	IsDirectory  bool
}

type ScanResult struct {
	Files    []FileInfo
	Errors   []error
	Duration time.Duration
}

type Scanner struct {
	svnPath    string
	maxWorkers int
	logger     *logger.Logger
}

func NewScanner(svnPath string, maxWorkers int) *Scanner {
	return &Scanner{
		svnPath:    svnPath,
		maxWorkers: maxWorkers,
	}
}

func (s *Scanner) ScanRepository(ctx context.Context, repo *config.Repository) (*ScanResult, error) {
	startTime := time.Now()
	result := &ScanResult{
		Files:  []FileInfo{},
		Errors: []error{},
	}

	jobs := make(chan string, 100)
	results := make(chan []FileInfo, 100)
	errors := make(chan error, 100)

	var wg sync.WaitGroup
	for i := 0; i < s.maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.worker(ctx, repo, jobs, results, errors)
		}()
	}

	go func() {
		for _, path := range repo.ScanPaths {
			fullURL := repo.URL + path
			jobs <- fullURL
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	for {
		select {
		case files, ok := <-results:
			if !ok {
				results = nil
			} else {
				result.Files = append(result.Files, files...)
			}
		case err, ok := <-errors:
			if !ok {
				errors = nil
			} else {
				result.Errors = append(result.Errors, err)
			}
		case <-ctx.Done():
			result.Duration = time.Since(startTime)
			return result, ctx.Err()
		}

		if results == nil && errors == nil {
			break
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

func (s *Scanner) worker(ctx context.Context, repo *config.Repository, jobs <-chan string, results chan<- []FileInfo, errors chan<- error) {
	for url := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			files, err := s.scanDirectory(ctx, url, repo.Username, repo.Password)
			if err != nil {
				errors <- fmt.Errorf("扫描目录 %s 失败: %w", url, err)
			} else {
				results <- files
			}
		}
	}
}

func (s *Scanner) scanDirectory(ctx context.Context, url, username, password string) ([]FileInfo, error) {
	args := []string{"list", "-R", "--xml"}
	if username != "" {
		args = append(args, "--username", username)
	}
	if password != "" {
		args = append(args, "--password", password)
	}
	args = append(args, url)

	output, err := s.execSVNCommand(ctx, args)
	if err != nil {
		return nil, err
	}

	return s.parseSVNListXML(output)
}

func (s *Scanner) execSVNCommand(ctx context.Context, args []string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.svnPath, args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("SVN命令执行失败: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("SVN命令执行失败: %w", err)
	}

	return string(output), nil
}

func (s *Scanner) parseSVNListXML(xmlOutput string) ([]FileInfo, error) {
	var files []FileInfo

	entryRegex := regexp.MustCompile(`<entry[^>]*>(?s:.*?)</entry>`)
	nameRegex := regexp.MustCompile(`<name>([^<]+)</name>`)
	sizeRegex := regexp.MustCompile(`<size>([^<]*)</size>`)
	dateRegex := regexp.MustCompile(`<date>([^<]+)</date>`)
	kindRegex := regexp.MustCompile(`kind="([^"]+)"`)

	entries := entryRegex.FindAllString(xmlOutput, -1)
	for _, entry := range entries {
		nameMatch := nameRegex.FindStringSubmatch(entry)
		if nameMatch == nil {
			continue
		}
		name := nameMatch[1]

		kind := "file"
		if kindMatch := kindRegex.FindStringSubmatch(entry); kindMatch != nil {
			kind = kindMatch[1]
		}

		var size int64 = 0
		if sizeMatch := sizeRegex.FindStringSubmatch(entry); sizeMatch != nil && sizeMatch[1] != "" {
			size, _ = strconv.ParseInt(sizeMatch[1], 10, 64)
		}

		var date time.Time
		if dateMatch := dateRegex.FindStringSubmatch(entry); dateMatch != nil {
			parsedDate, err := time.Parse(time.RFC3339, dateMatch[1])
			if err == nil {
				date = parsedDate
			}
		}

		files = append(files, FileInfo{
			Filename:     name,
			Size:         size,
			DateModified: date,
			IsDirectory:  kind == "dir",
		})
	}

	return files, nil
}

func (s *Scanner) parseSVNListOutput(output string) ([]FileInfo, error) {
	var files []FileInfo

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		file, err := s.parseSVNListLine(line)
		if err != nil {
			continue
		}
		if file != nil {
			files = append(files, *file)
		}
	}

	return files, nil
}

func (s *Scanner) parseSVNListLine(line string) (*FileInfo, error) {
	parts := strings.Fields(line)
	if len(parts) < 1 {
		return nil, fmt.Errorf("无效的行格式")
	}

	filename := parts[len(parts)-1]
	isDir := strings.HasSuffix(filename, "/")
	if isDir {
		filename = strings.TrimSuffix(filename, "/")
	}

	return &FileInfo{
		Filename:    filename,
		IsDirectory: isDir,
	}, nil
}
