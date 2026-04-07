CREATE TABLE IF NOT EXISTS network_elements (
    id TEXT PRIMARY KEY, -- NE ID -- like enb_001? 
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    vendor TEXT,
    status TEXT NOT NULL,
    capabilities TEXT[],
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- Inventory
CREATE TABLE IF NOT EXISTS inventory_snapshots (
    id TEXT PRIMARY KEY, -- Snapshot ID -- like snapshot_001 or ...?
    ne_id TEXT NOT NULL,
    synced_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (ne_id) REFERENCES network_elements(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS inventory_objects (
    id SERIAL PRIMARY KEY, -- like inventory_snapshot_id or ...?
    snapshot_id TEXT NOT NULL,
    dn TEXT NOT NULL,
    class TEXT NOT NULL,
    attributes JSONB,
    FOREIGN KEY (snapshot_id) REFERENCES inventory_snapshots(id) ON DELETE CASCADE
);

-- CM
CREATE TABLE IF NOT EXISTS cm_requests (
    id TEXT PRIMARY KEY, -- config id or request id -- like req_001 or ...?
    ne_id TEXT NOT NULL,
    parameter TEXT NOT NULL,
    value TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (ne_id) REFERENCES network_elements(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS cm_steps (
    id SERIAL PRIMARY KEY,
    request_id TEXT NOT NULL,
    name TEXT NOT NULL,
    status TEXT NOT NULL,
    message TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (request_id) REFERENCES cm_requests(id) ON DELETE CASCADE
);

-- Fault
CREATE TABLE IF NOT EXISTS fault_events (
    id TEXT PRIMARY KEY, -- like fault_001 or id or timestamp or ...?
    ne_id TEXT NOT NULL,
    severity TEXT NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (ne_id) REFERENCES network_elements(id) ON DELETE CASCADE
);

-- Heartbeat
CREATE TABLE IF NOT EXISTS heartbeats (
    ne_id TEXT PRIMARY KEY, 
    healthy BOOLEAN NOT NULL,
    checked_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (ne_id) REFERENCES network_elements(id) ON DELETE CASCADE
);

-- PM
CREATE TABLE IF NOT EXISTS pm_samples (
    id TEXT PRIMARY KEY,
    ne_id TEXT NOT NULL,
    metric TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    collected_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (ne_id) REFERENCES network_elements(id) ON DELETE CASCADE
);

-- Index
CREATE INDEX IF NOT EXISTS idx_fault_ne_id ON fault_events(ne_id);
CREATE INDEX IF NOT EXISTS idx_pm_ne_metric ON pm_samples(ne_id, metric);
CREATE INDEX IF NOT EXISTS idx_inventory_ne ON inventory_snapshots(ne_id);
CREATE INDEX IF NOT EXISTS idx_cm_ne ON cm_requests(ne_id);