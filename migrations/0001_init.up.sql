CREATE TABLE network_elements (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    vendor TEXT,
    status TEXT NOT NULL,
    capabilities TEXT[],
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Inventory
CREATE TABLE inventory_snapshots (
    id TEXT PRIMARY KEY,
    ne_id TEXT NOT NULL,
    synced_at TIMESTAMP NOT NULL,
    FOREIGN KEY (ne_id) REFERENCES network_elements(id) ON DELETE CASCADE
);

CREATE TABLE inventory_objects (
    id SERIAL PRIMARY KEY,
    snapshot_id TEXT NOT NULL,
    dn TEXT NOT NULL,
    class TEXT NOT NULL,
    attributes JSONB,
    FOREIGN KEY (snapshot_id) REFERENCES inventory_snapshots(id) ON DELETE CASCADE
);

-- CM
CREATE TABLE cm_requests (
    id TEXT PRIMARY KEY,
    ne_id TEXT NOT NULL,
    parameter TEXT NOT NULL,
    value TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (ne_id) REFERENCES network_elements(id) ON DELETE CASCADE
);

CREATE TABLE cm_steps (
    id SERIAL PRIMARY KEY,
    request_id TEXT NOT NULL,
    name TEXT NOT NULL,
    status TEXT NOT NULL,
    message TEXT,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (request_id) REFERENCES cm_requests(id) ON DELETE CASCADE
);

-- Fault
CREATE TABLE fault_events (
    id TEXT PRIMARY KEY,
    ne_id TEXT NOT NULL,
    severity TEXT NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (ne_id) REFERENCES network_elements(id) ON DELETE CASCADE
);

-- Heartbeat
CREATE TABLE heartbeats (
    ne_id TEXT PRIMARY KEY,
    healthy BOOLEAN NOT NULL,
    checked_at TIMESTAMP NOT NULL,
    FOREIGN KEY (ne_id) REFERENCES network_elements(id) ON DELETE CASCADE
);

-- PM
CREATE TABLE pm_samples (
    id TEXT PRIMARY KEY,
    ne_id TEXT NOT NULL,
    metric TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    collected_at TIMESTAMP NOT NULL,
    FOREIGN KEY (ne_id) REFERENCES network_elements(id) ON DELETE CASCADE
);

-- Index
CREATE INDEX idx_fault_ne_id ON fault_events(ne_id);
CREATE INDEX idx_pm_ne_metric ON pm_samples(ne_id, metric);
CREATE INDEX idx_inventory_ne ON inventory_snapshots(ne_id);
CREATE INDEX idx_cm_ne ON cm_requests(ne_id);