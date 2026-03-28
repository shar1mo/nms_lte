package postgres

import (
	"context"
	"errors"
	"nms_lte/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
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

func (s *Store) SaveNE (ne model.NetworkElement) error {
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
		return model.NetworkElement{}, false, err
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