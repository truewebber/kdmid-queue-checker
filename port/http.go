package port

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/truewebber/gopkg/log"

	"github.com/truewebber/kdmid-queue-checker/app"
)

type HTTPServer struct {
	server *http.Server
	app    *app.Application
	logger log.Logger
}

func NewHTTP(hostPort string, app *app.Application, logger log.Logger) *HTTPServer {
	httpServer := &HTTPServer{
		app:    app,
		logger: logger,
	}
	router := httpServer.configureRouter()

	httpServer.server = &http.Server{
		Addr:    hostPort,
		Handler: router,
	}

	return httpServer
}

func (s *HTTPServer) configureRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.openIndexPage)
	mux.HandleFunc("/user/{userID}/{date}", s.openCrawlListPage)

	return mux
}

const head = "<head><title>kdmid bot artifact viewer</title>" +
	"<style>" +
	"h2 {padding: 25px 0}" +
	"* {margin:0;padding:0;}" +
	".main {width: 80%; margin:0 auto; background-color: white;}" +
	".crawls_block {}" +
	".crawl { margin: 15px 0; padding: 15px; border-radius: 3px; }" +
	".crawl_general { background-color: #99ccff; }" +
	".crawl_error { background-color: #ff9999; }" +
	".crawl_interesting { background-color: #99ffcc; }" +
	".user { padding: 10px 0; }" +
	".screenshot { width: 300px; padding: 0 5px; }" +
	".extended { width: 1000px; }" +
	".captcha { width: 100px; padding: 0 5px; }" +
	".hr { margin: 0 0 10px 0; border-bottom: 2px solid #555 }" +
	"a { color: blue; text-decoration: none; }" +
	"a:visited { color: blue; }" +
	"a:hover { color: orange; }" +
	"</style>" +
	"</head>"

func (s *HTTPServer) openIndexPage(w http.ResponseWriter, r *http.Request) {
	users, err := s.app.Query.ListUsers.Handle(r.Context())
	if err != nil {
		s.responseError(http.StatusInternalServerError, err, w)

		return
	}

	date := time.Now()

	html := "<!doctype html><html>" + head + "<body>" +
		"<div class=\"main\">" +
		"<h2>kdmid bot artifact viewer</h2>" +
		"<p style=\"text-decoration: underline\">choose which user to browse</p>"

	for _, user := range users {
		active := "deactivated"
		if user.Active {
			active = "active"
		}

		crawls := "no crawls yet"
		if user.HasCrawls {
			crawls = "has crawls"
		}

		html += fmt.Sprintf(
			"<p class=\"user\"><a href=\"/user/%d/%s\">%d</a> | %s | %s</p>",
			user.TelegramID,
			date.Format(time.DateOnly),
			user.TelegramID,
			active,
			crawls,
		)
	}

	html += "</div></body></html>"

	s.responseHTML(html, w)
}

const (
	decimal   = 10
	sixtyFour = 64
)

func (s *HTTPServer) openCrawlListPage(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.PathValue("userID")
	userID, err := strconv.ParseInt(userIDVal, decimal, sixtyFour)
	if err != nil {
		s.responseError(http.StatusBadRequest, err, w)

		return
	}

	dateVal := r.PathValue("date")
	date, err := time.Parse(time.DateOnly, dateVal)
	if err != nil {
		s.responseError(http.StatusBadRequest, err, w)

		return
	}

	crawls, err := s.app.Query.ListCrawls.Handle(r.Context(), userID, date)
	if err != nil {
		s.responseError(http.StatusInternalServerError, err, w)

		return
	}

	pastVal := date.AddDate(0, 0, -1).Format(time.DateOnly)
	futureVal := date.AddDate(0, 0, 1).Format(time.DateOnly)

	html := "<!doctype html><html>" + head +
		"<body>" +
		"<div class=\"main\">" +
		"<h2>kdmid bot artifact viewer</h2>" +
		"<p>User \"" + userIDVal + "\" | Date: " + dateVal +
		" | <a href=\"/user/" + userIDVal + "/" + pastVal + "\">Past</a>" +
		" | <a href=\"/user/" + userIDVal + "/" + futureVal + "\">Future</a></p>" +
		"<p><a href=\"/\">Back</a></p>" +
		"<div class=\"crawls_block\">"

	sort.SliceStable(crawls, func(i, j int) bool {
		return crawls[i].CrawledAt.After(crawls[j].CrawledAt)
	})

	for _, c := range crawls {
		class := "crawl_general"
		text := ""

		switch {
		case c.Err != nil:
			class = "crawl_error"
			text = c.Err.Error()
		case c.SomethingInteresting:
			class = "crawl_interesting"
			text = "Success?"
		}

		html += "<div class=\"crawl " + class + "\">" +
			"<p>" + c.CrawledAt.Format(time.TimeOnly) + text + "</p>" +
			"<p class=\"hr\"></p>"

		for i := range c.Screenshots {
			html += "<img class=\"screenshot\" src=\"data:image/png;base64," +
				base64.StdEncoding.EncodeToString(c.Screenshots[i]) + "\">"
		}

		html += "<img class=\"captcha\" src=\"data:image/png;base64," +
			base64.StdEncoding.EncodeToString(c.Captch) + "\">"

		html += "</div>"
	}

	html += "</div></div>" +
		"<script type=\"text/javascript\">" +
		"var screenshots = document.querySelectorAll('.screenshot');" +
		"for (i = 0; i < screenshots.length; i++) {" +
		"  screenshots[i].addEventListener('click', function() {" +
		"    if (this.classList.contains(\"extended\")) {" +
		"      this.classList.remove(\"extended\");" +
		"" +
		"      return;" +
		"    }" +
		"" +
		"    shrinkEveryScreenshot();" +
		"    this.classList.add(\"extended\");" +
		"  });" +
		"}" +
		"" +
		"function shrinkEveryScreenshot() {" +
		"  for (i = 0; i < screenshots.length; i++) {" +
		"    screenshots[i].classList.remove(\"extended\");" +
		"  }" +
		"}" +
		"</script>" +
		"</body></html>"

	s.responseHTML(html, w)
}

func (s *HTTPServer) responseHTML(html string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")

	if _, err := w.Write([]byte(html)); err != nil {
		s.logger.Error("failed to write html", "error", err.Error())
	}
}

func (s *HTTPServer) responseError(code int, err error, w http.ResponseWriter) {
	w.WriteHeader(code)

	s.logger.Error("got error", "code", code, "err", err)
}

func (s *HTTPServer) Start(ctx context.Context) error {
	serverErr := make(chan error)
	defer close(serverErr)

	go func() {
		err := s.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server error: %w", err)
	case <-ctx.Done():
		if err := s.server.Close(); err != nil {
			return fmt.Errorf("http server close error: %w", err)
		}

		return nil
	}
}
