package httpapp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ttwms/internal/warehouse"
)

type setupPage struct {
	Kind     warehouse.Kind
	Heading  string
	Singular string
	Items    []warehouse.Asset
	Form     warehouse.Asset
	Capacity string
	Errors   warehouse.ValidationError
	Saved    bool
}

func (a *app) setupEnabled() bool {
	// Authentication is the next milestone. Do not expose these write operations
	// in staging or production before server-side authorization is implemented.
	return a.store != nil && (a.environment == "development" || a.environment == "test")
}

func (a *app) setupKind(w http.ResponseWriter, r *http.Request) (warehouse.Kind, bool) {
	kind := warehouse.Kind(r.PathValue("kind"))
	if !a.setupEnabled() || !warehouse.ValidKind(kind) {
		http.NotFound(w, r)
		return "", false
	}
	return kind, true
}

func (a *app) setupList(w http.ResponseWriter, r *http.Request) {
	kind, ok := a.setupKind(w, r)
	if !ok {
		return
	}
	a.renderSetup(w, r, kind, warehouse.Asset{Active: true, LocationType: "pick", Status: "active"}, "", nil, http.StatusOK)
}

func (a *app) setupEdit(w http.ResponseWriter, r *http.Request) {
	kind, ok := a.setupKind(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	asset, err := a.store.Get(ctx, kind, r.PathValue("id"))
	if err != nil {
		a.setupError(w, r, err)
		return
	}
	capacity := ""
	if asset.CapacityOrders != 0 {
		capacity = strconv.Itoa(asset.CapacityOrders)
	}
	a.renderSetup(w, r.WithContext(ctx), kind, asset, capacity, nil, http.StatusOK)
}

func (a *app) setupSave(w http.ResponseWriter, r *http.Request) {
	kind, ok := a.setupKind(w, r)
	if !ok {
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unable to read this form.", http.StatusBadRequest)
		return
	}
	asset := warehouse.Normalize(warehouse.Asset{
		ID: r.PathValue("id"), Code: r.PostForm.Get("code"), Barcode: r.PostForm.Get("barcode"),
		Zone: r.PostForm.Get("zone"), LocationType: r.PostForm.Get("location_type"), Active: r.PostForm.Get("active") == "on",
		Status: r.PostForm.Get("status"), PrinterName: r.PostForm.Get("printer_name"), PrinterFormat: r.PostForm.Get("printer_format"),
	})
	fields := warehouse.ValidationError{}
	capacity := strings.TrimSpace(r.PostForm.Get("capacity_orders"))
	if kind == warehouse.Trolleys && capacity != "" {
		n, err := strconv.ParseInt(capacity, 10, 32)
		if err != nil || n <= 0 {
			fields["capacity_orders"] = "Enter a positive whole number or leave blank."
		} else {
			asset.CapacityOrders = int(n)
		}
	}
	if err := warehouse.Validate(kind, asset); err != nil {
		var validation warehouse.ValidationError
		if errors.As(err, &validation) {
			for key, message := range validation {
				fields[key] = message
			}
		}
	}
	if len(fields) > 0 {
		a.renderSetup(w, r, kind, asset, capacity, fields, http.StatusUnprocessableEntity)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	_, err := a.store.Save(ctx, kind, asset)
	if errors.Is(err, warehouse.ErrConflict) {
		a.renderSetup(w, r.WithContext(ctx), kind, asset, capacity, warehouse.ValidationError{"form": "That code or barcode is already in use. Choose another value."}, http.StatusConflict)
		return
	}
	if err != nil {
		a.setupError(w, r, err)
		return
	}
	http.Redirect(w, r, "/setup/"+string(kind)+"?saved=1", http.StatusSeeOther)
}

func (a *app) renderSetup(w http.ResponseWriter, r *http.Request, kind warehouse.Kind, form warehouse.Asset, capacity string, fields warehouse.ValidationError, status int) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	items, err := a.store.List(ctx, kind)
	if err != nil {
		a.setupError(w, r, err)
		return
	}
	heading, singular := "Locations", "location"
	switch kind {
	case warehouse.Trolleys:
		heading, singular = "Trolleys", "trolley"
	case warehouse.PackingStations:
		heading, singular = "Packing stations", "packing station"
	}
	w.Header().Set("Cache-Control", "no-store")
	a.render(w, "setup", pageData{Title: heading + " · TT WMS", Environment: a.environment, SetupEnabled: true, Setup: &setupPage{
		Kind: kind, Heading: heading, Singular: singular, Items: items, Form: form, Capacity: capacity, Errors: fields, Saved: r.URL.Query().Get("saved") == "1",
	}}, status)
}

func (a *app) setupError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, warehouse.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	a.logger.Error("warehouse setup failed", "error", err)
	http.Error(w, "Warehouse setup is temporarily unavailable. Please try again.", http.StatusServiceUnavailable)
}

func (a *app) ready(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if a.store == nil || a.store.Ping(ctx) != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"not_ready"}`))
		return
	}
	_, _ = w.Write([]byte(`{"status":"ready"}`))
}
