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
	fmt.Println("First Run:")
	fmt.Println("  go controlbox <serverport>")
	fmt.Println()
	fmt.Println("General Usage:")
	fmt.Println("  go controlbox <serverport> <crtfile> <keyfile>")
}

func setupRoutes(h *controlbox) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(h, w, r)
	})
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
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
