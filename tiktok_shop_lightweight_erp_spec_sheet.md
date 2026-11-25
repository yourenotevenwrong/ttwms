NOTICE: This version 0.1 vision document has been superseded by
tiktok_shop_commerce_os_spec_v0.2.md. Retained for historical reference.

SPEC SHEET — TikTok-First Commerce Operating System
Version: 0.1 MVP
Target market: UK TikTok Shop sellers doing roughly 500–25,000+ orders/month

1. PRODUCT VISION
A low-cost, lightweight back-office platform for TikTok Shop sellers who have outgrown TikTok Seller Centre but do not need a traditional ecommerce website + ERP.

Core flow:
TikTok Shop → Orders → Inventory → Pick/Pack → Shipping → Returns → Reporting

Initial principle:
Do not build a full ERP. Build the 20% of ERP functionality that solves 80% of the operational problems of high-volume TikTok sellers.

2. TARGET CUSTOMER
Primary:
- UK TikTok Shop merchants
- Creators/brands selling their own products
- Approximately 500–25,000+ orders/month
- 10–2,000 SKUs
- Own warehouse or 3PL
- Small operational teams
- Businesses currently using spreadsheets, Seller Centre, or overly complex ERP systems

Secondary:
- Multi-TikTok-Shop operators
- TikTok-first brands
- Amazon/TikTok multichannel sellers
- Agencies managing multiple merchants

3. MVP MODULES

A. USER & COMPANY MANAGEMENT
- Account registration/login
- Company profile
- User roles
- Admin / manager / warehouse user
- Basic permissions
- Subscription/plan management

B. TIKTOK SHOP CONNECTION
- Connect TikTok Shop using TikTok OAuth
- Secure access-token management
- Import shop information
- Import products/SKUs/variants
- Import orders
- Import order status
- Import inventory
- Import shipping information
- Receive relevant TikTok webhooks/events
- Push inventory updates back to TikTok
- Push fulfilment/tracking updates
- Connection health/status dashboard
- Reconnect/re-authorise workflow

Important:
The application should use TikTok's approved APIs/OAuth and comply with TikTok Shop developer requirements. Never request or store TikTok passwords.

C. PRODUCT & SKU MANAGEMENT
Product fields:
- Product ID
- SKU
- Variant ID
- Product name
- Variant name
- Barcode
- Supplier
- Cost price
- Selling price
- Weight
- Dimensions
- Tax/VAT information
- Active/inactive status
- Reorder level
- Lead time
- Warehouse location

Functions:
- Search/filter products
- Bulk import/export
- SKU mapping
- Barcode management
- Cost updates
- Stock adjustments
- Product status
- Audit history

D. INVENTORY ENGINE
Core inventory concepts:
- Physical stock
- Available stock
- Reserved/committed stock
- Allocated stock
- Damaged stock
- Incoming stock
- Safety/buffer stock

Functions:
- Real-time stock ledger
- Stock adjustments
- Stock movement history
- Inventory reservations
- Low-stock alerts
- Reorder alerts
- Stock allocation rules
- Inventory sync to TikTok
- Prevent overselling where possible
- Multi-warehouse support architecture
- Inventory audit trail

Example:
Physical stock = 2,000
Reserved = 400
Safety buffer = 200
Available to sell = 1,400

E. ORDER MANAGEMENT
Dashboard:
- Total orders
- New
- Awaiting payment
- Awaiting fulfilment
- Picking
- Packed
- Shipped
- Delivered
- Cancelled
- Returned
- Exception

Functions:
- Import orders automatically
- Search orders
- Filter by status
- Filter by SKU
- Filter by date
- Filter by customer/order ID
- Order detail page
- Batch order processing
- Order notes
- Order exception handling
- Order audit history

F. WAREHOUSE / FULFILMENT
Keep this deliberately simple.

Workflow:
Order received
→ Pick list
→ Pick
→ Scan SKU/barcode
→ Pack
→ Shipping label/tracking
→ Mark shipped

Functions:
- Pick lists
- Batch picking
- Barcode scanning
- Packing confirmation
- Short-pick handling
- Stock discrepancy handling
- Shipping/tracking information
- Warehouse user interface optimised for desktop/mobile

G. PURCHASING
Basic ERP-style purchasing without building a full ERP.

Functions:
- Supplier records
- Supplier SKU mapping
- Supplier cost
- Lead time
- Minimum order quantity
- Reorder levels
- Suggested purchase orders
- Purchase order creation
- Purchase order status
- Goods received
- Incoming stock
- Partial deliveries

Suggested purchasing logic:
Forecast demand from recent sales
+ supplier lead time
+ safety stock
- available stock
- incoming stock
= suggested purchase quantity

H. RETURNS
- Import returns
- Return status
- Return reason
- Customer/order/SKU linkage
- Return received
- Inspection status
- Restock / damaged decision
- Inventory adjustment
- Refund status
- Return reporting

I. SHIPPING
MVP:
- Store shipping/tracking information
- Integrate with one or more shipping providers where commercially justified
- Batch fulfilment
- Tracking updates
- Shipping exceptions

Architecture should allow future integrations with:
- Royal Mail
- DPD
- Evri
- DHL
- Other UK carriers
- 3PLs

J. REPORTING & PROFITABILITY
Core dashboard:
- Gross sales
- Orders
- Units sold
- Average order value
- TikTok fees
- Product cost
- Shipping cost
- Returns
- Estimated contribution profit
- Profit margin

Views:
- By day/week/month
- By SKU
- By product
- By category
- By shop
- By warehouse
- By creator/affiliate where TikTok data permits
- Best sellers
- Slow sellers
- Low-margin products

K. ALERTS
- Low stock
- Out of stock
- Failed TikTok connection
- Order sync failure
- Inventory sync failure
- Fulfilment exception
- High return rate
- Purchase order overdue
- Stock discrepancy

4. CORE DASHBOARD

Home screen should immediately answer:

1. How much did I sell?
2. How many orders need attention?
3. What needs to be shipped?
4. What products are running out?
5. How much money am I making?
6. Is anything broken?

Example:
Today's sales: £18,420
Orders: 742
To fulfil: 381
Low-stock SKUs: 14
Returns: 23
Estimated contribution profit: £4,280
Sync issues: 0

5. DATA MODEL — INITIAL

Company
User
Role
Subscription
TikTokShop
TikTokProduct
Product
SKU
Variant
Warehouse
WarehouseLocation
InventoryBalance
InventoryTransaction
SalesOrder
SalesOrderLine
Customer
Fulfilment
Shipment
Return
ReturnLine
Supplier
SupplierSKU
PurchaseOrder
PurchaseOrderLine
StockReceipt
WebhookEvent
IntegrationLog
AuditLog

6. SYSTEM ARCHITECTURE

Recommended initial architecture:

Frontend:
- React / Next.js
- Responsive web application

Backend:
- Node.js / TypeScript or Python
- REST API initially
- Background job system for sync/processing

Database:
- PostgreSQL

Cache/queues:
- Redis

File/object storage:
- S3-compatible storage

Authentication:
- Secure email/password authentication
- Optional Google/Microsoft login later
- Role-based access control

Infrastructure:
- AWS / Azure / Google Cloud
- Containerised deployment
- Separate development/staging/production environments

7. ASYNCHRONOUS PROCESSING

Do not process large TikTok synchronisation jobs directly inside normal web requests.

Use background workers for:
- Order imports
- Product imports
- Inventory synchronisation
- Webhook processing
- Bulk updates
- Reporting calculations
- Shipping updates
- Notifications

Every integration job should have:
- Job ID
- Status
- Start/end time
- Retry count
- Error message
- Last successful sync

8. TIKTOK INTEGRATION REQUIREMENTS

The integration layer should be isolated from the rest of the application.

TikTok Adapter:
- OAuth
- Token refresh
- Shop authorisation
- Product API
- Order API
- Inventory API
- Fulfilment/shipping API
- Returns/refund API where available
- Webhooks/events
- Rate-limit handling
- Retry logic
- Idempotency
- API version management

Do not tightly couple business logic directly to TikTok API response formats.

Instead:

TikTok API
→ TikTok Adapter
→ Internal Commerce Model
→ Business Logic

This makes future Amazon/Shopify/etc. integrations much easier.

9. MULTI-CHANNEL ARCHITECTURE

Do not necessarily launch with multiple channels, but design the data model for them.

Channel abstraction:

Channel
ChannelAccount
ChannelProduct
ChannelOrder
ChannelInventory
ChannelFulfilment

Initial channel:
- TikTok Shop

Potential future channels:
- Amazon
- Shopify
- WooCommerce
- eBay
- Instagram/Meta
- Other marketplaces

10. SECURITY

Minimum requirements:
- HTTPS everywhere
- Encryption at rest where appropriate
- Secure secrets/token storage
- OAuth token encryption
- Role-based access
- Audit logging
- Rate limiting
- Input validation
- CSRF protection where applicable
- Secure webhook validation
- Database backups
- Monitoring
- Error logging
- GDPR-conscious data handling
- Data retention/deletion policies

11. PERFORMANCE TARGETS

Initial targets:
- Support 25,000+ orders/month/customer
- Support 100,000+ historical orders/customer
- Product search <1 second under normal load
- Dashboard <2 seconds under normal load
- Inventory updates processed asynchronously
- Webhook processing designed for burst traffic
- Automatic retries for transient API failures

Architecture should be capable of scaling beyond these numbers without a complete rewrite.

12. MVP EXCLUSIONS

Do NOT initially build:
- Ecommerce website builder
- Checkout
- CMS
- Blog
- CRM
- Loyalty programme
- Email marketing
- Full accounting system
- Payroll
- Full enterprise WMS
- Complex B2B pricing
- Hundreds of marketplace integrations
- Advanced demand forecasting
- AI features unless they solve a clearly demonstrated customer problem

13. DIFFERENTIATING FEATURES

The product should compete on:

1. TikTok-first
2. Extremely simple onboarding
3. Low monthly cost
4. No ecommerce website required
5. No ERP implementation required
6. Centralised inventory
7. Simple purchasing
8. Simple warehouse workflow
9. Clear profitability
10. Designed for high-volume TikTok operations

Core positioning:

"You've outgrown TikTok Seller Centre.
You don't need an ERP.
You need a simple commerce operating system."

14. INITIAL PRICING HYPOTHESIS

Possible pricing:

Free:
- Up to 500 orders/month
- 1 TikTok Shop
- Basic inventory
- Basic order management

Growth:
- £49/month
- Up to 2,500 orders
- Advanced inventory
- Purchasing
- Basic reporting

Pro:
- £99/month
- Up to 10,000 orders
- Warehouse/fulfilment
- Advanced reporting
- Multiple shops

Scale:
- £199/month
- Up to 25,000 orders
- Multiple warehouses
- API
- Advanced controls

Enterprise:
- £399+/month
- Higher volumes
- Dedicated support
- Custom integrations
- SLA

Pricing should be validated through customer interviews before development.

15. MVP SUCCESS CRITERIA

The MVP is successful if a TikTok seller can:

1. Create an account
2. Connect TikTok Shop
3. Import their catalogue
4. See accurate inventory
5. Receive orders automatically
6. Reserve stock
7. Generate pick lists
8. Pick/pack orders
9. Record tracking
10. Synchronise fulfilment
11. Handle returns
12. Create purchase orders
13. See basic profitability
14. Receive low-stock alerts

without needing:
- Shopify
- BigCommerce
- Magento
- A traditional ERP
- Spreadsheet-based order management

16. VALIDATION BEFORE FULL BUILD

Before committing to a large development budget:

Interview 20–30 UK TikTok Shop sellers.

Target:
- 1,000+ monthly orders
- Ideally 5,000+ monthly orders

Ask:
- What software do you use?
- What does it cost?
- What do you still manage in Excel?
- What breaks at high volume?
- How do you manage inventory?
- How do you purchase stock?
- How do you fulfil orders?
- How do you handle returns?
- What reports do you need?
- Why haven't you adopted an ERP?
- What would you pay monthly for a simpler system?

Build only the problems that appear repeatedly.

17. RECOMMENDED DEVELOPMENT PHASES

Phase 0 — Validation
- Customer interviews
- Competitor analysis
- TikTok API feasibility
- TikTok partner requirements
- Technical proof of concept

Phase 1 — TikTok MVP
- Authentication
- TikTok connection
- Product sync
- Order sync
- Inventory
- Basic dashboard

Phase 2 — Operations
- Picking
- Packing
- Shipping
- Returns
- Barcode scanning

Phase 3 — ERP-lite
- Suppliers
- Purchasing
- Purchase orders
- Goods received
- Reorder recommendations

Phase 4 — Commercial intelligence
- Profitability
- Advanced reporting
- Creator/affiliate reporting where supported
- Product performance

Phase 5 — Multichannel
- Amazon
- Shopify
- Additional channels

18. THE CORE PRODUCT LOOP

The entire system should ultimately revolve around:

SELL
↓
ORDER
↓
RESERVE STOCK
↓
PICK
↓
PACK
↓
SHIP
↓
RECONCILE
↓
ANALYSE PROFIT
↓
FORECAST STOCK
↓
PURCHASE
↓
REPEAT

19. STRATEGIC GOAL

Do not attempt to become another SAP, NetSuite, Shopify or traditional ERP.

Own the space between:

"TikTok Seller Centre is no longer enough"

and

"A full ecommerce/ERP stack is far too expensive and complicated."

That is the market gap this product should target.
