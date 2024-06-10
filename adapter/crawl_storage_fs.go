package adapter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"kdmid-queue-checker/domain/crawl"
	"kdmid-queue-checker/domain/log"
	"kdmid-queue-checker/domain/page"
)

type fileSystemCrawlStorage struct {
	dir    string
	logger log.Logger
}

const (
	decimal = 10
	bitSize = 64
)

func NewFileSystemCrawlStorage(dir string, logger log.Logger) (crawl.Storage, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create home directory: %w", err)
	}

	return &fileSystemCrawlStorage{
		logger: logger,
		dir:    dir,
	}, nil
}

func MustNewFileSystemCrawlStorage(dir string, logger log.Logger) crawl.Storage {
	storage, err := NewFileSystemCrawlStorage(dir, logger)
	if err != nil {
		panic(err)
	}

	return storage
}

func (f *fileSystemCrawlStorage) Save(_ context.Context, userID int64, result *crawl.Result) error {
	userDirName := strconv.FormatInt(userID, decimal)
	dateDirName := result.RanAt.Format(time.DateOnly)
	timeDirName := result.RanAt.Format(time.TimeOnly)

	crawlDir := filepath.Join(f.dir, userDirName, dateDirName, timeDirName)

	firstDir := filepath.Join(crawlDir, "1")
	if err := f.saveStat(firstDir, result.One); err != nil {
		return fmt.Errorf("save first stat: %w", err)
	}

	twoDir := filepath.Join(crawlDir, "2")
	if err := f.saveStat(twoDir, result.Two); err != nil {
		return fmt.Errorf("save second stat: %w", err)
	}

	threeDir := filepath.Join(crawlDir, "3")
	if err := f.saveStat(threeDir, result.Three); err != nil {
		return fmt.Errorf("save third stat: %w", err)
	}

	if result.SomethingInteresting {
		interestingFile := filepath.Join(crawlDir, "interesting.txt")
		if err := f.saveFile(interestingFile, []byte{}); err != nil {
			return fmt.Errorf("save interesting file: %w", err)
		}
	}

	return nil
}

func (f *fileSystemCrawlStorage) saveStat(dir string, stat page.Stat) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	htmlFile := filepath.Join(dir, "page.html")
	if err := f.saveFile(htmlFile, stat.HTML); err != nil {
		return fmt.Errorf("save html file: %w", err)
	}

	networkFile := filepath.Join(dir, "network.txt")
	if err := f.saveFile(networkFile, stat.Network); err != nil {
		return fmt.Errorf("save network file: %w", err)
	}

	screenshotFile := filepath.Join(dir, "screenshot.png")
	if err := f.saveFile(screenshotFile, stat.Screenshot); err != nil {
		return fmt.Errorf("save screenshot file: %w", err)
	}

	if stat.Captcha.Presented {
		captchaFile := filepath.Join(dir, "captcha.png")
		if err := f.saveFile(captchaFile, stat.Captcha.Image); err != nil {
			return fmt.Errorf("save captcha file: %w", err)
		}
	}

	return nil
}

func (f *fileSystemCrawlStorage) saveFile(filePath string, fileBytes []byte) error {
	fd, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	defer f.logger.CloseWithLog(fd)

	if _, err := fd.Write(fileBytes); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func (f *fileSystemCrawlStorage) ListUsers(_ context.Context) ([]int64, error) {
	entries, osErr := os.ReadDir(f.dir)
	if osErr != nil {
		return nil, fmt.Errorf("read directory: %w", osErr)
	}

	userIDs := make([]int64, 0, len(entries))

	for i := range entries {
		if !entries[i].IsDir() {
			continue
		}

		dirName := entries[i].Name()
		userID, err := strconv.ParseInt(dirName, decimal, bitSize)
		if err != nil {
			return nil, fmt.Errorf("parse user id - `%v`: %w", dirName, err)
		}

		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func (f *fileSystemCrawlStorage) ListResults(
	_ context.Context, userID int64, date time.Time,
) ([]crawl.Result, error) {
	//TODO implement me
	panic("implement me")
}
