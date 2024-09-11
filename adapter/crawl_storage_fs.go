package adapter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/truewebber/gopkg/log"

	"kdmid-queue-checker/domain/crawl"
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

	if result.Err != nil {
		errFile := filepath.Join(crawlDir, "error.txt")
		if err := f.saveFile(errFile, []byte(result.Err.Error())); err != nil {
			return fmt.Errorf("save error file: %w", err)
		}
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

	defer func() {
		if err := fd.Close(); err != nil {
			f.logger.Error("failed close", "error", err.Error())
		}
	}()

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
	ctx context.Context, userID int64, date time.Time,
) ([]crawl.Result, error) {
	userDirName := strconv.FormatInt(userID, decimal)
	dateDirName := date.Format(time.DateOnly)

	dailyCrawlsDir := filepath.Join(f.dir, userDirName, dateDirName)

	crawlTimes, osErr := os.ReadDir(dailyCrawlsDir)
	if osErr != nil {
		return nil, fmt.Errorf("read directory: %w", osErr)
	}

	crawlResults := make([]crawl.Result, 0, len(crawlTimes))

	for i := range crawlTimes {
		if !crawlTimes[i].IsDir() {
			continue
		}

		crawlResultDir := filepath.Join(dailyCrawlsDir, crawlTimes[i].Name())

		crawlResult, err := f.readCrawl(ctx, crawlResultDir)
		if err != nil {
			return nil, fmt.Errorf("read crawl: %w", err)
		}

		crawlResult.RanAt, err = time.Parse(time.DateTime, dateDirName+" "+crawlTimes[i].Name())
		if err != nil {
			return nil, fmt.Errorf("time parse: %w", err)
		}

		crawlResults = append(crawlResults, crawlResult)
	}

	return crawlResults, nil
}

func (f *fileSystemCrawlStorage) readCrawl(ctx context.Context, crawlDir string) (crawl.Result, error) {
	var (
		result = crawl.Result{}
		err    error
	)

	firstDir := filepath.Join(crawlDir, "1")
	result.One, err = f.readStat(ctx, firstDir)
	if err != nil {
		return crawl.Result{}, fmt.Errorf("read first stat: %w", err)
	}

	twoDir := filepath.Join(crawlDir, "2")
	result.Two, err = f.readStat(ctx, twoDir)
	if err != nil {
		return crawl.Result{}, fmt.Errorf("read second stat: %w", err)
	}

	threeDir := filepath.Join(crawlDir, "3")
	result.Three, err = f.readStat(ctx, threeDir)
	if err != nil {
		return crawl.Result{}, fmt.Errorf("save third stat: %w", err)
	}

	errorFile := filepath.Join(crawlDir, "error.txt")
	errText, err := f.readFile(ctx, errorFile)
	if err != nil {
		return crawl.Result{}, fmt.Errorf("read error file: %w", err)
	}

	if len(errText) > 0 {
		result.Err = fmt.Errorf(string(errText))
	}

	interestingFile := filepath.Join(crawlDir, "interesting.txt")
	result.SomethingInteresting, err = f.fileExists(ctx, interestingFile)
	if err != nil {
		return crawl.Result{}, fmt.Errorf("check interesting file exists: %w", err)
	}

	return result, nil
}

func (f *fileSystemCrawlStorage) readStat(ctx context.Context, statDir string) (page.Stat, error) {
	var (
		stat = page.Stat{}
		err  error
	)

	htmlFile := filepath.Join(statDir, "page.html")
	stat.HTML, err = f.readFile(ctx, htmlFile)
	if err != nil {
		return page.Stat{}, fmt.Errorf("read html file: %w", err)
	}

	networkFile := filepath.Join(statDir, "network.txt")
	stat.Network, err = f.readFile(ctx, networkFile)
	if err != nil {
		return page.Stat{}, fmt.Errorf("read network file: %w", err)
	}

	screenshotFile := filepath.Join(statDir, "screenshot.png")
	stat.Screenshot, err = f.readFile(ctx, screenshotFile)
	if err != nil {
		return page.Stat{}, fmt.Errorf("read screenshot file: %w", err)
	}

	captchaFile := filepath.Join(statDir, "captcha.png")
	stat.Captcha.Image, err = f.readFile(ctx, captchaFile)
	if err != nil {
		return page.Stat{}, fmt.Errorf("read screenshot file: %w", err)
	}

	stat.Captcha.Presented = stat.Captcha.Image != nil

	return stat, nil
}

func (f *fileSystemCrawlStorage) readFile(_ context.Context, filePath string) ([]byte, error) {
	fd, err := os.Open(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return []byte{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	defer func() {
		if err := fd.Close(); err != nil {
			f.logger.Error("failed close", "error", err.Error())
		}
	}()

	fileBytes, err := io.ReadAll(fd)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return fileBytes, nil
}

func (f *fileSystemCrawlStorage) fileExists(_ context.Context, filePath string) (bool, error) {
	_, err := os.Stat(filePath)

	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("stat file: %w", err)
	}

	return true, nil
}
