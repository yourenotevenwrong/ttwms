# TikTok-First Commerce Operating System

**Version:** 0.2 — Internal Prototype Specification  
**Status:** Working specification; Round 3 operational decisions remain open  
**Initial market:** Republic of Ireland  
**Initial customer:** One internal business operating its own warehouse  
**Expected volume:** Approximately 3,000 orders/month; a few hundred orders/day; fewer than 10,000 SKUs; 5–6 concurrent warehouse users

## 1. Product Objective

Build a lightweight operating system for a TikTok Shop seller without requiring a conventional ERP or ecommerce website. The application will start with fresh operational data rather than migrate Peoplevox records.

The initial operating loop is:

`TikTok order → reserve stock → batch pick → mixed trolley → scan-to-order pack → mock label → simulated fulfilment update → return/receive → reconcile`

The prototype is private and single-business. The data model must retain a company boundary so later multi-tenant SaaS conversion remains possible, but public registration, subscriptions, billing, tenant self-service, and multiple shops are excluded.

## 2. Prototype Success Criteria

During a four-week internal pilot, target:

- At least 99.5% correct item picks.
- At least 99% of labels generated without manual data re-entry.
- Inventory variance below 1%.
- No duplicate despatches.
- A measurable reduction in fulfilment time compared with Peoplevox.
- Complete auditability for stock movements, scans, exceptions, labels, and manual overrides.

## 3. Scope

### Included

- Admin and Warehouse User authentication.
- One company, Irish TikTok Shop, and warehouse.
- TikTok catalogue, order, cancellation, return, inventory, fulfilment, and webhook integration where approved APIs permit.
- Product, SKU, barcode, location, cost, and bundle management.
- Authoritative inventory ledger and TikTok availability synchronisation.
- All-or-nothing order reservations.
- Batch creation and locking for up to 50 orders.
- Bulk picking into a mixed trolley.
- Scan-driven order matching and packing.
- Multi-parcel fulfilment.
- Deterministic mock shipping adapter producing non-shippable test labels and tracking numbers.
- Purchase orders and partial goods receipt.
- Return receipt with restockable/damaged disposition.
- Estimated gross product margin.
- Exception centre and configurable critical email alerts.
- CSV-based initial setup and operational exports.
- Responsive web UI for PCs, tablets, phones, and Android scanners.

### Excluded

- Multiple shops, warehouses, channels, or 3PLs.
- Public SaaS onboarding, billing, plans, trials, and SLA.
- Native mobile applications.
- Live Shipster and direct carrier integrations.
- Ecommerce storefront, checkout, CMS, CRM, marketing, loyalty, or accounting.
- Automated reorder recommendations and advanced demand forecasting.
- Lots, batches, serial numbers, expiry dates, and preassembled kits.
- Packing slips.
- Public customer API.
- Automatic refund submission until explicitly approved after API and workflow validation.

## 4. Users and Permissions

### Admin

Full access, including:

- Users and roles.
- TikTok and shipping-adapter configuration and credentials.
- Product costs and margin.
- Imports and exports.
- Stock adjustments.
- Purchase-order administration.
- Audit logs, retention settings, and exception resolution.
- Cancellation/refund actions made available by integrations.

### Warehouse User

May:

- View operational orders and customer details needed for fulfilment.
- Create, claim, pick, suspend, pack, and complete batches.
- Receive returns and classify condition.
- Perform permitted purchase-order receiving.
- Perform audited stock adjustments and resolve shortages.
- Report shortages and request cancellation.

May not access integration credentials, user administration, costs/margins, exports, audit logs, integration configuration, or retention settings.

Warehouse users may operate purchase-order and goods-receipt workflows. All stock changes and overrides must be audited.

## 5. TikTok Shop Integration

Use an Irish TikTok Shop Seller Developer account and custom application for the private prototype. Start Partner Center registration and sandbox setup before integration development.

The TikTok adapter must isolate external schemas from internal commerce models and provide:

- OAuth authorization and encrypted token storage.
- Automatic token refresh and reauthorization workflow.
- Product, variant, SKU, barcode, order, cancellation, return, inventory, fulfilment, and tracking operations supported by approved scopes.
- Signed webhook validation.
- Idempotent webhook/event handling.
- Rate-limit handling, retries with backoff, and dead-letter handling.
- Periodic reconciliation in addition to webhooks.
- API version isolation and structured integration logs.
- Connection health and last-successful-sync visibility.

TikTok passwords must never be requested or stored.

## 6. Products, SKUs, and Bundles

Each SKU supports:

- Internal and channel identifiers.
- Product/variant name.
- One or more barcodes.
- Purchase cost and effective-date history.
- Weight and optional dimensions.
- Active state.
- Warehouse location.
- SKU-specific safety-buffer override.
- Supplier mappings.

A bundle is a virtual sellable SKU composed of component SKUs and quantities. It has no independent physical stock. Bundle availability equals the lowest whole number of bundles available from its components after reservations, damage, and buffers. Picking and stock consumption occur against components.

Example: Shampoo & Conditioner Bundle = 1 × Shampoo SKU + 1 × Conditioner SKU.

## 7. Inventory

This application is the inventory system of record.

Track:

- Physical stock.
- Reserved stock.
- Damaged stock.
- Incoming stock.
- Safety buffer.
- Available-to-sell.

`available-to-sell = physical − reserved − damaged − safety buffer`

The global safety buffer defaults to zero and may be changed globally or overridden per SKU.

Every inventory change creates an immutable ledger transaction with company, warehouse, SKU, quantity, reason, source entity, actor, timestamp, and idempotency reference.

When an order arrives, reservation is atomic and all-or-nothing. If all required component and ordinary SKU quantities cannot be reserved, reserve nothing and place the order in a shortage exception.

TikTok stock publication is asynchronous. Overselling is minimised through immediate local reservations, buffers, idempotency, retries, visible failures, and reconciliation; it cannot be guaranteed impossible.

## 8. Orders and State Management

Import orders automatically and snapshot the applicable purchase cost on each order line at import time.

Minimum internal states:

- Imported.
- Awaiting reservation.
- Reserved.
- Stock shortage.
- In picking batch.
- Picked.
- Packing.
- Packed.
- Label generated.
- Handed over.
- Shipped confirmed.
- Cancelled.
- Return in progress.
- Returned.
- Exception.

Search and filter by order ID, customer, barcode/SKU, status, time, despatch deadline, carrier/service, batch, and exception.

External status transitions must be idempotent and recorded separately from internal workflow state.

## 9. Batch Picking

Warehouse Users can create a batch of up to 50 eligible orders. Default selection is oldest eligible order first. Advanced filtering is supported for the fields confirmed in Section 18.

A picker claims a batch by selecting or scanning an available configured trolley. Store the trolley, user, session/device, and claim/activity times. Other users cannot pick a claimed batch, and a trolley cannot hold more than one active batch.

The picker bulk-picks all required units into one mixed trolley; orders are not separated into totes.

The UI must support barcode-first operation, clear quantities, progress, unexpected barcode errors, short picks, suspension, and safe recovery after connectivity loss. Picking defaults to scanning every unit individually. Batch filters are order time, despatch deadline, carrier/service, SKU, order type, and exception status.

Users may release their own claimed batch and Admin may force-release any batch. After 15 minutes without activity, warn the user and Admin but do not automatically unlock it.

## 10. Scan-to-Order Packing

At the packing station, every unit is scanned.

Proposed matching algorithm:

1. First scan identifies unpacked orders in the current batch requiring that SKU.
2. Select the oldest eligible matching order deterministically. The packer cannot reject or choose a different match in the prototype.
3. Lock that order to the packing session.
4. Display all remaining units and SKUs required for it.
5. Require every unit to be scanned.
6. Complete packing only when expected and scanned quantities match.
7. Permit suspension if remaining items cannot be found, preserving scan history.

The batch cannot close while orders or picked quantities remain unresolved. Wrong, duplicate, and excess scans must be rejected and audited. Final matching/suspension behaviour requires approval.

## 11. Shipping

The prototype uses a deterministic mock shipping adapter. It must never contact a carrier, purchase postage, produce a carrier-valid tracking number, or be enableable accidentally in production.

Expected flow:

`Packed order → mock shipment request → test label/tracking → print preview → simulated TikTok fulfilment update`

Requirements:

- Automatically request a label after packing.
- Store mock request/response identifiers without logging unnecessary customer data.
- Store each parcel, label, tracking number, carrier, and service independently.
- Allow a packer to add parcels to large orders.
- Retry safely without purchasing or printing duplicate labels.
- Void labels when permitted.
- Surface failures in the exception centre.
- Confirm TikTok's required multi-parcel line/quantity mapping.

The mock selects a configurable DPD or An Post test service. The default parcel quantity is 1 unit and default parcel weight is 300 g. Weight should ultimately be calculated from SKU weights plus packaging weight; raise a visible warning whenever the default is used.

The mock produces a clearly watermarked, non-shippable HTML label preview. Zebra printing and local print-bridge integration are deferred with live shipping.

## 12. Cancellations and Exceptions

Before picking/packing, an accepted TikTok cancellation releases the full reservation automatically.

After picking begins, cancellation creates a blocking operational exception. After label generation, it creates an urgent stop-despatch exception and voids the mock label. Stock must not be released until the parcel is recovered and its contents reconciled.

For the prototype, successful label generation and receipt of a tracking number marks the order as handed over/shipped for workflow purposes. This simplification does not prove physical courier collection and must be revisited before external production use.

Exception types include:

- Stock shortage or discrepancy.
- Unexpected/missing scan.
- Stale batch lock.
- TikTok sync/authentication failure.
- Mock label/void/preview failure.
- Cancellation during fulfilment.
- Unresolved picked quantity.
- Return mismatch.

## 13. Purchasing and Receiving

Prototype purchasing includes:

- Suppliers and supplier SKUs.
- Purchase cost.
- Purchase orders and lines.
- Draft, issued, part-received, received, and cancelled states.
- Partial receipts.
- All authenticated users may create, issue, cancel, and receive purchase orders.
- Over-receipts are permitted but visibly flagged and audited.
- Incoming quantities.
- Inventory ledger entries for accepted receipts.

Reorder recommendations, forecasting, supplier automation, and supplier portals are deferred.

## 14. Returns and Refunds

Locate/import the TikTok return and original order, then scan received units against it.

Each returned unit is classified as:

- Restockable: create a positive stock transaction.
- Damaged: create damaged stock; do not make it available to sell.

Inventory disposition and refund entitlement are separate decisions. Display TikTok refund status, but defer automatic refund submission. Both dispositions should be presented for explicit Admin refund action unless Section 18 changes this rule.

Unexpected returns and quantity/SKU mismatches become exceptions.

## 15. Reporting, Alerts, and Exports

Initial dashboard:

- Sales revenue.
- Orders and units.
- Orders awaiting attention/fulfilment.
- Picking and packing workload.
- Low/out-of-stock SKUs.
- Returns.
- Integration and fulfilment exceptions.
- Estimated gross product margin.

`estimated gross product margin = sales revenue − order-line purchase-cost snapshots`

This is not net profit and excludes VAT treatment, fees, affiliate commission, advertising, shipping, packaging, labour, duty, and returns unless later expanded.

In-app alerts cover sync failures, authentication, label/printing failures, stock discrepancies, short picks, overdue fulfilment, and low stock. Critical email recipients are configurable.

Admin CSV exports: orders, order lines, products, stock balances, stock movements, returns, shipments, and margin.

## 16. Initial Data Setup

No Peoplevox data migration will be performed. TikTok is the product-content source: import its catalogue and channel stock. Imported TikTok stock becomes a provisional audited opening balance until verified physically. Application-only fields such as internal barcodes, locations, costs, bundle definitions, suppliers, and buffers are completed locally.

Provide CSV templates and manual administration screens for products, SKUs, barcodes, locations, bundle definitions, suppliers, costs, and opening stock. Imports must support dry-run validation, duplicate and broken-reference reporting, atomic commit, error export, and audit history.

Admin setup must also create the warehouse's operational master data from scratch:

- Warehouse identity and address.
- Locations with unique code, scannable barcode, active state, optional zone, and type.
- Initial location types: receiving, storage/pick, packing/staging, returns, and damaged/quarantine.
- Reusable trolleys with unique code, scannable barcode, active/out-of-service state, and optional capacity.
- Packing stations with unique code and active state.
- Printer assignment per packing station, including printer name, print bridge/connection, and label format.

A trolley remains unavailable until its batch is completed, suspended and released, or force-released by an Admin. Trolley assignment records operational custody; it does not create order-specific compartments or inventory balances. Default codes are human-readable prefixed sequences such as `LOC-PICK-001`, `TROLLEY-001`, and `PACK-01`, editable by Admin before use.

Opening stock must be entered through a dedicated, audited opening-balance operation. Whether those quantities come from a physical count remains a business decision. Historical Peoplevox orders, stock transactions, suppliers, purchase orders, and rules are explicitly excluded.

## 17. Technical Architecture and Operations

### Application

- Go backend and web server.
- Server-rendered, responsive HTML enhanced with htmx.
- Fomantic UI CSS-only distribution for the base design system, including forms, buttons, tables, menus, labels, and messages.
- A small project stylesheet for dense warehouse tables, scanner prompts, touch targets, packing progress, and application-specific states.
- Do not use Tailwind CSS, Semantic UI, jQuery, or Fomantic's JavaScript components.
- Prefer native HTML controls and dialogs. Use focused JavaScript only where htmx and native HTML are insufficient, particularly barcode-scanner behaviour, resilient packing interactions, and label-print integration.
- Internal HTTP handlers and service boundaries should support later API extraction, but a public REST API is not required for the prototype.
- Separate background worker process.
- PostgreSQL from the beginning; SQLite is excluded except for isolated tests that do not validate production database behaviour.
- Redis for jobs/cache where required.
- Adapter boundaries for TikTok and shipping; the mock implements the shipping interface used by a future Shipster adapter.

### Hosting

- DigitalOcean EU region, provisionally Amsterdam.
- Codex may be installed on the private development/prototype server.
- Docker Compose defines the application, worker, PostgreSQL, Redis, and reverse proxy consistently.
- Initial internal prototype may run on one Droplet.
- Version 1 production uses a clean server/environment and should move PostgreSQL to DigitalOcean Managed PostgreSQL before operationally critical or external use.

### Security

- HTTPS, secure cookies, CSRF protection where applicable, input validation, and rate limiting.
- Password hashing and role-based authorization enforced server-side.
- Encrypted integration credentials and secret rotation.
- No secrets in Git, logs, or specification files.
- Webhook signature verification and replay protection.
- Audit logs for privileged and inventory-affecting actions.
- Least-privilege server, database, and DigitalOcean access.
- Dependency and operating-system security updates.

### Resilience

- Automated off-server PostgreSQL backups.
- Proposed daily backups retained for 30 days plus recovery coverage targeting no more than one hour of data loss.
- Target recovery within four hours.
- Regular restoration testing.
- Uptime, job-queue, integration, database, disk, and error monitoring.
- No formal uptime SLA for the private prototype.

### Data Protection

- EU hosting.
- Warehouse customer-data access limited to fulfilment needs.
- Customer-data exports restricted to Admin.
- Warehouse customer-data access window: 30 days.
- Admin operational customer-data access window: 90 days.
- After the approved period, anonymise unnecessary contact data while retaining minimum legally required commercial records.
- Implement access/export/deletion workflows before external launch.
- Obtain professional Irish data-protection/legal review before external use.

## 18. Open Decisions

1. Decide whether physical opening stock must be counted before TikTok-derived provisional balances become authoritative.
2. Supply real warehouse location, trolley, packing-station, bundle, supplier, and purchase-order examples as implementation fixtures.

## 19. External Dependencies and Required Inputs

- TikTok Seller Developer registration and Irish development shop.
- TikTok custom app, approved scopes, test credentials, and webhook configuration.
- 20–50 anonymised representative orders covering all important cases.
- Warehouse layout and location list, trolley identifiers, packing stations, and printer assignments.
- Real bundle definitions.
- Current supplier and purchase-order fields/workflows.
- DigitalOcean account and prototype Droplet.
- Later: domain/DNS, sender email/domain, monitoring accounts, legal documents, privacy policy, terms, and incident/support ownership.

## 20. Delivery Sequence

1. Integration feasibility: TikTok sandbox/scopes and deterministic shipping mock.
2. Foundation: repository, Docker Compose, PostgreSQL, authentication, roles, audit framework, and background jobs.
3. Initial setup: products, barcodes, locations, bundles, opening stock, suppliers, and costs.
4. Inventory and orders: ledger, reservation, TikTok import/webhooks, reconciliation, and availability publication.
5. Warehouse: batch creation/locking, picking, mixed-trolley packing, and exceptions.
6. Shipping: mock labels, preview, parcels, tracking, cancellation handling, and simulated TikTok fulfilment.
7. Purchasing and returns: purchase orders, receipts, return scanning, disposition, and refund visibility.
8. Reporting and hardening: dashboard, margin, CSV exports, alerts, backups, monitoring, security review, and pilot acceptance tests.
