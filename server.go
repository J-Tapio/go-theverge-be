package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

func serveMainNews() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := json.MarshalIndent(currentNews.Main, "", "\t")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}
	}
}

func serveFeedNews() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := json.MarshalIndent(currentNews.Feed, "", "\t")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}
	}
}


func initRouter() *mux.Router {
	appRouter := mux.NewRouter()
	appRouter.HandleFunc("/main-news", serveMainNews()).Methods("GET")
	appRouter.HandleFunc("/feed-news", serveFeedNews()).Methods("GET")
	return appRouter
}


func runServer() {
	server := &http.Server{
		Handler: initRouter(),
		Addr: ":" + os.Getenv("PORT"),
		// From mux docs; avoid Slowloris attacks by implementing timeouts.
		// Slowloris - partial HTTP requests.
		WriteTimeout: 15 * time.Second,
		ReadTimeout: 15 * time.Second,
		IdleTimeout:  time.Second * 60,
	}

	// Gracefully shutdown server
	// Referencing https://pkg.go.dev/net/http#Server.Shutdown
	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// Received an interrupt signal, shut down.
		if err := server.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
}