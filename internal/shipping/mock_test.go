package shipping

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestMockLabelIsDeterministicAndClearlyFake(t *testing.T) {
	mock, err := NewMock("test")
	if err != nil {
		t.Fatal(err)
	}
	req := CreateRequest{IdempotencyKey: "order-1-parcel-1", OrderID: "ORDER-1", Parcel: Parcel{Sequence: 1, Quantity: 1, WeightGram: 300}}
	first, err := mock.CreateLabel(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	second, err := mock.CreateLabel(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID || first.TrackingNumber != second.TrackingNumber {
		t.Fatalf("idempotent request changed label: %#v %#v", first, second)
	}
	if !strings.Contains(first.TrackingNumber, "DO-NOT-SHIP") || !strings.Contains(first.PreviewHTML, "DO NOT SHIP") {
		t.Fatalf("mock label is not visibly fake: %#v", first)
	}
}

func TestMockCannotRunInProduction(t *testing.T) {
	_, err := NewMock("production")
	if !errors.Is(err, ErrMockForbidden) {
		t.Fatalf("expected production guard, got %v", err)
	}
}

func TestMockVoid(t *testing.T) {
	mock, _ := NewMock("test")
	label, _ := mock.CreateLabel(context.Background(), CreateRequest{IdempotencyKey: "k", OrderID: "o", Parcel: Parcel{Sequence: 1, Quantity: 1, WeightGram: 300}})
	voided, err := mock.VoidLabel(context.Background(), label.ID)
	if err != nil || !voided.Voided {
		t.Fatalf("expected voided label, got %#v %v", voided, err)
	}
}
