package main

import (
	_ "github.com/go-sql-driver/mysql"

	"log"
	"net/http"
	"time"

	"rxsg-new-project/backend/internal/config"
	"rxsg-new-project/backend/internal/legacy"
	"rxsg-new-project/backend/internal/server"
	"rxsg-new-project/backend/internal/service"
)

func main() {
	cfg := config.Load()

	repo, err := legacy.NewRepository(cfg.LegacyDSN)
	if err != nil {
		log.Printf("legacy database warning: %v", err)
	}
	defer func() {
		if repo != nil {
			_ = repo.Close()
		}
	}()

	svc := service.New(repo)
	handler := server.New(cfg, svc).Handler()

	srv := &http.Server{
		Addr:              cfg.Address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("rxsg refactor api listening on %s", cfg.Address)
	log.Fatal(srv.ListenAndServe())
}
