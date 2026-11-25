package httpapp

import (
	"embed"
	"html/template"
	"log/slog"
	"net/http"
	"time"
)

//go:embed templates/*.html static/*
var assets embed.FS

type app struct {
	logger    *slog.Logger
	templates *template.Template
}

type pageData struct {
	Title       string
	Environment string
}

func New(logger *slog.Logger) http.Handler {
	templates := template.Must(template.ParseFS(assets, "templates/*.html"))
	a := &app{logger: logger, templates: templates}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", a.live)
	mux.HandleFunc("GET /", a.home)
	mux.Handle("GET /static/", http.FileServerFS(assets))
	return requestLog(logger, securityHeaders(mux))
}

func (a *app) live(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (a *app) home(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.templates.ExecuteTemplate(w, "home", pageData{Title: "TT WMS", Environment: "Foundation"}); err != nil {
		a.logger.Error("render home", "error", err)
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

func requestLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("http request", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(started).Milliseconds())
	})
}
