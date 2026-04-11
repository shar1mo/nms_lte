CREATE TABLE IF NOT EXISTS network_elements (
    id TEXT PRIMARY KEY, -- generated in application (e.g. ne-...)
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
    id TEXT PRIMARY KEY, -- generated in application (e.g. inv-...)
    ne_id TEXT NOT NULL,
    synced_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (ne_id) REFERENCES network_elements(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS inventory_objects (  
    id SERIAL PRIMARY KEY, -- internal ID (auto-increment), not used outside DB
    snapshot_id TEXT NOT NULL,
    dn TEXT NOT NULL,
    class TEXT NOT NULL,
    attributes JSONB,
    FOREIGN KEY (snapshot_id) REFERENCES inventory_snapshots(id) ON DELETE CASCADE
);

-- CM
CREATE TABLE IF NOT EXISTS cm_requests (
    id TEXT PRIMARY KEY, -- generated in application (e.g. req-...)
    ne_id TEXT NOT NULL,
    parameter TEXT NOT NULL,
    value TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (ne_id) REFERENCES network_elements(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS cm_steps (
    id SERIAL PRIMARY KEY, -- internal ID (auto-increment), not used outside DB
    request_id TEXT NOT NULL,
    name TEXT NOT NULL,
    status TEXT NOT NULL,
    message TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (request_id) REFERENCES cm_requests(id) ON DELETE CASCADE
);

-- Fault
CREATE TABLE IF NOT EXISTS fault_events (
    id TEXT PRIMARY KEY, -- generated in application (e.g. fault-...)
    ne_id TEXT NOT NULL,
    severity TEXT NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (ne_id) REFERENCES network_elements(id) ON DELETE CASCADE
);

-- Heartbeat
CREATE TABLE IF NOT EXISTS heartbeats (
    ne_id TEXT PRIMARY KEY, -- one heartbeat state per network element
    healthy BOOLEAN NOT NULL,
    checked_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (ne_id) REFERENCES network_elements(id) ON DELETE CASCADE
);

-- PM
CREATE TABLE IF NOT EXISTS pm_samples (
    id TEXT PRIMARY KEY, -- generated in application (e.g. pm-...)
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