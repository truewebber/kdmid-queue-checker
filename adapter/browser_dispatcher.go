package adapter

import (
	"fmt"
	"io"
	"net/url"

	"github.com/playwright-community/playwright-go"

	"kdmid-queue-checker/domain/page"
)

type BrowserDispatcher interface {
	io.Closer
	page.Dispatcher
}

type browserDispatcher struct {
	browser    playwright.Browser
	playwright *playwright.Playwright
}

func NewBrowserDispatcher(proxyURL *url.URL) (BrowserDispatcher, error) {
	proxy, err := configurePlaywrightProxy(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("configure playwright proxy: %w", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("run playwright: %w", err)
	}

	browser, err := pw.WebKit.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
		Proxy:    proxy,
	})
	if err != nil {
		return nil, fmt.Errorf("launch browser: %w", err)
	}

	return &browserDispatcher{
		browser:    browser,
		playwright: pw,
	}, nil
}

func configurePlaywrightProxy(proxyURL *url.URL) (*playwright.Proxy, error) {
	if proxyURL == nil {
		return nil, fmt.Errorf("no proxy URL provided")
	}

	const httpScheme = "http"

	if proxyURL.Scheme != httpScheme {
		return nil, fmt.Errorf("unsupported scheme `%s`", proxyURL.Scheme)
	}

	var (
		username *string
		password *string
	)

	if proxyURL.User != nil {
		u := proxyURL.User.Username()
		username = &u

		p, exists := proxyURL.User.Password()
		if exists {
			password = &p
		}
	}

	return &playwright.Proxy{
		Server:   proxyURL.Host,
		Bypass:   nil,
		Username: username,
		Password: password,
	}, nil
}

func MustNewBrowserDispatcher(proxyURL *url.URL) BrowserDispatcher {
	dispatcher, err := NewBrowserDispatcher(proxyURL)
	if err != nil {
		panic(err)
	}

	return dispatcher
}

func (c *browserDispatcher) NewNavigator(id, cd string) (page.Navigator, error) {
	ctx, err := c.browser.NewContext()
	if err != nil {
		return nil, fmt.Errorf("could not create new browser context: %w", err)
	}

	return &browserNavigator{ctx: ctx, id: id, cd: cd}, nil
}

func (c *browserDispatcher) Close() error {
	if err := c.browser.Close(); err != nil {
		return fmt.Errorf("could not close browser: %w", err)
	}

	if err := c.playwright.Stop(); err != nil {
		return fmt.Errorf("could not stop playwright: %w", err)
	}

	return nil
}
