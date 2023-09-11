package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/andrewhowdencom/courses.pito/delivery-service/server"
)

// flags that influence the programs behavior
var addr = flag.String("a", "localhost:9093", "the address on which the server should listen")

var log *slog.Logger

func main() {
	log.Info("application started")

	// Bind signal handlers
	ch := make(chan os.Signal, 1)

	// SIGINT is the signal to terminate ("interrupt") the process.
	signal.Notify(ch, syscall.SIGINT)

	// Setup the server
	srv := server.New()

	// Run the server, but in its own goroutine without blocking this thread.
	go func() {
		if err := srv.Listen(*addr); err != nil {
			log.Error("failed to start server", "error", err, "addr", *addr)
			os.Exit(1)
		}
	}()

	log.Info("awaiting shutdown signal (SIGINT)")
	<-ch
	log.Info("received shutdown signal")

	if err := srv.Shutdown(); err != nil {
		log.Error("failed to shutdown server", "error", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func init() {
	// Parse the flags
	flag.Parse()

	// Bootstrap the logger
	log = slog.New(slog.NewJSONHandler(
		os.Stderr, nil,
	))
}
