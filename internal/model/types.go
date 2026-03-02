package model

import "time"

type NetworkElement struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Address      string    `json:"address"`
	Vendor       string    `json:"vendor"`
	Status       string    `json:"status"`
	Capabilities []string  `json:"capabilities"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type InventoryObject struct {
	DN         string            `json:"dn"`
	Class      string            `json:"class"`
	Attributes map[string]string `json:"attributes"`
}

type InventorySnapshot struct {
	ID       string            `json:"id"`
	NEID     string            `json:"ne_id"`
	SyncedAt time.Time         `json:"synced_at"`
	Objects  []InventoryObject `json:"objects"`
}

type CMStep struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type CMRequest struct {
	ID        string    `json:"id"`
	NEID      string    `json:"ne_id"`
	Parameter string    `json:"parameter"`
	Value     string    `json:"value"`
	Status    string    `json:"status"`
	Steps     []CMStep  `json:"steps"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FaultEvent struct {
	ID        string    `json:"id"`
	NEID      string    `json:"ne_id"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type HeartbeatStatus struct {
	NEID      string    `json:"ne_id"`
	Healthy   bool      `json:"healthy"`
	CheckedAt time.Time `json:"checked_at"`
}

type PMSample struct {
	ID          string    `json:"id"`
	NEID        string    `json:"ne_id"`
	Metric      string    `json:"metric"`
	Value       float64   `json:"value"`
	CollectedAt time.Time `json:"collected_at"`
}
