package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"stockflow/internal/handler"
)

func main() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://stockflow:stockflow@localhost:5433/stockflow"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := pgxpool.New(context.Background(), url)
	if err != nil {
		log.Fatal("cannot create database pool: ", err)
	}
	defer db.Close()
	if err = db.Ping(context.Background()); err != nil {
		log.Fatal("cannot reach database: ", err)
	}

	h := &handler.Handler{DB: db}
	log.Println("listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, h.Routes()))
}
