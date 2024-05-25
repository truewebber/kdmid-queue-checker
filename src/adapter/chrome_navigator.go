package adapter

import (
	"bytes"
	"fmt"
	"log"
	"net/url"

	"github.com/playwright-community/playwright-go"

	"kdmid-queue-checker/domain/image"
	"kdmid-queue-checker/domain/page"
)

type chromeNavigator struct {
	ctx    playwright.BrowserContext
	id, cd string
}

func (c *chromeNavigator) buildURL() *url.URL {
	query := url.Values{}
	query.Set("id", c.id)
	query.Set("cd", c.cd)

	return &url.URL{
		Scheme:   "https",
		Host:     "barcelona.kdmid.ru",
		Path:     "/queue/OrderInfo.aspx",
		RawQuery: query.Encode(),
	}
}

func (c *chromeNavigator) OpenPageToAuthorize() (page.Stat, error) {
	if len(c.ctx.Pages()) != 0 {
		return page.Stat{}, fmt.Errorf("there're pages in context")
	}

	browserPage, err := c.ctx.NewPage()
	if err != nil {
		return page.Stat{}, fmt.Errorf("could not create page: %w", err)
	}

	networkBuffer := bytesBuffer{}

	browserPage.On("request", networkBuffer.onRequest)
	defer browserPage.RemoveListener("request", networkBuffer.onRequest)

	browserPage.On("response", networkBuffer.onResponse)
	defer browserPage.RemoveListener("response", networkBuffer.onResponse)

	response, err := browserPage.Goto(c.buildURL().String(), playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return page.Stat{}, fmt.Errorf("could not goto: %w", err)
	}

	pageHtml, err := browserPage.Content()
	if err != nil {
		return page.Stat{
			Network: networkBuffer.Bytes(),
		}, fmt.Errorf("page content: %w", err)
	}

	pageScreenshot, err := browserPage.Screenshot(playwright.PageScreenshotOptions{
		FullPage: playwright.Bool(true),
	})
	if err != nil {
		return page.Stat{
			Network: networkBuffer.Bytes(),
			HTML:    []byte(pageHtml),
		}, fmt.Errorf("could not take image: %w", err)
	}

	captchaScreenshot, err := c.takeCaptchaScreenshot(browserPage)
	if err != nil {
		return page.Stat{
			Network:    networkBuffer.Bytes(),
			HTML:       []byte(pageHtml),
			Screenshot: pageScreenshot,
		}, fmt.Errorf("could not take captcha screenshot: %w", err)
	}

	stat := page.Stat{
		Network:    networkBuffer.Bytes(),
		HTML:       []byte(pageHtml),
		Screenshot: pageScreenshot,
		Captcha: page.Captcha{
			Presented: true,
			Image:     captchaScreenshot,
		},
	}

	if !response.Ok() {
		return stat, fmt.Errorf("non-ok response: %v", response.Status())
	}

	return stat, nil
}

func (c *chromeNavigator) takeCaptchaScreenshot(browserPage playwright.Page) (image.PNG, error) {
	locator := browserPage.Locator("img")

	n, err := locator.Count()

	if err != nil {
		return nil, fmt.Errorf("count selected images: %w", err)
	}

	if n != 1 {
		return nil, fmt.Errorf("expected 1 image, got %d", n)
	}

	elemScreenshotBytes, err := locator.Screenshot()
	if err != nil {
		return nil, fmt.Errorf("could not take element image: %w", err)
	}

	croppedScreenshot, err := image.Crop(elemScreenshotBytes, func(height, width int) image.CroppingRect {
		return image.CroppingRect{
			X0: width / 3,
			Y0: 0,
			X1: width / 3 * 2,
			Y1: height,
		}
	})
	if err != nil {
		return nil, fmt.Errorf("crop image: %w", err)
	}

	return croppedScreenshot, nil
}

func (c *chromeNavigator) SubmitAuthorization(code string) (page.Stat, error) {
	pagesCount := len(c.ctx.Pages())
	if pagesCount != 1 {
		return page.Stat{}, fmt.Errorf("expected 1 page, got %d", pagesCount)
	}

	browserPage := c.ctx.Pages()[0]

	captchaLocator, err := c.getInputForCaptcha(browserPage)
	if err != nil {
		return page.Stat{}, fmt.Errorf("could not get captcha input: %w", err)
	}

	if err = captchaLocator.Fill(code); err != nil {
		return page.Stat{}, fmt.Errorf("could not fill input field: %w", err)
	}

	submitLocator, err := c.getSubmitButton(browserPage)
	if err != nil {
		return page.Stat{}, fmt.Errorf("could not get submit button: %w", err)
	}

	networkBuffer := bytesBuffer{}

	browserPage.On("request", networkBuffer.onRequest)
	defer browserPage.RemoveListener("request", networkBuffer.onRequest)

	browserPage.On("response", networkBuffer.onResponse)
	defer browserPage.RemoveListener("response", networkBuffer.onResponse)

	if err = submitLocator.Click(); err != nil {
		return page.Stat{
			Network: networkBuffer.Bytes(),
		}, fmt.Errorf("could not click submit button: %w", err)
	}

	err = browserPage.WaitForURL("https://barcelona.kdmid.ru/queue/OrderInfo.aspx*", playwright.PageWaitForURLOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	if err != nil {
		return page.Stat{
			Network: networkBuffer.Bytes(),
		}, fmt.Errorf("navigation failed: %w", err)
	}

	pageHtml, err := browserPage.Content()
	if err != nil {
		return page.Stat{
			Network: networkBuffer.Bytes(),
		}, fmt.Errorf("page content: %w", err)
	}

	screenshot, err := browserPage.Screenshot(playwright.PageScreenshotOptions{
		FullPage: playwright.Bool(true),
	})
	if err != nil {
		return page.Stat{
			Network: networkBuffer.Bytes(),
			HTML:    []byte(pageHtml),
		}, fmt.Errorf("could not take image: %w", err)
	}

	return page.Stat{
		Network:    networkBuffer.Bytes(),
		HTML:       []byte(pageHtml),
		Screenshot: screenshot,
	}, nil
}

func (c *chromeNavigator) getInputForCaptcha(page playwright.Page) (playwright.Locator, error) {
	inputLocator := page.Locator("div.inp > input")

	n, err := inputLocator.Count()
	if err != nil {
		return nil, fmt.Errorf("get count selected inputs: %w", err)
	}

	if n != 3 {
		return nil, fmt.Errorf("expected 3 input, got %d", n)
	}

	return inputLocator.Nth(2), nil
}

func (c *chromeNavigator) getSubmitButton(page playwright.Page) (playwright.Locator, error) {
	inputLocator := page.Locator("input[type=submit]")

	n, err := inputLocator.Count()
	if err != nil {
		return nil, fmt.Errorf("get count selected inputs: %w", err)
	}

	if n != 1 {
		return nil, fmt.Errorf("expected 1 input, got %d", n)
	}

	return inputLocator, nil
}

func (c *chromeNavigator) OpenSlotBookingPage() (page.Stat, error) {
	//TODO implement me
	panic("implement me")
}

func (c *chromeNavigator) Close() error {
	if err := c.ctx.Close(); err != nil {
		return fmt.Errorf("close chrome context: %w", err)
	}

	return nil
}

type bytesBuffer struct {
	bytes.Buffer
}

func (b *bytesBuffer) onRequest(request playwright.Request) {
	logEntry := "Request: " + request.URL() + "\n"
	_, err := b.WriteString(logEntry)
	if err != nil {
		log.Printf("could not write to log file: %v", err)
	}
}

func (b *bytesBuffer) onResponse(response playwright.Response) {
	logEntry := fmt.Sprintf("Response: %v, Status: %v, headers: %v\n", response.URL(), response.Status(), response.Headers())

	if _, err := b.WriteString(logEntry); err != nil {
		log.Printf("could not write to log file: %v", err)
	}
}
