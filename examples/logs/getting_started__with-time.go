package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Fprintf(os.Stderr, "%s: %s", time.Now(), "I'm here!")
}
