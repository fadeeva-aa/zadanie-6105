package storage

import (
	"context"
	"os"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	conn *pgxpool.Pool
}

func NewStorage(ctx context.Context) (*Storage, error) {

	conn, err := pgxpool.New(ctx, os.Getenv("POSTGRES_CONN"))
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	if err := create(ctx, conn); err != nil {
		return nil, err
	}

	return &Storage{
		conn: conn,
	}, nil

}

func (s *Storage) Ping(ctx context.Context) error {
	return s.conn.Ping(ctx)
}

func (s *Storage) Close() {
	s.conn.Close()
}

func (s *Storage) userId(ctx context.Context, username string) (uuid.UUID, error) {

	id := uuid.UUID{}
	query := `SELECT id FROM employee WHERE username = $1;`
	if err := s.conn.QueryRow(ctx, query, username).Scan(&id); err != nil {
		return uuid.UUID{}, ErrIncorrectUser
	}

	return id, nil

}

func (s *Storage) userOrgId(ctx context.Context, userId uuid.UUID) (uuid.UUID, error) {

	id := uuid.UUID{}
	query := `SELECT organization_id FROM organization_responsible WHERE user_id = $1;`
	if err := s.conn.QueryRow(ctx, query, userId).Scan(&id); err != nil {
		return uuid.UUID{}, err
	}

	return id, nil

}

func (s *Storage) checkUser(ctx context.Context, id uuid.UUID) error {

	res := 0
	query := `SELECT 1 FROM employee WHERE id = $1;`
	if err := s.conn.QueryRow(ctx, query, id).Scan(&res); err != nil {
		return ErrIncorrectUser
	}

	return nil

}

func (s *Storage) checkRelationToOrganization(ctx context.Context, userId, orgId uuid.UUID) bool {

	res := 0
	query := `SELECT 1 FROM organization_responsible WHERE user_id = $1 AND organization_id = $2;`
	err := s.conn.QueryRow(ctx, query, userId, orgId).Scan(&res)
	return err == nil

}

func (s *Storage) quorum(ctx context.Context, orgId uuid.UUID) (int, error) {

	res := 0
	query := `SELECT COUNT(*) FROM organization_responsible WHERE organization_id = $1;`
	if err := s.conn.QueryRow(ctx, query, orgId).Scan(&res); err != nil {
		return 0, err
	}

	return min(3, res), nil

}
