package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/andrewhowdencom/courses.pito/delivery-service/server"
)

// flags that influence the programs behavior
var addr = flag.String("a", "localhost:9093", "the address on which the server should listen")

func main() {

	// Bind signal handlers
	ch := make(chan os.Signal, 1)

	// SIGINT is the signal to terminate ("interrupt") the process.
	signal.Notify(ch, syscall.SIGINT)

	// Setup the server
	srv := server.New()

	// Run the server, but in its own goroutine without blocking this thread.
	go func() {
		if err := srv.Listen(*addr); err != nil {
			os.Exit(1)
		}
	}()

	<-ch

	if err := srv.Shutdown(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func init() {
	// Parse the flags
	flag.Parse()
}
