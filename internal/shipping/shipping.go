package shipping

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("shipment not found")

type Parcel struct {
	Sequence   int
	Quantity   int
	WeightGram int
}

type CreateRequest struct {
	IdempotencyKey string
	OrderID        string
	Carrier        string
	Service        string
	Parcel         Parcel
}

type Label struct {
	ID             string
	OrderID        string
	Carrier        string
	Service        string
	TrackingNumber string
	PreviewHTML    string
	Voided         bool
	CreatedAt      time.Time
}

type Provider interface {
	CreateLabel(context.Context, CreateRequest) (Label, error)
	VoidLabel(context.Context, string) (Label, error)
}
