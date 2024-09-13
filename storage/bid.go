package storage

import (
	"context"
	"fmt"
	"strings"
	"zadanie/model"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) bid(ctx context.Context, id uuid.UUID) (model.Bid, error) {

	query := `SELECT * FROM bid WHERE id = $1`
	row, err := s.conn.Query(ctx, query, id)
	if err != nil {
		return model.Bid{}, err
	}

	bid, err := pgx.CollectOneRow(row, pgx.RowToStructByNameLax[model.Bid])
	if err != nil {
		return model.Bid{}, ErrBidNotFound
	}

	return bid, nil

}

func (s *Storage) bidVer(ctx context.Context, id uuid.UUID, ver int) (model.Bid, error) {

	query := `SELECT * FROM bid_archive WHERE id = $1 AND version = $2;`
	row, err := s.conn.Query(ctx, query, id, ver)
	if err != nil {
		return model.Bid{}, err
	}

	bid, err := pgx.CollectOneRow(row, pgx.RowToStructByNameLax[model.Bid])
	if err != nil {
		return model.Bid{}, ErrVersionNotFound
	}

	return bid, nil

}

func (s *Storage) CreateBid(ctx context.Context, b model.Bid) (model.Bid, error) {

	if err := s.checkUser(ctx, b.AuthorId); err != nil {
		return model.Bid{}, err
	}

	tender, err := s.tender(ctx, b.TenderId)
	if err != nil {
		return model.Bid{}, err
	}

	if tender.Status != model.TenderStatusPublished && !s.checkRelationToOrganization(ctx, b.AuthorId, tender.OrganizationId) {
		return model.Bid{}, ErrNotEnoughPerm
	}

	insert := `	INSERT INTO bid(name, description, status, tender_id, author_type, author_id)
				VALUES ($1, $2, $3, $4, $5, $6)
				RETURNING *;`

	row, err := s.conn.Query(ctx, insert, b.Name, b.Description, model.BidStatusCreated, b.TenderId, b.AuthorType, b.AuthorId)
	if err != nil {
		return model.Bid{}, err
	}

	return pgx.CollectOneRow(row, pgx.RowToStructByNameLax[model.Bid])

}

func (s *Storage) ReadMyBids(ctx context.Context, username string, limit, offset int) ([]model.Bid, error) {

	userId, err := s.userId(ctx, username)
	if err != nil {
		return nil, err
	}

	row, err := s.conn.Query(ctx, `SELECT * FROM bid WHERE author_id = $1 ORDER BY name ASC LIMIT $2 OFFSET $3;`, userId, limit, offset)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(row, pgx.RowToStructByNameLax[model.Bid])

}

func (s *Storage) ReadBids(ctx context.Context, tenderId uuid.UUID, username string, limit, offset int) ([]model.Bid, error) {

	if _, err := s.tender(ctx, tenderId); err != nil {
		return nil, err
	}

	userId, err := s.userId(ctx, username)
	if err != nil {
		return nil, err
	}

	userOrgId, err := s.userOrgId(ctx, userId)
	if err != nil {
		return nil, err
	}

	query := `	SELECT bid.*
				FROM bid
				JOIN tender ON bid.tender_id = tender.id
				WHERE bid.tender_id = $1
				AND ( bid.author_id IN (
					SELECT user_id
					FROM organization_responsible 
					WHERE organization_id = $2)
					OR ( tender.organization_id = $2 
						AND bid.status = 'Published') )
				ORDER BY bid.name ASC
				LIMIT $3 OFFSET $4;`

	row, err := s.conn.Query(ctx, query, tenderId, userOrgId, limit, offset)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(row, pgx.RowToStructByNameLax[model.Bid])

}

func (s *Storage) ReadBidStatus(ctx context.Context, bidId uuid.UUID, username string) (model.BidStatus, error) {

	userId, err := s.userId(ctx, username)
	if err != nil {
		return "", err
	}

	bid, err := s.bid(ctx, bidId)
	if err != nil {
		return "", err
	}

	bidOrgId, err := s.userOrgId(ctx, bid.AuthorId)
	if err != nil {
		return "", err
	}

	tender, err := s.tender(ctx, bid.TenderId)
	if err != nil {
		return "", err
	}

	switch {
	case s.checkRelationToOrganization(ctx, userId, tender.OrganizationId):
		if bid.Status == model.BidStatusCreated || bid.Status == model.BidStatusCanceled {
			return "", ErrNotEnoughPerm
		}
	case s.checkRelationToOrganization(ctx, userId, bidOrgId):
	default:
		return "", ErrNotEnoughPerm
	}

	return bid.Status, nil

}

func (s *Storage) UpdateBidStatus(ctx context.Context, bidId uuid.UUID, username string, status model.BidStatus) (model.Bid, error) {

	userId, err := s.userId(ctx, username)
	if err != nil {
		return model.Bid{}, err
	}

	bid, err := s.bid(ctx, bidId)
	if err != nil {
		return model.Bid{}, err
	}

	bidOrgId, err := s.userOrgId(ctx, bid.AuthorId)
	if err != nil {
		return model.Bid{}, err
	}

	if !s.checkRelationToOrganization(ctx, userId, bidOrgId) {
		return model.Bid{}, ErrNotEnoughPerm
	}

	if bid.Status == model.BidStatusCanceled {
		return model.Bid{}, ErrStatusCantBeChanged
	}

	update := `	UPDATE bid 
				SET status = $1, 
					version = version + 1, 
					updated_at = now()::timestamp without time zone
				WHERE id = $2 
				RETURNING *;`

	row, err := s.conn.Query(ctx, update, status, bidId)
	if err != nil {
		return model.Bid{}, err
	}

	return pgx.CollectOneRow(row, pgx.RowToStructByNameLax[model.Bid])

}

func (s *Storage) UpdateBid(ctx context.Context, bidId uuid.UUID, username string, new model.Bid) (model.Bid, error) {

	userId, err := s.userId(ctx, username)
	if err != nil {
		return model.Bid{}, err
	}

	bid, err := s.bid(ctx, bidId)
	if err != nil {
		return model.Bid{}, err
	}

	bidOrgId, err := s.userOrgId(ctx, bid.AuthorId)
	if err != nil {
		return model.Bid{}, err
	}

	if !s.checkRelationToOrganization(ctx, userId, bidOrgId) {
		return model.Bid{}, ErrNotEnoughPerm
	}

	parts := []string{}
	if len(new.Name) != 0 {
		parts = append(parts, fmt.Sprintf("name = '%s' ", new.Name))
	}
	if len(new.Description) != 0 {
		parts = append(parts, fmt.Sprintf("description = '%s' ", new.Description))
	}

	parts = append(parts, "version = version + 1 ")
	parts = append(parts, "updated_at = now()::timestamp without time zone ")

	update := fmt.Sprintf(`UPDATE bid SET %s WHERE id = $1 RETURNING *;`, strings.Join(parts, ","))

	row, err := s.conn.Query(ctx, update, bidId)
	if err != nil {
		return model.Bid{}, err
	}

	return pgx.CollectOneRow(row, pgx.RowToStructByNameLax[model.Bid])

}

func (s *Storage) RollbackBid(ctx context.Context, bidId uuid.UUID, version int, username string) (model.Bid, error) {

	userId, err := s.userId(ctx, username)
	if err != nil {
		return model.Bid{}, err
	}

	bid, err := s.bid(ctx, bidId)
	if err != nil {
		return model.Bid{}, err
	}

	bidOrgId, err := s.userOrgId(ctx, bid.AuthorId)
	if err != nil {
		return model.Bid{}, err
	}

	if !s.checkRelationToOrganization(ctx, userId, bidOrgId) {
		return model.Bid{}, ErrNotEnoughPerm
	}

	oldBid, err := s.bidVer(ctx, bidId, version)
	if err != nil {
		return model.Bid{}, err
	}

	query := `	UPDATE bid
				SET name = $1,
					description = $2,
					status = $3,
					version = version + 1, 
					updated_at = now()::timestamp without time zone
				WHERE id = $4
				RETURNING *;`

	row, err := s.conn.Query(ctx, query, oldBid.Name, oldBid.Description, oldBid.Status, bidId)
	if err != nil {
		return model.Bid{}, err
	}

	return pgx.CollectOneRow(row, pgx.RowToStructByNameLax[model.Bid])

}

func (s *Storage) SubmitDecision(ctx context.Context, bidId uuid.UUID, decision model.BidStatus, username string) (model.Bid, error) {

	bid, err := s.bid(ctx, bidId)
	if err != nil {
		return model.Bid{}, err
	}

	if bid.Status != model.BidStatusPublished {
		return model.Bid{}, ErrStatusCantBeChanged
	}

	userId, err := s.userId(ctx, username)
	if err != nil {
		return model.Bid{}, err
	}

	tender, err := s.tender(ctx, bid.TenderId)
	if err != nil {
		return model.Bid{}, err
	}

	if !s.checkRelationToOrganization(ctx, userId, tender.OrganizationId) {
		return model.Bid{}, ErrNotEnoughPerm
	}

	if tender.Status == model.TenderStatusClosed {
		return model.Bid{}, ErrTenderClosed
	}

	if decision == model.BidStatusRejected {

		update := `	UPDATE bid 
					SET status = $1, 
						version = version + 1, 
						updated_at = now()::timestamp without time zone
					WHERE id = $2 
					RETURNING *;`

		row, err := s.conn.Query(ctx, update, model.BidStatusRejected, bidId)
		if err != nil {
			return model.Bid{}, err
		}

		return pgx.CollectOneRow(row, pgx.RowToStructByNameLax[model.Bid])

	}

	if _, err := s.conn.Exec(ctx, `INSERT INTO bid_approved_decision(bid_id, user_id) VALUES ($1, $2);`, bidId, userId); err != nil {
		return model.Bid{}, err
	}

	count := 0
	if err := s.conn.QueryRow(ctx, `SELECT COUNT(*) FROM bid_approved_decision WHERE bid_id = $1;`, bidId).Scan(&count); err != nil {
		return model.Bid{}, err
	}

	q, err := s.quorum(ctx, tender.OrganizationId)
	if err != nil {
		return model.Bid{}, err
	}

	if count < q {
		return bid, nil
	}

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return model.Bid{}, err
	}
	defer tx.Rollback(ctx)

	updateBid := `	UPDATE bid 
					SET status = $1, 
						version = version + 1, 
						updated_at = now()::timestamp without time zone
					WHERE id = $2 
					RETURNING *;`

	row, err := tx.Query(ctx, updateBid, model.BidStatusApproved, bidId)
	if err != nil {
		return model.Bid{}, err
	}

	updBid, err := pgx.CollectOneRow(row, pgx.RowToStructByNameLax[model.Bid])
	if err != nil {
		return model.Bid{}, err
	}

	updateTender := `	UPDATE tender 
						SET status = $1, 
							version = version + 1, 
							updated_at = now()::timestamp without time zone
						WHERE id = $2;`

	if _, err = tx.Exec(ctx, updateTender, model.TenderStatusClosed, updBid.TenderId); err != nil {
		return model.Bid{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return model.Bid{}, err
	}

	return updBid, nil
}

func (s *Storage) Feedback(ctx context.Context, bidId uuid.UUID, feedback string, username string) (model.Bid, error) {

	bid, err := s.bid(ctx, bidId)
	if err != nil {
		return model.Bid{}, err
	}

	userId, err := s.userId(ctx, username)
	if err != nil {
		return model.Bid{}, err
	}

	tender, err := s.tender(ctx, bid.TenderId)
	if err != nil {
		return model.Bid{}, err
	}

	if !s.checkRelationToOrganization(ctx, userId, tender.OrganizationId) {
		return model.Bid{}, ErrNotEnoughPerm
	}

	if bid.Status == model.BidStatusCreated || bid.Status == model.BidStatusCanceled {
		return model.Bid{}, ErrNotEnoughPerm
	}

	query := `INSERT INTO bid_feedback(bid_id, tender_id, user_id, description) VALUES ($1, $2, $3, $4);`
	if _, err := s.conn.Exec(ctx, query, bidId, tender.Id, userId, feedback); err != nil {
		return model.Bid{}, err
	}

	return bid, nil
}

func (s *Storage) BidReviews(ctx context.Context, tenderId uuid.UUID, authorUsername, requesterUsername string, limit, offset int) ([]model.BidFeedback, error) {

	tender, err := s.tender(ctx, tenderId)
	if err != nil {
		return nil, err
	}

	authorId, err := s.userId(ctx, authorUsername)
	if err != nil {
		return nil, err
	}

	requesterId, err := s.userId(ctx, requesterUsername)
	if err != nil {
		return nil, err
	}

	if !s.checkRelationToOrganization(ctx, requesterId, tender.OrganizationId) {
		return nil, ErrNotEnoughPerm
	}

	query := `	SELECT * 
				FROM bid_feedback
				WHERE tender_id = $1
				AND bid_id IN (
					SELECT id
					FROM bid
					WHERE author_id = $2);`
	row, err := s.conn.Query(ctx, query, tenderId, authorId)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(row, pgx.RowToStructByNameLax[model.BidFeedback])

}
