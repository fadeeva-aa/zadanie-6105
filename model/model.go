package model

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

type Tender struct {
	Id              uuid.UUID         `json:"id" db:"id"`
	Name            string            `json:"name" db:"name"`
	Description     string            `json:"description" db:"description"`
	ServiceType     TenderServiceType `json:"serviceType" db:"type"`
	Status          TenderStatus      `json:"status" db:"status"`
	OrganizationId  uuid.UUID         `json:"organizationId" db:"organization_id"`
	Version         uint              `json:"version" db:"version"`
	CreatedAt       time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time         `json:"updatedAt" db:"updated_at"`
	CreatorUsername string            `json:"creatorUsername" db:"-"`
}

func (t *Tender) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf(`{"id":"%s","name":"%s","description":"%s","status":"%s","serviceType":"%s","version":%d,"createdAt":"%s"}`,
		t.Id.String(), t.Name, t.Description, t.Status, t.ServiceType, t.Version, t.CreatedAt.Format(time.RFC3339))

	return []byte(str), nil
}

type Bid struct {
	Id          uuid.UUID     `json:"id" db:"id"`
	Name        string        `json:"name" db:"name"`
	Description string        `json:"description" db:"description"`
	Status      BidStatus     `json:"status" db:"status"`
	TenderId    uuid.UUID     `json:"tenderId" db:"tender_id"`
	AuthorType  BidAuthorType `json:"authorType" db:"author_type"`
	AuthorId    uuid.UUID     `json:"authorId" db:"author_id"`
	Version     int           `json:"version" db:"version"`
	CreatedAt   time.Time     `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time     `json:"updatedAt" db:"updated_at"`
}

func (b *Bid) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf(`{"id":"%s","name":"%s","status":"%s","authorType":"%s","authorId":"%s","version":%d,"createdAt":"%s"}`,
		b.Id.String(), b.Name, b.Status, b.AuthorType, b.AuthorId.String(), b.Version, b.CreatedAt.Format(time.RFC3339))

	return []byte(str), nil
}

type BidFeedback struct {
	Id          uuid.UUID `json:"id" db:"id"`
	TenderId    uuid.UUID `json:"tenderId" db:"tender_id"`
	BidId       uuid.UUID `json:"bidId" db:"bid_id"`
	UserId      uuid.UUID `json:"userId" db:"user_id"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

func (b *BidFeedback) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf(`{"id":"%s","description":"%s","createdAt":"%s"}`,
		b.Id.String(), b.Description, b.CreatedAt.Format(time.RFC3339))

	return []byte(str), nil
}

type TenderStatus string

const (
	TenderStatusCreated   TenderStatus = "Created"
	TenderStatusPublished TenderStatus = "Published"
	TenderStatusClosed    TenderStatus = "Closed"
)

func (ts TenderStatus) Validate() bool {
	switch ts {
	case TenderStatusCreated, TenderStatusPublished, TenderStatusClosed:
		return true
	default:
		return false
	}
}

type TenderServiceType string

const (
	TenderServiceTypeConstruction TenderServiceType = "Construction"
	TenderServiceTypeDelivery     TenderServiceType = "Delivery"
	TenderServiceTypeManufacture  TenderServiceType = "Manufacture"
)

func (tst TenderServiceType) Validate() bool {
	switch tst {
	case TenderServiceTypeConstruction, TenderServiceTypeDelivery, TenderServiceTypeManufacture:
		return true
	default:
		return false
	}
}

type BidAuthorType string

const (
	BidAuthorTypeOrganization BidAuthorType = "Organization"
	BidAuthorTypeUser         BidAuthorType = "User"
)

func (bat BidAuthorType) Validate() bool {
	switch bat {
	case BidAuthorTypeOrganization, BidAuthorTypeUser:
		return true
	default:
		return false
	}
}

type BidStatus string

const (
	BidStatusCreated   BidStatus = "Created"
	BidStatusPublished BidStatus = "Published"
	BidStatusCanceled  BidStatus = "Canceled"
	BidStatusApproved  BidStatus = "Approved"
	BidStatusRejected  BidStatus = "Rejected"
)

func (bs BidStatus) ValidateStatus() bool {
	switch bs {
	case BidStatusCreated, BidStatusPublished, BidStatusCanceled:
		return true
	default:
		return false
	}
}

func (bs BidStatus) ValidateDecision() bool {
	switch bs {
	case BidStatusApproved, BidStatusRejected:
		return true
	default:
		return false
	}
}
