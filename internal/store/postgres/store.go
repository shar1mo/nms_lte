package postgres

import (
	"context"
	"errors"
	"fmt"

	"nms_lte/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

var ConnString = "postgres://nms_user:nms_password@localhost:5432/nms_lte?sslmode=disable"

func New(connString string) (*Store, error) {
	db, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) SaveNE(ne model.NetworkElement) error {
	query := `
		INSERT INTO network_elements (
			id, name, address, vendor, status, capabilities, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE
		SET	name = EXCLUDED.name,
				address = EXCLUDED.address,
				vendor = EXCLUDED.vendor,
				status = EXCLUDED.status,
				capabilities = EXCLUDED.capabilities,
				created_at = EXCLUDED.created_at,
				updated_at = EXCLUDED.updated_at
	`

	if ne.Capabilities == nil {
		ne.Capabilities = []string{}
	}

	_, err := s.db.Exec(context.Background(),
		query,
		ne.ID,
		ne.Name,
		ne.Address,
		ne.Vendor,
		ne.Status,
		ne.Capabilities,
		ne.CreatedAt,
		ne.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteNE(id string) error {
	query := `
		DELETE FROM network_elements WHERE id = $1;
	`
	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetNE(id string) (model.NetworkElement, bool, error) {
	query := `
		SELECT id, name, address, vendor, status, capabilities, created_at, updated_at
		FROM network_elements
		WHERE id = $1;
	`

	var ne model.NetworkElement

	err := s.db.QueryRow(context.Background(), query, id).Scan(
		&ne.ID,
		&ne.Name,
		&ne.Address,
		&ne.Vendor,
		&ne.Status,
		&ne.Capabilities,
		&ne.CreatedAt,
		&ne.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.NetworkElement{}, false, nil
		}
		return model.NetworkElement{}, false, fmt.Errorf("get network element id=%s: %w", id, err)
	}

	return ne, true, nil
}

func (s *Store) ListNE() ([]model.NetworkElement, error) {
	query := `
		SELECT id, name, address, vendor, status, capabilities, created_at, updated_at
		FROM network_elements
		ORDER BY created_at
	`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.NetworkElement
	for rows.Next() {
		var ne model.NetworkElement
		err := rows.Scan(
			&ne.ID,
			&ne.Name,
			&ne.Address,
			&ne.Vendor,
			&ne.Status,
			&ne.Capabilities,
			&ne.CreatedAt,
			&ne.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, ne)
	}
	return out, nil
}

func (s *Store) SaveInventorySnapshot(snapshot model.InventorySnapshot) error {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	queryInvent := `
	INSERT INTO inventory_snapshots (
		id, ne_id, synced_at
	) VALUES ($1, $2, $3)
	ON CONFLICT (id) DO UPDATE
	SET synced_at = EXCLUDED.synced_at
	`

	queryObj := `
	INSERT INTO inventory_objects (
		snapshot_id, dn, class, attributes
	) VALUES ($1, $2, $3, $4)
	`

	_, err = tx.Exec(context.Background(),
		queryInvent,
		snapshot.ID,
		snapshot.NEID,
		snapshot.SyncedAt,
	)

	if err != nil {
		return err
	}

	for _, obj := range snapshot.Objects {
		_, err := tx.Exec(context.Background(),
			queryObj,
			snapshot.ID,
			obj.DN,
			obj.Class,
			obj.Attributes,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(context.Background())
}

func (s *Store) GetLatestInventorySnapshot(neID string) (model.InventorySnapshot, error) {
	query := `
	SELECT id, ne_id, synced_at 
	FROM inventory_snapshots
	WHERE ne_id = $1
	ORDER BY synced_at DESC 
	LIMIT 1;
	`

	queryObj := `
	SELECT dn, class, attributes
	FROM inventory_objects
	WHERE snapshot_id = $1;
	`

	var inventory model.InventorySnapshot
	err := s.db.QueryRow(context.Background(), query, neID).Scan(
		&inventory.ID,
		&inventory.NEID,
		&inventory.SyncedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.InventorySnapshot{}, nil
		}
		return model.InventorySnapshot{}, err
	}

	rows, err := s.db.Query(context.Background(), queryObj, inventory.ID)
	if err != nil {
		return model.InventorySnapshot{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var obj model.InventoryObject
		var attrs map[string]string
		err := rows.Scan(&obj.DN, &obj.Class, &attrs)
		if err != nil {
			return model.InventorySnapshot{}, err
		}
		obj.Attributes = attrs
		inventory.Objects = append(inventory.Objects, obj)
	}

	if err := rows.Err(); err != nil {
		return model.InventorySnapshot{}, err
	}

	return inventory, nil
}

func (s *Store) SaveCMRequest(req model.CMRequest) error {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	queryRequest := `
	INSERT INTO cm_requests (
		id, ne_id, parameter, value, status, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (id) DO UPDATE
	SET created_at = EXCLUDED.created_at,
			updated_at = EXCLUDED.updated_at
	`

	querySteps := `
	INSERT INTO cm_steps (
		request_id, name, status, message, created_at
	) VALUES ($1, $2, $3, $4, $5)
	`

	_, err = tx.Exec(context.Background(),
		queryRequest,
		req.ID,
		req.NEID,
		req.Parameter,
		req.Value,
		req.Status,
		req.CreatedAt,
		req.UpdatedAt,
	)
	if err != nil {
		return err
	}

	for _, step := range req.Steps {
		_, err := tx.Exec(context.Background(),
			querySteps,
			req.ID,
			step.Name,
			step.Status,
			step.Message,
			step.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(context.Background())
}

func (s *Store) ListCMRequests() ([]model.CMRequest, error) {
	query := `
	SELECT id, ne_id, parameter, value, status, created_at, updated_at
	FROM cm_requests
	ORDER BY created_at
	`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	var out []model.CMRequest
	for rows.Next() {
		var req model.CMRequest
		err := rows.Scan(
			&req.ID,
			&req.NEID,
			&req.Parameter,
			&req.Value,
			&req.Status,
			&req.CreatedAt,
			&req.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, req)
	}
	return out, nil
}

func (s *Store) GetCMRequest(id string) (model.CMRequest, error) {
	query := `
	SELECT id, ne_id, parameter, value, status, created_at, updated_at
	FROM cm_requests
	WHERE id = $1;
	`

	var req model.CMRequest
	
	err := s.db.QueryRow(context.Background(), query, id).Scan(
		&req.ID,
		&req.NEID,
		&req.Parameter,
		&req.Value,
		&req.Status,
		&req.CreatedAt,
		&req.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.CMRequest{}, nil
		}
		return model.CMRequest{}, fmt.Errorf("get cm request id=%s: %w", id, err)
	}

	return req, nil
}

func (s *Store) CreateUser(ctx context.Context, user model.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := s.db.Exec(ctx,
		query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
	)

	return err
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`

	var user model.User
	err := s.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (model.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at
		FROM users
		WHERE username = $1
	`

	var user model.User
	err := s.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (s *Store) AddFaultEvent(event model.FaultEvent) {
}

func (s *Store) AddPMSample(sample model.PMSample) {
}

func (s *Store) ListPMSamples(neID, metric string, limit int) []model.PMSample {

	return nil
}

func (s *Store) SaveHeartbeat(hb model.HeartbeatStatus) {
}

func (s *Store) GetHeartbeat(neID string) (model.HeartbeatStatus, bool) {
	return model.HeartbeatStatus{}, false
}

func (s *Store) ListFaultEvents(neID string) []model.FaultEvent {
	return nil
}
