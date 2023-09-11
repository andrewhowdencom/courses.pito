package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {

	w := json.NewEncoder(os.Stderr)

	if len(os.Args[1:]) == 0 {
		w.Encode(map[string]interface{}{
			"level":   "INFO",
			"time":    time.Now(),
			"message": "No argument supplied.",
		})

		return
	}

	total := 1
	for _, in := range os.Args[1:] {
		i, err := strconv.Atoi(in)
		if err != nil {
			w.Encode(map[string]interface{}{
				"level":   "ERROR",
				"time":    time.Now(),
				"message": "wrong argument supplied",
				"context": map[string]interface{}{
					"error": err,
					"input": in,
				},
			})
			return
		}

		total = total * i
	}

	fmt.Println(total)
}
