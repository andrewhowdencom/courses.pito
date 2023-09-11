package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args[1:]) == 0 {
		fmt.Fprintf(os.Stderr, "INFO: %s: %s", time.Now(), "No argument supplied.\n")
		return
	}

	total := 1
	for _, in := range os.Args[1:] {
		i, err := strconv.Atoi(in)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s: %s, %s (%s)", time.Now(), "wrong argument supplied\n", err, in)
			return
		}

		total = total * i
	}

	fmt.Println(total)
}
