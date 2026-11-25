package warehouse

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Kind string

const (
	Locations       Kind = "locations"
	Trolleys        Kind = "trolleys"
	PackingStations Kind = "packing-stations"
)

var (
	ErrNotFound = errors.New("warehouse asset not found")
	ErrConflict = errors.New("that code or barcode is already in use in this warehouse")
)

// Asset contains the editable master data. CapacityOrders zero means unspecified.
type Asset struct {
	ID             string
	Code           string
	Barcode        string
	Zone           string
	LocationType   string
	Active         bool
	Status         string
	CapacityOrders int
	PrinterName    string
	PrinterFormat  string
}

type Repository interface {
	Ping(context.Context) error
	List(context.Context, Kind) ([]Asset, error)
	Get(context.Context, Kind, string) (Asset, error)
	Save(context.Context, Kind, Asset) (Asset, error)
}

type ValidationError map[string]string

func (ValidationError) Error() string { return "check the highlighted fields" }

func ValidKind(kind Kind) bool {
	return kind == Locations || kind == Trolleys || kind == PackingStations
}

func Normalize(asset Asset) Asset {
	asset.Code = strings.TrimSpace(asset.Code)
	asset.Barcode = strings.TrimSpace(asset.Barcode)
	asset.Zone = strings.TrimSpace(asset.Zone)
	asset.PrinterName = strings.TrimSpace(asset.PrinterName)
	asset.PrinterFormat = strings.TrimSpace(asset.PrinterFormat)
	return asset
}

func Validate(kind Kind, asset Asset) error {
	if !ValidKind(kind) {
		return ErrNotFound
	}
	fields := ValidationError{}
	text := func(field, value string, required bool, limit int) {
		if required && strings.TrimSpace(value) == "" {
			fields[field] = "This field is required."
		} else if utf8.RuneCountInString(value) > limit || strings.ContainsAny(value, "\x00\r\n") {
			fields[field] = "Use a single line of at most " + strconv.Itoa(limit) + " characters."
		}
	}
	text("code", asset.Code, true, 100)
	switch kind {
	case Locations:
		text("barcode", asset.Barcode, true, 200)
		text("zone", asset.Zone, false, 100)
		switch asset.LocationType {
		case "receiving", "pick", "staging", "returns", "damaged":
		default:
			fields["location_type"] = "Choose a location type."
		}
	case Trolleys:
		text("barcode", asset.Barcode, true, 200)
		if asset.CapacityOrders < 0 || asset.CapacityOrders > 2147483647 {
			fields["capacity_orders"] = "Enter a positive whole number or leave blank."
		}
	}
	if kind != Locations && asset.Status != "active" && asset.Status != "out_of_service" {
		fields["status"] = "Choose active or out of service."
	}
	if kind == PackingStations {
		text("printer_name", asset.PrinterName, false, 200)
		text("printer_format", asset.PrinterFormat, false, 100)
	}
	if len(fields) != 0 {
		return fields
	}
	return nil
}
