package adapter

import (
	"fmt"
	"io"

	"github.com/playwright-community/playwright-go"

	"kdmid-queue-checker/domain/page"
)

type ChromeDispatcher interface {
	io.Closer
	page.Dispatcher
}

type chromeDispatcher struct {
	browser    playwright.Browser
	playwright *playwright.Playwright
}

func NewChromeDispatcher() (ChromeDispatcher, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not run playwright: %w", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("could not launch browser: %w", err)
	}

	return &chromeDispatcher{
		browser:    browser,
		playwright: pw,
	}, nil
}

func MustNewChromeDispatcher() ChromeDispatcher {
	dispatcher, err := NewChromeDispatcher()
	if err != nil {
		panic(err)
	}

	return dispatcher
}

func (c *chromeDispatcher) NewNavigator(id, cd string) (page.Navigator, error) {
	ctx, err := c.browser.NewContext()
	if err != nil {
		return nil, fmt.Errorf("could not create new browser context: %w", err)
	}

	return &chromeNavigator{ctx: ctx, id: id, cd: cd}, nil
}

func (c *chromeDispatcher) Close() error {
	if err := c.browser.Close(); err != nil {
		return fmt.Errorf("could not close browser: %w", err)
	}

	if err := c.playwright.Stop(); err != nil {
		return fmt.Errorf("could not stop playwright: %w", err)
	}

	return nil
}
