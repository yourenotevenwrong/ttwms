package httpapp

import (
	"bytes"
	"embed"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"ttwms/internal/warehouse"
)

//go:embed templates/*.html static/*
var assets embed.FS

type app struct {
	logger      *slog.Logger
	templates   *template.Template
	store       warehouse.Repository
	environment string
}

type pageData struct {
	Title        string
	Environment  string
	SetupEnabled bool
	Setup        *setupPage
}

func New(logger *slog.Logger) http.Handler {
	return NewWithOptions(logger, Options{Environment: "development"})
}

type Options struct {
	Environment string
	Warehouse   warehouse.Repository
}

func NewWithOptions(logger *slog.Logger, options Options) http.Handler {
	templates := template.Must(template.ParseFS(assets, "templates/*.html"))
	a := &app{logger: logger, templates: templates, store: options.Warehouse, environment: options.Environment}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", a.live)
	mux.HandleFunc("GET /health/ready", a.ready)
	mux.HandleFunc("GET /{$}", a.home)
	mux.HandleFunc("GET /setup/{kind}", a.setupList)
	mux.HandleFunc("GET /setup/{kind}/{id}", a.setupEdit)
	mux.HandleFunc("POST /setup/{kind}", a.setupSave)
	mux.HandleFunc("POST /setup/{kind}/{id}", a.setupSave)
	mux.Handle("GET /static/", http.FileServerFS(assets))
	return requestLog(logger, securityHeaders(http.NewCrossOriginProtection().Handler(mux)))
}

func (a *app) live(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (a *app) home(w http.ResponseWriter, _ *http.Request) {
	a.render(w, "home", pageData{Title: "TT WMS", Environment: a.environment, SetupEnabled: a.setupEnabled()}, http.StatusOK)
}

func (a *app) render(w http.ResponseWriter, name string, data pageData, status int) {
	var output bytes.Buffer
	if err := a.templates.ExecuteTemplate(&output, name, data); err != nil {
		a.logger.Error("render page", "template", name, "error", err)
		http.Error(w, "Unable to display this page.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(output.Bytes())
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
