package handlers

import (
	"context"
	"zadanie/model"

	"github.com/gofrs/uuid"
)

type Storage interface {
	Pinger
	Tenderer
	Bidder
}

type Pinger interface {
	Ping(ctx context.Context) error
}

type Tenderer interface {
	CreateTender(ctx context.Context, tender model.Tender, username string) (model.Tender, error)
	ReadTenders(ctx context.Context, limit int, offset int, types []model.TenderServiceType) ([]model.Tender, error)
	ReadMyTenders(ctx context.Context, username string, limit int, offset int) ([]model.Tender, error)
	ReadTenderStatus(ctx context.Context, tenderId uuid.UUID, username string) (model.TenderStatus, error)
	UpdateTender(ctx context.Context, tenderId uuid.UUID, username string, new model.Tender) (model.Tender, error)
	UpdateTenderStatus(ctx context.Context, tenderId uuid.UUID, username string, status model.TenderStatus) (model.Tender, error)
	RollbackTender(ctx context.Context, tenderId uuid.UUID, username string, ver int) (model.Tender, error)
}

type Bidder interface {
	CreateBid(ctx context.Context, b model.Bid) (model.Bid, error)
	ReadBids(ctx context.Context, tenderId uuid.UUID, username string, limit int, offset int) ([]model.Bid, error)
	ReadMyBids(ctx context.Context, username string, limit int, offset int) ([]model.Bid, error)
	ReadBidStatus(ctx context.Context, bidId uuid.UUID, username string) (model.BidStatus, error)
	UpdateBid(ctx context.Context, bidId uuid.UUID, username string, new model.Bid) (model.Bid, error)
	UpdateBidStatus(ctx context.Context, bidId uuid.UUID, username string, status model.BidStatus) (model.Bid, error)
	SubmitDecision(ctx context.Context, bidId uuid.UUID, decision model.BidStatus, username string) (model.Bid, error)
	Feedback(ctx context.Context, bidId uuid.UUID, feedback string, username string) (model.Bid, error)
	BidReviews(ctx context.Context, tenderId uuid.UUID, authorUsername string, requesterUsername string, limit int, offset int) ([]model.BidFeedback, error)
	RollbackBid(ctx context.Context, bidId uuid.UUID, version int, username string) (model.Bid, error)
}
