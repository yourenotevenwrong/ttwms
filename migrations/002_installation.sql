-- The prototype has one company and warehouse. Keep their identities explicit
-- so every setup query and audit event has a stable company/warehouse boundary.
CREATE TABLE app_installation (
  singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
  company_id uuid NOT NULL REFERENCES companies(id),
  warehouse_id uuid NOT NULL REFERENCES warehouses(id)
);

WITH company AS (
  INSERT INTO companies(name) VALUES ('My company') RETURNING id
), warehouse AS (
  INSERT INTO warehouses(company_id, code, name)
  SELECT id, 'MAIN', 'Main warehouse' FROM company
  RETURNING id, company_id
)
INSERT INTO app_installation(company_id, warehouse_id)
SELECT company_id, id FROM warehouse;
