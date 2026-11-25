package warehouse

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Postgres struct{ DB *sql.DB }

func (p *Postgres) Ping(ctx context.Context) error { return p.DB.PingContext(ctx) }

// All SQL identifiers come from this fixed allowlist; form values are parameters.
type assetTable struct {
	name    string
	columns []string
}

func tableFor(kind Kind) (assetTable, error) {
	switch kind {
	case Locations:
		return assetTable{"locations", []string{"code", "barcode", "zone", "type", "active"}}, nil
	case Trolleys:
		return assetTable{"trolleys", []string{"code", "barcode", "status", "capacity_orders"}}, nil
	case PackingStations:
		return assetTable{"packing_stations", []string{"code", "status", "printer_name", "printer_format"}}, nil
	default:
		return assetTable{}, ErrNotFound
	}
}

func (t assetTable) selectColumns() string {
	return "id::text," + strings.Join(t.columns, ",")
}

type scanner interface{ Scan(...any) error }

func scanAsset(row scanner, kind Kind) (Asset, error) {
	var asset Asset
	var zone, printerName, printerFormat sql.NullString
	var capacity sql.NullInt64
	var err error
	switch kind {
	case Locations:
		err = row.Scan(&asset.ID, &asset.Code, &asset.Barcode, &zone, &asset.LocationType, &asset.Active)
		asset.Zone = zone.String
	case Trolleys:
		err = row.Scan(&asset.ID, &asset.Code, &asset.Barcode, &asset.Status, &capacity)
		asset.CapacityOrders = int(capacity.Int64)
	case PackingStations:
		err = row.Scan(&asset.ID, &asset.Code, &asset.Status, &printerName, &printerFormat)
		asset.PrinterName, asset.PrinterFormat = printerName.String, printerFormat.String
	default:
		return Asset{}, ErrNotFound
	}
	return asset, mapError(err)
}

func (p *Postgres) List(ctx context.Context, kind Kind) ([]Asset, error) {
	table, err := tableFor(kind)
	if err != nil {
		return nil, err
	}
	rows, err := p.DB.QueryContext(ctx, "SELECT "+table.selectColumns()+" FROM "+table.name+" WHERE warehouse_id=(SELECT warehouse_id FROM app_installation WHERE singleton) ORDER BY code,id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	assets := []Asset{}
	for rows.Next() {
		asset, err := scanAsset(rows, kind)
		if err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}
	return assets, rows.Err()
}

func (p *Postgres) Get(ctx context.Context, kind Kind, id string) (Asset, error) {
	table, err := tableFor(kind)
	if err != nil {
		return Asset{}, err
	}
	return scanAsset(p.DB.QueryRowContext(ctx, "SELECT "+table.selectColumns()+" FROM "+table.name+" WHERE id=$1 AND warehouse_id=(SELECT warehouse_id FROM app_installation WHERE singleton)", id), kind)
}

func (p *Postgres) Save(ctx context.Context, kind Kind, asset Asset) (Asset, error) {
	asset = Normalize(asset)
	if err := Validate(kind, asset); err != nil {
		return Asset{}, err
	}
	table, _ := tableFor(kind)
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return Asset{}, err
	}
	defer tx.Rollback()
	var companyID, warehouseID string
	if err := tx.QueryRowContext(ctx, `SELECT company_id::text,warehouse_id::text FROM app_installation WHERE singleton`).Scan(&companyID, &warehouseID); err != nil {
		return Asset{}, err
	}
	var before *Asset
	if asset.ID != "" {
		existing, err := scanAsset(tx.QueryRowContext(ctx, "SELECT "+table.selectColumns()+" FROM "+table.name+" WHERE id=$1 AND warehouse_id=$2 FOR UPDATE", asset.ID, warehouseID), kind)
		if err != nil {
			return Asset{}, err
		}
		before = &existing
	}
	var values []any
	switch kind {
	case Locations:
		values = []any{asset.Code, asset.Barcode, asset.Zone, asset.LocationType, asset.Active}
	case Trolleys:
		var capacity any
		if asset.CapacityOrders != 0 {
			capacity = asset.CapacityOrders
		}
		values = []any{asset.Code, asset.Barcode, asset.Status, capacity}
	case PackingStations:
		values = []any{asset.Code, asset.Status, asset.PrinterName, asset.PrinterFormat}
	}
	var query, action string
	if before == nil {
		placeholders := make([]string, len(values))
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}
		values = append(values, warehouseID)
		query = fmt.Sprintf("INSERT INTO %s (%s,warehouse_id) VALUES (%s,$%d)", table.name, strings.Join(table.columns, ","), strings.Join(placeholders, ","), len(values))
		action = "setup.create"
	} else {
		assignments := make([]string, len(values))
		for i, column := range table.columns {
			assignments[i] = fmt.Sprintf("%s=$%d", column, i+1)
		}
		values = append(values, asset.ID, warehouseID)
		query = fmt.Sprintf("UPDATE %s SET %s WHERE id=$%d AND warehouse_id=$%d", table.name, strings.Join(assignments, ","), len(values)-1, len(values))
		action = "setup.update"
	}
	stored, err := scanAsset(tx.QueryRowContext(ctx, query+" RETURNING "+table.selectColumns(), values...), kind)
	if err != nil {
		return Asset{}, err
	}
	detail, err := json.Marshal(struct {
		Before *Asset `json:"before,omitempty"`
		After  Asset  `json:"after"`
	}{before, stored})
	if err != nil {
		return Asset{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO audit_log(company_id,action,entity_type,entity_id,detail) VALUES ($1,$2,$3,$4,$5::jsonb)`, companyID, action, string(kind), stored.ID, string(detail)); err != nil {
		return Asset{}, err
	}
	if err := tx.Commit(); err != nil {
		return Asset{}, err
	}
	return stored, nil
}

func mapError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	var state interface{ SQLState() string }
	if errors.As(err, &state) {
		switch state.SQLState() {
		case "23505":
			return ErrConflict
		case "22P02":
			return ErrNotFound
		}
	}
	return err
}
