package warehouse

import (
	"errors"
	"strings"
	"testing"
)

func TestValidationRejectsInvalidMasterData(t *testing.T) {
	tests := []struct {
		name  string
		kind  Kind
		asset Asset
		field string
	}{
		{"missing code", Locations, Asset{Barcode: "B", LocationType: "pick"}, "code"},
		{"missing barcode", Locations, Asset{Code: "L", LocationType: "pick"}, "barcode"},
		{"unknown location type", Locations, Asset{Code: "L", Barcode: "B", LocationType: "unknown"}, "location_type"},
		{"multiline barcode", Trolleys, Asset{Code: "T", Barcode: "B\nC", Status: "active"}, "barcode"},
		{"negative capacity", Trolleys, Asset{Code: "T", Barcode: "B", Status: "active", CapacityOrders: -1}, "capacity_orders"},
		{"capacity overflow", Trolleys, Asset{Code: "T", Barcode: "B", Status: "active", CapacityOrders: 2147483648}, "capacity_orders"},
		{"unknown status", PackingStations, Asset{Code: "P", Status: "broken"}, "status"},
		{"oversize printer", PackingStations, Asset{Code: "P", Status: "active", PrinterName: strings.Repeat("a", 201)}, "printer_name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fields ValidationError
			if err := Validate(tt.kind, tt.asset); !errors.As(err, &fields) || fields[tt.field] == "" {
				t.Fatalf("expected validation error for %s, got %v", tt.field, err)
			}
		})
	}
}

func TestValidMasterData(t *testing.T) {
	for _, test := range []struct {
		kind  Kind
		asset Asset
	}{
		{Locations, Asset{Code: "LOC-PICK-001", Barcode: "LOC-PICK-001", LocationType: "pick", Active: true}},
		{Trolleys, Asset{Code: "TROLLEY-001", Barcode: "TROLLEY-001", Status: "active"}},
		{PackingStations, Asset{Code: "PACK-01", Status: "out_of_service", PrinterFormat: "100 × 150 mm"}},
	} {
		if err := Validate(test.kind, test.asset); err != nil {
			t.Fatalf("%s: %v", test.kind, err)
		}
	}
	if err := Validate(Kind("users"), Asset{}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown asset kind accepted: %v", err)
	}
}
