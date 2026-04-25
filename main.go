package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	_ "github.com/joho/godotenv/autoload"
)

var remoteSki string

var frontend WebsocketClient

// main app
func usage() {
	fmt.Println("Usage: controlbox <port>")
	fmt.Println()
	fmt.Println("Certificate configuration via .env file:")
	fmt.Println("  CERT_PEM + KEY_PEM   inline PEM content")
	fmt.Println("  (auto-generated and persisted on first run if absent)")
}

func setupRoutes(h *controlbox) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(h, w, r)
	})
}

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(1)
	}

	srv := new(controlbox)
	srv.run()
	setupRoutes(srv)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		os.Exit(0)
	}()

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(httpdPort), nil))
}
