package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"zadanie/model"
	"zadanie/storage"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
)

func writeErrorResponse(w http.ResponseWriter, err error, statusCode int, method string) (int, error) {
	defer log.Println("%s: ", method, err.Error())

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
	return w.Write([]byte(fmt.Sprintf(`{"reason":%q}`, err.Error())))
}

func errStatusCode(err error) (code int) {
	switch {
	case errors.Is(err, storage.ErrIncorrectUser), errors.Is(err, ErrPassUsername):
		code = 401
	case errors.Is(err, storage.ErrNotEnoughPerm):
		code = 403
	case errors.Is(err, storage.ErrTenderNotFound),
		errors.Is(err, storage.ErrBidNotFound),
		errors.Is(err, storage.ErrVersionNotFound):
		code = 404
	default:
		code = 400
	}

	return
}

func Ping(ctx context.Context, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if err := s.Ping(ctx); err != nil {
			writeErrorResponse(w, err, 500, "ping")
			return
		}

		w.Write([]byte("ok"))

	}
}

func Tenders(ctx context.Context, s Storage) http.HandlerFunc {
	method := "tenders"

	return func(w http.ResponseWriter, r *http.Request) {

		limit := 5
		if tmp, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
			limit = tmp
		}

		offset := 0
		if tmp, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
			offset = tmp
		}

		serviceTypes := []model.TenderServiceType{}

		for _, t := range r.URL.Query()["service_type"] {
			tst := model.TenderServiceType(t)
			if !tst.Validate() {
				continue
			}
			serviceTypes = append(serviceTypes, tst)
		}

		tenders, err := s.ReadTenders(ctx, limit, offset, serviceTypes)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		bytes, err := json.Marshal(tenders)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(bytes)

	}
}

func NewTender(ctx context.Context, s Storage) http.HandlerFunc {
	method := "new tender"

	return func(w http.ResponseWriter, r *http.Request) {

		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}
		_ = r.Body.Close()

		t := model.Tender{}
		if err := json.Unmarshal(bytes, &t); err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		tenders, err := s.CreateTender(ctx, t, t.CreatorUsername)
		if err != nil {
			writeErrorResponse(w, err, errStatusCode(err), method)
			return
		}

		bytes, err = json.Marshal(tenders)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(bytes)

	}
}

func MyTenders(ctx context.Context, s Storage) http.HandlerFunc {
	method := "my tenders"

	return func(w http.ResponseWriter, r *http.Request) {

		limit := 5
		if tmp, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
			limit = tmp
		}

		offset := 0
		if tmp, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
			offset = tmp
		}

		username := r.URL.Query().Get("username")
		if len(username) == 0 {
			writeErrorResponse(w, ErrPassUsername, 401, method)
			return
		}

		tenders, err := s.ReadMyTenders(ctx, username, limit, offset)
		if err != nil {
			writeErrorResponse(w, err, errStatusCode(err), method)
			return
		}

		bytes, err := json.Marshal(tenders)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(bytes)

	}
}

func TenderStatus(ctx context.Context, s Storage) http.HandlerFunc {
	method := "tender status"

	return func(w http.ResponseWriter, r *http.Request) {

		tenderId, err := uuid.FromString(chi.URLParam(r, "tenderId"))
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		username := r.URL.Query().Get("username")
		if len(username) == 0 {
			writeErrorResponse(w, ErrPassUsername, 401, method)
			return
		}

		tender, err := s.ReadTenderStatus(ctx, tenderId, username)
		if err != nil {
			writeErrorResponse(w, err, errStatusCode(err), method)
			return
		}

		bytes, err := json.Marshal(tender)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(bytes)

	}
}

func UpdateTenderStatus(ctx context.Context, s Storage) http.HandlerFunc {
	method := "update tender status"

	return func(w http.ResponseWriter, r *http.Request) {

		tenderId, err := uuid.FromString(chi.URLParam(r, "tenderId"))
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		status := model.TenderStatus(r.URL.Query().Get("status"))
		if !status.Validate() {
			writeErrorResponse(w, ErrIncorrectStatus, 400, method)
			return
		}

		username := r.URL.Query().Get("username")
		if len(username) == 0 {
			writeErrorResponse(w, ErrPassUsername, 401, method)
			return
		}

		tender, err := s.UpdateTenderStatus(ctx, tenderId, username, status)
		if err != nil {
			writeErrorResponse(w, err, errStatusCode(err), method)
			return
		}

		bytes, err := json.Marshal(tender)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(bytes)

	}
}

func EditTender(ctx context.Context, s Storage) http.HandlerFunc {
	method := "edit tender"

	return func(w http.ResponseWriter, r *http.Request) {

		tenderId, err := uuid.FromString(chi.URLParam(r, "tenderId"))
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		username := r.URL.Query().Get("username")
		if len(username) == 0 {
			writeErrorResponse(w, ErrPassUsername, 401, method)
			return
		}

		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}
		_ = r.Body.Close()

		t := model.Tender{}
		if err := json.Unmarshal(bytes, &t); err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		if len(t.Name) == 0 && len(t.Description) == 0 && len(t.ServiceType) == 0 {
			writeErrorResponse(w, ErrNothingToDo, 400, method)
			return
		}

		tender, err := s.UpdateTender(ctx, tenderId, username, t)
		if err != nil {
			writeErrorResponse(w, err, errStatusCode(err), method)
			return
		}

		bytes, err = json.Marshal(tender)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(bytes)

	}

}
func RollbackTender(ctx context.Context, s Storage) http.HandlerFunc {
	method := "rollback tender version"

	return func(w http.ResponseWriter, r *http.Request) {

		tenderId, err := uuid.FromString(chi.URLParam(r, "tenderId"))
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		version, err := strconv.Atoi(chi.URLParam(r, "version"))
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		username := r.URL.Query().Get("username")
		if len(username) == 0 {
			writeErrorResponse(w, ErrPassUsername, 401, method)
			return
		}

		tender, err := s.RollbackTender(ctx, tenderId, username, version)
		if err != nil {
			writeErrorResponse(w, err, errStatusCode(err), method)
			return
		}

		bytes, err := json.Marshal(tender)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(bytes)

	}
}
func NewBid(ctx context.Context, s Storage) http.HandlerFunc {
	method := "new bid"

	return func(w http.ResponseWriter, r *http.Request) {

		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}
		_ = r.Body.Close()

		b := model.Bid{}
		if err := json.Unmarshal(bytes, &b); err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		bid, err := s.CreateBid(ctx, b)
		if err != nil {
			writeErrorResponse(w, err, errStatusCode(err), method)
			return
		}

		bytes, err = json.Marshal(bid)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(bytes)

	}
}

func MyBids(ctx context.Context, s Storage) http.HandlerFunc {
	method := "my bids"

	return func(w http.ResponseWriter, r *http.Request) {

		limit := 5
		if tmp, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
			limit = tmp
		}

		offset := 0
		if tmp, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
			offset = tmp
		}

		username := r.URL.Query().Get("username")
		if len(username) == 0 {
			writeErrorResponse(w, ErrPassUsername, 401, method)
			return
		}

		bids, err := s.ReadMyBids(ctx, username, limit, offset)
		if err != nil {
			writeErrorResponse(w, err, errStatusCode(err), method)
			return
		}

		bytes, err := json.Marshal(bids)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(bytes)

	}
}

func BidsList(ctx context.Context, s Storage) http.HandlerFunc {
	method := "bids list"

	return func(w http.ResponseWriter, r *http.Request) {

		tenderId, err := uuid.FromString(chi.URLParam(r, "tenderId"))
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		username := r.URL.Query().Get("username")
		if len(username) == 0 {
			writeErrorResponse(w, ErrPassUsername, 401, method)
			return
		}

		limit := 5
		if tmp, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
			limit = tmp
		}

		offset := 0
		if tmp, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
			offset = tmp
		}

		bids, err := s.ReadBids(ctx, tenderId, username, limit, offset)
		if err != nil {
			writeErrorResponse(w, err, errStatusCode(err), method)
			return
		}

		bytes, err := json.Marshal(bids)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(bytes)

	}
}

func BidStatus(ctx context.Context, s Storage) http.HandlerFunc {
	method := "bid status"

	return func(w http.ResponseWriter, r *http.Request) {

		bidId, err := uuid.FromString(chi.URLParam(r, "bidId"))
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		username := r.URL.Query().Get("username")
		if len(username) == 0 {
			writeErrorResponse(w, ErrPassUsername, 401, method)
			return
		}

		bids, err := s.ReadBidStatus(ctx, bidId, username)
		if err != nil {
			writeErrorResponse(w, err, errStatusCode(err), method)
			return
		}

		bytes, err := json.Marshal(bids)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(bytes)

	}
}

func UpdateBidStatus(ctx context.Context, s Storage) http.HandlerFunc {
	method := "update bid status"

	return func(w http.ResponseWriter, r *http.Request) {

		bidId, err := uuid.FromString(chi.URLParam(r, "bidId"))
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		status := model.BidStatus(r.URL.Query().Get("status"))
		if !status.ValidateStatus() {
			writeErrorResponse(w, ErrIncorrectStatus, 400, method)
			return
		}

		username := r.URL.Query().Get("username")
		if len(username) == 0 {
			writeErrorResponse(w, ErrPassUsername, 401, method)
			return
		}

		bid, err := s.UpdateBidStatus(ctx, bidId, username, status)
		if err != nil {
			writeErrorResponse(w, err, errStatusCode(err), method)
			return
		}

		bytes, err := json.Marshal(bid)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(bytes)

	}
}

func EditBid(ctx context.Context, s Storage) http.HandlerFunc {
	method := "edit bid"

	return func(w http.ResponseWriter, r *http.Request) {

		bidId, err := uuid.FromString(chi.URLParam(r, "bidId"))
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		username := r.URL.Query().Get("username")
		if len(username) == 0 {
			writeErrorResponse(w, ErrPassUsername, 401, method)
			return
		}

		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}
		_ = r.Body.Close()

		b := model.Bid{}
		if err := json.Unmarshal(bytes, &b); err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		bid, err := s.UpdateBid(ctx, bidId, username, b)
		if err != nil {
			writeErrorResponse(w, err, errStatusCode(err), method)
			return
		}

		bytes, err = json.Marshal(bid)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(bytes)

	}
}

func SubmitDecision(ctx context.Context, s Storage) http.HandlerFunc {
	method := "submit decosion"

	return func(w http.ResponseWriter, r *http.Request) {

		bidId, err := uuid.FromString(chi.URLParam(r, "bidId"))
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		decision := model.BidStatus(r.URL.Query().Get("decision"))
		if !decision.ValidateDecision() {
			writeErrorResponse(w, ErrIncorrectStatus, 400, method)
			return
		}

		username := r.URL.Query().Get("username")
		if len(username) == 0 {
			writeErrorResponse(w, ErrPassUsername, 401, method)
			return
		}

		bid, err := s.SubmitDecision(ctx, bidId, decision, username)
		if err != nil {
			writeErrorResponse(w, err, errStatusCode(err), method)
			return
		}

		bytes, err := json.Marshal(bid)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(bytes)

	}
}

func Feedback(ctx context.Context, s Storage) http.HandlerFunc {
	method := "feedback"

	return func(w http.ResponseWriter, r *http.Request) {

		bidId, err := uuid.FromString(chi.URLParam(r, "bidId"))
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		feedback := r.URL.Query().Get("bidFeedback")
		if len(feedback) == 0 {
			writeErrorResponse(w, ErrPassFeedback, 400, method)
			return
		}

		username := r.URL.Query().Get("username")
		if len(username) == 0 {
			writeErrorResponse(w, ErrPassUsername, 401, method)
			return
		}

		bid, err := s.Feedback(ctx, bidId, feedback, username)
		if err != nil {
			writeErrorResponse(w, err, errStatusCode(err), method)
			return
		}

		bytes, err := json.Marshal(bid)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(bytes)

	}
}
func RollbackBid(ctx context.Context, s Storage) http.HandlerFunc {
	method := "rollback bid version"

	return func(w http.ResponseWriter, r *http.Request) {

		bidId, err := uuid.FromString(chi.URLParam(r, "bidId"))
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		version, err := strconv.Atoi(chi.URLParam(r, "version"))
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		username := r.URL.Query().Get("username")
		if len(username) == 0 {
			writeErrorResponse(w, ErrPassUsername, 401, method)
			return
		}

		bid, err := s.RollbackBid(ctx, bidId, version, username)
		if err != nil {
			writeErrorResponse(w, err, errStatusCode(err), method)
			return
		}

		bytes, err := json.Marshal(bid)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(bytes)

	}
}

func ReviewsBids(ctx context.Context, s Storage) http.HandlerFunc {
	method := "reviews bids"

	return func(w http.ResponseWriter, r *http.Request) {

		tenderId, err := uuid.FromString(chi.URLParam(r, "tenderId"))
		if err != nil {
			log.Println(err)
			return
		}

		authorUsername := r.URL.Query().Get("authorUsername")
		if len(authorUsername) == 0 {
			log.Println(err)
			return
		}

		requesterUsername := r.URL.Query().Get("requesterUsername")
		if len(requesterUsername) == 0 {
			log.Println(err)
			return
		}

		limit := 5
		if tmp, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
			limit = tmp
		}

		offset := 0
		if tmp, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
			offset = tmp
		}

		bid, err := s.BidReviews(ctx, tenderId, authorUsername, requesterUsername, limit, offset)
		if err != nil {
			writeErrorResponse(w, err, errStatusCode(err), method)
			return
		}

		bytes, err := json.Marshal(bid)
		if err != nil {
			writeErrorResponse(w, err, 400, method)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(bytes)

	}
}
