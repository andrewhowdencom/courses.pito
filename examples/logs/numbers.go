package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args[1:]) == 0 {
		fmt.Fprintf(os.Stderr, "%s: %s", time.Now(), "No arguments supplied")
		return
	}

	total := 1
	for _, in := range os.Args[1:] {
		i, err := strconv.Atoi(in)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s: %s (input: %s)", time.Now(), "failed to create number", err, in)
			return
		}

		total = total * i
	}

	fmt.Println(total)
}
