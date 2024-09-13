package storage

import (
	"context"
	"fmt"
	"strings"
	"zadanie/model"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) tender(ctx context.Context, id uuid.UUID) (model.Tender, error) {

	query := `SELECT * FROM tender WHERE id = $1`
	row, err := s.conn.Query(ctx, query, id)
	if err != nil {
		return model.Tender{}, err
	}

	tender, err := pgx.CollectOneRow(row, pgx.RowToStructByNameLax[model.Tender])
	if err != nil {
		return model.Tender{}, ErrTenderNotFound
	}

	return tender, nil

}

func (s *Storage) tenderVer(ctx context.Context, id uuid.UUID, ver int) (model.Tender, error) {

	query := `SELECT * FROM tender_archive WHERE id = $1 AND version = $2;`
	row, err := s.conn.Query(ctx, query, id, ver)
	if err != nil {
		return model.Tender{}, err
	}

	tender, err := pgx.CollectOneRow(row, pgx.RowToStructByNameLax[model.Tender])
	if err != nil {
		return model.Tender{}, ErrVersionNotFound
	}

	return tender, nil

}

func (s *Storage) CreateTender(ctx context.Context, tender model.Tender, username string) (model.Tender, error) {

	userId, err := s.userId(ctx, username)
	if err != nil {
		return model.Tender{}, err
	}

	if !s.checkRelationToOrganization(ctx, userId, tender.OrganizationId) {
		return model.Tender{}, ErrNotEnoughPerm
	}

	insert := `	INSERT INTO tender(name, description, type, status, organization_id)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING *;`

	row, err := s.conn.Query(ctx, insert, tender.Name, tender.Description, tender.ServiceType, model.TenderStatusCreated, tender.OrganizationId)
	if err != nil {
		return model.Tender{}, err
	}

	return pgx.CollectOneRow(row, pgx.RowToStructByNameLax[model.Tender])

}

func (s *Storage) ReadTenders(ctx context.Context, limit, offset int, types []model.TenderServiceType) ([]model.Tender, error) {

	query := `	SELECT * FROM tender WHERE status = 'Published' `

	if len(types) != 0 {

		parts := []string{}

		for _, t := range types {
			parts = append(parts, fmt.Sprintf("type = '%s' ", string(t)))
		}

		query += fmt.Sprintf("AND ( %s ) ", strings.Join(parts, "OR "))
	}

	query += `	ORDER BY name ASC LIMIT $1 OFFSET $2 ;`

	row, err := s.conn.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(row, pgx.RowToStructByNameLax[model.Tender])

}

func (s *Storage) ReadMyTenders(ctx context.Context, username string, limit, offset int) ([]model.Tender, error) {

	userId, err := s.userId(ctx, username)
	if err != nil {
		return nil, err
	}

	orgId, err := s.userOrgId(ctx, userId)
	if err != nil {
		return nil, err
	}

	row, err := s.conn.Query(ctx, `SELECT * FROM tender WHERE organization_id = $1 ORDER BY name ASC LIMIT $2 OFFSET $3;`, orgId, limit, offset)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(row, pgx.RowToStructByNameLax[model.Tender])

}

func (s *Storage) ReadTenderStatus(ctx context.Context, tenderId uuid.UUID, username string) (model.TenderStatus, error) {

	tender, err := s.tender(ctx, tenderId)
	if err != nil {
		return "", err
	}

	if tender.Status == model.TenderStatusPublished {
		return tender.Status, nil
	}

	userId, err := s.userId(ctx, username)
	if err != nil {
		return "", err
	}

	if !s.checkRelationToOrganization(ctx, userId, tender.OrganizationId) {
		return "", ErrNotEnoughPerm
	}

	return tender.Status, nil

}

func (s *Storage) UpdateTenderStatus(ctx context.Context, tenderId uuid.UUID, username string, status model.TenderStatus) (model.Tender, error) {

	tender, err := s.tender(ctx, tenderId)
	if err != nil {
		return model.Tender{}, err
	}

	userId, err := s.userId(ctx, username)
	if err != nil {
		return model.Tender{}, err
	}

	if !s.checkRelationToOrganization(ctx, userId, tender.OrganizationId) {
		return model.Tender{}, ErrNotEnoughPerm
	}

	if tender.Status == model.TenderStatusClosed {
		return model.Tender{}, ErrTenderClosed
	}

	update := `	UPDATE tender 
				SET status = $1, 
					version = version + 1, 
					updated_at = now()::timestamp without time zone
				WHERE id = $2 
				RETURNING *;`

	row, err := s.conn.Query(ctx, update, status, tenderId)
	if err != nil {
		return model.Tender{}, err
	}

	return pgx.CollectOneRow(row, pgx.RowToStructByNameLax[model.Tender])

}

func (s *Storage) UpdateTender(ctx context.Context, tenderId uuid.UUID, username string, new model.Tender) (model.Tender, error) {

	tender, err := s.tender(ctx, tenderId)
	if err != nil {
		return model.Tender{}, err
	}

	userId, err := s.userId(ctx, username)
	if err != nil {
		return model.Tender{}, err
	}

	if !s.checkRelationToOrganization(ctx, userId, tender.OrganizationId) {
		return model.Tender{}, ErrNotEnoughPerm
	}

	if tender.Status == model.TenderStatusClosed {
		return model.Tender{}, ErrTenderClosed
	}

	parts := []string{}
	if len(new.Name) != 0 {
		parts = append(parts, fmt.Sprintf("name = '%s' ", new.Name))
	}
	if len(new.Description) != 0 {
		parts = append(parts, fmt.Sprintf("description = '%s' ", new.Description))
	}
	if len(new.ServiceType) != 0 {
		parts = append(parts, fmt.Sprintf("type = '%s' ", new.ServiceType))
	}
	parts = append(parts, "version = version + 1 ")
	parts = append(parts, "updated_at = now()::timestamp without time zone ")

	update := fmt.Sprintf(`UPDATE tender SET %s WHERE id = $1 RETURNING *;`, strings.Join(parts, ","))

	row, err := s.conn.Query(ctx, update, tenderId)
	if err != nil {
		return model.Tender{}, err
	}

	return pgx.CollectOneRow(row, pgx.RowToStructByNameLax[model.Tender])

}

func (s *Storage) RollbackTender(ctx context.Context, tenderId uuid.UUID, username string, ver int) (model.Tender, error) {

	tender, err := s.tender(ctx, tenderId)
	if err != nil {
		return model.Tender{}, err
	}

	userId, err := s.userId(ctx, username)
	if err != nil {
		return model.Tender{}, err
	}

	if !s.checkRelationToOrganization(ctx, userId, tender.OrganizationId) {
		return model.Tender{}, ErrNotEnoughPerm
	}

	oldTender, err := s.tenderVer(ctx, tenderId, ver)
	if err != nil {
		return model.Tender{}, err
	}

	query := `	UPDATE tender
				SET name = $1,
					description = $2,
					type = $3,
					status = $4,
					version = version + 1, 
					updated_at = now()::timestamp without time zone
				WHERE id = $5
				RETURNING *;`

	row, err := s.conn.Query(ctx, query, oldTender.Name, oldTender.Description, oldTender.ServiceType, oldTender.Status, tenderId)
	if err != nil {
		return model.Tender{}, err
	}

	return pgx.CollectOneRow(row, pgx.RowToStructByNameLax[model.Tender])

}
