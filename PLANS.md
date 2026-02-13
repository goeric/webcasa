<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# Finance Detail UX Design (Issue #93)

## Scope

Minimal finance tracking for dogfooding: quote acceptance + payments.

### Schema

1. **Quote** -- add `AcceptedAt *time.Time`. Toggle marks the winning bid.
   Only one quote per project can be accepted at a time.

2. **ProjectPayment** -- new entity:
   - `ProjectID uint` (required)
   - `VendorID *uint` (optional)
   - `AmountCents int64`
   - `PaidAt time.Time`
   - `Method string` (check/card/transfer/cash)
   - `Reference string` (check #, confirmation code)
   - `Notes string`

### UX

- Projects tab: add `Pay` drilldown column (count of payments)
- Quotes: add `Accepted` column showing date; toggle via inline edit
- Payment detail view: ID, Vendor (→Vendors), Amount, Date, Method, Ref, Notes
- Dashboard: show paid total alongside budget in project rows

### Not included (future)

- Invoice entity (add when outstanding-balance tracking is needed)
- Auto-compute Actual from payments (keep manual for now)
