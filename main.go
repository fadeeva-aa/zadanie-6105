package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"zadanie/handlers"
	"zadanie/storage"

	"github.com/go-chi/chi/v5"
)

const PORT = ":8080"

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	storage, err := storage.NewStorage(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer storage.Close()

	router := chi.NewRouter()

	router.Route("/api", func(r chi.Router) {
		r.Get("/ping", handlers.Ping(ctx, storage))

		r.Route("/tenders", func(r chi.Router) {
			r.Get("/", handlers.Tenders(ctx, storage))
			r.Post("/new", handlers.NewTender(ctx, storage))
			r.Get("/my", handlers.MyTenders(ctx, storage))
			r.Get("/{tenderId}/status", handlers.TenderStatus(ctx, storage))
			r.Put("/{tenderId}/status", handlers.UpdateTenderStatus(ctx, storage))
			r.Patch("/{tenderId}/edit", handlers.EditTender(ctx, storage))
			r.Put("/{tenderId}/rollback/{version}", handlers.RollbackTender(ctx, storage))
		})

		r.Route("/bids", func(r chi.Router) {
			r.Post("/new", handlers.NewBid(ctx, storage))
			r.Get("/my", handlers.MyBids(ctx, storage))
			r.Get("/{tenderId}/list", handlers.BidsList(ctx, storage))
			r.Get("/{tenderId}/reviews", handlers.ReviewsBids(ctx, storage))
			r.Get("/{bidId}/status", handlers.BidStatus(ctx, storage))
			r.Put("/{bidId}/status", handlers.UpdateBidStatus(ctx, storage))
			r.Patch("/{bidId}/edit", handlers.EditBid(ctx, storage))
			r.Put("/{bidId}/submit_decision", handlers.SubmitDecision(ctx, storage))
			r.Put("/{bidId}/feedback", handlers.Feedback(ctx, storage))
			r.Put("/{bidId}/rollback/{version}", handlers.RollbackBid(ctx, storage))

		})
	})

	srv := &http.Server{
		Addr:    PORT,
		Handler: router,
	}

	go func() {
		log.Println(`server started on ":8080"`)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

}
