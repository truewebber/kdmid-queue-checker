package port

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"kdmid-queue-checker/app"
	"kdmid-queue-checker/domain/log"
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

	return mux
}

func (s *HTTPServer) openIndexPage(w http.ResponseWriter, r *http.Request) {
	users, err := s.app.Query.ListUsers.Handle(r.Context())
	if err != nil {
		s.responseError(http.StatusInternalServerError, err, w)

		return
	}

	html := "<!doctype html><html><head><title>kdmid bot artifact viewer</title></head><body>" +
		"<h2>kdmid bot artifact viewer</h2>" +
		"<p>choose which user to browse</p>" +
		"<ul>"

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
			"<li><a href=\"/user/%d\">%d</a> | %s | %s</li>",
			user.TelegramID,
			user.TelegramID,
			active,
			crawls,
		)
	}

	html += "</ul></body></html>"

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
