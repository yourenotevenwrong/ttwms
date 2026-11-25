# Project Discovery — Condensed Decision Record

**Updated:** 3 September 2026  
**Detailed requirements:** `tiktok_shop_commerce_os_spec_v0.2.md`

This records discovery outcomes and remaining business inputs without preserving the original question-and-answer transcript.

## Progress

- [x] TikTok Shop Partner Center Seller Developer registration.
- [x] Custom app created; App Key and App Secret obtained and retained privately.
- [x] Required MVP API scopes enabled/requested.
- [x] Irish Core Function development shop created.
- [x] Custom app authorized against the Irish development shop.
- [x] API Testing Tool authorization succeeded and `Get Authorized Shops` returned the Irish local sandbox shop.
- [ ] Application-side OAuth callback, token exchange, encrypted access/refresh-token storage, and refresh flow implemented.
- [ ] TikTok webhook subscriptions configured and verified.
- [x] Live Shipster integration deferred; deterministic mock shipping selected and implemented for prototype testing.
- [ ] DigitalOcean prototype environment created.

## Confirmed Decisions

### Product and launch

- Internal prototype for the business's own Republic of Ireland TikTok Shop and warehouse.
- Codex will generate the initial codebase.
- Preserve a path to multi-tenant SaaS without building SaaS onboarding or billing now.
- Initially support one company, shop, and warehouse; UK and additional channels are deferred.
- Deliver as soon as practical using capability-based milestones.

### Scale

- Approximately 3,000 orders/month and a few hundred/day.
- Fewer than 10,000 SKUs, 5–6 concurrent warehouse users, and up to 50 orders per batch.

### Technology

- DigitalOcean EU region, provisionally Amsterdam.
- Backend and web server: Go.
- User interface: server-rendered HTML enhanced with htmx.
- Styling: Fomantic UI's CSS-only distribution plus a small application-specific stylesheet.
- Do not use Tailwind CSS, Semantic UI, jQuery, or Fomantic's JavaScript components.
- Prefer native HTML controls and dialogs. Use focused JavaScript only where browser APIs or hardware integrations require it, such as barcode scanning and label printing.
- PostgreSQL from the outset; no SQLite application phase.
- Docker Compose for repeatable local and hosted environments.
- The internal prototype may run the app, worker, PostgreSQL, and Redis on one Droplet.
- Codex may run on the private prototype/development server.
- Version 1 gets a clean production environment and should use Managed PostgreSQL before becoming operationally critical or serving external customers.
- Responsive web only; no native mobile apps.

### Access

- Roles: Admin and Warehouse User.
- Admin has full access.
- Warehouse Users may perform audited stock adjustments and operate purchase-order and receiving workflows.
- Warehouse Users cannot access credentials, users, costs/margins, exports, audit logs, integration configuration, or retention settings.
- Warehouse Users may view customer details required for fulfilment; exports are Admin-only.

### Products and inventory

- This application is the inventory authority.
- Support ordinary SKUs and virtual bundles assembled from components at pick time.
- Exclude lots, serial numbers, expiry dates, and specialist stock types.
- Availability is `physical − reserved − damaged − safety buffer`.
- Safety buffer defaults globally to zero with per-SKU overrides.
- Reservations are all-or-nothing per order; shortages block the whole order.
- Store purchase cost per SKU and snapshot it when importing each order.
- Initial reporting uses estimated gross product margin: sales revenue minus purchase cost.

### Warehouse

- Warehouse Users create batches; oldest eligible orders are selected first with optional filters.
- Starting a batch locks it to that picker.
- Admin creates locations, reusable trolleys, packing stations, and printer assignments from scratch.
- Locations have unique codes/barcodes and types such as receiving, storage/pick, packing/staging, returns, and damaged/quarantine.
- Trolleys have unique codes/barcodes and active/out-of-service states.
- Claiming a batch requires selecting or scanning an available trolley; one trolley may hold only one active batch.
- Products are bulk-picked into a mixed trolley without order-specific totes.
- Every item is scanned at packing.
- The first scan selects a matching order; all remaining items for it are scanned before despatch.
- Packing slips are excluded; multi-quantity, multi-SKU, and multi-parcel orders are supported.
- Shortages require a visible manual-resolution workflow.

### Shipping

- Prototype shipping uses a deterministic mock with DPD and An Post test-service options.
- It generates clearly non-shippable test labels/tracking after packing and supports multiple parcels.
- It must make no Shipster/carrier calls or postage purchases and is forbidden in production mode.
- Live Shipster and Zebra ZD421 printing are deferred.
- Default parcel quantity is 1 unit and default weight is 300 g; SKU plus packaging weight is preferred later.

### Purchasing, returns, and reporting

- Purchase orders and goods receipt are required; automated reorder recommendations are deferred.
- Scan returns against the original order/TikTok return and classify as restockable or damaged.
- Refund decisions remain separate from inventory disposition; automatic refunds are provisionally deferred.
- CSV exports: orders, order lines, products, stock balances, stock movements, returns, shipments, and margin.
- Provide an in-app exception centre and configurable critical email recipients.
- Proposed pilot targets: 99.5% correct picks, 99% of labels without manual re-entry, inventory variance below 1%, no duplicate despatches, and measurably faster fulfilment than Peoplevox.

## Remaining Clarifications

- Decide whether opening TikTok stock is trusted provisionally or must wait for a physical count. The recommended approach is to import it as provisional stock, then reconcile it through an audited opening count before relying on it for fulfilment.
- Supply real warehouse, trolley, packing-station, supplier, purchase-order, bundle, and anonymised order fixtures when available.

## Inputs Needed

- Application-side TikTok OAuth/token lifecycle and verified webhook subscriptions.
- 20–50 anonymised orders covering normal and exceptional workflows.
- At least two bundle definitions with component quantities.
- Current supplier and purchase-order fields/workflow.
- Warehouse layout/location list, trolley identifiers, packing stations, and printer assignments.
- DigitalOcean account and prototype Droplet.

Do not place production secrets in this document or commit them to Git.

## Deferred

- UK rollout and additional shops, channels, warehouses, or 3PLs.
- Public SaaS onboarding, tenant administration, billing, pricing, trials, SLA, and customer API.
- Native apps, live Shipster/carrier integration, and Zebra production printing.
- Automated reordering, advanced forecasting, and full profitability/accounting integrations.
- Final branding, domain, legal terms, privacy notice, and external support model.
