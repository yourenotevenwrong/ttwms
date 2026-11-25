package httpapp

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"ttwms/internal/warehouse"
)

type setupStore struct {
	items     []warehouse.Asset
	saved     []warehouse.Asset
	saveError error
	readError error
	pingError error
}

func (s *setupStore) Ping(context.Context) error { return s.pingError }
func (s *setupStore) List(context.Context, warehouse.Kind) ([]warehouse.Asset, error) {
	return s.items, s.readError
}
func (s *setupStore) Get(_ context.Context, _ warehouse.Kind, id string) (warehouse.Asset, error) {
	for _, asset := range s.items {
		if asset.ID == id {
			return asset, nil
		}
	}
	return warehouse.Asset{}, warehouse.ErrNotFound
}
func (s *setupStore) Save(_ context.Context, _ warehouse.Kind, asset warehouse.Asset) (warehouse.Asset, error) {
	if s.saveError != nil {
		return warehouse.Asset{}, s.saveError
	}
	s.saved = append(s.saved, asset)
	return asset, nil
}

func setupHandler(store warehouse.Repository, environment string) http.Handler {
	return NewWithOptions(slog.New(slog.NewTextHandler(io.Discard, nil)), Options{Environment: environment, Warehouse: store})
}

func submit(h http.Handler, path string, values url.Values) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, path, strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestSetupPagesAndEditForms(t *testing.T) {
	for _, kind := range []string{"locations", "trolleys", "packing-stations"} {
		t.Run(kind, func(t *testing.T) {
			store := &setupStore{items: []warehouse.Asset{{ID: "asset-1", Code: "EXISTING", Barcode: "BARCODE", Status: "active", LocationType: "pick", Active: true, CapacityOrders: 50, PrinterName: "Printer 1"}}}
			h := setupHandler(store, "development")
			for _, path := range []string{"/setup/" + kind, "/setup/" + kind + "/asset-1"} {
				w := httptest.NewRecorder()
				h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
				if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "EXISTING") || !strings.Contains(w.Body.String(), "<form") {
					t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
				}
			}
		})
	}
}

func TestCreateAndEditRedirectAfterSave(t *testing.T) {
	for _, path := range []string{"/setup/locations", "/setup/locations/asset-1"} {
		t.Run(path, func(t *testing.T) {
			store := &setupStore{}
			w := submit(setupHandler(store, "development"), path, url.Values{"code": {" LOC-001 "}, "barcode": {" B001 "}, "location_type": {"pick"}, "active": {"on"}})
			if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/setup/locations?saved=1" || len(store.saved) != 1 {
				t.Fatalf("save failed: %d %s", w.Code, w.Body.String())
			}
			if store.saved[0].Code != "LOC-001" || store.saved[0].Barcode != "B001" || !store.saved[0].Active {
				t.Fatalf("invalid saved values: %#v", store.saved[0])
			}
			if strings.HasSuffix(path, "asset-1") && store.saved[0].ID != "asset-1" {
				t.Fatal("edit did not retain the asset identity")
			}
		})
	}
}

func TestInvalidFormDoesNotSaveAndRetainsInput(t *testing.T) {
	store := &setupStore{}
	w := submit(setupHandler(store, "development"), "/setup/trolleys", url.Values{"code": {"TROLLEY-001"}, "barcode": {"B001"}, "status": {"active"}, "capacity_orders": {"0"}})
	if w.Code != http.StatusUnprocessableEntity || len(store.saved) != 0 || !strings.Contains(w.Body.String(), `value="TROLLEY-001"`) || !strings.Contains(w.Body.String(), "positive whole number") {
		t.Fatalf("invalid form result: %d %s", w.Code, w.Body.String())
	}
}

func TestDuplicateCodeKeepsFormAndEscapesInput(t *testing.T) {
	store := &setupStore{saveError: warehouse.ErrConflict}
	w := submit(setupHandler(store, "development"), "/setup/packing-stations", url.Values{"code": {"<script>alert(1)</script>"}, "status": {"active"}})
	if w.Code != http.StatusConflict || strings.Contains(w.Body.String(), "<script>alert(1)</script>") || !strings.Contains(w.Body.String(), "&lt;script&gt;") {
		t.Fatalf("conflict response unsafe or incorrect: %d %s", w.Code, w.Body.String())
	}
}

func TestSetupUnavailableOutsideDevelopmentAndTest(t *testing.T) {
	for _, environment := range []string{"staging", "production"} {
		store := &setupStore{}
		w := submit(setupHandler(store, environment), "/setup/packing-stations", url.Values{"code": {"PACK-01"}, "status": {"active"}})
		if w.Code != http.StatusNotFound || len(store.saved) != 0 {
			t.Fatalf("setup exposed in %s", environment)
		}
	}
}

func TestNotFoundRoutesAndMissingAssets(t *testing.T) {
	h := setupHandler(&setupStore{}, "development")
	for _, path := range []string{"/unknown", "/setup/users", "/setup/locations/missing"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("%s returned %d", path, w.Code)
		}
	}
}

func TestCrossOriginFormIsRejected(t *testing.T) {
	store := &setupStore{}
	r := httptest.NewRequest(http.MethodPost, "/setup/locations", strings.NewReader("code=LOC-1&barcode=B1&location_type=pick"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Origin", "https://untrusted.example")
	r.Header.Set("Sec-Fetch-Site", "cross-site")
	w := httptest.NewRecorder()
	setupHandler(store, "development").ServeHTTP(w, r)
	if w.Code != http.StatusForbidden || len(store.saved) != 0 {
		t.Fatalf("cross-origin write accepted: %d", w.Code)
	}
}

func TestReadinessAndDatabaseFailure(t *testing.T) {
	for _, test := range []struct {
		store warehouse.Repository
		want  int
	}{
		{nil, http.StatusServiceUnavailable},
		{&setupStore{}, http.StatusOK},
		{&setupStore{pingError: errors.New("offline")}, http.StatusServiceUnavailable},
	} {
		w := httptest.NewRecorder()
		setupHandler(test.store, "development").ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
		if w.Code != test.want {
			t.Fatalf("readiness: got %d, want %d", w.Code, test.want)
		}
	}
	w := httptest.NewRecorder()
	setupHandler(&setupStore{readError: errors.New("private database details")}, "development").ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/setup/locations", nil))
	if w.Code != http.StatusServiceUnavailable || strings.Contains(w.Body.String(), "private database details") {
		t.Fatalf("database error response: %d %s", w.Code, w.Body.String())
	}
}
