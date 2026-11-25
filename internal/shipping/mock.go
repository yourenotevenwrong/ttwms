package shipping

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"strings"
	"sync"
	"time"
)

var ErrMockForbidden = errors.New("mock shipping is forbidden in production")

type Mock struct {
	mu      sync.Mutex
	byKey   map[string]Label
	keyByID map[string]string
	now     func() time.Time
}

func NewMock(environment string) (*Mock, error) {
	switch strings.ToLower(strings.TrimSpace(environment)) {
	case "production", "prod":
		return nil, ErrMockForbidden
	}
	return &Mock{byKey: make(map[string]Label), keyByID: make(map[string]string), now: time.Now}, nil
}

func (m *Mock) CreateLabel(_ context.Context, req CreateRequest) (Label, error) {
	if strings.TrimSpace(req.IdempotencyKey) == "" || strings.TrimSpace(req.OrderID) == "" {
		return Label{}, errors.New("idempotency key and order ID are required")
	}
	if req.Parcel.Sequence < 1 || req.Parcel.Quantity < 1 || req.Parcel.WeightGram < 1 {
		return Label{}, errors.New("parcel sequence, quantity, and weight must be positive")
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, ok := m.byKey[req.IdempotencyKey]; ok {
		return existing, nil
	}

	digest := sha256.Sum256([]byte(req.IdempotencyKey + "|" + req.OrderID))
	short := strings.ToUpper(hex.EncodeToString(digest[:6]))
	label := Label{
		ID:             "MOCK-LBL-" + short,
		OrderID:        req.OrderID,
		Carrier:        defaultString(req.Carrier, "AN_POST_TEST"),
		Service:        defaultString(req.Service, "STANDARD_TEST"),
		TrackingNumber: "TEST-DO-NOT-SHIP-" + short,
		CreatedAt:      m.now().UTC(),
	}
	label.PreviewHTML = fmt.Sprintf(`<section class="mock-label"><strong>TEST LABEL — DO NOT SHIP</strong><dl><dt>Order</dt><dd>%s</dd><dt>Tracking</dt><dd>%s</dd><dt>Parcel</dt><dd>%d</dd><dt>Weight</dt><dd>%d g</dd></dl></section>`, html.EscapeString(label.OrderID), label.TrackingNumber, req.Parcel.Sequence, req.Parcel.WeightGram)
	m.byKey[req.IdempotencyKey] = label
	m.keyByID[label.ID] = req.IdempotencyKey
	return label, nil
}

func (m *Mock) VoidLabel(_ context.Context, id string) (Label, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key, ok := m.keyByID[id]
	if !ok {
		return Label{}, ErrNotFound
	}
	label := m.byKey[key]
	label.Voided = true
	m.byKey[key] = label
	return label, nil
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
