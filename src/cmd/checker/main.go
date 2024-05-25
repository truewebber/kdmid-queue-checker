package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"kdmid-queue-checker/adapter"
	"kdmid-queue-checker/domain/captcha"
	"kdmid-queue-checker/domain/page"
)

func main() {
	cfg := mustLoadConfig()

	dispatcher := adapter.MustNewChromeDispatcher()
	solver := adapter.NewTwoCaptchaSolver(cfg.TwoCaptcha.APIKey)

	if err := run(cfg.Application.ID, cfg.Application.Secret, dispatcher, solver); err != nil {
		panic(err)
	}

	if err := dispatcher.Close(); err != nil {
		panic(err)
	}
}

func run(id, cd string, dispatcher page.Dispatcher, solver captcha.Solver) error {
	navigator, err := dispatcher.NewNavigator(id, cd)
	if err != nil {
		return fmt.Errorf("new navigator: %w", err)
	}

	stat, err := navigator.OpenPageToAuthorize()
	if err != nil {
		return fmt.Errorf("open slot booking page: %w", err)
	}

	fmt.Printf("%v\n", stat)

	return nil
}

func contextClosableOnSignals(sig ...os.Signal) context.Context {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, sig...)

	ctx, cancel := context.WithCancel(context.Background())

	go func(signals <-chan os.Signal) {
		<-signals

		cancel()
	}(signals)

	return ctx
}

//func openPage(browser playwright.Browser) error {
//	ctx, err := browser.NewContext()
//	if err != nil {
//		return fmt.Errorf("could not create context: %w", err)
//	}
//
//	defer ctx.Close()
//
//	page, err := ctx.NewPage()
//	if err != nil {
//		return fmt.Errorf("could not create page: %w", err)
//	}
//
//	// Create or open a file to save network logs
//	networkFile, err := os.Create("network_log.txt")
//	if err != nil {
//		return fmt.Errorf("could not create log file: %w", err)
//	}
//	defer networkFile.Close()
//
//	// Listen to request and response events
//	page.OnRequest(func(request playwright.Request) {
//		logEntry := "Request: " + request.URL() + "\n"
//		_, err := networkFile.WriteString(logEntry)
//		if err != nil {
//			log.Printf("could not write to log file: %v", err)
//		}
//	})
//
//	page.OnResponse(func(response playwright.Response) {
//		logEntry := fmt.Sprintf("Response: %v, Status: %v, headers: %v\n", response.URL(), response.Status(), response.Headers())
//
//		_, err = networkFile.WriteString(logEntry)
//		if err != nil {
//			log.Printf("could not write to log file: %v", err)
//		}
//	})
//
//	println("ready to request")
//
//	response, err := page.Goto(url, playwright.PageGotoOptions{
//		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
//	})
//	if err != nil {
//		return fmt.Errorf("could not goto: %w", err)
//	}
//
//	println("loaded")
//
//	if !response.Ok() {
//		return fmt.Errorf("non-ok response: %v", response.Status())
//	}
//
//	println("ok")
//
//	_, err = page.Screenshot(playwright.PageScreenshotOptions{
//		Path:     playwright.String("example.png"),
//		FullPage: playwright.Bool(true),
//	})
//	if err != nil {
//		return fmt.Errorf("could not take image: %w", err)
//	}
//
//	locator := page.Locator("img")
//
//	n, err := locator.Count()
//
//	if err != nil {
//		return fmt.Errorf("count selected images: %w", err)
//	}
//
//	if n != 1 {
//		return fmt.Errorf("expected 1 image, got %d", n)
//	}
//
//	elemScreenshotBytes, err := locator.Screenshot()
//	if err != nil {
//		return fmt.Errorf("could not take element image: %w", err)
//	}
//
//	croppedScreenshot, err := image.Crop(elemScreenshotBytes, func(height, width int) image.CroppingRect {
//		return image.CroppingRect{
//			X0: width / 3,
//			Y0: 0,
//			X1: width / 3 * 2,
//			Y1: height,
//		}
//	})
//	if err != nil {
//		return fmt.Errorf("crop image: %w", err)
//	}
//
//	elemScreenshotFile, err := os.Create("element.png")
//	if err != nil {
//		return fmt.Errorf("could not create image file: %w", err)
//	}
//	defer elemScreenshotFile.Close()
//
//	if _, err := elemScreenshotFile.Write(croppedScreenshot); err != nil {
//		return fmt.Errorf("write file: %w", err)
//	}
//
//	println("image done")
//
//	htmlPageFile, err := os.Create("page.html")
//	if err != nil {
//		return fmt.Errorf("could not create log file: %w", err)
//	}
//	defer htmlPageFile.Close()
//
//	content, err := page.Content()
//	if err != nil {
//		return fmt.Errorf("page content: %w", err)
//	}
//
//	if _, err := htmlPageFile.WriteString(content); err != nil {
//		return fmt.Errorf("write string: %w", err)
//	}
//
//	code, err := solveCapcha(croppedScreenshot)
//	if err != nil {
//		return fmt.Errorf("solve capcha: %w", err)
//	}
//
//	println("capcha", code)
//
//	capchaLocator, err := getInputForCapcha(page)
//	if err != nil {
//		return fmt.Errorf("could not get capcha input: %w", err)
//	}
//
//	fmt.Printf("%#v\n", capchaLocator)
//
//	if err = capchaLocator.Fill(code); err != nil {
//		return fmt.Errorf("could not fill input field: %w", err)
//	}
//
//	submitLocator, err := getSubmitButton(page)
//	if err != nil {
//		return fmt.Errorf("could not get submit button: %w", err)
//	}
//
//	fmt.Printf("%v\n", submitLocator)
//
//	if err = submitLocator.Click(); err != nil {
//		return fmt.Errorf("could not click submit button: %w", err)
//	}
//
//	err = page.WaitForURL("https://barcelona.kdmid.ru/queue/OrderInfo.aspx*", playwright.PageWaitForURLOptions{
//		WaitUntil: playwright.WaitUntilStateLoad,
//	})
//	if err != nil {
//		return fmt.Errorf("navigation failed: %w", err)
//	}
//
//	_, err = page.Screenshot(playwright.PageScreenshotOptions{
//		Path:     playwright.String("example2.png"),
//		FullPage: playwright.Bool(true),
//	})
//	if err != nil {
//		return fmt.Errorf("could not take image: %w", err)
//	}
//
//	htmlPage2File, err := os.Create("page2.html")
//	if err != nil {
//		return fmt.Errorf("could not create log file: %w", err)
//	}
//	defer htmlPageFile.Close()
//
//	content2, err := page.Content()
//	if err != nil {
//		return fmt.Errorf("page content: %w", err)
//	}
//
//	if _, err := htmlPage2File.WriteString(content2); err != nil {
//		return fmt.Errorf("write string: %w", err)
//	}
//
//	return nil
//}
//
//func getInputForCapcha(page playwright.Page) (playwright.Locator, error) {
//	inputLocator := page.Locator("div.inp > input")
//
//	n, err := inputLocator.Count()
//	if err != nil {
//		return nil, fmt.Errorf("get count selected inputs: %w", err)
//	}
//
//	if n != 3 {
//		return nil, fmt.Errorf("expected 3 input, got %d", n)
//	}
//
//	return inputLocator.Nth(2), nil
//}
//
//func getSubmitButton(page playwright.Page) (playwright.Locator, error) {
//	inputLocator := page.Locator("input[type=submit]")
//
//	n, err := inputLocator.Count()
//	if err != nil {
//		return nil, fmt.Errorf("get count selected inputs: %w", err)
//	}
//
//	if n != 1 {
//		return nil, fmt.Errorf("expected 1 input, got %d", n)
//	}
//
//	return inputLocator, nil
//}
//
